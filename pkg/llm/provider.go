package llm

// Provider LLM提供商类型
type Provider string

const (
	ProviderGLM       Provider = "glm"       // 智谱AI (GLM)
	ProviderOpenAI    Provider = "openai"    // OpenAI
	ProviderAnthropic Provider = "anthropic" // Anthropic (Claude)
	ProviderAzure     Provider = "azure"     // Azure OpenAI
	ProviderGoogle    Provider = "google"    // Google (Gemini)
	ProviderCustom    Provider = "custom"    // 自定义端点
)

// Config LLM客户端配置
type Config struct {
	// 提供商
	Provider Provider `json:"provider"`
	// API密钥
	APIKey string `json:"api_key"`
	// 自定义端点URL（可选）
	BaseURL string `json:"base_url,omitempty"`
	// 模型名称
	Model string `json:"model"`
	// 最大token数
	MaxTokens int `json:"max_tokens"`
	// 温度参数
	Temperature float64 `json:"temperature"`
	// 请求超时时间（秒）
	Timeout int `json:"timeout"`
}
