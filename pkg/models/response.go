package models

import "time"

// OptimizationResponse GEO优化响应
type OptimizationResponse struct {
	// 优化后的内容
	OptimizedContent string `json:"optimized_content"`
	Title            string `json:"title"`

	// 生成的组件
	SchemaMarkup string `json:"schema_markup"`
	FAQSection   string `json:"faq_section"`
	Summary      string `json:"summary"`

	// 产品植入分析
	ProductMentions []ProductMention `json:"product_mentions"`
	MentionCount    int              `json:"mention_count"`

	// 竞品差异化
	DifferentiationPoints []string `json:"differentiation_points"`

	// 优化分析
	AppliedStrategies []StrategyType `json:"applied_strategies"`
	Optimizations     []Optimization `json:"optimizations"`
	ScoreBefore       float64        `json:"score_before"`
	ScoreAfter        float64        `json:"score_after"`

	// 建议
	Recommendations []string `json:"recommendations"`

	// 元数据
	GeneratedAt time.Time `json:"generated_at"`
	Version     string    `json:"version"`
	LLMModel    string    `json:"llm_model"`
	TokensUsed  int       `json:"tokens_used"`
}

// ProductMention 产品植入信息
type ProductMention struct {
	Position    int    `json:"position"`     // 位置
	Context     string `json:"context"`      // 上下文
	ImpactLevel string `json:"impact_level"` // "high", "medium", "low"
}

// Optimization 优化详情
type Optimization struct {
	Type        StrategyType `json:"type"`
	Description string       `json:"description"`
	Before      string       `json:"before"`
	After       string       `json:"after"`
}
