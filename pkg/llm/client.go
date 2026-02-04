package llm

import (
	"context"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// LLMClient LLM客户端接口
type LLMClient interface {
	// Chat 通用对话接口
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}

// ContentAnalysis 内容分析结果
type ContentAnalysis struct {
	// 结构化评分
	StructureScore float64 `json:"structure_score"`
	// 权威性评分
	AuthorityScore float64 `json:"authority_score"`
	// 清晰度评分
	ClarityScore float64 `json:"clarity_score"`
	// 可引用性评分
	CitationScore float64 `json:"citation_score"`
	// Schema完整性评分
	SchemaScore float64 `json:"schema_score"`
	// 总分
	TotalScore float64 `json:"total_score"`
	// 关键词提取
	Keywords []string `json:"keywords"`
	// 建议改进点
	Suggestions []string `json:"suggestions"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Message 聊天消息
type Message struct {
	Role    string `json:"role"` // system, user, assistant, tool
	Content string `json:"content"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Content      string `json:"content"`
	TokensUsed   int    `json:"tokens_used"`
	Model        string `json:"model"`
	FinishReason string `json:"finish_reason"`
}

// NewClient 创建LLM客户端
func NewClient(config Config) (LLMClient, error) {
	switch config.Provider {
	case ProviderGLM:
		return NewGLMClient(config)
	default:
		return NewGLMClient(config)
	}
}

// DeprecatedClient 包含已弃用方法的客户端接口
// Deprecated: Use optimizer.Optimizer instead of these methods.
type DeprecatedClient interface {
	LLMClient

	// GenerateOptimization 生成内容优化
	// Deprecated: Use optimizer.Optimizer.Optimize instead.
	GenerateOptimization(ctx context.Context, req *models.OptimizationRequest) (*models.OptimizationResponse, error)

	// GenerateSchema 生成Schema标记
	// Deprecated: Use optimizer.Optimizer with schema strategy instead.
	GenerateSchema(ctx context.Context, content string, schemaType string) (string, error)

	// AnalyzeContent 分析内容
	// Deprecated: Use analyzer.Scorer instead.
	AnalyzeContent(ctx context.Context, content string) (*ContentAnalysis, error)
}
