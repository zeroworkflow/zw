package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	APIBaseURL string
	UserAgent  string
	Timeout    time.Duration
	Model      string
}

// AIParams holds AI-specific parameters
type AIParams struct {
	Temperature float64
	TopP        float64
	MaxTokens   int
}

// UserContext holds user-specific context variables
type UserContext struct {
	Name     string
	Location string
	Language string
	Timezone string
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		APIBaseURL: "https://chat.z.ai/api",
		UserAgent:  "Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0",
		Timeout:    120 * time.Second,
		Model:      "0727-360B-API",
	}
}

// DefaultAIParams returns default AI parameters
func DefaultAIParams() *AIParams {
	return &AIParams{
		Temperature: 0.8,
		TopP:        0.95,
		MaxTokens:   4000,
	}
}

// DefaultUserContext returns default user context
func DefaultUserContext() *UserContext {
	return &UserContext{
		Name:     "Developer",
		Location: "Russia",
		Language: "ru-RU",
		Timezone: "Europe/Moscow",
	}
}

// ValidateToken validates AI token format
func ValidateToken(token string) error {
	if token == "" {
		return fmt.Errorf("AI_TOKEN environment variable not set")
	}
	if len(token) < 10 {
		return fmt.Errorf("invalid token format: too short")
	}
	return nil
}

// LoadEnv loads environment variables from .env file if it exists
func LoadEnv() error {
	// Try to find .env file in current directory and parent directories
	envPaths := []string{
		".env",
		"../.env",
		"../../.env",
	}

	for _, envPath := range envPaths {
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err != nil {
				return fmt.Errorf("failed to load %s: %w", envPath, err)
			}
			return nil
		}
	}

	// Try to find .env in executable directory
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		envPath := filepath.Join(execDir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			if err := godotenv.Load(envPath); err != nil {
				return fmt.Errorf("failed to load %s: %w", envPath, err)
			}
			return nil
		}
	}

	// .env file not found, but that's okay
	return nil
}

// GetToken retrieves and validates AI token from environment
func GetToken() (string, error) {
	// Try to load .env file first
	LoadEnv()

	token := os.Getenv("AI_TOKEN")
	if err := ValidateToken(token); err != nil {
		return "", err
	}
	return token, nil
}
