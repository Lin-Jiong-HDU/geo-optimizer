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

// priorityOrder defines the priority weights for sorting suggestions.
var priorityOrder = map[string]int{"high": 3, "medium": 2, "low": 1}

// ScoreWithAI performs AI-based scoring using LLM.
func (s *Scorer) ScoreWithAI(ctx context.Context, content string) (*models.ScoreResult, error) {
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
		Temperature: 0.3,
	}

	resp, err := s.llmClient.Chat(ctx, req)
	if err != nil {
		return s.degradeToRuleScore(content, fmt.Sprintf("LLM call failed: %v", err), 0), nil
	}

	score, err := s.parseAIResponse(resp.Content)
	if err != nil {
		return s.degradeToRuleScore(content, fmt.Sprintf("Failed to parse LLM response: %v", err), resp.TokensUsed), nil
	}

	return &models.ScoreResult{
		GeoScore:   score,
		ScoreType:  "ai",
		Degraded:   false,
		TokensUsed: resp.TokensUsed,
	}, nil
}

// CompareWithAI compares scores before and after optimization using AI scoring.
func (s *Scorer) CompareWithAI(ctx context.Context, before, after string) (*ScoreComparisonResult, error) {
	scoreBefore, err := s.ScoreWithAI(ctx, before)
	if err != nil {
		return nil, fmt.Errorf("failed to score before content: %w", err)
	}

	scoreAfter, err := s.ScoreWithAI(ctx, after)
	if err != nil {
		return nil, fmt.Errorf("failed to score after content: %w", err)
	}

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

// parseAIResponse parses the AI scoring response.
func (s *Scorer) parseAIResponse(content string) (*models.GeoScore, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var resp prompts.AiScoreResponse
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	score := &models.GeoScore{
		Structure: clampScore(resp.Structure),
		Authority: clampScore(resp.Authority),
		Clarity:   clampScore(resp.Clarity),
		Citation:  clampScore(resp.Citation),
		Schema:    clampScore(resp.Schema),
	}

	return score, nil
}

// degradeToRuleScore falls back to rule-based scoring when AI scoring fails.
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

// clampScore ensures the score is within 0-100 range.
func clampScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// ScoreComparisonResult represents the result of AI score comparison.
type ScoreComparisonResult struct {
	Before       *models.ScoreResult `json:"before"`
	After        *models.ScoreResult `json:"after"`
	Improvements map[string]float64  `json:"improvements"`
	TotalChange  float64             `json:"total_change"`
	TokensUsed   int                 `json:"tokens_used"`
}

// ScoreWithSuggestions performs scoring and returns improvement suggestions in a single LLM call.
func (s *Scorer) ScoreWithSuggestions(ctx context.Context, content string) (*models.ScoreResultWithSuggestions, error) {
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

	resp, err := s.llmClient.Chat(ctx, req)
	if err != nil {
		return s.degradeToRuleScoreWithSuggestions(content, fmt.Sprintf("LLM call failed: %v", err)), nil
	}

	result, err := s.parseAIResponseWithSuggestions(resp.Content)
	if err != nil {
		return s.degradeToRuleScoreWithSuggestions(content, fmt.Sprintf("Failed to parse LLM response: %v", err)), nil
	}

	return result, nil
}

// parseAIResponseWithSuggestions parses the AI response with suggestions.
func (s *Scorer) parseAIResponseWithSuggestions(content string) (*models.ScoreResultWithSuggestions, error) {
	jsonContent, err := extractJSONFromText(content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract JSON from LLM response: %w", err)
	}

	var resp prompts.AiScoreWithSuggestionsResponse
	if err := json.Unmarshal([]byte(jsonContent), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	geoScore := &models.GeoScore{
		Structure: clampScore(resp.Scores.Structure),
		Authority: clampScore(resp.Scores.Authority),
		Clarity:   clampScore(resp.Scores.Clarity),
		Citation:  clampScore(resp.Scores.Citation),
		Schema:    clampScore(resp.Scores.Schema),
	}

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

	sort.Slice(topSuggestions, func(i, j int) bool {
		pi := priorityOrder[strings.ToLower(topSuggestions[i].Priority)]
		pj := priorityOrder[strings.ToLower(topSuggestions[j].Priority)]
		if pi != pj {
			return pi > pj
		}
		return topSuggestions[i].EstimatedGain > topSuggestions[j].EstimatedGain
	})

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

// degradeToRuleScoreWithSuggestions falls back to rule-based scoring without suggestions.
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

// extractJSONFromText extracts the first valid JSON substring from text containing markdown code blocks.
func extractJSONFromText(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", fmt.Errorf("empty LLM response")
	}

	start := strings.Index(content, "{")
	if start == -1 {
		return "", fmt.Errorf("no JSON object start found in LLM response")
	}

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
