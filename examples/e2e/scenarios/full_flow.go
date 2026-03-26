package scenarios

import (
	"context"
	"fmt"
	"time"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer"
)

// FullFlowScenario tests the complete optimization workflow.
type FullFlowScenario struct{}

// NewFullFlowScenario creates a new FullFlowScenario.
func NewFullFlowScenario() *FullFlowScenario {
	return &FullFlowScenario{}
}

// Name returns the scenario name.
func (s *FullFlowScenario) Name() string {
	return "Full Business Flow"
}

// Description returns the scenario description.
func (s *FullFlowScenario) Description() string {
	return "Tests complete Optimize() pipeline with realistic enterprise content"
}

// Run executes the full business flow test.
func (s *FullFlowScenario) Run(ctx context.Context, client llm.LLMClient, opt *optimizer.Optimizer) (*ScenarioResult, error) {
	result := &ScenarioResult{
		ScenarioName: s.Name(),
		Description:  s.Description(),
		Results:      make([]Result, 0),
		AllPassed:    true,
	}

	startTime := time.Now()

	// Run the full optimization test
	testResult := s.testFullOptimize(ctx, opt)
	result.Results = append(result.Results, testResult)
	if !testResult.Success {
		result.AllPassed = false
	}

	result.TotalDuration = time.Since(startTime)
	return result, nil
}

func (s *FullFlowScenario) testFullOptimize(ctx context.Context, opt *optimizer.Optimizer) Result {
	result := NewResult("TestOptimize")
	start := time.Now()

	// Use realistic enterprise content
	req := &models.OptimizationRequest{
		Title: "企业数字化转型解决方案",
		Content: `企业数字化转型是当今商业环境中的重要议题。

随着技术的快速发展，企业需要适应新的市场环境。数字化转型可以帮助企业提高效率、降低成本、增强竞争力。

我们的解决方案包括：
- 云计算基础设施
- 数据分析平台
- 人工智能应用

这些技术可以帮助企业实现业务目标。`,
		Enterprise: models.EnterpriseInfo{
			CompanyName:        "智云科技",
			CompanyWebsite:     "https://example.com",
			CompanyDescription: "专注于企业数字化转型解决方案",
			ProductName:        "智云企业平台",
			ProductURL:         "https://example.com/platform",
			ProductDescription: "一站式企业数字化解决方案",
			ProductFeatures: []string{
				"云端部署",
				"数据可视化",
				"AI智能分析",
				"多端协同",
			},
			USP: []string{
				"行业领先的AI技术",
				"99.9%服务可用性",
				"7x24小时技术支持",
			},
			BrandVoice:       "专业、可靠、创新",
			TargetAudience:   "中大型企业决策者",
			ValueProposition: "帮助企业快速实现数字化转型",
		},
		Competitors: []models.CompetitorInfo{
			{
				Name:       "竞争对手A",
				Website:    "https://competitor-a.com",
				Weaknesses: []string{"功能单一", "缺乏AI支持"},
			},
		},
		TargetAI:   []string{"chatgpt", "perplexity"},
		Keywords:   []string{"数字化转型", "企业解决方案", "云计算", "AI"},
		Strategies: []models.StrategyType{models.StrategyStructure, models.StrategySchema, models.StrategyFAQ},
		Tone:       "professional",
		Industry:   "technology",
	}

	resp, err := opt.Optimize(ctx, req)
	if err != nil {
		result.SetFailure(err.Error(), time.Since(start))
		return *result
	}

	// Validate response fields
	var validationErrors []string
	if resp.OptimizedContent == "" {
		validationErrors = append(validationErrors, "OptimizedContent is empty")
	}
	if resp.SchemaMarkup == "" {
		validationErrors = append(validationErrors, "SchemaMarkup is empty")
	}
	if resp.FAQSection == "" {
		validationErrors = append(validationErrors, "FAQSection is empty")
	}
	if resp.TokensUsed == 0 {
		validationErrors = append(validationErrors, "TokensUsed is 0")
	}

	if len(validationErrors) > 0 {
		result.SetFailure(fmt.Sprintf("Validation failed: %v", validationErrors), time.Since(start))
		return *result
	}

	result.Input = req.Content
	result.Output = fmt.Sprintf("Content length: %d, Schema: %t, FAQ: %t",
		len(resp.OptimizedContent), resp.SchemaMarkup != "", resp.FAQSection != "")
	result.ScoreBefore = resp.ScoreBefore
	result.ScoreAfter = resp.ScoreAfter
	result.SetSuccess(fmt.Sprintf("All fields validated, Tokens: %d", resp.TokensUsed), time.Since(start))
	return *result
}
