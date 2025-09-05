package ai

import (
	"fmt"
	"sync"
)

// DefaultFactory is the default AI factory instance
var DefaultFactory = NewFactory()

// factory implements the Factory interface
type factory struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewFactory creates a new AI factory
func NewFactory() Factory {
	return &factory{
		providers: make(map[string]Provider),
	}
}

// CreateClient creates a client for the specified provider
func (f *factory) CreateClient(provider string, token string) (Client, error) {
	f.mu.RLock()
	p, exists := f.providers[provider]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider '%s' not registered", provider)
	}

	return p.CreateClient(token)
}

// RegisterProvider registers a new provider
func (f *factory) RegisterProvider(name string, provider Provider) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providers[name] = provider
}

// ListProviders returns available providers
func (f *factory) ListProviders() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	providers := make([]string, 0, len(f.providers))
	for name := range f.providers {
		providers = append(providers, name)
	}
	return providers
}

// MustCreateClient creates a client or panics on error
func MustCreateClient(provider string, token string) Client {
	client, err := DefaultFactory.CreateClient(provider, token)
	if err != nil {
		panic(fmt.Sprintf("failed to create AI client: %v", err))
	}
	return client
}
