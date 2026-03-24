# AI评分功能实现计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有规则评分基础上，新增 `ScoreWithAI` 方法，通过LLM进行更准确的语义评分。

**Architecture:** 新增 `ScoreResult` 结构体包含评分结果和元信息，新增 `ScoreWithAI` 和 `CompareWithAI` 方法，失败时自动降级到规则评分。

**Tech Stack:** Go 1.25, 现有 llm.LLMClient 接口

---

## 文件结构

```
pkg/models/
├── content.go             # 修改：新增 ScoreResult 结构体

pkg/llm/prompts/
├── score_prompt.go        # 新增：AI评分prompt模板

pkg/analyzer/
├── scorer.go              # 修改：新增 ScoreWithAI 方法
├── scorer_ai.go           # 新增：AI评分核心逻辑（解析LLM响应）
├── scorer_test.go         # 修改：新增AI评分测试
```

---

## Chunk 1: 数据模型

### Task 1: ScoreResult 数据模型

**Files:**
- Modify: `pkg/models/content.go:36-47`

- [ ] **Step 1: 写失败测试**

Run: `go test ./pkg/models -run TestScoreResult -v`
Expected: FAIL with "undefined: ScoreResult"

- [ ] **Step 2: 添加 ScoreResult 结构体**

在 `pkg/models/content.go` 的 `GeoScore` 结构体后添加：

```go
// ScoreResult 评分结果（支持AI评分和规则评分）
type ScoreResult struct {
	*GeoScore // 复用现有5维度评分

	// 元信息
	ScoreType    string `json:"score_type"`    // "ai" 或 "rules"
	Degraded     bool   `json:"degraded"`      // 是否从AI降级到规则
	ErrorMessage string `json:"error_message"` // 降级时的错误信息（可选）
}
```

- [ ] **Step 3: 运行测试验证**

Run: `go test ./pkg/models -v`
Expected: PASS

- [ ] **Step 4: 提交**

```bash
git add pkg/models/content.go
git commit -m "feat(models): add ScoreResult struct for AI scoring"
```

---

## Chunk 2: Prompt模板

### Task 2: AI评分Prompt模板

**Files:**
- Create: `pkg/llm/prompts/score_prompt.go`

- [ ] **Step 1: 创建评分prompt文件**

创建 `pkg/llm/prompts/score_prompt.go`：

```go
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

// aiScoreResponse AI评分响应结构（内部使用）
type aiScoreResponse struct {
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
```

- [ ] **Step 2: 运行测试验证语法**

Run: `go build ./pkg/llm/prompts`
Expected: 无错误

- [ ] **Step 3: 提交**

```bash
git add pkg/llm/prompts/score_prompt.go
git commit -m "feat(prompts): add AI scoring prompt template"
```

---

## Chunk 3: AI评分核心逻辑

### Task 3: AI评分解析器

**Files:**
- Create: `pkg/analyzer/scorer_ai.go`

- [ ] **Step 1: 创建AI评分逻辑文件**

创建 `pkg/analyzer/scorer_ai.go`：

