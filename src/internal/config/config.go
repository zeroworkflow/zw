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
	apiURL := os.Getenv("ZW_API_URL")
	if apiURL == "" {
		apiURL = "https://chat.z.ai/api"
	}
	
	userAgent := os.Getenv("ZW_USER_AGENT")
	if userAgent == "" {
		userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0"
	}
	
	model := os.Getenv("ZW_MODEL")
	if model == "" {
		model = "0727-360B-API"
	}

	return &Config{
		APIBaseURL: apiURL,
		UserAgent:  userAgent,
		Timeout:    120 * time.Second,
		Model:      model,
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

// Load loads configuration
func Load() (*Config, error) {
	LoadEnv()
	return DefaultConfig(), nil
}

// GetToken retrieves AI token from environment
func GetToken() (string, error) {
	// Try to load .env file first
	LoadEnv()

	token := os.Getenv("AI_TOKEN")
	if token == "" {
		return "", fmt.Errorf("AI_TOKEN environment variable not set")
	}
	return token, nil
}

// LoadEnv loads environment variables from .env file if it exists
func LoadEnv() error {
    // Try to find .env file in current directory, parent directories, and XDG/HOME config locations
    envPaths := []string{
        ".env",
        "../.env",
        "../../.env",
    }

    // XDG config path: $XDG_CONFIG_HOME/zw/.env
    if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
        envPaths = append(envPaths, filepath.Join(xdg, "zw", ".env"))
    }

    // HOME config path: $HOME/.config/zw/.env
    if home := os.Getenv("HOME"); home != "" {
        envPaths = append(envPaths, filepath.Join(home, ".config", "zw", ".env"))
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

