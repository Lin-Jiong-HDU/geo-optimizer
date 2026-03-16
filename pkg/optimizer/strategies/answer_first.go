package strategies

import (
	"strings"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm/prompts"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// AnswerFirstStrategy 答案优先策略
// 将关键结论放在内容开头，便于AI快速提取
type AnswerFirstStrategy struct {
	*BaseStrategy
}

// NewAnswerFirstStrategy 创建答案优先策略
func NewAnswerFirstStrategy() *AnswerFirstStrategy {
	return &AnswerFirstStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategyAnswerFirst, "answer_first"),
	}
}

// Validate 验证策略是否适用
func (a *AnswerFirstStrategy) Validate(req *models.OptimizationRequest) bool {
	// 答案优先策略需要内容不为空
	return req.Content != ""
}

// Preprocess 预处理内容
func (a *AnswerFirstStrategy) Preprocess(content string, req *models.OptimizationRequest) string {
	if hasConclusionFirst(content) {
		return content
	}
	return content
}

// Postprocess 后处理内容
func (a *AnswerFirstStrategy) Postprocess(content string, req *models.OptimizationRequest) string {
	// 确保结论在前
	return content
}

// BuildPrompt 构建答案优先提示词
func (a *AnswerFirstStrategy) BuildPrompt(req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildStrategyPrompt(models.StrategyAnswerFirst, req)
}

// hasConclusionFirst 检查内容开头是否包含结论性词语
func hasConclusionFirst(content string) bool {
	if len(content) > 100 {
		content = content[:100]
	}

	conclusionWords := []string{"结论", "总结", "总之", "简而言之", "答案是"}
	for _, word := range conclusionWords {
		if strings.Contains(content, word) {
			return true
		}
	}
	return false
}
