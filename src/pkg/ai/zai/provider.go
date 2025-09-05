package zai

import (
	"fmt"
	"zero-workflow/src/pkg/interfaces"
)

// Provider implements the AI provider for Z.ai
type Provider struct{}

// NewProvider creates a new Z.ai provider
func NewProvider() *Provider {
	return &Provider{}
}

// CreateClient creates a new Z.ai client with the given token
func (p *Provider) CreateClient(token string) (interfaces.AIClient, error) {
	return NewClient(token)
}

// ValidateToken validates the token format for Z.ai
func (p *Provider) ValidateToken(token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	return nil
}

// GetName returns the provider name
func (p *Provider) GetName() string {
	return "z.ai"
}
