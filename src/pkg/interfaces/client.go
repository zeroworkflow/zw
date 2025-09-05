package interfaces

import (
	"context"
	"zero-workflow/src/pkg/types"
)

// AIClient defines the interface for AI clients
type AIClient interface {
	// Chat sends a message and returns the complete response
	Chat(ctx context.Context, message string) (string, error)
	
	// ChatStream sends a message and streams the response via callback
	ChatStream(ctx context.Context, message string, callback types.StreamCallback) (string, error)
	
	// ChatWithMessages sends multiple messages and returns response
	ChatWithMessages(ctx context.Context, messages []types.Message) (string, error)
	
	// ChatStreamWithMessages sends multiple messages and streams response
	ChatStreamWithMessages(ctx context.Context, messages []types.Message, callback types.StreamCallback) (string, error)
}

// HTTPClient defines interface for HTTP operations
type HTTPClient interface {
	Do(req *HTTPRequest) (*HTTPResponse, error)
}

// HTTPRequest wraps HTTP request
type HTTPRequest interface {
	SetHeader(key, value string)
	SetBody(body []byte)
	GetURL() string
}

// HTTPResponse wraps HTTP response
type HTTPResponse interface {
	GetStatusCode() int
	GetBody() []byte
	Close() error
}

// StreamProcessor handles streaming responses
type StreamProcessor interface {
	ProcessStream(reader interface{}, callback types.StreamCallback) (string, error)
}

// ConfigProvider provides configuration
type ConfigProvider interface {
	GetAPIURL() string
	GetUserAgent() string
	GetModel() string
	GetTimeout() int
}
