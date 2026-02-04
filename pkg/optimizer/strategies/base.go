package strategies

import (
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// Strategy 优化策略接口（复制自父包，避免循环导入）
type Strategy interface {
	// Name 返回策略名称
	Name() string

	// Type 返回策略类型
	Type() models.StrategyType

	// Preprocess 预处理内容（在调用 LLM 前执行）
	Preprocess(content string, req *models.OptimizationRequest) string

	// Postprocess 后处理内容（在 LLM 返回后执行）
	Postprocess(content string, req *models.OptimizationRequest) string

	// BuildPrompt 构建策略特定的 Prompt
	BuildPrompt(req *models.OptimizationRequest) string

	// Validate 验证策略是否适用于当前请求
	Validate(req *models.OptimizationRequest) bool
}

// BaseStrategy 基础策略实现
type BaseStrategy struct {
	strategyType models.StrategyType
	name         string
}

// NewBaseStrategy 创建基础策略
func NewBaseStrategy(strategyType models.StrategyType, name string) *BaseStrategy {
	return &BaseStrategy{
		strategyType: strategyType,
		name:         name,
	}
}

// Name 返回策略名称
func (b *BaseStrategy) Name() string {
	return b.name
}

// Type 返回策略类型
func (b *BaseStrategy) Type() models.StrategyType {
	return b.strategyType
}

// Preprocess 预处理内容（默认实现：不处理）
func (b *BaseStrategy) Preprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// Postprocess 后处理内容（默认实现：不处理）
func (b *BaseStrategy) Postprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// Validate 验证策略是否适用（默认实现：总是适用）
func (b *BaseStrategy) Validate(req *models.OptimizationRequest) bool {
	return true
}

// BuildPrompt 构建策略特定的 Prompt（默认实现：需要在子类中覆盖）
func (b *BaseStrategy) BuildPrompt(req *models.OptimizationRequest) string {
	return req.Content
}
