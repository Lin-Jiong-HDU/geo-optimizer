# Token Tracking Fix Design

**Date:** 2026-03-24
**Author:** Claude
**Status:** Approved

## Problem

Several methods that call the LLM do not properly track or return token usage:

| Method | Issue |
|--------|-------|
| `OptimizeWithStrategy()` | Discards `chatResp`, loses token info |
| `ScoreWithAI()` | Returns `ScoreResult` without `TokensUsed` |
| `CompareWithAI()` | Makes 2 LLM calls, no token tracking |

## Solution

Add `TokensUsed` field to return structures and fix token accumulation in methods.

## Changes

### 1. ScoreResult (`pkg/models/scorer.go`)

Add `TokensUsed` field:

```go
type ScoreResult struct {
    *GeoScore
    ScoreType    string  // "rules" or "ai"
    Degraded     bool    // whether degraded to rule scoring
    ErrorMessage string  // degradation reason
    TokensUsed   int     // NEW: tokens consumed (0 for rule-based scoring)
}
```

### 2. ScoreComparisonResult (`pkg/analyzer/scorer_ai.go`)

Add `TokensUsed` field:

```go
type ScoreComparisonResult struct {
    Before       *ScoreResult
    After        *ScoreResult
    Improvements map[string]float64
    TotalChange  float64
    TokensUsed   int  // NEW: total tokens from both ScoreWithAI calls
}
```

### 3. ScoreWithAI (`pkg/analyzer/scorer_ai.go`)

Update to return token info:

```go
func (s *Scorer) ScoreWithAI(ctx context.Context, content string) (*models.ScoreResult, error) {
    // ... existing code ...
    resp, err := s.llmClient.Chat(ctx, req)
    // ...
    return &models.ScoreResult{
        GeoScore:  score,
        ScoreType: "ai",
        Degraded:  false,
        TokensUsed: resp.TokensUsed,  // NEW
    }, nil
}
```

### 4. CompareWithAI (`pkg/analyzer/scorer_ai.go`)

Update to track and return total tokens:

```go
func (s *Scorer) CompareWithAI(ctx context.Context, before, after string) (*ScoreComparisonResult, error) {
    scoreBefore, err := s.ScoreWithAI(ctx, before)
    // ...
    scoreAfter, err := s.ScoreWithAI(ctx, after)
    // ...
    return &ScoreComparisonResult{
        Before:       scoreBefore,
        After:        scoreAfter,
        Improvements: improvements,
        TotalChange:  totalChange,
        TokensUsed:   scoreBefore.TokensUsed + scoreAfter.TokensUsed,  // NEW
    }, nil
}
```

### 5. OptimizeWithStrategy (`pkg/optimizer/optimizer.go`)

Fix to track tokens:

```go
func (o *Optimizer) OptimizeWithStrategy(...) (*models.OptimizationResponse, error) {
    // ... existing code ...

    // Before: optimizedContent, _, err := o.executeStrategy(...)
    // After:
    optimizedContent, chatResp, err := o.executeStrategy(ctx, reqCopy, strategy)
    // ...

    response := &models.OptimizationResponse{
        // ... existing fields ...
        TokensUsed: chatResp.TokensUsed,  // NEW
        LLMModel:   chatResp.Model,       // NEW
    }
    return response, nil
}
```

## Backward Compatibility

All changes are **fully backward compatible**:

- Adding fields to structs does not break existing code in Go
- The business code (`bookish-train`) only uses `Optimize()` which already works correctly
- No function signatures are changed

## Testing

- Update existing tests to verify `TokensUsed` is populated
- Add tests for edge cases (degraded scoring, errors)
