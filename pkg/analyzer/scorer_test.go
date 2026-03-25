package analyzer

import (
	"context"
	"fmt"
	"testing"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// mockLLMClient 模拟LLM客户端
type mockLLMClient struct{}

func (m *mockLLMClient) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	return &llm.ChatResponse{
		Content:      "{}",
		TokensUsed:   100,
		Model:        "test-model",
		FinishReason: "stop",
	}, nil
}

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

func TestNewScorer(t *testing.T) {
	client := &mockLLMClient{}
	scorer := NewScorer(client)

	if scorer == nil {
		t.Fatal("NewScorer should return a non-nil scorer")
	}

	if scorer.llmClient == nil {
		t.Error("NewScorer should set the LLM client")
	}
}

func TestScorer_Score(t *testing.T) {
	scorer := NewScorer(&mockLLMClient{})

	content := `# 云服务选择指南

## 如何选择合适的云服务提供商

在选择云服务提供商时，需要考虑以下因素：

1. 价格：根据预算选择合适的服务
2. 性能：确保服务稳定可靠
3. 支持：良好的技术支持至关重要

### 研究显示

根据2024年云计算报告显示，85%的企业选择云服务时最看重性价比。

### 总结

因此，建议企业在选择云服务时，优先考虑价格和性能的平衡。`

	score, err := scorer.Score(context.Background(), content)
	if err != nil {
		t.Fatalf("Score should not return error: %v", err)
	}

	if score == nil {
		t.Fatal("Score should return a non-nil GeoScore")
	}

	// 验证各维度分数在合理范围内
	if score.Structure < 0 || score.Structure > 100 {
		t.Errorf("Structure score should be between 0-100, got: %.2f", score.Structure)
	}
	if score.Authority < 0 || score.Authority > 100 {
		t.Errorf("Authority score should be between 0-100, got: %.2f", score.Authority)
	}
	if score.Clarity < 0 || score.Clarity > 100 {
		t.Errorf("Clarity score should be between 0-100, got: %.2f", score.Clarity)
	}
	if score.Citation < 0 || score.Citation > 100 {
		t.Errorf("Citation score should be between 0-100, got: %.2f", score.Citation)
	}
	if score.Schema < 0 || score.Schema > 100 {
		t.Errorf("Schema score should be between 0-100, got: %.2f", score.Schema)
	}

	// 这个内容应该有中等偏上的分数
	overall := score.OverallScore()
	if overall < 25 {
		t.Errorf("Overall score should be at least 25 for this content, got: %.2f", overall)
	}

	t.Logf("Structure: %.2f, Authority: %.2f, Clarity: %.2f, Citation: %.2f, Schema: %.2f",
		score.Structure, score.Authority, score.Clarity, score.Citation, score.Schema)
	t.Logf("Overall: %.2f", overall)
}

func TestScorer_Score_EmptyContent(t *testing.T) {
	scorer := NewScorer(&mockLLMClient{})

	score, err := scorer.Score(context.Background(), "")
	if err != nil {
		t.Fatalf("Score should not return error for empty content: %v", err)
	}

	if score.OverallScore() != 0 {
		t.Errorf("Empty content should have score of 0, got: %.2f", score.OverallScore())
	}
}

func TestScorer_Score_WithSchema(t *testing.T) {
	scorer := NewScorer(&mockLLMClient{})

	content := `# 云服务选择指南

<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@type": "Article",
  "name": "云服务选择指南",
  "headline": "如何选择合适的云服务提供商",
  "description": "详细介绍如何选择云服务"
}
</script>

本文介绍如何选择云服务。`

	score, err := scorer.Score(context.Background(), content)
	if err != nil {
		t.Fatalf("Score should not return error: %v", err)
	}

	if score.Schema < 50 {
		t.Errorf("Content with Schema should have high Schema score, got: %.2f", score.Schema)
	}

	t.Logf("Schema score: %.2f", score.Schema)
}

