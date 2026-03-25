package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm/prompts"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// priorityOrder 优先级排序权重
var priorityOrder = map[string]int{"high": 3, "medium": 2, "low": 1}

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
		// 降级到规则评分
		return s.degradeToRuleScore(content, fmt.Sprintf("LLM调用失败: %v", err)), nil
	}

	// 解析响应
	score, err := s.parseAIResponse(resp.Content)
	if err != nil {
		// 降级到规则评分
		return s.degradeToRuleScore(content, fmt.Sprintf("解析LLM响应失败: %v", err)), nil
	}

	return &models.ScoreResult{
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
func (s *Scorer) degradeToRuleScore(content string, errMsg string) *models.ScoreResult {
	score := s.scoreByRules(content)
	return &models.ScoreResult{
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
	Before       *models.ScoreResult `json:"before"`
	After        *models.ScoreResult `json:"after"`
	Improvements map[string]float64  `json:"improvements"`
	TotalChange  float64             `json:"total_change"`
}

// ScoreWithSuggestions 评分并返回改进建议（一次LLM调用）
func (s *Scorer) ScoreWithSuggestions(ctx context.Context, content string) (*models.ScoreResultWithSuggestions, error) {
	// 构建评分请求
	req := &llm.ChatRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: prompts.SystemPromptScoringWithSuggestions,
			},
			{
				Role:    "user",
				Content: prompts.BuildScoreWithSuggestionsPrompt(content),
			},
		},
		Temperature: 0.3,
	}

	// 调用LLM
	resp, err := s.llmClient.Chat(ctx, req)
	if err != nil {
		// 降级：只返回规则评分，无建议
		return s.degradeToRuleScoreWithSuggestions(content, fmt.Sprintf("LLM调用失败: %v", err)), nil
	}

	// 解析响应
	result, err := s.parseAIResponseWithSuggestions(resp.Content)
	if err != nil {
		// 降级：只返回规则评分，无建议
		return s.degradeToRuleScoreWithSuggestions(content, fmt.Sprintf("解析LLM响应失败: %v", err)), nil
	}

	return result, nil
}

// parseAIResponseWithSuggestions 解析带建议的AI评分响应
func (s *Scorer) parseAIResponseWithSuggestions(content string) (*models.ScoreResultWithSuggestions, error) {
	// 从LLM响应中提取有效的 JSON 子串，避免因为说明文字/多段代码块导致解析失败
	jsonContent, err := extractJSONFromText(content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract JSON from LLM response: %w", err)
	}

	var resp prompts.AiScoreWithSuggestionsResponse
	if err := json.Unmarshal([]byte(jsonContent), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// 构建GeoScore
	geoScore := &models.GeoScore{
		Structure: clampScore(resp.Scores.Structure),
		Authority: clampScore(resp.Scores.Authority),
		Clarity:   clampScore(resp.Scores.Clarity),
		Citation:  clampScore(resp.Scores.Citation),
		Schema:    clampScore(resp.Scores.Schema),
	}

	// 转换建议
	dimensionSuggestions := make(map[string][]models.Suggestion)
	for dim, suggestions := range resp.DimensionSuggestions {
		for _, sugg := range suggestions {
			dimensionSuggestions[dim] = append(dimensionSuggestions[dim], models.Suggestion{
				Issue:         sugg.Issue,
				Direction:     sugg.Direction,
				Priority:      sugg.Priority,
				EstimatedGain: sugg.EstimatedGain,
				Example:       sugg.Example,
			})
		}
	}

	// 转换并排序 TopSuggestions
	topSuggestions := make([]models.Suggestion, 0, len(resp.TopSuggestions))
	for _, sugg := range resp.TopSuggestions {
		topSuggestions = append(topSuggestions, models.Suggestion{
			Issue:         sugg.Issue,
			Direction:     sugg.Direction,
			Priority:      sugg.Priority,
			EstimatedGain: sugg.EstimatedGain,
			Example:       sugg.Example,
		})
	}

	// 按 priority + estimated_gain 排序
	sort.Slice(topSuggestions, func(i, j int) bool {
		pi := priorityOrder[strings.ToLower(topSuggestions[i].Priority)]
		pj := priorityOrder[strings.ToLower(topSuggestions[j].Priority)]
		if pi != pj {
			return pi > pj // high > medium > low
		}
		return topSuggestions[i].EstimatedGain > topSuggestions[j].EstimatedGain
	})

	// 截断到最多5条
	if len(topSuggestions) > 5 {
		topSuggestions = topSuggestions[:5]
	}

	return &models.ScoreResultWithSuggestions{
		ScoreResult: &models.ScoreResult{
			GeoScore:  geoScore,
			ScoreType: "ai",
			Degraded:  false,
		},
		DimensionSuggestions: dimensionSuggestions,
		TopSuggestions:       topSuggestions,
	}, nil
}

// degradeToRuleScoreWithSuggestions 降级到规则评分（无建议）
func (s *Scorer) degradeToRuleScoreWithSuggestions(content string, errMsg string) *models.ScoreResultWithSuggestions {
	score := s.scoreByRules(content)
	return &models.ScoreResultWithSuggestions{
		ScoreResult: &models.ScoreResult{
			GeoScore:     score,
			ScoreType:    "rules",
			Degraded:     true,
			ErrorMessage: errMsg,
		},
		DimensionSuggestions: make(map[string][]models.Suggestion),
		TopSuggestions:       []models.Suggestion{},
	}
}

// extractJSONFromText 从包含说明文字/markdown代码块的文本中提取第一个有效的JSON子串
func extractJSONFromText(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", fmt.Errorf("empty LLM response")
	}

	// 查找第一个可能的JSON起始符号（对象）
	start := strings.Index(content, "{")
	if start == -1 {
		return "", fmt.Errorf("no JSON object start found in LLM response")
	}

	// 使用括号匹配找到完整的JSON对象
	depth := 0
	end := -1
	for i := start; i < len(content); i++ {
		if content[i] == '{' {
			depth++
		} else if content[i] == '}' {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}

	if end == -1 {
		return "", fmt.Errorf("unbalanced JSON braces in LLM response")
	}

	return content[start:end], nil
}
