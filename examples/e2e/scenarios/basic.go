package scenarios

import (
	"context"
	"fmt"
	"time"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"
)

// BasicScenario tests basic functionality of the geo-optimizer.
type BasicScenario struct{}

// NewBasicScenario creates a new BasicScenario.
func NewBasicScenario() *BasicScenario {
	return &BasicScenario{}
}

// Name returns the scenario name.
func (s *BasicScenario) Name() string {
	return "Basic Functionality"
}

// Description returns the scenario description.
func (s *BasicScenario) Description() string {
	return "Validates core APIs: Chat, Score, and OptimizeWithStrategy"
}

// Run executes the basic functionality tests.
func (s *BasicScenario) Run(ctx context.Context, client llm.LLMClient, opt *optimizer.Optimizer) (*ScenarioResult, error) {
	result := &ScenarioResult{
		ScenarioName: s.Name(),
		Description:  s.Description(),
		Results:      make([]Result, 0),
		AllPassed:    true,
	}

	startTime := time.Now()

	// Test 1: Chat
	chatResult := s.testChat(ctx, client)
	result.Results = append(result.Results, chatResult)
	if !chatResult.Success {
		result.AllPassed = false
	}

	// Test 2: Score
	scoreResult := s.testScore(ctx, opt)
	result.Results = append(result.Results, scoreResult)
	if !scoreResult.Success {
		result.AllPassed = false
	}

	// Test 3: OptimizeWithStrategy
	optResult := s.testOptimizeWithStrategy(ctx, opt)
	result.Results = append(result.Results, optResult)
	if !optResult.Success {
		result.AllPassed = false
	}

	result.TotalDuration = time.Since(startTime)
	return result, nil
}

func (s *BasicScenario) testChat(ctx context.Context, client llm.LLMClient) Result {
	result := NewResult("TestChat")
	start := time.Now()

	input := "什么是AI？"
	resp, err := client.Chat(ctx, &llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "user", Content: input},
		},
	})

	if err != nil {
		result.SetFailure(err.Error(), time.Since(start))
		return *result
	}

	result.Input = input
	result.SetSuccess(resp.Content, time.Since(start))
	return *result
}

func (s *BasicScenario) testScore(ctx context.Context, opt *optimizer.Optimizer) Result {
	result := NewResult("TestScore")
	start := time.Now()

	content := "人工智能正在改变世界。它可以帮助我们解决很多问题。"
	score, err := opt.Scorer().Score(ctx, content)

	if err != nil {
		result.SetFailure(err.Error(), time.Since(start))
		return *result
	}

	result.Input = content
	result.ScoreBefore = score.OverallScore()
	result.SetSuccess(fmt.Sprintf("Score: %.1f", score.OverallScore()), time.Since(start))
	return *result
}

func (s *BasicScenario) testOptimizeWithStrategy(ctx context.Context, opt *optimizer.Optimizer) Result {
	result := NewResult("TestOptimizeWithStrategy")
	start := time.Now()

	req := &models.OptimizationRequest{
		Title:   "AI技术简介",
		Content: "AI是人工智能的缩写，是计算机科学的一个分支。",
		Enterprise: models.EnterpriseInfo{
			CompanyName:        "测试公司",
			ProductName:        "测试产品",
			ProductDescription: "这是一个测试产品",
		},
		Strategies: []models.StrategyType{models.StrategyStructure},
	}

	resp, err := opt.OptimizeWithStrategy(ctx, req, models.StrategyStructure)
	if err != nil {
		result.SetFailure(err.Error(), time.Since(start))
		return *result
	}

	result.Input = req.Content
	result.Output = resp.OptimizedContent
	result.ScoreBefore = resp.ScoreBefore
	result.ScoreAfter = resp.ScoreAfter
	result.SetSuccess(fmt.Sprintf("Tokens used: %d", resp.TokensUsed), time.Since(start))
	return *result
}