```go
package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm/prompts"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// ScoreWithAI 使用LLM进行AI评分
func (s *Scorer) ScoreWithAI(ctx context.Context, content string) (*ScoreResult, error) {
	// 构建评分请求
	req := &llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: prompts.SystemPromptScoring,
			},
			{
				Role:    "user",
				Content: prompts.BuildScorePrompt(content),
			},
		},
		Temperature: 0.3, // 评分需要较低的温度以保证一致性
	}

	// 调用LLM
	resp, err := s.llmClient.Chat(ctx, req)
	if err != nil {
		// 降级到规则评分
		return s.degradeToRuleScore(content, fmt.Sprintf("LLM调用失败: %v", err)), nil
	}

	// 解析响应
	score, err := s.parseAIResponse(resp.Content)
	if err != nil {
		// 降级到规则评分
		return s.degradeToRuleScore(content, fmt.Sprintf("解析LLM响应失败: %v", err)), nil
	}

	return &ScoreResult{
		GeoScore:  score,
		ScoreType: "ai",
		Degraded:  false,
	}, nil
}

// CompareWithAI 使用AI评分对比优化前后内容
func (s *Scorer) CompareWithAI(ctx context.Context, before, after string) (*ScoreComparisonResult, error) {
	// 评分前
	scoreBefore, err := s.ScoreWithAI(ctx, before)
	if err != nil {
		return nil, fmt.Errorf("failed to score before content: %w", err)
	}

	// 评分后
	scoreAfter, err := s.ScoreWithAI(ctx, after)
	if err != nil {
		return nil, fmt.Errorf("failed to score after content: %w", err)
	}

	// 计算提升幅度
	improvements := map[string]float64{
		"structure": scoreAfter.Structure - scoreBefore.Structure,
		"authority": scoreAfter.Authority - scoreBefore.Authority,
		"clarity":   scoreAfter.Clarity - scoreBefore.Clarity,
		"citation":  scoreAfter.Citation - scoreBefore.Citation,
		"schema":    scoreAfter.Schema - scoreBefore.Schema,
	}

	totalChange := scoreAfter.OverallScore() - scoreBefore.OverallScore()

	return &ScoreComparisonResult{
		Before:       scoreBefore,
		After:        scoreAfter,
		Improvements: improvements,
		TotalChange:  totalChange,
	}, nil
}

// parseAIResponse 解析AI评分响应
func (s *Scorer) parseAIResponse(content string) (*models.GeoScore, error) {
	// 清理可能的markdown代码块标记
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var resp prompts.AiScoreResponse
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// 验证分数范围
	score := &models.GeoScore{
		Structure: clampScore(resp.Structure),
		Authority: clampScore(resp.Authority),
		Clarity:   clampScore(resp.Clarity),
		Citation:  clampScore(resp.Citation),
		Schema:    clampScore(resp.Schema),
	}

	return score, nil
}

// degradeToRuleScore 降级到规则评分
func (s *Scorer) degradeToRuleScore(content string, errMsg string) *ScoreResult {
	score := s.scoreByRules(content)
	return &ScoreResult{
		GeoScore:     score,
		ScoreType:    "rules",
		Degraded:     true,
		ErrorMessage: errMsg,
	}
}

// clampScore 确保分数在0-100范围内
func clampScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// ScoreComparisonResult AI评分对比结果
type ScoreComparisonResult struct {
	Before       *ScoreResult       `json:"before"`
	After        *ScoreResult       `json:"after"`
	Improvements map[string]float64 `json:"improvements"`
	TotalChange  float64            `json:"total_change"`
}
```

- [ ] **Step 2: 修复prompts中的类型名**

修改 `pkg/llm/prompts/score_prompt.go` 中的结构体名称为大写开头：

```go
// AiScoreResponse AI评分响应结构（内部使用）
type AiScoreResponse struct {
```

- [ ] **Step 3: 运行构建验证**

Run: `go build ./pkg/analyzer`
Expected: 无错误

- [ ] **Step 4: 提交**

```bash
git add pkg/analyzer/scorer_ai.go
git commit -m "feat(analyzer): add ScoreWithAI and CompareWithAI methods"
```

---

## Chunk 4: 测试

### Task 4: AI评分测试

**Files:**
- Modify: `pkg/analyzer/scorer_test.go`

- [ ] **Step 1: 添加mock LLM客户端（返回评分JSON）**

在 `pkg/analyzer/scorer_test.go` 中添加新的mock客户端：

```go
// mockLLMClientForScoring 模拟返回评分JSON的LLM客户端
type mockLLMClientForScoring struct {
	response string
	err      error
}

func (m *mockLLMClientForScoring) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &llm.ChatResponse{
		Content:      m.response,
		TokensUsed:   100,
		Model:        "test-model",
		FinishReason: "stop",
	}, nil
}
```

- [ ] **Step 2: 添加TestScorer_ScoreWithAI测试**

