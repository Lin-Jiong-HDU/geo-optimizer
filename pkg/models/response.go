package models

import "time"

// OptimizationResponse represents a GEO optimization response.
type OptimizationResponse struct {
	OptimizedContent string `json:"optimized_content"`
	Title            string `json:"title"`

	SchemaMarkup string `json:"schema_markup"`
	FAQSection   string `json:"faq_section"`
	Summary      string `json:"summary"`

	ProductMentions []ProductMention `json:"product_mentions"`
	MentionCount    int              `json:"mention_count"`

	DifferentiationPoints []string `json:"differentiation_points"`

	AppliedStrategies []StrategyType `json:"applied_strategies"`
	Optimizations     []Optimization `json:"optimizations"`
	ScoreBefore       float64        `json:"score_before"`
	ScoreAfter        float64        `json:"score_after"`

	Recommendations []string `json:"recommendations"`

	GeneratedAt time.Time `json:"generated_at"`
	Version     string    `json:"version"`
	LLMModel    string    `json:"llm_model"`
	TokensUsed  int       `json:"tokens_used"`
}

// ProductMention represents a product mention in content.
type ProductMention struct {
	Position    int    `json:"position"`
	Context     string `json:"context"`
	ImpactLevel string `json:"impact_level"`
}

// Optimization represents optimization details.
type Optimization struct {
	Type        StrategyType `json:"type"`
	Description string       `json:"description"`
	Before      string       `json:"before"`
	After       string       `json:"after"`
}
