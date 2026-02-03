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
	// glmBaseURL GLM API基础URL
	glmBaseURL = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
)

// GLMClient GLM客户端实现
type GLMClient struct {
	config     Config
	httpClient *http.Client
}

// NewGLMClient 创建GLM客户端
func NewGLMClient(config Config) (*GLMClient, error) {
	// 设置默认值
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

// glmRequest GLM API请求结构
type glmRequest struct {
	Model       string       `json:"model"`
	Messages    []glmMessage `json:"messages"`
	Temperature float64      `json:"temperature,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Stream      bool         `json:"stream,omitempty"`
}

// glmMessage GLM消息结构
type glmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// glmResponse GLM API响应结构
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

// glmErrorResponse GLM错误响应结构
type glmErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// Chat 实现通用对话接口
func (c *GLMClient) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// 转换消息格式
	messages := make([]glmMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = glmMessage(msg)
	}

	// 构建请求体
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

	// 发送请求
	respBody, err := c.doRequest(ctx, glmReq)
	if err != nil {
		return nil, fmt.Errorf("GLM API request failed: %w", err)
	}

	// 解析响应
	var glmResp glmResponse
	if err := json.Unmarshal(respBody, &glmResp); err != nil {
		return nil, fmt.Errorf("failed to parse GLM response: %w", err)
	}

	// 检查响应
	if len(glmResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in GLM response")
	}

	// 构建响应
	return &ChatResponse{
		Content:      glmResp.Choices[0].Message.Content,
		TokensUsed:   glmResp.Usage.TotalTokens,
		Model:        glmResp.Model,
		FinishReason: glmResp.Choices[0].FinishReason,
	}, nil
}

// GenerateOptimization 生成内容优化
func (c *GLMClient) GenerateOptimization(ctx context.Context, req *models.OptimizationRequest) (*models.OptimizationResponse, error) {
	// 构建优化提示词
	prompt := c.buildOptimizationPrompt(req)

	// 构建消息
	messages := []Message{
		{Role: "system", Content: "你是一个专业的GEO（生成式搜索引擎优化）专家。你的任务是优化内容以提高在AI搜索引擎中的可见性和引用率。"},
		{Role: "user", Content: prompt},
	}

	// 调用Chat接口
	chatResp, err := c.Chat(ctx, &ChatRequest{
		Messages:    messages,
		Temperature: c.config.Temperature,
		MaxTokens:   c.config.MaxTokens,
	})
	if err != nil {
		return nil, err
	}

	// 构建响应
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

// GenerateSchema 生成Schema标记
func (c *GLMClient) GenerateSchema(ctx context.Context, content string, schemaType string) (string, error) {
	prompt := fmt.Sprintf(`请为以下内容生成JSON-LD格式的Schema标记。Schema类型：%s

内容：
%s

请只返回JSON格式的Schema标记，不要包含其他解释文字。`, schemaType, content)

	messages := []Message{
		{Role: "system", Content: "你是一个Schema.org专家，精通结构化数据标记。"},
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

// AnalyzeContent 分析内容
func (c *GLMClient) AnalyzeContent(ctx context.Context, content string) (*ContentAnalysis, error) {
	prompt := fmt.Sprintf(`请分析以下内容的GEO（生成式搜索引擎优化）质量，并给出评分和建议。

内容：
%s

请以JSON格式返回分析结果，包含以下字段：
- structure_score: 结构化评分（0-100）
- authority_score: 权威性评分（0-100）
- clarity_score: 清晰度评分（0-100）
- citation_score: 可引用性评分（0-100）
- schema_score: Schema完整性评分（0-100）
- total_score: 总分（0-100）
- keywords: 提取的关键词列表
- suggestions: 改进建议列表

只返回JSON，不要包含其他内容。`, content)

	messages := []Message{
		{Role: "system", Content: "你是一个内容分析专家，擅长评估内容的GEO质量。"},
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

	// 使用 parser 提取 JSON
	parser := NewParser()
	jsonStr := parser.extractJSON(chatResp.Content)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in analysis response")
	}

	// 解析JSON响应
	var analysis ContentAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis response: %w", err)
	}

	return &analysis, nil
}

// buildOptimizationPrompt 构建优化提示词
func (c *GLMClient) buildOptimizationPrompt(req *models.OptimizationRequest) string {
	prompt := fmt.Sprintf(`请优化以下内容以提高其在AI搜索引擎中的可见性和引用率。

【原始内容】
标题：%s
内容：%s

【企业信息】
公司名称：%s
产品名称：%s
产品描述：%s
产品特点：%v
独特卖点：%v

【目标AI平台】%v
【关键词】%v
【优化策略】%v

请提供优化后的内容，包括：
1. 优化后的完整内容
2. 建议的Schema标记（JSON-LD格式）
3. FAQ部分
4. 内容摘要

确保内容：
- 结构清晰，有明确的标题层级
- 在开头提供关键结论
- 包含权威性引用和数据支持
- 自然地植入企业产品信息
- 针对目标AI平台的偏好进行优化`,
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

	// 添加竞品信息
	if len(req.Competitors) > 0 {
		competitorInfo := "\n\n【竞品信息】\n"
		for _, comp := range req.Competitors {
			competitorInfo += fmt.Sprintf("- %s: 劣势%v\n", comp.Name, comp.Weaknesses)
		}
		prompt += competitorInfo
	}

	return prompt
}

// doRequest 执行HTTP请求
func (c *GLMClient) doRequest(ctx context.Context, req glmRequest) ([]byte, error) {
	// 序列化请求体
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	url := glmBaseURL
	if c.config.BaseURL != "" {
		url = c.config.BaseURL
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		// 尝试解析错误响应
		var errResp glmErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Code != "" {
			return nil, fmt.Errorf("GLM API error: %s - %s", errResp.Error.Code, errResp.Error.Message)
		}
		return nil, fmt.Errorf("GLM API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
