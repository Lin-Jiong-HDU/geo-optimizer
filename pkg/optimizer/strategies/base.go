package strategies

import (
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// Strategy defines the interface for optimization strategies.
type Strategy interface {
	Name() string
	Type() models.StrategyType
	Preprocess(content string, req *models.OptimizationRequest) string
	Postprocess(content string, req *models.OptimizationRequest) string
	BuildPrompt(req *models.OptimizationRequest) string
	BuildPromptWithContent(content string, req *models.OptimizationRequest) string
	Validate(req *models.OptimizationRequest) bool
}

// BaseStrategy provides a base implementation for optimization strategies.
type BaseStrategy struct {
	strategyType models.StrategyType
	name         string
}

// NewBaseStrategy creates a new base strategy.
func NewBaseStrategy(strategyType models.StrategyType, name string) *BaseStrategy {
	return &BaseStrategy{
		strategyType: strategyType,
		name:         name,
	}
}

// Name returns the strategy name.
func (b *BaseStrategy) Name() string {
	return b.name
}

// Type returns the strategy type.
func (b *BaseStrategy) Type() models.StrategyType {
	return b.strategyType
}

// Preprocess preprocesses content before LLM call.
func (b *BaseStrategy) Preprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// Postprocess postprocesses content after LLM response.
func (b *BaseStrategy) Postprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// Validate checks if the strategy is applicable to the request.
func (b *BaseStrategy) Validate(req *models.OptimizationRequest) bool {
	return true
}

// BuildPrompt builds the strategy-specific prompt.
func (b *BaseStrategy) BuildPrompt(req *models.OptimizationRequest) string {
	return req.Content
}

// BuildPromptWithContent builds the prompt with specified content for cumulative optimization.
func (b *BaseStrategy) BuildPromptWithContent(content string, req *models.OptimizationRequest) string {
	tempReq := &models.OptimizationRequest{
		Content:       content,
		Title:         req.Title,
		Enterprise:    req.Enterprise,
		Competitors:   req.Competitors,
		TargetAI:      req.TargetAI,
		AIPreferences: req.AIPreferences,
		Keywords:      req.Keywords,
		Strategies:    req.Strategies,
	}
	return b.BuildPrompt(tempReq)
}
