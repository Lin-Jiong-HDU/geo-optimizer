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
func (s *Scorer) ScoreWithAI(ctx context.Context, content string) (*models.ScoreResult, error) {
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
		// 降级到规则评分（LLM调用失败，无token消耗）
		return s.degradeToRuleScore(content, fmt.Sprintf("LLM调用失败: %v", err), 0), nil
	}

	// 解析响应
	score, err := s.parseAIResponse(resp.Content)
	if err != nil {
		// 降级到规则评分（LLM调用成功但解析失败，保留token计数）
		return s.degradeToRuleScore(content, fmt.Sprintf("解析LLM响应失败: %v", err), resp.TokensUsed), nil
	}

	return &models.ScoreResult{
		GeoScore:   score,
		ScoreType:  "ai",
		Degraded:   false,
		TokensUsed: resp.TokensUsed,
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
		TokensUsed:   scoreBefore.TokensUsed + scoreAfter.TokensUsed,
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
func (s *Scorer) degradeToRuleScore(content string, errMsg string, tokensUsed int) *models.ScoreResult {
	score := s.scoreByRules(content)
	return &models.ScoreResult{
		GeoScore:     score,
		ScoreType:    "rules",
		Degraded:     true,
		ErrorMessage: errMsg,
		TokensUsed:   tokensUsed,
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
	Before       *models.ScoreResult `json:"before"`
	After        *models.ScoreResult `json:"after"`
	Improvements map[string]float64  `json:"improvements"`
	TotalChange  float64             `json:"total_change"`
	TokensUsed   int                 `json:"tokens_used"` // 两次评分总token数
}