func TestScorer_ScoreWithAnalysis(t *testing.T) {
	scorer := NewScorer(&mockLLMClient{})

	content := "# 测试内容\n\n## 章节一\n\n测试内容。"

	analysis, err := scorer.ScoreWithAnalysis(context.Background(), content)
	if err != nil {
		t.Fatalf("ScoreWithAnalysis should not return error: %v", err)
	}

	if analysis == nil {
		t.Fatal("ScoreWithAnalysis should return non-nil analysis")
	}

	// 验证返回的分析结果在合理范围内
	if analysis.StructureScore < 0 || analysis.StructureScore > 100 {
		t.Errorf("StructureScore should be between 0-100, got: %.2f", analysis.StructureScore)
	}
	if analysis.AuthorityScore < 0 || analysis.AuthorityScore > 100 {
		t.Errorf("AuthorityScore should be between 0-100, got: %.2f", analysis.AuthorityScore)
	}

	t.Logf("StructureScore: %.2f, AuthorityScore: %.2f, GeoScore: %.2f",
		analysis.StructureScore, analysis.AuthorityScore, analysis.GeoScore)
}

func TestScorer_Compare(t *testing.T) {
	scorer := NewScorer(&mockLLMClient{})

	before := `简单的内容
没有任何结构。`

	after := `# 优化后的内容

## 有结构的内容

- 列表项1
- 列表项2

根据研究显示，结构化内容更容易被引用。`

	comparison, err := scorer.Compare(context.Background(), before, after)
	if err != nil {
		t.Fatalf("Compare should not return error: %v", err)
	}

	if comparison.Before == nil {
		t.Fatal("Before score should not be nil")
	}
	if comparison.After == nil {
		t.Fatal("After score should not be nil")
	}

	// 优化后的内容应该分数更高
	if comparison.TotalChange <= 0 {
		t.Logf("Warning: TotalChange is %.2f (expected positive)", comparison.TotalChange)
	}

	// 验证各维度提升
	if comparison.Improvements == nil {
		t.Fatal("Improvements should not be nil")
	}

	t.Logf("Before: %.2f, After: %.2f, Change: %.2f",
		comparison.Before.OverallScore(),
		comparison.After.OverallScore(),
		comparison.TotalChange)

	summary := comparison.GetImprovementSummary()
	if summary == "" {
		t.Error("Improvement summary should not be empty")
	}

	t.Logf("\nImprovement Summary:\n%s", summary)
}

func TestScorer_calculateStructureScore(t *testing.T) {
	scorer := NewScorer(&mockLLMClient{})

	tests := []struct {
		name     string
		content  string
		minScore float64
	}{
		{
			name:     "Empty content",
			content:  "",
			minScore: 0,
		},
		{
			name:     "Content with headings",
			content:  "# Heading 1\n## Heading 2\n### Heading 3",
			minScore: 15,
		},
		{
			name:     "Content with lists",
			content:  "- Item 1\n- Item 2\n- Item 3",
			minScore: 10,
		},
		{
			name:     "Content with sections",
			content:  "## Section 1\nContent\n## Section 2\nContent",
			minScore: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateStructureScore(tt.content)
			if score < float64(tt.minScore) {
				t.Errorf("Expected score >= %.2f, got: %.2f", tt.minScore, score)
			}
			t.Logf("Score: %.2f", score)
		})
	}
}

func TestScorer_calculateAuthorityScore(t *testing.T) {
	scorer := NewScorer(&mockLLMClient{})

	tests := []struct {
		name     string
		content  string
		minScore float64
	}{
		{
			name:     "Empty content",
			content:  "",
			minScore: 0,
		},
		{
			name:     "Content with data",
			content:  "研究显示85%的用户选择此产品",
			minScore: 5,
		},
		{
			name:     "Content with sources",
			content:  "根据来源：研究机构报道，数据显示结果",
			minScore: 5,
		},
		{
			name:     "Content with professional terms",
			content:  "优化策略分析评估效果",
			minScore: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateAuthorityScore(tt.content)
			if score < float64(tt.minScore) {
				t.Errorf("Expected score >= %.2f, got: %.2f", tt.minScore, score)
			}
			t.Logf("Score: %.2f", score)
		})
	}
}

