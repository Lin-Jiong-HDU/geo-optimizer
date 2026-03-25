package models

import "time"

// OptimizationRequest represents a GEO optimization request.
type OptimizationRequest struct {
	Content string `json:"content"`
	Title   string `json:"title"`
	URL     string `json:"url"`

	Enterprise  EnterpriseInfo   `json:"enterprise"`
	Competitors []CompetitorInfo `json:"competitors"`

	TargetAI   []string       `json:"target_ai"`
	Keywords   []string       `json:"keywords"`
	Strategies []StrategyType `json:"strategies"`

	AIPreferences map[string]AIPreference `json:"ai_preferences"`

	Tone     string `json:"tone"`
	Industry string `json:"industry"`

	IncludeProductMention bool   `json:"include_product_mention"`
	MentionFrequency      string `json:"mention_frequency"`
	CallToAction          string `json:"call_to_action"`

	Author      string            `json:"author"`
	PublishDate time.Time         `json:"publish_date"`
	CustomData  map[string]string `json:"custom_data"`
}

// StrategyType represents the type of optimization strategy.
type StrategyType string

// Strategy type constants.
const (
	StrategyStructure   StrategyType = "structure"
	StrategySchema      StrategyType = "schema"
	StrategyAnswerFirst StrategyType = "answer_first"
	StrategyAuthority   StrategyType = "authority"
	StrategyFAQ         StrategyType = "faq"
)

// EnterpriseInfo contains enterprise and product information.
type EnterpriseInfo struct {
	CompanyName        string `json:"company_name"`
	CompanyWebsite     string `json:"company_website"`
	CompanyDescription string `json:"company_description"`

	ProductName        string   `json:"product_name"`
	ProductURL         string   `json:"product_url"`
	ProductDescription string   `json:"product_description"`
	ProductFeatures    []string `json:"product_features"`
	USP                []string `json:"usp"`

	BrandVoice       string `json:"brand_voice"`
	TargetAudience   string `json:"target_audience"`
	ValueProposition string `json:"value_proposition"`

	Certifications []string `json:"certifications"`
	Awards         []string `json:"awards"`
	CaseStudies    []string `json:"case_studies"`
}

// CompetitorInfo contains competitor information.
type CompetitorInfo struct {
	Name             string   `json:"name"`
	Website          string   `json:"website"`
	Weaknesses       []string `json:"weaknesses"`
	CommonObjections []string `json:"common_objections"`
}

// AIPreference contains AI platform preference configuration.
type AIPreference struct {
	ContentStyle   string `json:"content_style"`
	ResponseFormat string `json:"response_format"`
	CitationStyle  string `json:"citation_style"`

	SectionStructure []string `json:"section_structure"`
	PreferredHeading string   `json:"preferred_heading"`

	KeywordDensity float64 `json:"keyword_density"`
	SynonymUsage   bool    `json:"synonym_usage"`

	CodeBlocks     bool   `json:"code_blocks"`
	TechnicalDepth string `json:"technical_depth"`

	CustomInstructions string `json:"custom_instructions"`
}
