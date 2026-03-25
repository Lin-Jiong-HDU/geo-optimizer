package llm

// Provider represents the type of LLM provider.
type Provider string

// LLM provider constants.
const (
	ProviderGLM       Provider = "glm"       // Zhipu AI (GLM)
	ProviderOpenAI    Provider = "openai"    // OpenAI
	ProviderAnthropic Provider = "anthropic" // Anthropic (Claude)
	ProviderAzure     Provider = "azure"     // Azure OpenAI
	ProviderGoogle    Provider = "google"    // Google (Gemini)
	ProviderCustom    Provider = "custom"    // Custom endpoint
)

// Config holds the LLM client configuration.
type Config struct {
	Provider    Provider `json:"provider"`
	APIKey      string   `json:"api_key"`
	BaseURL     string   `json:"base_url,omitempty"`
	Model       string   `json:"model"`
	MaxTokens   int      `json:"max_tokens"`
	Temperature float64  `json:"temperature"`
	Timeout     int      `json:"timeout"` // Request timeout in seconds
}
