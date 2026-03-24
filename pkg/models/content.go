package models

// ContentAnalysis 内容分析结果
type ContentAnalysis struct {
	// 内容质量评分
	QualityScore float64 `json:"quality_score"`

	// 结构分析
	StructureScore float64 `json:"structure_score"`
	HeadingCount   int     `json:"heading_count"`
	ParagraphCount int     `json:"paragraph_count"`

	// 权威性分析
	AuthorityScore float64 `json:"authority_score"`
	CitationCount  int     `json:"citation_count"`
	SourceCount    int     `json:"source_count"`

	// 清晰度分析
	ClarityScore float64 `json:"clarity_score"`
	Readability  string  `json:"readability"`

	// 可引用性分析
	CitationScore float64 `json:"citation_score"`

	// Schema完整性
	SchemaScore float64 `json:"schema_score"`

	// 总体GEO评分
	GeoScore float64 `json:"geo_score"`

	// 建议
	Suggestions []string `json:"suggestions"`
}

// GeoScore GEO评分明细
type GeoScore struct {
	Structure float64 `json:"structure"` // 结构化程度 (0-100)
	Authority float64 `json:"authority"` // 权威性 (0-100)
	Clarity   float64 `json:"clarity"`   // 清晰度 (0-100)
	Citation  float64 `json:"citation"`  // 可引用性 (0-100)
	Schema    float64 `json:"schema"`    // Schema完整性 (0-100)
}

// OverallScore 计算总体评分
func (g *GeoScore) OverallScore() float64 {
	return (g.Structure + g.Authority + g.Clarity + g.Citation + g.Schema) / 5.0
}

// ScoreResult 评分结果（支持AI评分和规则评分）
type ScoreResult struct {
	*GeoScore // 复用现有5维度评分

	// 元信息
	ScoreType    string `json:"score_type"`    // "ai" 或 "rules"
	Degraded     bool   `json:"degraded"`      // 是否从AI降级到规则
	ErrorMessage string `json:"error_message"` // 降级时的错误信息（可选）
	TokensUsed   int    `json:"tokens_used"`   // AI评分消耗的token数（规则评分为0）
}
