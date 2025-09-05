package ai

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	keyring "github.com/zalando/go-keyring"
)

var (
	// ErrTokenNotFound is returned when no AI token can be resolved from any source.
	ErrTokenNotFound = errors.New("AI token not found")
)

const (
	keyringService = "ZeroWorkflow"
	keyringUser    = "ai_token"
)

type config struct {
	AIToken string `json:"ai_token"`
}

func getConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		// Fallback to ~/.config
		home, herr := os.UserHomeDir()
		if herr != nil {
			return "", fmt.Errorf("cannot resolve config dir: %w", err)
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "zw", "config.json"), nil
}

func loadConfig() (*config, error) {
	p, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &config{}, nil
		}
		return nil, err
	}
	var c config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func saveConfig(c *config) (string, error) {
	p, err := getConfigPath()
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.Write(b); err != nil {
		return "", err
	}
	return p, nil
}

// ResolveToken returns the token and its source: env | keyring | file.
func ResolveToken() (string, string, error) {
	if tok := strings.TrimSpace(os.Getenv("AI_TOKEN")); tok != "" {
		return tok, "env", nil
	}
	if tok, err := keyring.Get(keyringService, keyringUser); err == nil {
		if strings.TrimSpace(tok) != "" {
			return tok, "keyring", nil
		}
	}
	cfg, err := loadConfig()
	if err == nil && strings.TrimSpace(cfg.AIToken) != "" {
		return cfg.AIToken, "file", nil
	}
	return "", "", ErrTokenNotFound
}

// GetToken resolves token or returns ErrTokenNotFound.
func GetToken() (string, error) {
	tok, _, err := ResolveToken()
	return tok, err
}

// SaveToken stores token, preferring keyring; falls back to config file.
// Returns storage location: keyring | file.
func SaveToken(token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", errors.New("empty token")
	}
	if err := keyring.Set(keyringService, keyringUser, token); err == nil {
		return "keyring", nil
	}
	cfg, err := loadConfig()
	if err != nil {
		return "", err
	}
	cfg.AIToken = token
	path, err := saveConfig(cfg)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file:%s", path), nil
}

// DeleteToken removes token from keyring and config file.
func DeleteToken() error {
	_ = keyring.Delete(keyringService, keyringUser)
	cfg, err := loadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if cfg.AIToken == "" {
		return nil
	}
	cfg.AIToken = ""
	_, err = saveConfig(cfg)
	return err
}
