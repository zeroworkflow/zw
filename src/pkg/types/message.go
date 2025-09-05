package types

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// StreamCallback is called for each delta during streaming
type StreamCallback func(delta string)
