# GEO 优化器设计文档

## 概述

本产品是一个面向 B 端的 GEO（Generative Engine Optimization）优化引擎，以 Go 包形式提供给后端其他程序调用。通过 LLM 驱动的内容优化，提升企业在 AI 搜索引擎（ChatGPT、Perplexity、Google AI）中的引用率和推荐质量。

## 核心功能

- **内容结构化优化**：添加清晰的标题层级、信息要点、FAQ 格式
- **Schema 标记生成**：生成 JSON-LD 格式的结构化数据
- **答案优先架构**：在开头提供明确结论，后续展开说明
- **权威性增强**：添加数据来源、可信引用、作者/机构信息
- **企业/产品植入**：自然地在内容中植入企业产品和品牌信息
- **竞品差异化**：针对竞品劣势进行差异化表达

## 项目结构

```
geo-optimizer/
├── pkg/
│   ├── optimizer/              # 核心优化引擎
│   │   ├── optimizer.go        # 主优化器
│   │   ├── strategies/         # 优化策略实现
│   │   │   ├── structure.go
│   │   │   ├── authority.go
│   │   │   ├── schema.go
│   │   │   └── answer_first.go
│   │   └── enhancer/           # 内容增强器
│   ├── llm/                    # LLM 相关模块
│   │   ├── client.go           # LLM 客户端接口
│   │   ├── provider.go         # 提供商配置
│   │   ├── prompts/            # Prompt 模板
│   │   └── parser.go           # LLM 响应解析
│   ├── models/                 # 数据模型
│   │   ├── request.go
│   │   ├── response.go
│   │   ├── enterprise.go
│   │   └── content.go
│   ├── config/                 # 配置管理
│   │   ├── config.go
│   │   └── ai_profiles.go      # AI 偏好配置
│   └── analyzer/               # 内容分析器
│       └── scorer.go           # 评分系统
└── go.mod
```

## 数据模型

### 1. OptimizationRequest（优化请求）

```go
type OptimizationRequest struct {
    // 基础内容
    Content      string            `json:"content"`       // 原始内容
    Title        string            `json:"title"`         // 内容标题
    URL          string            `json:"url"`           // 发布URL

    // 企业与产品信息
    Enterprise   EnterpriseInfo    `json:"enterprise"`
    Competitors  []CompetitorInfo  `json:"competitors"`

    // 优化目标
    TargetAI     []string          `json:"target_ai"`     // ["chatgpt", "perplexity", "google_ai"]
    Keywords     []string          `json:"keywords"`

    // AI 偏好配置
    AIPreferences map[string]AIPreference `json:"ai_preferences"`

    // 优化配置
    Strategies   []StrategyType    `json:"strategies"`
    Tone         string            `json:"tone"`
    Industry     string            `json:"industry"`

    // 内容增强选项
    IncludeProductMention bool      `json:"include_product_mention"`
    MentionFrequency      string   `json:"mention_frequency"`  // "low", "medium", "high"
    CallToAction          string   `json:"call_to_action"`

    // 元数据
    Author       string            `json:"author"`
    PublishDate  time.Time         `json:"publish_date"`
    CustomData   map[string]string `json:"custom_data"`
}
```

### 2. EnterpriseInfo（企业信息）

```go
type EnterpriseInfo struct {
    // 企业基本信息
    CompanyName        string   `json:"company_name"`
    CompanyWebsite     string   `json:"company_website"`
    CompanyDescription string   `json:"company_description"`

    // 产品信息
    ProductName        string   `json:"product_name"`
    ProductURL         string   `json:"product_url"`
    ProductDescription string   `json:"product_description"`
    ProductFeatures    []string `json:"product_features"`
    USP                []string `json:"usp"`  // 独特卖点

    // 品牌信息
    BrandVoice         string   `json:"brand_voice"`
    TargetAudience     string   `json:"target_audience"`
    ValueProposition   string   `json:"value_proposition"`

    // 认证与权威性
    Certifications     []string `json:"certifications"`
    Awards             []string `json:"awards"`
    CaseStudies        []string `json:"case_studies"`
}
```

### 3. CompetitorInfo（竞品信息）

```go
type CompetitorInfo struct {
    Name               string   `json:"name"`
    Website            string   `json:"website"`
    Weaknesses         []string `json:"weaknesses"`           // 竞品劣势
    CommonObjections   []string `json:"common_objections"`    // 常见异议点
}
```

### 4. AIPreference（AI 偏好配置）

```go
type AIPreference struct {
    // 内容偏好
    ContentStyle      string   `json:"content_style"`       // "professional", "casual", "academic"
    ResponseFormat    string   `json:"response_format"`
    CitationStyle     string   `json:"citation_style"`

    // 结构偏好
    SectionStructure  []string `json:"section_structure"`
    PreferredHeading  string   `json:"preferred_heading"`

    // 关键词偏好
    KeywordDensity    float64  `json:"keyword_density"`     // 0.0-1.0
    SynonymUsage      bool     `json:"synonym_usage"`

    // 技术偏好
    CodeBlocks        bool     `json:"code_blocks"`
    TechnicalDepth    string   `json:"technical_depth"`     // "beginner", "intermediate", "advanced"

    // 自定义指令
    CustomInstructions string `json:"custom_instructions"`
}
```

