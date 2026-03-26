package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the e2e test configuration.
type Config struct {
	APIKey  string
	Model   string
	BaseURL string
	Timeout int
}

// Load loads configuration from .env file.
func Load() (*Config, error) {
	cfg := &Config{
		Model:   "glm-4-flash",
		Timeout: 300,
	}

	// Try multiple paths to find .env
	paths := []string{
		".env",
		"../.env",
		"../../.env",
	}

	var envPath string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			envPath = p
			break
		}
	}

	if envPath == "" {
		return nil, fmt.Errorf(".env file not found, please copy .env.example to .env and configure")
	}

	if err := loadEnvFile(envPath); err != nil {
		return nil, fmt.Errorf("failed to load .env: %w", err)
	}

	cfg.APIKey = os.Getenv("GLM_API_KEY")
	if model := os.Getenv("GLM_MODEL"); model != "" {
		cfg.Model = model
	}
	if baseURL := os.Getenv("GLM_BASE_URL"); baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if timeout := os.Getenv("GLM_TIMEOUT"); timeout != "" {
		fmt.Sscanf(timeout, "%d", &cfg.Timeout)
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("GLM_API_KEY is required in .env file")
	}

	return cfg, nil
}

// loadEnvFile loads environment variables from a file.
func loadEnvFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	file, err := os.Open(absPath)
	if err != nil {
		return err
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
