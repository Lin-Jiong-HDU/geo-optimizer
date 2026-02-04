package strategies

import (
	"strings"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm/prompts"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// SchemaStrategy Schema标记策略
// 生成JSON-LD格式的Schema标记
type SchemaStrategy struct {
	*BaseStrategy
	schemaType string
}

// NewSchemaStrategy 创建Schema策略
func NewSchemaStrategy() *SchemaStrategy {
	return &SchemaStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategySchema, "schema"),
		schemaType:   "Article",
	}
}

// NewSchemaStrategyWithType 创建指定类型的Schema策略
func NewSchemaStrategyWithType(schemaType string) *SchemaStrategy {
	return &SchemaStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategySchema, "schema"),
		schemaType:   schemaType,
	}
}

// Validate 验证策略是否适用
func (s *SchemaStrategy) Validate(req *models.OptimizationRequest) bool {
	// Schema策略需要内容不为空
	return req.Content != ""
}

// Preprocess 预处理内容
func (s *SchemaStrategy) Preprocess(content string, req *models.OptimizationRequest) string {
	// Schema策略不需要预处理
	return content
}

// Postprocess 后处理内容
func (s *SchemaStrategy) Postprocess(content string, req *models.OptimizationRequest) string {
	// Schema策略不修改内容，只生成Schema标记
	return content
}

// BuildPrompt 构建Schema生成提示词
func (s *SchemaStrategy) BuildPrompt(req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildSchemaPrompt(req.Content, s.schemaType)
}

// SetSchemaType 设置Schema类型
func (s *SchemaStrategy) SetSchemaType(schemaType string) {
	s.schemaType = schemaType
}

// GetSchemaType 获取Schema类型
func (s *SchemaStrategy) GetSchemaType() string {
	return s.schemaType
}

// InferSchemaType 根据内容推断Schema类型
func (s *SchemaStrategy) InferSchemaType(content string) string {
	// 根据内容特征推断Schema类型
	// 这里使用简单的规则推断

	// 检查是否包含产品相关信息
	if strings.Contains(content, "产品") || strings.Contains(content, "价格") ||
		strings.Contains(content, "购买") || strings.Contains(content, "规格") {
		return "Product"
	}

	// 检查是否包含FAQ
	if strings.Contains(content, "常见问题") || strings.Contains(content, "FAQ") {
		return "FAQPage"
	}

	// 检查是否包含操作步骤
	if strings.Contains(content, "步骤") || strings.Contains(content, "方法") ||
		strings.Contains(content, "如何") || strings.Contains(content, "教程") {
		return "HowTo"
	}

	// 检查是否包含组织信息
	if strings.Contains(content, "公司") || strings.Contains(content, "企业") ||
		strings.Contains(content, "团队") || strings.Contains(content, "我们") {
		return "Organization"
	}

	// 默认使用Article
	return "Article"
}
