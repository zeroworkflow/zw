package zai

import (
	"zero-workflow/src/internal/config"
)

// Provider implements the AI provider for Z.ai
type Provider struct{}

// NewProvider creates a new Z.ai provider
func NewProvider() *Provider {
	return &Provider{}
}

// CreateClient creates a new Z.ai client with the given token
func (p *Provider) CreateClient(token string) (interface{}, error) {
	return NewClient(token)
}

// ValidateToken validates the token format for Z.ai
func (p *Provider) ValidateToken(token string) error {
	return config.ValidateToken(token)
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "z.ai"
}
