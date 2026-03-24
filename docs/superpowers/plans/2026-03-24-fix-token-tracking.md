# Token Tracking Fix Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix token tracking in `OptimizeWithStrategy`, `ScoreWithAI`, and `CompareWithAI` methods.

**Architecture:** Add `TokensUsed` field to return structures and fix token accumulation in methods. All changes are backward compatible (adding fields to structs).

**Tech Stack:** Go 1.25

---

## File Structure

| File | Action | Purpose |
|------|--------|---------|
| `pkg/models/content.go` | Modify | Add `TokensUsed` to `ScoreResult` |
| `pkg/analyzer/scorer_ai.go` | Modify | Return token info in `ScoreWithAI`, `CompareWithAI` |
| `pkg/analyzer/scorer_test.go` | Modify | Add token verification tests |
| `pkg/optimizer/optimizer.go` | Modify | Fix `OptimizeWithStrategy` token tracking |
| `pkg/optimizer/optimizer_test.go` | Modify | Add `OptimizeWithStrategy` token test |

---

## Chunk 1: ScoreResult Token Tracking

### Task 1: Add TokensUsed to ScoreResult

**Files:**
- Modify: `pkg/models/content.go:49-57`
- Test: `pkg/analyzer/scorer_test.go`

- [ ] **Step 1: Write the failing test**

Add test to verify `TokensUsed` is populated in `ScoreWithAI` result:

```go
// In pkg/analyzer/scorer_test.go, add after TestScorer_ScoreWithAI function

func TestScorer_ScoreWithAI_TokensUsed(t *testing.T) {
	mockResp := `{
		"structure": 85,
		"authority": 70,
		"clarity": 90,
		"citation": 75,
		"schema": 60
	}`

	scorer := NewScorer(&mockLLMClientForScoring{response: mockResp})
	content := "# 测试内容\n\n测试内容。"

	result, err := scorer.ScoreWithAI(context.Background(), content)
	if err != nil {
		t.Fatalf("ScoreWithAI should not return error: %v", err)
	}

	// 验证 TokensUsed 被正确返回
	if result.TokensUsed != 100 {
		t.Errorf("Expected TokensUsed 100, got: %d", result.TokensUsed)
	}

	t.Logf("TokensUsed: %d", result.TokensUsed)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/analyzer -run TestScorer_ScoreWithAI_TokensUsed -v`
Expected: FAIL with "result.TokensUsed undefined" or similar

- [ ] **Step 3: Add TokensUsed field to ScoreResult**

Modify `pkg/models/content.go`:

```go
// ScoreResult 评分结果（支持AI评分和规则评分）
type ScoreResult struct {
	*GeoScore // 复用现有5维度评分

	// 元信息
	ScoreType    string `json:"score_type"`    // "ai" 或 "rules"
	Degraded     bool   `json:"degraded"`      // 是否从AI降级到规则
	ErrorMessage string `json:"error_message"` // 降级时的错误信息（可选）
	TokensUsed   int    `json:"tokens_used"`   // AI评分消耗的token数（规则评分为0）
}
```

- [ ] **Step 4: Update ScoreWithAI to populate TokensUsed**

Modify `pkg/analyzer/scorer_ai.go:45-49`:

```go
// Before:
return &models.ScoreResult{
	GeoScore:  score,
	ScoreType: "ai",
	Degraded:  false,
}, nil

// After:
return &models.ScoreResult{
	GeoScore:   score,
	ScoreType:  "ai",
	Degraded:   false,
	TokensUsed: resp.TokensUsed,
}, nil
```

- [ ] **Step 5: Update degradeToRuleScore to set TokensUsed to 0**

Modify `pkg/analyzer/scorer_ai.go:112-119`:

```go
// Before:
return &models.ScoreResult{
	GeoScore:     score,
	ScoreType:    "rules",
	Degraded:     true,
	ErrorMessage: errMsg,
}

// After:
return &models.ScoreResult{
	GeoScore:     score,
	ScoreType:    "rules",
	Degraded:     true,
	ErrorMessage: errMsg,
	TokensUsed:   0,
}
```

- [ ] **Step 6: Run test to verify it passes**

