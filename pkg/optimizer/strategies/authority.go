package strategies

import (
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm/prompts"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// AuthorityStrategy 权威性增强策略
// 通过添加数据支撑、引用来源等方式增强内容的权威性
type AuthorityStrategy struct {
	*BaseStrategy
}

// NewAuthorityStrategy 创建权威性策略
func NewAuthorityStrategy() *AuthorityStrategy {
	return &AuthorityStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategyAuthority, "authority"),
	}
}

// Validate 验证策略是否适用
func (a *AuthorityStrategy) Validate(req *models.OptimizationRequest) bool {
	// 权威性策略需要内容不为空
	return req.Content != ""
}

// Preprocess 预处理内容
func (a *AuthorityStrategy) Preprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// Postprocess 后处理内容
func (a *AuthorityStrategy) Postprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// BuildPrompt 构建权威性增强提示词
func (a *AuthorityStrategy) BuildPrompt(req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildAuthorityPrompt(req.Content, req.Enterprise)
}

// BuildPromptWithContent 使用指定内容构建 Prompt
func (a *AuthorityStrategy) BuildPromptWithContent(content string, req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildAuthorityPrompt(content, req.Enterprise)
}
