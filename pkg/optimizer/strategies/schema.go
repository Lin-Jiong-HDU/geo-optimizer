package strategies

import (
	"strings"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm/prompts"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// SchemaStrategy generates JSON-LD format Schema markup.
type SchemaStrategy struct {
	*BaseStrategy
	schemaType string
}

// NewSchemaStrategy creates a new Schema strategy.
func NewSchemaStrategy() *SchemaStrategy {
	return &SchemaStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategySchema, "schema"),
		schemaType:   "Article",
	}
}

// NewSchemaStrategyWithType creates a Schema strategy with a specific type.
func NewSchemaStrategyWithType(schemaType string) *SchemaStrategy {
	return &SchemaStrategy{
		BaseStrategy: NewBaseStrategy(models.StrategySchema, "schema"),
		schemaType:   schemaType,
	}
}

// Validate checks if the strategy is applicable.
func (s *SchemaStrategy) Validate(req *models.OptimizationRequest) bool {
	return req.Content != ""
}

// Preprocess preprocesses content before optimization.
func (s *SchemaStrategy) Preprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// Postprocess postprocesses content after optimization.
func (s *SchemaStrategy) Postprocess(content string, req *models.OptimizationRequest) string {
	return content
}

// BuildPrompt builds the Schema generation prompt.
func (s *SchemaStrategy) BuildPrompt(req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildSchemaPrompt(req.Content, s.schemaType)
}

// BuildPromptWithContent builds the prompt with specified content.
func (s *SchemaStrategy) BuildPromptWithContent(content string, req *models.OptimizationRequest) string {
	builder := prompts.NewBuilder()
	return builder.BuildSchemaPrompt(content, s.schemaType)
}

// SetSchemaType sets the Schema type.
func (s *SchemaStrategy) SetSchemaType(schemaType string) {
	s.schemaType = schemaType
}

// GetSchemaType returns the Schema type.
func (s *SchemaStrategy) GetSchemaType() string {
	return s.schemaType
}

// InferSchemaType infers the Schema type based on content features.
func (s *SchemaStrategy) InferSchemaType(content string) string {
	if strings.Contains(content, "product") || strings.Contains(content, "price") ||
		strings.Contains(content, "buy") || strings.Contains(content, "specification") ||
		strings.Contains(content, "产品") || strings.Contains(content, "价格") ||
		strings.Contains(content, "购买") || strings.Contains(content, "规格") {
		return "Product"
	}

	if strings.Contains(content, "FAQ") || strings.Contains(content, "常见问题") {
		return "FAQPage"
	}

	if strings.Contains(content, "step") || strings.Contains(content, "how to") ||
		strings.Contains(content, "tutorial") ||
		strings.Contains(content, "步骤") || strings.Contains(content, "方法") ||
		strings.Contains(content, "如何") || strings.Contains(content, "教程") {
		return "HowTo"
	}

	if strings.Contains(content, "company") || strings.Contains(content, "organization") ||
		strings.Contains(content, "team") || strings.Contains(content, "about us") ||
		strings.Contains(content, "公司") || strings.Contains(content, "企业") ||
		strings.Contains(content, "团队") || strings.Contains(content, "我们") {
		return "Organization"
	}

	return "Article"
}