func TestScorer_calculateClarityScore(t *testing.T) {
	scorer := NewScorer(&mockLLMClient{})

	tests := []struct {
		name     string
		content  string
		minScore float64
	}{
		{
			name:     "Empty content",
			content:  "",
			minScore: 0,
		},
		{
			name:     "Short sentences",
			content:  "这是第一句。这是第二句。这是第三句。",
			minScore: 20,
		},
		{
			name:     "Content with connectors",
			content:  "因此，但是，此外，另外",
			minScore: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateClarityScore(tt.content)
			if score < float64(tt.minScore) {
				t.Errorf("Expected score >= %.2f, got: %.2f", tt.minScore, score)
			}
			t.Logf("Score: %.2f", score)
		})
	}
}

func TestScorer_calculateCitationScore(t *testing.T) {
	scorer := NewScorer(&mockLLMClient{})

	tests := []struct {
		name     string
		content  string
		minScore float64
	}{
		{
			name:     "Empty content",
			content:  "",
			minScore: 0,
		},
		{
			name:     "Content with conclusion",
			content:  "总结：这是一项重要研究。因此得出结论。",
			minScore: 10,
		},
		{
			name:     "Content with facts",
			content:  "2024年是云计算发展的关键年份",
			minScore: 5,
		},
		{
			name:     "Content with recommendations",
			content:  "建议选择更好的服务，推荐使用云服务，可以提高效率",
			minScore: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateCitationScore(tt.content)
			if score < float64(tt.minScore) {
				t.Errorf("Expected score >= %.2f, got: %.2f", tt.minScore, score)
			}
			t.Logf("Score: %.2f", score)
		})
	}
}

func TestScorer_calculateSchemaScore(t *testing.T) {
	scorer := NewScorer(&mockLLMClient{})

	tests := []struct {
		name     string
		content  string
		minScore float64
	}{
		{
			name:     "Empty content",
			content:  "",
			minScore: 0,
		},
		{
			name:     "Content with JSON-LD",
			content:  `<script type="application/ld+json">{"@type": "Article"}</script>`,
			minScore: 70,
		},
		{
			name:     "Content with plain JSON",
			content:  `{"@context": "https://schema.org", "@type": "WebPage"}`,
			minScore: 60,
		},
		{
			name:     "Content without Schema",
			content:  "普通内容",
			minScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateSchemaScore(tt.content)
			if score < float64(tt.minScore) {
				t.Errorf("Expected score >= %.2f, got: %.2f", tt.minScore, score)
			}
			t.Logf("Score: %.2f", score)
		})
	}
}

func TestGeoScore_OverallScore(t *testing.T) {
	tests := []struct {
		name     string
		score    models.GeoScore
		expected float64
	}{
		{
			name: "All zeros",
			score: models.GeoScore{
				Structure: 0,
				Authority: 0,
				Clarity:   0,
				Citation:  0,
				Schema:    0,
			},
			expected: 0,
		},
		{
			name: "All 100",
			score: models.GeoScore{
				Structure: 100,
				Authority: 100,
				Clarity:   100,
				Citation:  100,
				Schema:    100,
			},
			expected: 100,
		},
		{
			name: "Mixed scores",
			score: models.GeoScore{
				Structure: 80,
				Authority: 60,
				Clarity:   90,
				Citation:  70,
				Schema:    50,
			},
			expected: (80 + 60 + 90 + 70 + 50) / 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.score.OverallScore()
			if actual != tt.expected {
				t.Errorf("Expected %.2f, got: %.2f", tt.expected, actual)
			}
		})
	}
}

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

