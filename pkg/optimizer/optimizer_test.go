package optimizer

import (
	"context"
	"testing"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/llm"
	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
	strategiespkg "github.com/Lin-Jiong-HDU/geo-optimizer/pkg/optimizer/strategies"
)

// mockLLMClient 模拟LLM客户端
type mockLLMClient struct {
	llm.LLMClient
	response string
	err      error
}

func (m *mockLLMClient) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
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

// TestNewOptimizer 测试创建优化器
func TestNewOptimizer(t *testing.T) {
	client := &mockLLMClient{}
	opt := New(client)

	if opt == nil {
		t.Fatal("Expected non-nil optimizer")
	}

	if opt.llmClient == nil {
		t.Error("Expected llmClient to be set")
	}

	if opt.scorer == nil {
		t.Error("Expected scorer to be set")
	}

	if opt.strategies == nil {
		t.Error("Expected strategies map to be initialized")
	}
}

// TestRegisterStrategy 测试注册策略
func TestRegisterStrategy(t *testing.T) {
	client := &mockLLMClient{}
	opt := New(client)

	initialCount := len(opt.strategies)

	// 注册新策略
	strategy := newMockStrategy()
	opt.RegisterStrategy(strategy)

	if len(opt.strategies) != initialCount+1 {
		t.Errorf("Expected strategy count to be %d, got %d", initialCount+1, len(opt.strategies))
	}
}

// TestValidateRequest 测试请求验证
func TestValidateRequest(t *testing.T) {
	client := &mockLLMClient{}
	opt := New(client)

	tests := []struct {
		name    string
		req     *models.OptimizationRequest
		wantErr bool
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "empty content",
			req: &models.OptimizationRequest{
				Title:   "Test",
				Content: "",
			},
			wantErr: true,
		},
		{
			name: "empty title",
			req: &models.OptimizationRequest{
				Title:   "",
				Content: "Test content",
			},
			wantErr: true,
		},
		{
			name: "valid request",
			req: &models.OptimizationRequest{
				Title:   "Test Title",
				Content: "Test content",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := opt.validateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestApplyDefaults 测试应用默认值
func TestApplyDefaults(t *testing.T) {
	client := &mockLLMClient{}
	opt := New(client)

	req := &models.OptimizationRequest{
		Title:   "Test Title",
		Content: "Test content",
	}

	opt.applyDefaults(req)

	if len(req.Strategies) == 0 {
		t.Error("Expected default strategies to be set")
	}

	if len(req.TargetAI) == 0 {
		t.Error("Expected default target AI to be set")
	}

	if req.AIPreferences == nil {
		t.Error("Expected AI preferences to be initialized")
	}
}

// TestExtractSchema 测试Schema提取
func TestExtractSchema(t *testing.T) {
	client := &mockLLMClient{}
	opt := New(client)

	tests := []struct {
		name        string
		content     string
		shouldExist bool
	}{
		{
			name:        "JSON-LD script tag",
			content:     `<script type="application/ld+json">{"@context": "https://schema.org"}</script>`,
			shouldExist: true,
		},
		{
			name:        "no schema",
			content:     "Just some text without schema",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := opt.extractSchema(tt.content)
			if tt.shouldExist && result == "" {
				t.Errorf("extractSchema() should return schema, got empty")
			}
			if !tt.shouldExist && result != "" {
				t.Errorf("extractSchema() should return empty, got %v", result)
			}
		})
	}
}

// TestExtractFAQ 测试FAQ提取
func TestExtractFAQ(t *testing.T) {
	client := &mockLLMClient{}
	opt := New(client)

	tests := []struct {
		name        string
		content     string
		shouldExist bool
	}{
		{
			name:        "FAQ section with ##",
			content:     "## 常见问题\n\nQ1: What is this?\nA: This is a test.\n\n## Other Section",
			shouldExist: true,
		},
		{
			name:        "FAQ section with ###",
			content:     "### FAQ\n\nQ1: Question\nA: Answer",
			shouldExist: true,
		},
		{
			name:        "no FAQ",
			content:     "Just some text without FAQ",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := opt.extractFAQ(tt.content)
			if tt.shouldExist && result == "" {
				t.Errorf("extractFAQ() should return FAQ, got empty")
			}
			if !tt.shouldExist && result != "" {
				t.Errorf("extractFAQ() should return empty, got %v", result)
			}
		})
	}
}

// TestExtractSummary 测试摘要提取
func TestExtractSummary(t *testing.T) {
	client := &mockLLMClient{}
	opt := New(client)

	tests := []struct {
		name        string
		content     string
		shouldExist bool
	}{
		{
			name:        "Summary section",
			content:     "## 摘要\n\nThis is a summary of the content.",
			shouldExist: true,
		},
		{
			name:        "Summary section with ###",
			content:     "### 总结\n\nThis is a conclusion.",
			shouldExist: true,
		},
		{
			name:        "no summary section",
			content:     "This is the first paragraph.\n\nThis is the second paragraph.",
			shouldExist: true,
		},
		{
			name:        "short content",
			content:     "Short content",
			shouldExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := opt.extractSummary(tt.content)
			if tt.shouldExist && result == "" {
				t.Errorf("extractSummary() should return summary, got empty")
			}
		})
	}
}

// TestCountProductMentions 测试产品提及计数
func TestCountProductMentions(t *testing.T) {
	client := &mockLLMClient{}
	opt := New(client)

	tests := []struct {
		name        string
		content     string
		productName string
		expected    int
	}{
		{
			name:        "single mention",
			content:     "We recommend ProductX for this use case.",
			productName: "ProductX",
			expected:    1,
		},
		{
			name:        "multiple mentions",
			content:     "ProductX is great. ProductX helps you save time.",
			productName: "ProductX",
			expected:    2,
		},
		{
			name:        "case insensitive",
			content:     "productx and ProductX are the same.",
			productName: "ProductX",
			expected:    2,
		},
		{
			name:        "no mention",
			content:     "This is about something else.",
			productName: "ProductX",
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := opt.countProductMentions(tt.content, tt.productName)
			if result != tt.expected {
				t.Errorf("countProductMentions() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestExtractDifferentiationPoints 测试差异化要点提取
func TestExtractDifferentiationPoints(t *testing.T) {
	client := &mockLLMClient{}
	opt := New(client)

	content := "我们的产品有独特的功能，使其与竞争对手不同。" +
		"主要优势是更好的性能。相比其他产品，我们提供卓越的质量。"

	points := opt.extractDifferentiationPoints(content, []models.CompetitorInfo{})

	// 只验证函数能正常执行，不强制要求提取到要点
	t.Logf("Extracted %d differentiation points", len(points))
	for i, p := range points {
		t.Logf("Point %d: %s", i+1, p)
	}
}

// Mock strategy for testing
type mockStrategy struct {
	*strategiespkg.BaseStrategy
}

func newMockStrategy() *mockStrategy {
	return &mockStrategy{
		BaseStrategy: strategiespkg.NewBaseStrategy("mock", "mock"),
	}
}

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
