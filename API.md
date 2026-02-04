# GEO Optimizer - API 使用文档

## 概述

`geo-optimizer` 是一个 Go 语言编写的 GEO（生成式搜索引擎优化）优化引擎包，用于优化内容以提升在 AI 搜索引擎（如 ChatGPT、Perplexity、Google AI 等）中的可见性和引用率。

**模块路径**: `github.com/Lin-Jiong-HDU/geo-optimizer`

**Go 版本要求**: 1.21+

---

## 目录

- [安装](#安装)
- [快速开始](#快速开始)
- [核心概念](#核心概念)
- [API 参考](#api-参考)
  - [LLM 客户端](#llm-客户端)
  - [优化器](#优化器)
  - [评分器](#评分器)
  - [解析器](#解析器)
  - [数据模型](#数据模型)
- [优化策略](#优化策略)
- [AI 平台预设](#ai-平台预设)
- [完整示例](#完整示例)

---

## 安装

```bash
go get github.com/Lin-Jiong-HDU/geo-optimizer
```

---

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"
)

func main() {
    // 1. 创建 LLM 客户端
    client, err := llm.NewClient(llm.Config{
        Provider: llm.ProviderGLM,
        APIKey:   "your-api-key",
        Model:    "glm-4.7",
    })
    if err != nil {
        log.Fatal(err)
    }

    // 2. 创建优化器
    opt := optimizer.New(client)

    // 3. 构建优化请求
    req := &models.OptimizationRequest{
        Content: "如何选择合适的云服务提供商？需要考虑性能、价格、可靠性等因素...",
        Title:   "云服务选择指南",

        Enterprise: models.EnterpriseInfo{
            CompanyName:        "CloudTech",
            ProductName:        "CloudTech Pro",
            ProductDescription: "企业级云计算解决方案",
            USP:                []string{"行业最低价格", "5分钟快速部署"},
        },

        TargetAI:   []string{"chatgpt", "perplexity"},
        Keywords:   []string{"云服务", "云计算", "云服务器"},
        Strategies: []models.StrategyType{
            models.StrategyStructure,
            models.StrategySchema,
            models.StrategyFAQ,
        },
    }

    // 4. 执行优化
    ctx := context.Background()
    resp, err := opt.Optimize(ctx, req)
    if err != nil {
        log.Fatal(err)
    }

    // 5. 使用优化结果
    fmt.Printf("优化前评分: %.2f\n", resp.ScoreBefore)
    fmt.Printf("优化后评分: %.2f\n", resp.ScoreAfter)
    fmt.Printf("优化内容:\n%s\n", resp.OptimizedContent)
}
```

---

## 核心概念

### GEO 优化流程

```
┌─────────────────┐
│  原始内容输入    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  优化前评分      │  analyzer.Scorer
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  应用优化策略    │  Strategy 1 → Strategy 2 → ...
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  LLM 生成内容    │  llm.LLMClient.Chat
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  解析与后处理    │  llm.Parser
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  优化后评分      │  analyzer.Scorer
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  返回优化响应    │  OptimizationResponse
└─────────────────┘
```

### 架构分层

| 层级 | 包路径 | 职责 |
|------|--------|------|
| **模型层** | `pkg/models` | 数据结构定义 |
| **LLM 层** | `pkg/llm` | LLM 客户端抽象、解析器、Prompt 构建 |
| **分析层** | `pkg/analyzer` | 内容评分系统 |
| **优化层** | `pkg/optimizer` | 优化引擎、策略实现 |
| **配置层** | `pkg/config` | AI 平台预设配置 |

---

## API 参考

### LLM 客户端

LLM 客户端提供与大语言模型交互的统一接口。

#### 创建客户端

```go
import "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"

// Config LLM 客户端配置
type Config struct {
    Provider    Provider // 提供商类型
    APIKey      string   // API 密钥
    BaseURL     string   // 自定义 API 端点（可选）
    Model       string   // 模型名称
    MaxTokens   int      // 最大 token 数
    Temperature float64  // 温度参数 (0.0-1.0)
    Timeout     int      // 请求超时时间（秒）
}

// Provider LLM 提供商类型
type Provider string

const (
    ProviderGLM Provider = "glm" // 智谱 GLM
)

// NewClient 创建 LLM 客户端
func NewClient(config Config) (LLMClient, error)
```

**示例**:

```go
// 使用 GLM
client, err := llm.NewClient(llm.Config{
    Provider:    llm.ProviderGLM,
    APIKey:      "your-glm-api-key",
    Model:       "glm-4.7",
    MaxTokens:   8000,
    Temperature: 0.7,
    Timeout:     300,
})
```

#### 通用对话接口

```go
// LLMClient LLM 客户端接口
type LLMClient interface {
    // Chat 通用对话接口
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}

// ChatRequest 聊天请求
type ChatRequest struct {
    Messages    []Message // 消息列表
    Temperature float64   // 温度参数（可选）
    MaxTokens   int       // 最大 token 数（可选）
    Stream      bool      // 是否流式输出（可选）
}

// Message 聊天消息
type Message struct {
    Role    string // system, user, assistant, tool
    Content string // 消息内容
}

// ChatResponse 聊天响应
type ChatResponse struct {
    Content      string // 响应内容
    TokensUsed   int    // 使用的 token 数
    Model        string // 模型名称
    FinishReason string // 结束原因
}
```

**示例**:

```go
resp, err := client.Chat(ctx, &llm.ChatRequest{
    Messages: []llm.Message{
        {Role: "system", Content: "你是一个专业的内容优化专家"},
        {Role: "user", Content: "请优化这段内容..."},
    },
    Temperature: 0.7,
    MaxTokens:   8000,
})

fmt.Println(resp.Content)
```

---

### 优化器

优化器是 GEO 优化的核心入口，协调评分、策略执行和 LLM 调用。

#### 创建优化器

```go
import "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"

// New 创建优化器
func New(client llm.LLMClient) *Optimizer
```

**示例**:

```go
opt := optimizer.New(client)
```

#### 执行优化

```go
// Optimize 执行完整优化流程
func (o *Optimizer) Optimize(ctx context.Context, req *models.OptimizationRequest) (*models.OptimizationResponse, error)

// OptimizeWithStrategy 使用指定策略执行优化
func (o *Optimizer) OptimizeWithStrategy(ctx context.Context, req *models.OptimizationRequest, strategyType models.StrategyType) (*models.OptimizationResponse, error)
```

**示例**:

```go
// 完整优化（使用多个策略）
resp, err := opt.Optimize(ctx, req)

// 单策略优化
resp, err := opt.OptimizeWithStrategy(ctx, req, models.StrategyStructure)
```

#### 注册自定义策略

```go
// RegisterStrategy 注册策略
func (o *Optimizer) RegisterStrategy(strategy strategies.Strategy)

// RegisterDefaultStrategies 注册默认策略（自动调用）
func (o *Optimizer) RegisterDefaultStrategies()
```

**示例**:

```go
// 注册自定义策略
opt.RegisterStrategy(myCustomStrategy)
```

---

### 评分器

评分器对内容进行 GEO 质量评分，支持规则基础的快速评分。

#### 创建评分器

```go
import "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/analyzer"

// Scorer 内容评分器
type Scorer struct {
    llmClient llm.LLMClient
}

// NewScorer 创建评分器
func NewScorer(client llm.LLMClient) *Scorer
```

**示例**:

```go
scorer := analyzer.NewScorer(client)
```

#### 评分接口

```go
// Score 对内容进行 GEO 评分（基于规则的快速评分）
func (s *Scorer) Score(ctx context.Context, content string) (*models.GeoScore, error)

// Compare 对比优化前后的评分
func (s *Scorer) Compare(ctx context.Context, before, after string) (*ScoreComparison, error)
```

**示例**:

```go
// 单次评分
score, err := scorer.Score(ctx, content)
fmt.Printf("结构化: %.2f, 权威性: %.2f, 总分: %.2f\n",
    score.Structure, score.Authority, score.OverallScore())

// 对比评分
comparison, err := scorer.Compare(ctx, originalContent, optimizedContent)
fmt.Printf("总分提升: %.2f → %.2f (%.2f%%)\n",
    comparison.Before.OverallScore(),
    comparison.After.OverallScore(),
    comparison.TotalChange)
```

#### 评分维度

| 维度 | 范围 | 说明 |
|------|------|------|
| Structure | 0-100 | 标题层级、段落结构、列表使用 |
| Authority | 0-100 | 数据引用、来源标注、专业术语 |
| Clarity | 0-100 | 句子长度、段落长度、逻辑连接词 |
| Citation | 0-100 | 结论性陈述、事实陈述、可操作性建议 |
| Schema | 0-100 | JSON-LD 标记、Schema.org 类型、必需字段 |

---

### 解析器

解析器从 LLM 响应中提取结构化信息。

#### 创建解析器

```go
import "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"

// Parser LLM 响应解析器
type Parser struct{}

// NewParser 创建解析器
func NewParser() *Parser
```

**示例**:

```go
parser := llm.NewParser()
```

#### 解析接口

```go
// ParseOptimizationResponse 解析优化响应
func (p *Parser) ParseOptimizationResponse(content string) (*ParseResult, error)

// ParseSchemaResponse 解析 Schema 响应
func (p *Parser) ParseSchemaResponse(content string) (map[string]interface{}, error)
```

#### 工具函数

```go
// ExtractMarkdownCodeBlock 提取 Markdown 代码块
func ExtractMarkdownCodeBlock(content string, language string) string

// CleanMarkdown 清理 Markdown 格式
func CleanMarkdown(content string) string

// ValidateJSON 验证 JSON 格式
func ValidateJSON(jsonStr string) bool

// ExtractKeyPoints 提取关键点
func ExtractKeyPoints(content string) []string
```

**示例**:

```go
parser := llm.NewParser()

// 解析优化响应
result, err := parser.ParseOptimizationResponse(llmResponse)
fmt.Println("优化内容:", result.OptimizedContent)
fmt.Println("FAQ 章节:", result.Sections["faq"])
fmt.Println("Schema 标记:", result.Sections["schema"])

// 提取 JSON 代码块
jsonBlock := llm.ExtractMarkdownCodeBlock(content, "json")

// 清理 Markdown 格式
cleanContent := llm.CleanMarkdown(content)

// 提取关键点
keyPoints := llm.ExtractKeyPoints(content)
```

---

### 数据模型

#### OptimizationRequest - 优化请求

```go
// OptimizationRequest GEO 优化请求
type OptimizationRequest struct {
    // 基础内容
    Content string // 原始内容（必填）
    Title   string // 内容标题（必填）
    URL     string // 发布 URL（可选）

    // 企业与产品信息
    Enterprise  EnterpriseInfo   // 企业信息（必填）
    Competitors []CompetitorInfo // 竞品信息（可选）

    // 优化目标
    TargetAI   []string       // 目标 AI 平台，如 ["chatgpt", "perplexity"]
    Keywords   []string       // 关键词列表
    Strategies []StrategyType // 优化策略列表

    // AI 偏好配置
    AIPreferences map[string]AIPreference // AI 平台偏好配置

    // 优化配置
    Tone     string // 内容语气
    Industry string // 行业类型

    // 内容增强选项
    IncludeProductMention bool   // 是否包含产品提及
    MentionFrequency      string // 提及频率: "low", "medium", "high"
    CallToAction          string // 行动号召

    // 元数据
    Author      string            // 作者
    PublishDate time.Time         // 发布日期
    CustomData  map[string]string // 自定义数据
}
```

**完整示例**:

```go
req := &models.OptimizationRequest{
    Content: "原始内容...",
    Title:   "内容标题",
    URL:     "https://example.com/article",

    Enterprise: models.EnterpriseInfo{
        CompanyName:        "示例公司",
        CompanyWebsite:     "https://example.com",
        CompanyDescription: "公司描述",
        ProductName:        "产品名称",
        ProductURL:         "https://example.com/product",
        ProductDescription: "产品描述",
        ProductFeatures:    []string{"功能1", "功能2"},
        USP:                []string{"独特卖点1", "独特卖点2"},
        BrandVoice:         "专业",
        TargetAudience:     "企业用户",
        Certifications:     []string{"ISO9001"},
        Awards:             []string{"2023年度最佳产品"},
    },

    Competitors: []models.CompetitorInfo{
        {
            Name:             "竞品A",
            Website:          "https://competitor-a.com",
            Weaknesses:       []string{"价格高", "配置复杂"},
            CommonObjections: []string{"服务响应慢"},
        },
    },

    TargetAI:   []string{"chatgpt", "perplexity", "google_ai"},
    Keywords:   []string{"关键词1", "关键词2"},
    Strategies: []models.StrategyType{
        models.StrategyStructure,
        models.StrategySchema,
        models.StrategyFAQ,
    },

    Tone:     "professional",
    Industry: "technology",

    IncludeProductMention: true,
    MentionFrequency:      "medium",
    CallToAction:          "立即联系我们",
}
```

#### OptimizationResponse - 优化响应

```go
// OptimizationResponse GEO 优化响应
type OptimizationResponse struct {
    // 优化后的内容
    OptimizedContent string // 优化后的完整内容
    Title            string // 标题

    // 生成的组件
    SchemaMarkup string // Schema 标记（JSON-LD 格式）
    FAQSection   string // FAQ 章节内容
    Summary      string // 内容摘要

    // 产品植入分析
    ProductMentions []ProductMention // 产品提及详情列表
    MentionCount    int              // 产品提及次数

    // 竞品差异化
    DifferentiationPoints []string // 差异化要点列表

    // 优化分析
    AppliedStrategies []StrategyType // 应用的策略列表
    Optimizations     []Optimization // 优化详情列表
    ScoreBefore       float64        // 优化前评分
    ScoreAfter        float64        // 优化后评分

    // 建议
    Recommendations []string // 改进建议列表

    // 元数据
    GeneratedAt time.Time // 生成时间
    Version     string    // 版本号
    LLMModel    string    // 使用的 LLM 模型
    TokensUsed  int       // 消耗的 token 数
}
```

**使用示例**:

```go
resp, err := opt.Optimize(ctx, req)

// 访问优化内容
fmt.Println(resp.OptimizedContent)

// 访问 Schema 标记
fmt.Println(resp.SchemaMarkup)

// 访问 FAQ 章节
fmt.Println(resp.FAQSection)

// 访问评分提升
improvement := resp.ScoreAfter - resp.ScoreBefore
fmt.Printf("评分提升: %.2f\n", improvement)

// 访问产品提及
for _, mention := range resp.ProductMentions {
    fmt.Printf("位置: %d, 影响级别: %s, 上下文: %s\n",
        mention.Position, mention.ImpactLevel, mention.Context)
}
```

#### EnterpriseInfo - 企业信息

```go
// EnterpriseInfo 企业信息
type EnterpriseInfo struct {
    // 企业基本信息
    CompanyName        string // 公司名称（必填）
    CompanyWebsite     string // 公司网站
    CompanyDescription string // 公司描述

    // 产品信息
    ProductName        string   // 产品名称
    ProductURL         string   // 产品链接
    ProductDescription string   // 产品描述
    ProductFeatures    []string // 产品特点列表
    USP                []string // 独特卖点列表

    // 品牌信息
    BrandVoice       string // 品牌语气: professional, casual, friendly
    TargetAudience   string // 目标受众
    ValueProposition string // 价值主张

    // 认证与权威性
    Certifications []string // 认证列表
    Awards         []string // 获奖列表
    CaseStudies    []string // 案例研究列表
}
```

#### CompetitorInfo - 竞品信息

```go
// CompetitorInfo 竞品信息
type CompetitorInfo struct {
    Name             string   // 竞品名称
    Website          string   // 竞品网站
    Weaknesses       []string // 竞品劣势列表
    CommonObjections []string // 常见异议点列表
}
```

#### AIPreference - AI 偏好配置

```go
// AIPreference AI 偏好配置
type AIPreference struct {
    // 内容偏好
    ContentStyle   string // 内容风格: professional, casual, academic
    ResponseFormat string // 响应格式: structured, direct, comprehensive
    CitationStyle  string // 引用风格: inline, footnote, academic

    // 结构偏好
    SectionStructure  []string // 章节结构列表
    PreferredHeading string   // 首选标题格式: ##, ###

    // 关键词偏好
    KeywordDensity float64 // 关键词密度 (0.0-1.0)
    SynonymUsage   bool    // 是否使用同义词

    // 技术偏好
    CodeBlocks     bool   // 是否包含代码块
    TechnicalDepth string // 技术深度: beginner, intermediate, advanced

    // 自定义指令
    CustomInstructions string // 自定义指令文本
}
```

#### GeoScore - GEO 评分

```go
// GeoScore GEO 评分明细
type GeoScore struct {
    Structure float64 // 结构化程度 (0-100)
    Authority float64 // 权威性 (0-100)
    Clarity   float64 // 清晰度 (0-100)
    Citation  float64 // 可引用性 (0-100)
    Schema    float64 // Schema 完整性 (0-100)
}

// OverallScore 计算总体评分（五个维度的平均值）
func (g *GeoScore) OverallScore() float64
```

#### StrategyType - 策略类型

```go
// StrategyType 优化策略类型
type StrategyType string

const (
    StrategyStructure   StrategyType = "structure"    // 结构化优化
    StrategySchema      StrategyType = "schema"       // Schema 标记生成
    StrategyAnswerFirst StrategyType = "answer_first" // 答案优先架构
    StrategyAuthority   StrategyType = "authority"    // 权威性增强
    StrategyFAQ         StrategyType = "faq"          // FAQ 生成
)
```

---

## 优化策略

优化器支持 5 种内置策略，可以单独使用或组合使用。

### 1. Structure - 结构化优化

优化内容的结构，添加标题层级、列表、章节划分。

**特性**:
- 自动添加 H1/H2/H3 标题
- 使用列表组织要点
- 控制段落长度（100-300字）
- 清理多余空行

**适用场景**: 所有类型的内容

**示例**:

```go
resp, err := opt.OptimizeWithStrategy(ctx, req, models.StrategyStructure)
```

### 2. Schema - Schema 标记生成

生成 JSON-LD 格式的 Schema.org 结构化数据标记。

**支持的 Schema 类型**:
- `Article` - 文章（默认）
- `Product` - 产品
- `HowTo` - 教程/指南
- `FAQPage` - FAQ 页面
- `Organization` - 组织

**示例**:

```go
// 使用默认类型（Article）
resp, err := opt.OptimizeWithStrategy(ctx, req, models.StrategySchema)

// 指定 Schema 类型
schemaStrategy := strategies.NewSchemaStrategyWithType("Product")
opt.RegisterStrategy(schemaStrategy)
```

### 3. AnswerFirst - 答案优先

将关键结论移到内容开头，采用"答案优先"架构。

**特性**:
- 开头直接给出核心结论
- 然后展开详细解释
- 使用简练语言

**适用场景**: 问答型、教程型内容

**示例**:

```go
resp, err := opt.OptimizeWithStrategy(ctx, req, models.StrategyAnswerFirst)
```

### 4. Authority - 权威性增强

增强内容的权威性，添加数据支撑和专业元素。

**特性**:
- 添加数据支撑（使用 `[数据]` 标注）
- 引用来源（使用 `[来源]` 标注）
- 添加专业术语
- 增加案例或证据

**适用场景**: 专业性要求高的内容

**示例**:

```go
resp, err := opt.OptimizeWithStrategy(ctx, req, models.StrategyAuthority)
```

### 5. FAQ - FAQ 生成

生成常见问题（FAQ）章节。

**特性**:
- 生成 3-5 个常见问题
- 每个问题给出简洁准确的答案
- 问题覆盖内容的核心要点

**适用场景**: 需要补充 FAQ 的内容

**示例**:

```go
resp, err := opt.OptimizeWithStrategy(ctx, req, models.StrategyFAQ)
```

### 组合使用策略

```go
// 组合多个策略
req.Strategies = []models.StrategyType{
    models.StrategyStructure,
    models.StrategySchema,
    models.StrategyFAQ,
    models.StrategyAnswerFirst,
    models.StrategyAuthority,
}

resp, err := opt.Optimize(ctx, req)
```

### 自定义策略

```go
import strategiespkg "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer/strategies"

// 实现策略接口
type CustomStrategy struct {
    *strategiespkg.BaseStrategy
}

func (s *CustomStrategy) BuildPrompt(req *models.OptimizationRequest) string {
    return fmt.Sprintf("自定义优化指令: %s", req.Content)
}

func (s *CustomStrategy) Validate(req *models.OptimizationRequest) bool {
    return req.Content != ""
}

// 注册策略
customStrategy := &CustomStrategy{
    BaseStrategy: strategiespkg.NewBaseStrategy("custom", "自定义策略"),
}
opt.RegisterStrategy(customStrategy)
```

---

## AI 平台预设

系统内置了主流 AI 平台的偏好配置，可自动应用。

### 内置平台

#### ChatGPT

```go
"chatgpt": models.AIPreference{
    ContentStyle:      "professional",
    ResponseFormat:    "structured",
    CitationStyle:     "inline",
    PreferredHeading:  "##",
    KeywordDensity:    0.02,
    TechnicalDepth:    "intermediate",
    CustomInstructions: "请使用专业的语言风格，确保内容结构清晰...",
}
```

**适用场景**: 专业性强、需要结构化的内容

#### Perplexity

```go
"perplexity": models.AIPreference{
    ContentStyle:      "concise",
    ResponseFormat:    "direct",
    CitationStyle:     "source_links",
    PreferredHeading:  "###",
    KeywordDensity:    0.015,
    TechnicalDepth:    "beginner",
    CustomInstructions: "请直接回答问题，使用简洁的语言...",
}
```

**适用场景**: 快速问答、简洁回复

#### Google AI

```go
"google_ai": models.AIPreference{
    ContentStyle:      "structured",
    ResponseFormat:    "comprehensive",
    CitationStyle:     "academic",
    PreferredHeading:  "##",
    KeywordDensity:    0.025,
    TechnicalDepth:    "advanced",
    CustomInstructions: "请提供全面详细的内容，使用学术风格的引用...",
}
```

**适用场景**: 技术性强、需要深度分析的内容

#### Claude

```go
"claude": models.AIPreference{
    ContentStyle:      "natural",
    ResponseFormat:    "conversational",
    CitationStyle:     "contextual",
    PreferredHeading:  "##",
    KeywordDensity:    0.018,
    TechnicalDepth:    "intermediate",
    CustomInstructions: "请使用自然对话式的语言，注重内容的可读性...",
}
```

**适用场景**: 对话式、可读性要求高的内容

### 使用预设配置

```go
import "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/config"

// 方式 1: 通过 TargetAI 自动应用
req.TargetAI = []string{"chatgpt", "perplexity"}
// 优化器会自动应用对应平台的预设配置

// 方式 2: 手动应用预设配置
chatgptPref, _ := config.GetAIProfile("chatgpt")
req.AIPreferences = map[string]models.AIPreference{
    "chatgpt": chatgptPref,
}

// 方式 3: 自定义配置并覆盖预设
req.AIPreferences = map[string]models.AIPreference{
    "chatgpt": {
        ContentStyle: "casual", // 覆盖默认的 professional
        TechnicalDepth: "advanced",
    },
}
// 优化器会合并用户配置与默认配置
```

### 获取可用平台

```go
platforms := config.GetAvailablePlatforms()
// ["chatgpt", "perplexity", "google_ai", "claude"]
```

---

## 完整示例

### 示例 1: 基础优化

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"
)

func main() {
    // 创建 LLM 客户端
    client, err := llm.NewClient(llm.Config{
        Provider:    llm.ProviderGLM,
        APIKey:      "your-api-key",
        Model:       "glm-4.7",
        Temperature: 0.7,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 创建优化器
    opt := optimizer.New(client)

    // 构建请求
    req := &models.OptimizationRequest{
        Content: "如何选择合适的云服务提供商？在选择云服务时，需要考虑多个因素...",
        Title:   "云服务选择指南",
        Enterprise: models.EnterpriseInfo{
            CompanyName: "CloudTech",
            ProductName: "CloudTech Pro",
            USP:         []string{"行业最低价格", "5分钟快速部署"},
        },
        Strategies: []models.StrategyType{
            models.StrategyStructure,
            models.StrategySchema,
        },
    }

    // 执行优化
    resp, err := opt.Optimize(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }

    // 输出结果
    fmt.Printf("优化前评分: %.2f\n", resp.ScoreBefore)
    fmt.Printf("优化后评分: %.2f\n", resp.ScoreAfter)
    fmt.Printf("提升幅度: %.2f%%\n", (resp.ScoreAfter-resp.ScoreBefore)/resp.ScoreBefore*100)
    fmt.Printf("\n优化内容:\n%s\n", resp.OptimizedContent)
    fmt.Printf("\nSchema 标记:\n%s\n", resp.SchemaMarkup)
}
```

### 示例 2: 多平台优化

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"
)

func main() {
    client, _ := llm.NewClient(llm.Config{
        Provider: llm.ProviderGLM,
        APIKey:   "your-api-key",
    })
    opt := optimizer.New(client)

    baseReq := &models.OptimizationRequest{
        Content: "人工智能技术在医疗领域的应用...",
        Title:   "AI 医疗应用指南",
        Enterprise: models.EnterpriseInfo{
            CompanyName: "MedAI",
            ProductName: "MedAI Pro",
        },
        Keywords: []string{"人工智能", "医疗", "AI诊断"},
        Strategies: []models.StrategyType{
            models.StrategyStructure,
            models.StrategyFAQ,
        },
    }

    // 为不同平台优化
    platforms := []string{"chatgpt", "perplexity", "google_ai"}

    for _, platform := range platforms {
        // 复制请求并设置目标平台
        req := *baseReq
        req.TargetAI = []string{platform}

        // 执行优化
        resp, err := opt.Optimize(context.Background(), &req)
        if err != nil {
            log.Printf("优化失败 (%s): %v", platform, err)
            continue
        }

        fmt.Printf("\n=== %s 优化结果 ===\n", platform)
        fmt.Printf("评分: %.2f → %.2f\n", resp.ScoreBefore, resp.ScoreAfter)
        fmt.Printf("内容:\n%s\n", resp.OptimizedContent)
    }
}
```

### 示例 3: 内容评分对比

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/analyzer"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
)

func main() {
    client, _ := llm.NewClient(llm.Config{
        Provider: llm.ProviderGLM,
        APIKey:   "your-api-key",
    })

    scorer := analyzer.NewScorer(client)

    originalContent := `原始内容...`
    optimizedContent := `优化后内容...`

    // 对比评分
    comparison, err := scorer.Compare(context.Background(), originalContent, optimizedContent)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(comparison.GetImprovementSummary())
    // 输出:
    // 总分变化: 45.20 → 78.50 (33.30)
    // structure: +25.00
    // authority: +18.50
    // clarity: +12.30
    // citation: +20.00
    // schema: +50.00
}
```

### 示例 4: 解析器使用

```go
package main

import (
    "fmt"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
)

func main() {
    parser := llm.NewParser()

    llmResponse := `
## 云服务选择指南

优化后的内容...

### 常见问题

Q: 如何选择云服务？
A: 需要考虑性能、价格...

\`\`\`json
{
  "@context": "https://schema.org",
  "@type": "Article",
  "headline": "云服务选择指南"
}
\`\`\`
`

    // 解析响应
    result, err := parser.ParseOptimizationResponse(llmResponse)
    if err != nil {
        fmt.Printf("解析失败: %v\n", err)
        return
    }

    fmt.Println("优化内容:", result.OptimizedContent)
    fmt.Println("FAQ 章节:", result.Sections["faq"])
    fmt.Println("Schema 标记:", result.Sections["schema"])

    // 提取关键点
    keyPoints := llm.ExtractKeyPoints(llmResponse)
    fmt.Println("关键点:", keyPoints)
}
```

### 示例 5: 自定义策略

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
    "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"
    strategiespkg "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer/strategies"
)

// SEOKwStrategy 关键词优化策略
type SEOKwStrategy struct {
    *strategiespkg.BaseStrategy
    keywords []string
}

func NewSEOKwStrategy(keywords []string) *SEOKwStrategy {
    return &SEOKwStrategy{
        BaseStrategy: strategiespkg.NewBaseStrategy("seo_keywords", "关键词优化"),
        keywords:     keywords,
    }
}

func (s *SEOKwStrategy) Validate(req *models.OptimizationRequest) bool {
    return len(s.keywords) > 0
}

func (s *SEOKwStrategy) BuildPrompt(req *models.OptimizationRequest) string {
    prompt := fmt.Sprintf("请优化以下内容，自然地融入关键词: %v\n\n", s.keywords)
    prompt += fmt.Sprintf("原始内容:\n%s\n\n", req.Content)
    prompt += "要求:\n"
    prompt += "- 关键词密度控制在 2-3%\n"
    prompt += "- 在标题、开头、结尾自然出现关键词\n"
    prompt += "- 不要堆砌关键词\n"
    return prompt
}

func main() {
    client, _ := llm.NewClient(llm.Config{
        Provider: llm.ProviderGLM,
        APIKey:   "your-api-key",
    })

    opt := optimizer.New(client)

    // 注册自定义策略
    keywordStrategy := NewSEOKwStrategy([]string{"云服务", "云计算", "云服务器"})
    opt.RegisterStrategy(keywordStrategy)

    req := &models.OptimizationRequest{
        Content: "选择云服务提供商...",
        Title:   "云服务指南",
        Enterprise: models.EnterpriseInfo{
            CompanyName: "CloudTech",
        },
    }

    resp, err := opt.OptimizeWithStrategy(context.Background(), req, "seo_keywords")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.OptimizedContent)
}
```

---

## 开发命令

### 构建

```bash
go build ./...
```

### 运行测试

```bash
# 运行所有测试
go test ./... -v

# 运行特定包的测试
go test ./pkg/optimizer -v

# 运行单个测试
go test ./pkg/models -run TestOptimizationRequest -v
```

### 代码格式化

```bash
go fmt ./...
```

### 依赖管理

```bash
go mod tidy
```

---

## 常见问题

### Q: 如何切换 LLM 提供商？

A: 当前版本支持 GLM 提供商。如需支持其他提供商，可以实现 `LLMClient` 接口：

```go
type CustomClient struct {
    config llm.Config
}

func (c *CustomClient) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
    // 实现自定义的 Chat 逻辑
    return &llm.ChatResponse{}, nil
}
```

### Q: 评分系统是基于 LLM 还是规则？

A: 当前版本使用**基于规则的快速评分**，无需调用 LLM，速度快且成本低。如需深度分析，可以自行调用 LLM 的 `AnalyzeContent` 方法（已弃用但可用）。

### Q: 如何处理 LLM 返回的不规范内容？

A: 使用 `llm.Parser` 的工具函数进行清理：

```go
// 清理 Markdown 格式
cleanContent := llm.CleanMarkdown(llmResponse)

// 提取 JSON 代码块
jsonStr := llm.ExtractMarkdownCodeBlock(llmResponse, "json")

// 验证 JSON 格式
valid := llm.ValidateJSON(jsonStr)
```

### Q: 如何控制优化成本？

A: 可以通过以下方式控制成本：

1. 使用 `Score()` 方法预先评分，跳过已优化的内容
2. 单独使用策略而非组合策略
3. 调整 `MaxTokens` 和 `Temperature` 参数
4. 使用规则评分而非 LLM 分析

### Q: 是否支持流式输出？

A: 当前版本不支持流式输出。`ChatRequest` 的 `Stream` 字段已预留但未实现。

---

## 版本历史

### v1.0.0

- 初始版本
- 支持 GLM 提供商
- 5 种优化策略
- 基于规则的内容评分
- AI 平台预设配置
- Prompt 构建器
- LLM 响应解析器

---

## 许可证

[请根据实际项目添加许可证信息]

---

## 联系方式

如有问题或建议，请提交 Issue 或 Pull Request。