```go
func TestScorer_ScoreWithAI(t *testing.T) {
	// 模拟LLM返回的评分JSON
	mockResp := `{
		"structure": 85,
		"structure_reason": "内容结构清晰",
		"authority": 70,
		"authority_reason": "有数据支撑",
		"clarity": 90,
		"clarity_reason": "表达简洁",
		"citation": 75,
		"citation_reason": "有可引用结论",
		"schema": 60,
		"schema_reason": "缺少结构化标记"
	}`

	scorer := NewScorer(&mockLLMClientForScoring{response: mockResp})
	content := "# 测试内容\n\n## 章节\n\n测试内容。"

	result, err := scorer.ScoreWithAI(context.Background(), content)
	if err != nil {
		t.Fatalf("ScoreWithAI should not return error: %v", err)
	}

	if result == nil {
		t.Fatal("ScoreWithAI should return non-nil result")
	}

	// 验证评分类型
	if result.ScoreType != "ai" {
		t.Errorf("Expected ScoreType 'ai', got: %s", result.ScoreType)
	}

	// 验证未降级
	if result.Degraded {
		t.Error("Expected Degraded to be false")
	}

	// 验证分数范围
	if result.Structure < 0 || result.Structure > 100 {
		t.Errorf("Structure should be 0-100, got: %.2f", result.Structure)
	}

	t.Logf("ScoreType: %s, Degraded: %v", result.ScoreType, result.Degraded)
	t.Logf("Structure: %.2f, Authority: %.2f, Clarity: %.2f, Citation: %.2f, Schema: %.2f",
		result.Structure, result.Authority, result.Clarity, result.Citation, result.Schema)
}
```

- [ ] **Step 3: 添加降级测试**

```go
func TestScorer_ScoreWithAI_Degraded(t *testing.T) {
	// 模拟LLM调用失败
	scorer := NewScorer(&mockLLMClientForScoring{err: fmt.Errorf("timeout")})
	content := "# 测试内容\n\n测试内容。"

	result, err := scorer.ScoreWithAI(context.Background(), content)
	if err != nil {
		t.Fatalf("ScoreWithAI should not return error even on LLM failure: %v", err)
	}

	// 验证已降级
	if !result.Degraded {
		t.Error("Expected Degraded to be true")
	}

	if result.ScoreType != "rules" {
		t.Errorf("Expected ScoreType 'rules', got: %s", result.ScoreType)
	}

	if result.ErrorMessage == "" {
		t.Error("Expected ErrorMessage to be set")
	}

	t.Logf("Degraded: %v, ScoreType: %s, ErrorMessage: %s",
		result.Degraded, result.ScoreType, result.ErrorMessage)
}
```

- [ ] **Step 4: 添加CompareWithAI测试**

```go
func TestScorer_CompareWithAI(t *testing.T) {
	mockResp := `{
		"structure": 80,
		"authority": 70,
		"clarity": 85,
		"citation": 75,
		"schema": 60
	}`

	scorer := NewScorer(&mockLLMClientForScoring{response: mockResp})

	before := "简单内容"
	after := "# 优化后内容\n\n## 章节\n\n优化后的内容。"

	comparison, err := scorer.CompareWithAI(context.Background(), before, after)
	if err != nil {
		t.Fatalf("CompareWithAI should not return error: %v", err)
	}

	if comparison.Before == nil {
		t.Fatal("Before should not be nil")
	}
	if comparison.After == nil {
		t.Fatal("After should not be nil")
	}

	// 验证Improvements
	if comparison.Improvements == nil {
		t.Fatal("Improvements should not be nil")
	}

	t.Logf("Before: %.2f, After: %.2f, Change: %.2f",
		comparison.Before.OverallScore(),
		comparison.After.OverallScore(),
		comparison.TotalChange)
}
```

- [ ] **Step 5: 运行所有测试**

Run: `go test ./pkg/analyzer -v`
Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add pkg/analyzer/scorer_test.go
git commit -m "test(analyzer): add tests for ScoreWithAI and CompareWithAI"
```

---

## Chunk 5: 最终验证

### Task 5: 完整测试和格式化

- [ ] **Step 1: 运行完整测试套件**

Run: `go test ./... -v`
Expected: PASS

- [ ] **Step 2: 格式化代码**

Run: `go fmt ./...`
Expected: 无输出

- [ ] **Step 3: 运行go mod tidy**

Run: `go mod tidy`
Expected: 无错误

- [ ] **Step 4: 最终构建验证**

Run: `go build ./...`
Expected: 无错误

- [ ] **Step 5: 查看变更状态**

Run: `git status`
Expected: working tree clean

---

## 实现完成后

实现完成后，调用方可以按以下方式使用：

```go
// 使用规则评分（现有方式，不变）
score, err := scorer.Score(ctx, content)

// 使用AI评分（新方式）
result, err := scorer.ScoreWithAI(ctx, content)
if result.Degraded {
    log.Printf("AI评分降级: %s", result.ErrorMessage)
}

// 对比（AI评分）
comparison, err := scorer.CompareWithAI(ctx, before, after)
```
