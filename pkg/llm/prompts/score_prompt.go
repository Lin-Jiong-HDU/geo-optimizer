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
