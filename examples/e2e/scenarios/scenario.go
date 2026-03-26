package scenarios

import (
	"context"
	"time"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"
)

// Result represents the result of a single test within a scenario.
type Result struct {
	Name        string
	Success     bool
	Duration    time.Duration
	Input       string
	Output      string
	ScoreBefore float64
	ScoreAfter  float64
	Error       string
}

// ScenarioResult represents the result of an entire scenario.
type ScenarioResult struct {
	ScenarioName  string
	Description   string
	Results       []Result
	TotalDuration time.Duration
	AllPassed     bool
}

// Scenario defines the interface for test scenarios.
type Scenario interface {
	Name() string
	Description() string
	Run(ctx context.Context, client llm.LLMClient, opt *optimizer.Optimizer) (*ScenarioResult, error)
}

// NewResult creates a new Result with default values.
func NewResult(name string) *Result {
	return &Result{
		Name:    name,
		Success: false,
	}
}

// SetSuccess marks the result as successful with output.
func (r *Result) SetSuccess(output string, duration time.Duration) {
	r.Success = true
	r.Output = output
	r.Duration = duration
}

// SetFailure marks the result as failed with error message.
func (r *Result) SetFailure(errMsg string, duration time.Duration) {
	r.Success = false
	r.Error = errMsg
	r.Duration = duration
}
