# AI评分功能设计文档

## 概述

在现有规则评分基础上，新增AI评分功能，通过LLM对内容进行更准确的语义评分。采用新增方法的方式，确保不影响现有业务。

## 需求决策

| 项目 | 决策 |
|------|------|
| 定位 | 新增 `ScoreWithAI` 方法，不影响现有接口 |
| 评分维度 | 复用现有5个维度（Structure/Authority/Clarity/Citation/Schema） |
| 返回结构 | `ScoreResult` 包含 `GeoScore` + 元信息 |
| 降级策略 | 自动降级到规则评分，标记 `Degraded=true` |
| Prompt | 返回JSON格式，包含分数和简短理由 |

## 数据模型

### ScoreResult

```go
// ScoreResult 评分结果（支持AI评分和规则评分）
type ScoreResult struct {
    *models.GeoScore                          // 复用现有5维度评分

    // 元信息
    ScoreType    string `json:"score_type"`    // "ai" 或 "rules"
    Degraded     bool   `json:"degraded"`      // 是否从AI降级到规则
    ErrorMessage string `json:"error_message"` // 降级时的错误信息（可选）
}
```

### ScoreComparisonResult

```go
type ScoreComparisonResult struct {
    Before       *ScoreResult       `json:"before"`
    After        *ScoreResult       `json:"after"`
    Improvements map[string]float64 `json:"improvements"`
    TotalChange  float64            `json:"total_change"`
}
```

## 接口设计

```go
// ScoreWithAI 使用LLM进行AI评分
func (s *Scorer) ScoreWithAI(ctx context.Context, content string) (*ScoreResult, error)

// CompareWithAI 使用AI评分对比优化前后内容
func (s *Scorer) CompareWithAI(ctx context.Context, before, after string) (*ScoreComparisonResult, error)
```

## AI评分Prompt

```
你是一个GEO（生成引擎优化）评分专家。请对以下内容进行评分。

内容:
%s

请从以下5个维度进行评分（0-100分），并给出简短理由：

1. **Structure（结构化）**: 标题层级、列表使用、段落组织是否清晰
2. **Authority（权威性）**: 数据支撑、来源引用、专业术语使用
3. **Clarity（清晰度）**: 句子长度、逻辑连接、可读性
4. **Citation（可引用性）**: 结论性陈述、事实陈述、可操作建议
5. **Schema（结构化数据）**: JSON-LD/Schema.org标记的完整性

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
}
```

## 实现流程

```
ScoreWithAI(content)
        │
        ▼
┌─────────────────────┐
│ 1. 构建评分Prompt    │
└─────────┬───────────┘
          ▼
┌─────────────────────┐
│ 2. 调用LLM API       │
└─────────┬───────────┘
          ▼
   ┌──────┴──────┐
   │ LLM成功?     │
   └──────┬──────┘
     Yes  │   No
          │    ▼
          │  ┌────────────────┐
          │  │ 降级到规则评分  │
          │  │ Degraded=true  │
          │  └───────┬────────┘
          │          │
          ▼          ▼
┌─────────────────────────┐
│ 3. 构建ScoreResult返回   │
└─────────────────────────┘
```

## 错误处理

以下场景触发降级到规则评分：
- LLM调用超时
- API返回非200状态码
- 响应JSON解析失败

降级时：
- `ScoreType` 设为 "rules"
- `Degraded` 设为 true
- `ErrorMessage` 记录具体错误

## 文件结构

```
pkg/analyzer/
├── scorer.go              # 现有，新增 ScoreWithAI 方法
├── scorer_ai.go           # 新增，AI评分核心逻辑
├── scorer_test.go         # 现有，新增AI评分测试

pkg/llm/prompts/
├── score_prompt.go        # 新增，评分prompt模板

pkg/models/
├── content.go             # 现有，新增 ScoreResult 结构体
```

## 代码风格要求

- 遵循现有错误处理模式：`fmt.Errorf("...: %w", err)`
- 命名风格与现有代码一致
- 测试覆盖新增方法
- 使用 `go fmt ./...` 格式化代码