### 5. OptimizationResponse（优化响应）

```go
type OptimizationResponse struct {
    // 优化后的内容
    OptimizedContent    string              `json:"optimized_content"`
    Title               string              `json:"title"`

    // 生成的组件
    SchemaMarkup        string              `json:"schema_markup"`
    FAQSection          string              `json:"faq_section"`
    Summary             string              `json:"summary"`

    // 产品植入分析
    ProductMentions     []ProductMention    `json:"product_mentions"`
    MentionCount        int                 `json:"mention_count"`

    // 竞品差异化
    DifferentiationPoints []string          `json:"differentiation_points"`

    // 优化分析
    AppliedStrategies   []StrategyType      `json:"applied_strategies"`
    Optimizations       []Optimization      `json:"optimizations"`
    ScoreBefore         float64             `json:"score_before"`
    ScoreAfter          float64             `json:"score_after"`

    // 建议
    Recommendations     []string            `json:"recommendations"`

    // 元数据
    GeneratedAt         time.Time           `json:"generated_at"`
    Version             string              `json:"version"`
    LLMModel            string              `json:"llm_model"`
    TokensUsed          int                 `json:"tokens_used"`
}
```

## LLM 模块设计

### LLM 客户端接口

```go
type LLMClient interface {
    GenerateOptimization(ctx context.Context, req *OptimizationRequest) (*OptimizationResponse, error)
    GenerateSchema(ctx context.Context, content string, schemaType string) (string, error)
    AnalyzeContent(ctx context.Context, content string) (*ContentAnalysis, error)
    GenerateStream(ctx context.Context, prompt string) (<-chan string, error)
}
```

### 支持的 LLM 提供商

- OpenAI
- Anthropic (Claude)
- Azure OpenAI
- Google (Gemini)
- 自定义端点

### 配置结构

```go
type Config struct {
    Provider    Provider `json:"provider"`
    APIKey      string   `json:"api_key"`
    BaseURL     string   `json:"base_url"`     // 自定义端点
    Model       string   `json:"model"`        // 模型名称
    MaxTokens   int      `json:"max_tokens"`
    Temperature float64  `json:"temperature"`
}
```

## 优化策略类型

```go
type StrategyType string

const (
    StrategyStructure   StrategyType = "structure"    // 结构化优化
    StrategySchema      StrategyType = "schema"       // Schema 标记生成
    StrategyAnswerFirst StrategyType = "answer_first" // 答案优先架构
    StrategyAuthority   StrategyType = "authority"    // 权威性增强
    StrategyFAQ         StrategyType = "faq"          // FAQ 生成
)
```

## 评分系统

```go
type GeoScore struct {
    Structure    float64  // 结构化程度 (0-100)
    Authority    float64  // 权威性 (0-100)
    Clarity      float64  // 清晰度 (0-100)
    Citation     float64  // 可引用性 (0-100)
    Schema       float64  // Schema 完整性 (0-100)
}
```

## 使用示例

```go
package main

import (
    "context"
    "github.com/your-org/geo-optimizer/pkg/llm"
    "github.com/your-org/geo-optimizer/pkg/models"
    "github.com/your-org/geo-optimizer/pkg/optimizer"
)

func main() {
    // 初始化 LLM 客户端
    llmClient, _ := llm.NewClient(llm.Config{
        Provider:   llm.ProviderOpenAI,
        APIKey:     "your-api-key",
        Model:      "gpt-4",
        MaxTokens:  4000,
        Temperature: 0.7,
    })

    // 创建优化器
    opt := optimizer.New(llmClient)

    // 构建优化请求
    request := &models.OptimizationRequest{
        Content: "如何选择合适的云服务提供商...",
        Title:   "云服务选择指南",

        Enterprise: models.EnterpriseInfo{
            CompanyName:     "CloudTech",
            ProductName:     "CloudTech Pro",
            ProductURL:      "https://cloudtech.com/pro",
            USP:             []string{"行业最低价格", "5分钟快速部署"},
        },

        Competitors: []models.CompetitorInfo{
            {
                Name:      "AWS",
                Weaknesses: []string{"学习曲线陡峭", "定价复杂"},
            },
        },

        TargetAI:   []string{"chatgpt", "perplexity"},
        Keywords:   []string{"云服务", "云计算"},
        Strategies: []models.StrategyType{
            models.StrategyStructure,
            models.StrategySchema,
        },

        IncludeProductMention: true,
        MentionFrequency:      "medium",
    }

    // 执行优化
    response, _ := opt.Optimize(context.Background(), request)

    // 使用优化结果
    println(response.OptimizedContent)
    println(response.SchemaMarkup)
}
```

## 技术栈

- **语言**: Go 1.21+
- **LLM SDK**: 支持多厂商（OpenAI、Anthropic、Azure、Google）
- **配置管理**: Viper
- **日志**: Zap 或 Logrus
- **测试**: Testify

## 扩展功能（可选）

1. **批量优化**: 支持批量内容处理，并发优化控制
2. **缓存机制**: 优化结果缓存，减少重复计算
3. **多模态优化**: 图片和视频内容优化建议
4. **A/B 测试**: 生成多个优化版本，提供对比分析
5. **流式生成**: 实时流式输出优化内容
