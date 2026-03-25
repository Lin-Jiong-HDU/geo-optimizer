package models

// ContentAnalysis represents content analysis results.
type ContentAnalysis struct {
	QualityScore float64 `json:"quality_score"`

	StructureScore float64 `json:"structure_score"`
	HeadingCount   int     `json:"heading_count"`
	ParagraphCount int     `json:"paragraph_count"`

	AuthorityScore float64 `json:"authority_score"`
	CitationCount  int     `json:"citation_count"`
	SourceCount    int     `json:"source_count"`

	ClarityScore float64 `json:"clarity_score"`
	Readability  string  `json:"readability"`

	CitationScore float64 `json:"citation_score"`

	SchemaScore float64 `json:"schema_score"`

	GeoScore float64 `json:"geo_score"`

	Suggestions []string `json:"suggestions"`
}

// GeoScore represents GEO scoring details across five dimensions.
type GeoScore struct {
	Structure float64 `json:"structure"`
	Authority float64 `json:"authority"`
	Clarity   float64 `json:"clarity"`
	Citation  float64 `json:"citation"`
	Schema    float64 `json:"schema"`
}

// OverallScore calculates the overall GEO score.
func (g *GeoScore) OverallScore() float64 {
	return (g.Structure + g.Authority + g.Clarity + g.Citation + g.Schema) / 5.0
}

// ScoreResult represents a scoring result supporting both AI and rule-based scoring.
type ScoreResult struct {
	*GeoScore

	ScoreType    string `json:"score_type"`
	Degraded     bool   `json:"degraded"`
	ErrorMessage string `json:"error_message"`
	TokensUsed   int    `json:"tokens_used"`
}

// Suggestion represents a single improvement suggestion.
type Suggestion struct {
	Issue         string  `json:"issue"`
	Direction     string  `json:"direction"`
	Priority      string  `json:"priority"`
	EstimatedGain float64 `json:"estimated_gain"`
	Example       string  `json:"example"`
}

// ScoreResultWithSuggestions represents a scoring result with improvement suggestions.
type ScoreResultWithSuggestions struct {
	*ScoreResult
	DimensionSuggestions map[string][]Suggestion `json:"dimension_suggestions"`
	TopSuggestions       []Suggestion            `json:"top_suggestions"`
}
