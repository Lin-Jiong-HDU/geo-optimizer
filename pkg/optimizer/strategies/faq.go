package strategies

import (
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm/prompts"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// FAQStrategy generates FAQ sections to improve content citability.
type FAQStrategy struct {
	*BaseStrategy
	faqCount int
}

// NewFAQStrategy creates a new FAQ strategy.
func NewFAQStrategy() *FAQStrategy {
	return &FAQStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategyFAQ, "faq"),
		faqCount:     5,
	}
}

// NewFAQStrategyWithCount creates an FAQ strategy with a specific count.
func NewFAQStrategyWithCount(count int) *FAQStrategy {
	if count <= 0 {
		count = 5
	}
	return &FAQStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategyFAQ, "faq"),
		faqCount:     count,
	}
}

// Validate checks if the strategy is applicable.
func (f *FAQStrategy) Validate(req *models.OptimizationRequest) bool {
	return req.Content != ""
}

// Preprocess preprocesses content before optimization.
func (f *FAQStrategy) Preprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// Postprocess postprocesses content after optimization.
func (f *FAQStrategy) Postprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// BuildPrompt builds the FAQ generation prompt.
func (f *FAQStrategy) BuildPrompt(req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildFAQPromptWithEnterprise(req.Content, f.faqCount, req.Enterprise)
}

// BuildPromptWithContent builds the prompt with specified content.
func (f *FAQStrategy) BuildPromptWithContent(content string, req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildFAQPromptWithEnterprise(content, f.faqCount, req.Enterprise)
}

// SetFAQCount sets the FAQ count.
func (f *FAQStrategy) SetFAQCount(count int) {
	if count > 0 {
		f.faqCount = count
	}
}

// GetFAQCount returns the FAQ count.
func (f *FAQStrategy) GetFAQCount() int {
	return f.faqCount
}
