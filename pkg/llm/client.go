package llm

import (
	"context"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// LLMClient defines the interface for LLM client implementations.
type LLMClient interface {
	// Chat sends a chat request and returns the response.
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}

// ContentAnalysis represents the result of content analysis.
type ContentAnalysis struct {
	StructureScore float64  `json:"structure_score"`
	AuthorityScore float64  `json:"authority_score"`
	ClarityScore   float64  `json:"clarity_score"`
	CitationScore  float64  `json:"citation_score"`
	SchemaScore    float64  `json:"schema_score"`
	TotalScore     float64  `json:"total_score"`
	Keywords       []string `json:"keywords"`
	Suggestions    []string `json:"suggestions"`
}

// ChatRequest represents a chat request to the LLM.
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Message represents a single message in a chat conversation.
type Message struct {
	Role    string `json:"role"` // system, user, assistant, tool
	Content string `json:"content"`
}

// ChatResponse represents the response from an LLM chat request.
type ChatResponse struct {
	Content      string `json:"content"`
	TokensUsed   int    `json:"tokens_used"`
	Model        string `json:"model"`
	FinishReason string `json:"finish_reason"`
}

// NewClient creates an LLM client based on the provided configuration.
func NewClient(config Config) (LLMClient, error) {
	switch config.Provider {
	case ProviderGLM:
		return NewGLMClient(config)
	default:
		return NewGLMClient(config)
	}
}

// DeprecatedClient contains deprecated methods.
// Deprecated: Use optimizer.Optimizer instead of these methods.
type DeprecatedClient interface {
	LLMClient

	// GenerateOptimization generates content optimization.
	// Deprecated: Use optimizer.Optimizer.Optimize instead.
	GenerateOptimization(ctx context.Context, req *models.OptimizationRequest) (*models.OptimizationResponse, error)

	// GenerateSchema generates Schema markup.
	// Deprecated: Use optimizer.Optimizer with schema strategy instead.
	GenerateSchema(ctx context.Context, content string, schemaType string) (string, error)

	// AnalyzeContent analyzes content.
	// Deprecated: Use analyzer.Scorer instead.
	AnalyzeContent(ctx context.Context, content string) (*ContentAnalysis, error)
}
