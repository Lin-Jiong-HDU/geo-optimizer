package strategies

import (
	"strings"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm/prompts"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// AnswerFirstStrategy places key conclusions at the beginning of content for easy AI extraction.
type AnswerFirstStrategy struct {
	*BaseStrategy
}

// NewAnswerFirstStrategy creates a new answer-first strategy.
func NewAnswerFirstStrategy() *AnswerFirstStrategy {
	return &AnswerFirstStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategyAnswerFirst, "answer_first"),
	}
}

// Validate checks if the strategy is applicable.
func (a *AnswerFirstStrategy) Validate(req *models.OptimizationRequest) bool {
	return req.Content != ""
}

// Preprocess preprocesses content before optimization.
func (a *AnswerFirstStrategy) Preprocess(content string, req *models.OptimizationRequest) string {
	if hasConclusionFirst(content) {
		return content
	}
	return content
}

// Postprocess postprocesses content after optimization.
func (a *AnswerFirstStrategy) Postprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// BuildPrompt builds the answer-first optimization prompt.
func (a *AnswerFirstStrategy) BuildPrompt(req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildStrategyPrompt(models.StrategyAnswerFirst, req)
}

// BuildPromptWithContent builds the prompt with specified content.
func (a *AnswerFirstStrategy) BuildPromptWithContent(content string, req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildStrategyPromptWithContent(models.StrategyAnswerFirst, content, req)
}

// hasConclusionFirst checks if content starts with conclusion words.
func hasConclusionFirst(content string) bool {
	if len(content) > 100 {
		content = content[:100]
	}

	conclusionWords := []string{"conclusion", "summary", "in short", "the answer is", "结论", "总结", "总之", "简而言之", "答案是"}
	for _, word := range conclusionWords {
		if strings.Contains(content, word) {
			return true
		}
	}
	return false
}
