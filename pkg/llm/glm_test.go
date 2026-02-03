package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewGLMClient(t *testing.T) {
	config := Config{
		APIKey:      "test-key",
		Model:       "glm-4.7",
		MaxTokens:   2048,
		Temperature: 0.5,
	}

	client, err := NewGLMClient(config)
	if err != nil {
		t.Fatalf("NewGLMClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("client is nil")
	}

	if client.config.Model != "glm-4.7" {
		t.Errorf("Model = %v, want glm-4.7", client.config.Model)
	}
}

func TestGLMClientChat(t *testing.T) {
	// 创建 mock 服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Authorization header missing")
		}

		// 返回 mock 响应
		resp := glmResponse{
			ID:    "test-123",
			Model: "glm-4.7",
			Choices: []struct {
				Index   int `json:"index"`
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			}{
				{
					Index: 0,
					Message: struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					}{
						Role:    "assistant",
						Content: "测试响应内容",
					},
					FinishReason: "stop",
				},
			},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// 创建客户端
	client, _ := NewGLMClient(Config{
		APIKey:  "test-key",
		BaseURL: mockServer.URL,
		Model:   "glm-4.7",
	})

	// 测试 Chat 方法
	resp, err := client.Chat(context.Background(), &ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "你好"},
		},
	})

	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if resp.Content != "测试响应内容" {
		t.Errorf("Content = %v, want 测试响应内容", resp.Content)
	}

	if resp.TokensUsed != 30 {
		t.Errorf("TokensUsed = %v, want 30", resp.TokensUsed)
	}
}

func TestGLMClientChatWithError(t *testing.T) {
	// 创建返回错误的 mock 服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "401",
				"message": "Invalid API key",
			},
		})
	}))
	defer mockServer.Close()

	client, _ := NewGLMClient(Config{
		APIKey:  "invalid-key",
		BaseURL: mockServer.URL,
	})

	_, err := client.Chat(context.Background(), &ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "测试"},
		},
	})

	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestMessageTypeConversion(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "test content",
	}

	// 测试类型转换
	glmMsg := glmMessage(msg)

	if glmMsg.Role != "user" {
		t.Errorf("Role = %v, want user", glmMsg.Role)
	}

	if glmMsg.Content != "test content" {
		t.Errorf("Content = %v, want test content", glmMsg.Content)
	}
}
