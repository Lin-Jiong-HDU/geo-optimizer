package strategies

import (
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm/prompts"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// FAQStrategy FAQ生成策略
// 为内容生成常见问题部分，提高内容的可引用性
type FAQStrategy struct {
	*BaseStrategy
	faqCount int
}

// NewFAQStrategy 创建FAQ策略
func NewFAQStrategy() *FAQStrategy {
	return &FAQStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategyFAQ, "faq"),
		faqCount:     5,
	}
}

// NewFAQStrategyWithCount 创建指定数量的FAQ策略
func NewFAQStrategyWithCount(count int) *FAQStrategy {
	if count <= 0 {
		count = 5
	}
	return &FAQStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategyFAQ, "faq"),
		faqCount:     count,
	}
}

// Validate 验证策略是否适用
func (f *FAQStrategy) Validate(req *models.OptimizationRequest) bool {
	// FAQ策略需要内容不为空
	return req.Content != ""
}

// Preprocess 预处理内容
func (f *FAQStrategy) Preprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// Postprocess 后处理内容
func (f *FAQStrategy) Postprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// BuildPrompt 构建FAQ生成提示词
func (f *FAQStrategy) BuildPrompt(req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildFAQPromptWithEnterprise(req.Content, f.faqCount, req.Enterprise)
}

// SetFAQCount 设置FAQ数量
func (f *FAQStrategy) SetFAQCount(count int) {
	if count > 0 {
		f.faqCount = count
	}
}

// GetFAQCount 获取FAQ数量
func (f *FAQStrategy) GetFAQCount() int {
	return f.faqCount
}