Run: `go test ./pkg/analyzer -run TestScorer_ScoreWithAI -v`
Expected: All 3 tests PASS (ScoreWithAI, ScoreWithAI_Degraded, ScoreWithAI_TokensUsed)

- [ ] **Step 7: Commit**

```bash
git add pkg/models/content.go pkg/analyzer/scorer_ai.go pkg/analyzer/scorer_test.go
git commit -m "fix(analyzer): add TokensUsed to ScoreResult in ScoreWithAI"
```

---

## Chunk 2: CompareWithAI Token Tracking

### Task 2: Add TokensUsed to ScoreComparisonResult

**Files:**
- Modify: `pkg/analyzer/scorer_ai.go:133-139`
- Test: `pkg/analyzer/scorer_test.go`

- [ ] **Step 1: Write the failing test**

Add test to verify `TokensUsed` in `CompareWithAI`:

```go
// In pkg/analyzer/scorer_test.go, add after TestScorer_CompareWithAI function

func TestScorer_CompareWithAI_TokensUsed(t *testing.T) {
	mockResp := `{
		"structure": 80,
		"authority": 70,
		"clarity": 85,
		"citation": 75,
		"schema": 60
	}`

	scorer := NewScorer(&mockLLMClientForScoring{response: mockResp})

	before := "简单内容"
	after := "# 优化后内容\n\n优化后的内容。"

	comparison, err := scorer.CompareWithAI(context.Background(), before, after)
	if err != nil {
		t.Fatalf("CompareWithAI should not return error: %v", err)
	}

	// 验证 TokensUsed 是两次调用的总和 (100 + 100 = 200)
	if comparison.TokensUsed != 200 {
		t.Errorf("Expected TokensUsed 200, got: %d", comparison.TokensUsed)
	}

	// 验证 Before 和 After 的 TokensUsed 也被正确设置
	if comparison.Before.TokensUsed != 100 {
		t.Errorf("Expected Before.TokensUsed 100, got: %d", comparison.Before.TokensUsed)
	}
	if comparison.After.TokensUsed != 100 {
		t.Errorf("Expected After.TokensUsed 100, got: %d", comparison.After.TokensUsed)
	}

	t.Logf("Total TokensUsed: %d", comparison.TokensUsed)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/analyzer -run TestScorer_CompareWithAI_TokensUsed -v`
Expected: FAIL with "comparison.TokensUsed undefined"

- [ ] **Step 3: Add TokensUsed to ScoreComparisonResult**

Modify `pkg/analyzer/scorer_ai.go:133-139`:

```go
// Before:
type ScoreComparisonResult struct {
	Before       *models.ScoreResult `json:"before"`
	After        *models.ScoreResult `json:"after"`
	Improvements map[string]float64  `json:"improvements"`
	TotalChange  float64             `json:"total_change"`
}

// After:
type ScoreComparisonResult struct {
	Before       *models.ScoreResult `json:"before"`
	After        *models.ScoreResult `json:"after"`
	Improvements map[string]float64  `json:"improvements"`
	TotalChange  float64             `json:"total_change"`
	TokensUsed   int                 `json:"tokens_used"` // 两次评分总token数
}
```

- [ ] **Step 4: Update CompareWithAI to populate TokensUsed**

Modify `pkg/analyzer/scorer_ai.go:77-82`:

```go
// Before:
return &ScoreComparisonResult{
	Before:       scoreBefore,
	After:        scoreAfter,
	Improvements: improvements,
	TotalChange:  totalChange,
}, nil

// After:
return &ScoreComparisonResult{
	Before:       scoreBefore,
	After:        scoreAfter,
	Improvements: improvements,
	TotalChange:  totalChange,
	TokensUsed:   scoreBefore.TokensUsed + scoreAfter.TokensUsed,
}, nil
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./pkg/analyzer -run TestScorer_CompareWithAI -v`
Expected: All 2 tests PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/analyzer/scorer_ai.go pkg/analyzer/scorer_test.go
git commit -m "fix(analyzer): add TokensUsed to ScoreComparisonResult in CompareWithAI"
```

---

## Chunk 3: OptimizeWithStrategy Token Tracking

### Task 3: Fix OptimizeWithStrategy token tracking

**Files:**
- Modify: `pkg/optimizer/optimizer.go:90-141`
- Test: `pkg/optimizer/optimizer_test.go`

- [ ] **Step 1: Write the failing test**

Add test for `OptimizeWithStrategy` token tracking in `pkg/optimizer/optimizer_test.go`:

```go
// In pkg/optimizer/optimizer_test.go, add after existing tests

