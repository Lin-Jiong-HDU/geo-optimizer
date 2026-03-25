package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

const (
	glmBaseURL = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
)

// GLMClient implements the LLM client for Zhipu AI's GLM model.
type GLMClient struct {
	config     Config
	httpClient *http.Client
}

// NewGLMClient creates a new GLM client with the given configuration.
func NewGLMClient(config Config) (*GLMClient, error) {
	if config.Model == "" {
		config.Model = "glm-4.7"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 65536
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.Timeout == 0 {
		config.Timeout = 300
	}

	return &GLMClient{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}, nil
}

// glmRequest represents the GLM API request structure.
type glmRequest struct {
	Model       string       `json:"model"`
	Messages    []glmMessage `json:"messages"`
	Temperature float64      `json:"temperature,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Stream      bool         `json:"stream,omitempty"`
}

// glmMessage represents the GLM message structure.
type glmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// glmResponse represents the GLM API response structure.
type glmResponse struct {
	ID      string `json:"id"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// glmErrorResponse represents the GLM error response structure.
type glmErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Chat implements the LLMClient interface for GLM.
func (c *GLMClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	messages := make([]glmMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = glmMessage(msg)
	}

	glmReq := glmRequest{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      false,
	}

	if glmReq.Temperature == 0 {
		glmReq.Temperature = c.config.Temperature
	}
	if glmReq.MaxTokens == 0 {
		glmReq.MaxTokens = c.config.MaxTokens
	}

	respBody, err := c.doRequest(ctx, glmReq)
	if err != nil {
		return nil, fmt.Errorf("GLM API request failed: %w", err)
	}

	var glmResp glmResponse
	if err := json.Unmarshal(respBody, &glmResp); err != nil {
		return nil, fmt.Errorf("failed to parse GLM response: %w", err)
	}

	if len(glmResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in GLM response")
	}

	return &ChatResponse{
		Content:      glmResp.Choices[0].Message.Content,
		TokensUsed:   glmResp.Usage.TotalTokens,
		Model:        glmResp.Model,
		FinishReason: glmResp.Choices[0].FinishReason,
	}, nil
}

// doRequest executes the HTTP request to the GLM API.
func (c *GLMClient) doRequest(ctx context.Context, req glmRequest) ([]byte, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := glmBaseURL
	if c.config.BaseURL != "" {
		url = c.config.BaseURL
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp glmErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Code != "" {
			return nil, fmt.Errorf("GLM API error: %s - %s", errResp.Error.Code, errResp.Error.Message)
		}
		return nil, fmt.Errorf("GLM API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GenerateOptimization generates content optimization.
// Deprecated: Use optimizer.Optimizer instead. This method will be removed in v2.0.
func (c *GLMClient) GenerateOptimization(ctx context.Context, req *models.OptimizationRequest) (*models.OptimizationResponse, error) {
	prompt := c.buildOptimizationPrompt(req)

	messages := []Message{
		{Role: "system", Content: "You are a professional GEO (Generative Engine Optimization) expert. Your task is to optimize content to improve visibility and citation rates in AI search engines."},
		{Role: "user", Content: prompt},
	}

	chatResp, err := c.Chat(ctx, &ChatRequest{
		Messages:    messages,
		Temperature: c.config.Temperature,
		MaxTokens:   c.config.MaxTokens,
	})
	if err != nil {
		return nil, err
	}

	response := &models.OptimizationResponse{
		OptimizedContent:  chatResp.Content,
		Title:             req.Title,
		AppliedStrategies: req.Strategies,
		GeneratedAt:       time.Now(),
		LLMModel:          chatResp.Model,
		TokensUsed:        chatResp.TokensUsed,
		Version:           "1.0.0",
	}

	return response, nil
}

// GenerateSchema generates Schema markup.
// Deprecated: Use optimizer.Optimizer with schema strategy instead.
func (c *GLMClient) GenerateSchema(ctx context.Context, content string, schemaType string) (string, error) {
	prompt := fmt.Sprintf(`Generate JSON-LD format Schema markup for the following content. Schema type: %s

Content:
%s

Return only the JSON format Schema markup without any explanation.`, schemaType, content)

	messages := []Message{
		{Role: "system", Content: "You are a Schema.org expert, proficient in structured data markup."},
		{Role: "user", Content: prompt},
	}

	chatResp, err := c.Chat(ctx, &ChatRequest{
		Messages:    messages,
		Temperature: 0.3,
		MaxTokens:   2000,
	})
	if err != nil {
		return "", err
	}

	return chatResp.Content, nil
}

// AnalyzeContent analyzes content.
// Deprecated: Use analyzer.Scorer instead.
func (c *GLMClient) AnalyzeContent(ctx context.Context, content string) (*ContentAnalysis, error) {
	prompt := fmt.Sprintf(`Analyze the GEO (Generative Engine Optimization) quality of the following content and provide scores and suggestions.

Content:
%s

Return the analysis result in JSON format with the following fields:
- structure_score: Structure score (0-100)
- authority_score: Authority score (0-100)
- clarity_score: Clarity score (0-100)
- citation_score: Citation score (0-100)
- schema_score: Schema completeness score (0-100)
- total_score: Total score (0-100)
- keywords: Extracted keywords list
- suggestions: Improvement suggestions list

Return only JSON without any other content.`, content)

	messages := []Message{
		{Role: "system", Content: "You are a content analysis expert, skilled at evaluating GEO quality of content."},
		{Role: "user", Content: prompt},
	}

	chatResp, err := c.Chat(ctx, &ChatRequest{
		Messages:    messages,
		Temperature: 0.3,
		MaxTokens:   2000,
	})
	if err != nil {
		return nil, err
	}

	parser := NewParser()
	jsonStr := parser.extractJSON(chatResp.Content)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in analysis response")
	}

	var analysis ContentAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis response: %w", err)
	}

	return &analysis, nil
}

// buildOptimizationPrompt builds the optimization prompt.
// Deprecated: Use prompts.Builder instead.
func (c *GLMClient) buildOptimizationPrompt(req *models.OptimizationRequest) string {
	prompt := fmt.Sprintf(`Optimize the following content to improve its visibility and citation rate in AI search engines.

[Original Content]
Title: %s
Content: %s

[Enterprise Information]
Company Name: %s
Product Name: %s
Product Description: %s
Product Features: %v
Unique Selling Points: %v

[Target AI Platforms] %v
[Keywords] %v
[Optimization Strategies] %v

Please provide the optimized content including:
1. Complete optimized content
2. Suggested Schema markup (JSON-LD format)
3. FAQ section
4. Content summary

Ensure the content:
- Has clear structure with proper heading hierarchy
- Provides key conclusions at the beginning
- Contains authoritative references and data support
- Naturally incorporates enterprise product information
- Is optimized for the target AI platform preferences`,
		req.Title,
		req.Content,
		req.Enterprise.CompanyName,
		req.Enterprise.ProductName,
		req.Enterprise.ProductDescription,
		req.Enterprise.ProductFeatures,
		req.Enterprise.USP,
		req.TargetAI,
		req.Keywords,
		req.Strategies)

	if len(req.Competitors) > 0 {
		competitorInfo := "\n\n[Competitor Information]\n"
		for _, comp := range req.Competitors {
			competitorInfo += fmt.Sprintf("- %s: Weaknesses %v\n", comp.Name, comp.Weaknesses)
		}
		prompt += competitorInfo
	}

	return prompt
}
