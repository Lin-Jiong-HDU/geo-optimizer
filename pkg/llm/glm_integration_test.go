package llm

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Lin-Jiong-HDU/geo-optimizer/pkg/models"
)

// loadEnv 从 .env 文件加载环境变量
func loadEnv() error {
	// 尝试多个可能的路径
	paths := []string{
		".env",
		"../.env",
		"../../.env",
	}

	var file *os.File
	var err error

	for _, path := range paths {
		file, err = os.Open(path)
		if err == nil {
			break
		}
	}

	if file == nil {
		return fmt.Errorf(" .env file not found")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

func TestGLMClientIntegration(t *testing.T) {
	if err := loadEnv(); err != nil {
		t.Skip("Failed to load .env file, skipping integration test")
	}

	apiKey := os.Getenv("GLM_API_KEY")
	if apiKey == "" {
		t.Skip("GLM_API_KEY not set in .env, skipping integration test")
	}

	client, err := NewGLMClient(Config{
		APIKey:      apiKey,
		Model:       "glm-4-flash",
		MaxTokens:   2000,
		Temperature: 0.7,
		Timeout:     120,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Run("Chat", func(t *testing.T) {
		resp, err := client.Chat(context.Background(), &ChatRequest{
			Messages: []Message{
				{Role: "user", Content: "什么是AI？"},
			},
		})

		if err != nil {
			t.Fatalf("Chat failed: %v", err)
		}

		output := fmt.Sprintf(`# Chat 接口测试

## 输入
什么是AI？

## 输出
%s

## Token使用
%d
`, resp.Content, resp.TokensUsed)

		os.WriteFile("test_chat.md", []byte(output), 0644)
		t.Logf("Chat test saved to test_chat.md")
	})

	t.Run("GenerateSchema", func(t *testing.T) {
		schema, err := client.GenerateSchema(context.Background(), "人工智能是计算机科学的一个分支", "Article")
		if err != nil {
			t.Fatalf("GenerateSchema failed: %v", err)
		}

		output := fmt.Sprintf(`# GenerateSchema 接口测试

## 输入内容
人工智能是计算机科学的一个分支

## Schema类型
Article

## 生成的Schema
%s
`, schema)

		os.WriteFile("test_schema.md", []byte(output), 0644)
		t.Logf("Schema test saved to test_schema.md")
	})

	t.Run("AnalyzeContent", func(t *testing.T) {
		analysis, err := client.AnalyzeContent(context.Background(), "人工智能正在改变世界")
		if err != nil {
			t.Fatalf("AnalyzeContent failed: %v", err)
		}

		output := fmt.Sprintf(`# AnalyzeContent 接口测试

## 输入内容
人工智能正在改变世界

## 分析结果
- 结构化评分: %.0f
- 权威性评分: %.0f
- 清晰度评分: %.0f
- 可引用性评分: %.0f
- Schema评分: %.0f
- 总分: %.0f

## 关键词
%v

## 建议
%v
`,
			analysis.StructureScore,
			analysis.AuthorityScore,
			analysis.ClarityScore,
			analysis.CitationScore,
			analysis.SchemaScore,
			analysis.TotalScore,
			analysis.Keywords,
			analysis.Suggestions,
		)

		os.WriteFile("test_analysis.md", []byte(output), 0644)
		t.Logf("Analysis test saved to test_analysis.md")
	})

	t.Run("GenerateOptimization", func(t *testing.T) {
		req := &models.OptimizationRequest{
			Title:   "AI技术简介",
			Content: "AI是人工智能的缩写",
			Enterprise: models.EnterpriseInfo{
				CompanyName:        "测试公司",
				ProductName:        "测试产品",
				ProductDescription: "这是一个测试产品",
			},
			TargetAI:   []string{"chatgpt"},
			Keywords:   []string{"AI", "人工智能"},
			Strategies: []models.StrategyType{models.StrategyStructure},
		}

		resp, err := client.GenerateOptimization(context.Background(), req)
		if err != nil {
			t.Fatalf("GenerateOptimization failed: %v", err)
		}

		output := fmt.Sprintf(`# GenerateOptimization 接口测试

## 原始内容
AI是人工智能的缩写

## 优化后内容
%s

## 策略
%v

## 模型
%s

## Token使用
%d

## 生成时间
%s
`,
			resp.OptimizedContent,
			resp.AppliedStrategies,
			resp.LLMModel,
			resp.TokensUsed,
			resp.GeneratedAt.Format("2006-01-02 15:04:05"),
		)

		os.WriteFile("test_optimization.md", []byte(output), 0644)
		t.Logf("Optimization test saved to test_optimization.md")
	})
}
