package ai

import (
	"zero-workflow/src/pkg/interfaces"
)

// Client defines the interface for AI clients (alias for interfaces.AIClient)
type Client = interfaces.AIClient

// Provider defines the interface for AI providers (alias for interfaces.Provider)
type Provider = interfaces.Provider

// Factory creates AI clients
type Factory interface {
	// CreateClient creates a client for the specified provider
	CreateClient(provider string, token string) (Client, error)
	
	// RegisterProvider registers a new provider
	RegisterProvider(name string, provider Provider)
	
	// ListProviders returns available providers
	ListProviders() []string
}