func TestScorer_ScoreWithSuggestions(t *testing.T) {
	// 模拟LLM返回的评分和建议JSON
	mockResp := `{
		"scores": {
			"structure": 65,
			"structure_reason": "结构不够清晰",
			"authority": 70,
			"authority_reason": "有部分数据支撑",
			"clarity": 80,
			"clarity_reason": "表达较清晰",
			"citation": 55,
			"citation_reason": "缺少可引用结论",
			"schema": 40,
			"schema_reason": "无结构化标记"
		},
		"dimension_suggestions": {
			"structure": [
				{"issue": "缺少二级标题", "direction": "添加更多章节划分", "priority": "high", "estimated_gain": 15, "example": "## 核心概念"}
			],
			"citation": [
				{"issue": "无明确结论", "direction": "在开头添加结论摘要", "priority": "high", "estimated_gain": 20, "example": "结论：..."}
			],
			"schema": [
				{"issue": "无JSON-LD标记", "direction": "添加Schema.org结构化数据", "priority": "medium", "estimated_gain": 10, "example": ""}
			]
		},
		"top_suggestions": [
			{"issue": "缺少可引用结论", "direction": "在开头添加结论摘要", "priority": "high", "estimated_gain": 20, "example": "结论：..."},
			{"issue": "缺少二级标题", "direction": "添加更多章节划分", "priority": "high", "estimated_gain": 15, "example": "## 核心概念"}
		]
	}`

	scorer := NewScorer(&mockLLMClientForScoring{response: mockResp})
	content := "# 测试内容\n\n一些测试内容。"

	result, err := scorer.ScoreWithSuggestions(context.Background(), content)
	if err != nil {
		t.Fatalf("ScoreWithSuggestions should not return error: %v", err)
	}

	if result == nil {
		t.Fatal("ScoreWithSuggestions should return non-nil result")
	}

	// 验证评分类型
	if result.ScoreType != "ai" {
		t.Errorf("Expected ScoreType 'ai', got: %s", result.ScoreType)
	}

	// 验证未降级
	if result.Degraded {
		t.Error("Expected Degraded to be false")
	}

	// 验证有维度建议
	if len(result.DimensionSuggestions) == 0 {
		t.Error("Expected dimension suggestions")
	}

	// 验证有Top建议
	if len(result.TopSuggestions) == 0 {
		t.Error("Expected top suggestions")
	}

	// 验证Top建议按estimated_gain排序
	if len(result.TopSuggestions) >= 2 {
		if result.TopSuggestions[0].EstimatedGain < result.TopSuggestions[1].EstimatedGain {
			t.Error("Top suggestions should be sorted by estimated_gain descending")
		}
	}

	t.Logf("ScoreType: %s, Degraded: %v", result.ScoreType, result.Degraded)
	t.Logf("Dimensions with suggestions: %d", len(result.DimensionSuggestions))
	t.Logf("Top suggestions: %d", len(result.TopSuggestions))
	for i, s := range result.TopSuggestions {
		t.Logf("  %d. [%s] %s (gain: %.0f)", i+1, s.Priority, s.Issue, s.EstimatedGain)
	}
}

func TestScorer_ScoreWithSuggestions_Degraded(t *testing.T) {
	// 模拟LLM调用失败
	scorer := NewScorer(&mockLLMClientForScoring{err: fmt.Errorf("timeout")})
	content := "# 测试内容\n\n测试内容。"

	result, err := scorer.ScoreWithSuggestions(context.Background(), content)
	if err != nil {
		t.Fatalf("ScoreWithSuggestions should not return error even on LLM failure: %v", err)
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

	// 验证无建议
	if len(result.DimensionSuggestions) != 0 {
		t.Error("Expected no dimension suggestions on degraded")
	}

	if len(result.TopSuggestions) != 0 {
		t.Error("Expected no top suggestions on degraded")
	}

	t.Logf("Degraded: %v, ScoreType: %s, ErrorMessage: %s",
		result.Degraded, result.ScoreType, result.ErrorMessage)
}
