package strategies

import (
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm/prompts"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// AuthorityStrategy enhances content authority by adding data support and citations.
type AuthorityStrategy struct {
	*BaseStrategy
}

// NewAuthorityStrategy creates a new authority strategy.
func NewAuthorityStrategy() *AuthorityStrategy {
	return &AuthorityStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategyAuthority, "authority"),
	}
}

// Validate checks if the strategy is applicable.
func (a *AuthorityStrategy) Validate(req *models.OptimizationRequest) bool {
	return req.Content != ""
}

// Preprocess preprocesses content before optimization.
func (a *AuthorityStrategy) Preprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// Postprocess postprocesses content after optimization.
func (a *AuthorityStrategy) Postprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// BuildPrompt builds the authority enhancement prompt.
func (a *AuthorityStrategy) BuildPrompt(req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildAuthorityPrompt(req.Content, req.Enterprise)
}

// BuildPromptWithContent builds the prompt with specified content.
func (a *AuthorityStrategy) BuildPromptWithContent(content string, req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildAuthorityPrompt(content, req.Enterprise)
}
