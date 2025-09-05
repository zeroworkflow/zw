package ai

import (
	"context"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// StreamCallback is called for each delta during streaming
type StreamCallback func(delta string)

// Client defines the interface for AI clients
type Client interface {
	// Chat sends a message and returns the complete response
	Chat(ctx context.Context, message string) (string, error)
	
	// ChatStream sends a message and streams the response via callback
	ChatStream(ctx context.Context, message string, callback StreamCallback) (string, error)
	
	// ChatWithMessages sends multiple messages and returns response
	ChatWithMessages(ctx context.Context, messages []Message) (string, error)
	
	// ChatStreamWithMessages sends multiple messages and streams response
	ChatStreamWithMessages(ctx context.Context, messages []Message, callback StreamCallback) (string, error)
}

// Provider defines the interface for AI providers
type Provider interface {
	// CreateClient creates a new AI client with the given token
	CreateClient(token string) (Client, error)
	
	// ValidateToken validates the token format
	ValidateToken(token string) error
	
	// GetName returns the provider name
	GetName() string
}

// Factory creates AI clients
type Factory interface {
	// CreateClient creates a client for the specified provider
	CreateClient(provider string, token string) (Client, error)
	
	// RegisterProvider registers a new provider
	RegisterProvider(name string, provider Provider)
	
	// ListProviders returns available providers
	ListProviders() []string
}