func TestOptimizer_OptimizeWithStrategy_TokensUsed(t *testing.T) {
	mockClient := &mockLLMClient{}
	opt := New(mockClient)

	req := &models.OptimizationRequest{
		Content: "测试内容",
		Title:   "测试标题",
		Strategies: []models.StrategyType{
			models.StrategyStructure,
		},
	}

	resp, err := opt.OptimizeWithStrategy(context.Background(), req, models.StrategyStructure)
	if err != nil {
		t.Fatalf("OptimizeWithStrategy should not return error: %v", err)
	}

	// 验证 TokensUsed 被正确返回
	if resp.TokensUsed != 100 {
		t.Errorf("Expected TokensUsed 100, got: %d", resp.TokensUsed)
	}

	// 验证 LLMModel 被正确返回
	if resp.LLMModel != "test-model" {
		t.Errorf("Expected LLMModel 'test-model', got: %s", resp.LLMModel)
	}

	t.Logf("TokensUsed: %d, LLMModel: %s", resp.TokensUsed, resp.LLMModel)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/optimizer -run TestOptimizer_OptimizeWithStrategy_TokensUsed -v`
Expected: FAIL with "resp.TokensUsed is 0" (currently not tracked)

- [ ] **Step 3: Fix OptimizeWithStrategy to track tokens**

Modify `pkg/optimizer/optimizer.go:119` and `131-139`:

```go
// Before (line 119):
optimizedContent, _, err := o.executeStrategy(ctx, reqCopy, strategy)

// After:
optimizedContent, chatResp, err := o.executeStrategy(ctx, reqCopy, strategy)

// Before (lines 131-139):
response := &models.OptimizationResponse{
	OptimizedContent:  optimizedContent,
	Title:             reqCopy.Title,
	AppliedStrategies: []models.StrategyType{strategyType},
	ScoreBefore:       scoreBefore.OverallScore(),
	ScoreAfter:        scoreAfter.OverallScore(),
	GeneratedAt:       time.Now(),
	Version:           "1.0.0",
}

// After:
response := &models.OptimizationResponse{
	OptimizedContent:  optimizedContent,
	Title:             reqCopy.Title,
	AppliedStrategies: []models.StrategyType{strategyType},
	ScoreBefore:       scoreBefore.OverallScore(),
	ScoreAfter:        scoreAfter.OverallScore(),
	GeneratedAt:       time.Now(),
	Version:           "1.0.0",
	LLMModel:          chatResp.Model,
	TokensUsed:        chatResp.TokensUsed,
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/optimizer -run TestOptimizer_OptimizeWithStrategy_TokensUsed -v`
Expected: PASS

- [ ] **Step 5: Run all optimizer tests to ensure no regression**

Run: `go test ./pkg/optimizer -v`
Expected: All tests PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/optimizer/optimizer.go pkg/optimizer/optimizer_test.go
git commit -m "fix(optimizer): track tokens in OptimizeWithStrategy"
```

---

## Chunk 4: Final Verification

### Task 4: Run all tests and verify backward compatibility

- [ ] **Step 1: Run all tests**

Run: `go test ./... -v`
Expected: All tests PASS

- [ ] **Step 2: Format code**

Run: `go fmt ./...`

- [ ] **Step 3: Final commit (if any formatting changes)**

```bash
git add -A
git commit -m "style: format code" || echo "No formatting changes needed"
```

- [ ] **Step 4: Verify backward compatibility**

The following APIs remain unchanged (only new fields added):
- `models.ScoreResult` - added `TokensUsed int`
- `analyzer.ScoreComparisonResult` - added `TokensUsed int`
- `models.OptimizationResponse` - already has `TokensUsed`, now populated in `OptimizeWithStrategy`

All existing code will continue to work without modification.
