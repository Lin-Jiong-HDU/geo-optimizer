package prompts

import "fmt"

// AI评分系统提示词
const SystemPromptScoring = "你是一个GEO（生成引擎优化）评分专家，擅长评估内容在AI搜索引擎中的可见性和引用潜力。请基于内容质量而非格式进行评分。"

// AI评分Prompt模板
const PromptGEOScore = `请对以下内容进行GEO评分，评估其在AI搜索引擎中的可见性和引用潜力。

内容:
%s

请从以下5个维度进行评分（0-100分）：

1. **Structure（结构化）**: 内容组织是否清晰，是否有明确的逻辑层次
2. **Authority（权威性）**: 是否有数据支撑、来源引用、专业深度
3. **Clarity（清晰度）**: 表达是否简洁明了，是否易于理解
4. **Citation（可引用性）**: 是否有可直接引用的结论、事实和建议
5. **Schema（结构化数据）**: 是否包含Schema.org/JSON-LD标记

请以JSON格式返回，不要包含markdown代码块标记：
{
  "structure": <0-100>,
  "structure_reason": "<简短理由>",
  "authority": <0-100>,
  "authority_reason": "<简短理由>",
  "clarity": <0-100>,
  "clarity_reason": "<简短理由>",
  "citation": <0-100>,
  "citation_reason": "<简短理由>",
  "schema": <0-100>,
  "schema_reason": "<简短理由>"
}`

// BuildScorePrompt 构建评分Prompt
func BuildScorePrompt(content string) string {
	return fmt.Sprintf(PromptGEOScore, content)
}

// AiScoreResponse AI评分响应结构（内部使用）
type AiScoreResponse struct {
	Structure       float64 `json:"structure"`
	StructureReason string  `json:"structure_reason"`
	Authority       float64 `json:"authority"`
	AuthorityReason string  `json:"authority_reason"`
	Clarity         float64 `json:"clarity"`
	ClarityReason   string  `json:"clarity_reason"`
	Citation        float64 `json:"citation"`
	CitationReason  string  `json:"citation_reason"`
	Schema          float64 `json:"schema"`
	SchemaReason    string  `json:"schema_reason"`
}

// SystemPromptScoringWithSuggestions 带建议的评分系统提示词
const SystemPromptScoringWithSuggestions = "你是一个GEO（生成引擎优化）评分专家，擅长评估内容在AI搜索引擎中的可见性和引用潜力，并给出具体可行的改进建议。"

// PromptGEOScoreWithSuggestions 带建议的评分Prompt模板
const PromptGEOScoreWithSuggestions = `请对以下内容进行GEO评分，并给出改进建议。

内容:
%s

请从以下5个维度进行评分（0-100分），并根据实际需要给出改进建议：
- 分数高的维度可以少给或不给建议
- 分数低的维度应多给具体建议

请以JSON格式返回，不要包含markdown代码块标记：
{
  "scores": {
    "structure": <0-100>,
    "structure_reason": "<简短理由>",
    "authority": <0-100>,
    "authority_reason": "<简短理由>",
    "clarity": <0-100>,
    "clarity_reason": "<简短理由>",
    "citation": <0-100>,
    "citation_reason": "<简短理由>",
    "schema": <0-100>,
    "schema_reason": "<简短理由>"
  },
  "dimension_suggestions": {
    "structure": [
      {"issue": "<问题描述>", "direction": "<改进方向>", "priority": "<high/medium/low>", "estimated_gain": <0-20>, "example": "<示例片段>"}
    ]
  },
  "top_suggestions": [
    {"issue": "<问题描述>", "direction": "<改进方向>", "priority": "<high/medium/low>", "estimated_gain": <0-20>, "example": "<示例片段>"}
  ]
}

注意：
- dimension_suggestions 中只包含需要改进的维度，高分的维度可以省略
- top_suggestions 按优先级和 estimated_gain 排序，最多5条
- example 字段可选，提供简短的改进示例
- estimated_gain 表示改进后该维度预估可提升的分数`

// BuildScoreWithSuggestionsPrompt 构建带建议的评分Prompt
func BuildScoreWithSuggestionsPrompt(content string) string {
	return fmt.Sprintf(PromptGEOScoreWithSuggestions, content)
}

// AiScoreWithSuggestionsResponse 带建议的评分响应结构
type AiScoreWithSuggestionsResponse struct {
	Scores               AiScoreResponse               `json:"scores"`
	DimensionSuggestions map[string][]AiSuggestionItem `json:"dimension_suggestions"`
	TopSuggestions       []AiSuggestionItem            `json:"top_suggestions"`
}

// AiSuggestionItem AI建议项
type AiSuggestionItem struct {
	Issue         string  `json:"issue"`
	Direction     string  `json:"direction"`
	Priority      string  `json:"priority"`
	EstimatedGain float64 `json:"estimated_gain"`
	Example       string  `json:"example"`
}
