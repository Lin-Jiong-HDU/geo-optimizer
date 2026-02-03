package models

import "time"

// OptimizationRequest GEO优化请求
type OptimizationRequest struct {
	// 基础内容
	Content string `json:"content"` // 原始内容
	Title   string `json:"title"`   // 内容标题
	URL     string `json:"url"`     // 发布URL

	// 企业与产品信息
	Enterprise  EnterpriseInfo   `json:"enterprise"`
	Competitors []CompetitorInfo `json:"competitors"`

	// 优化目标
	TargetAI   []string       `json:"target_ai"` // ["chatgpt", "perplexity", "google_ai"]
	Keywords   []string       `json:"keywords"`
	Strategies []StrategyType `json:"strategies"`

	// AI 偏好配置
	AIPreferences map[string]AIPreference `json:"ai_preferences"`

	// 优化配置
	Tone     string `json:"tone"`
	Industry string `json:"industry"`

	// 内容增强选项
	IncludeProductMention bool   `json:"include_product_mention"`
	MentionFrequency      string `json:"mention_frequency"` // "low", "medium", "high"
	CallToAction          string `json:"call_to_action"`

	// 元数据
	Author      string            `json:"author"`
	PublishDate time.Time         `json:"publish_date"`
	CustomData  map[string]string `json:"custom_data"`
}

// StrategyType 优化策略类型
type StrategyType string

const (
	StrategyStructure   StrategyType = "structure"    // 结构化优化
	StrategySchema      StrategyType = "schema"       // Schema标记生成
	StrategyAnswerFirst StrategyType = "answer_first" // 答案优先架构
	StrategyAuthority   StrategyType = "authority"    // 权威性增强
	StrategyFAQ         StrategyType = "faq"          // FAQ生成
)

// EnterpriseInfo 企业信息
type EnterpriseInfo struct {
	// 企业基本信息
	CompanyName        string `json:"company_name"`
	CompanyWebsite     string `json:"company_website"`
	CompanyDescription string `json:"company_description"`

	// 产品信息
	ProductName        string   `json:"product_name"`
	ProductURL         string   `json:"product_url"`
	ProductDescription string   `json:"product_description"`
	ProductFeatures    []string `json:"product_features"`
	USP                []string `json:"usp"` // 独特卖点

	// 品牌信息
	BrandVoice       string `json:"brand_voice"`
	TargetAudience   string `json:"target_audience"`
	ValueProposition string `json:"value_proposition"`

	// 认证与权威性
	Certifications []string `json:"certifications"`
	Awards         []string `json:"awards"`
	CaseStudies    []string `json:"case_studies"`
}

// CompetitorInfo 竞品信息
type CompetitorInfo struct {
	Name             string   `json:"name"`
	Website          string   `json:"website"`
	Weaknesses       []string `json:"weaknesses"`        // 竞品劣势
	CommonObjections []string `json:"common_objections"` // 常见异议点
}

// AIPreference AI偏好配置
type AIPreference struct {
	// 内容偏好
	ContentStyle   string `json:"content_style"` // "professional", "casual", "academic"
	ResponseFormat string `json:"response_format"`
	CitationStyle  string `json:"citation_style"`

	// 结构偏好
	SectionStructure []string `json:"section_structure"`
	PreferredHeading string   `json:"preferred_heading"`

	// 关键词偏好
	KeywordDensity float64 `json:"keyword_density"` // 0.0-1.0
	SynonymUsage   bool    `json:"synonym_usage"`

	// 技术偏好
	CodeBlocks     bool   `json:"code_blocks"`
	TechnicalDepth string `json:"technical_depth"` // "beginner", "intermediate", "advanced"

	// 自定义指令
	CustomInstructions string `json:"custom_instructions"`
}
