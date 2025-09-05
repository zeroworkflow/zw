package stream

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"zero-workflow/src/pkg/types"
)

const (
	DefaultBufferSize = 8192
	MaxBufferSize     = 65536
)

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	Type string `json:"type"`
	Data struct {
		DeltaContent string      `json:"delta_content,omitempty"`
		Phase        string      `json:"phase"`
		Done         bool        `json:"done,omitempty"`
		Usage        interface{} `json:"usage,omitempty"`
	} `json:"data"`
}

// Processor handles streaming responses efficiently
type Processor struct {
	bufferSize int
	pool       sync.Pool
}

// NewProcessor creates new stream processor
func NewProcessor() *Processor {
	return &Processor{
		bufferSize: DefaultBufferSize,
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, DefaultBufferSize)
			},
		},
	}
}

// ProcessStream processes SSE stream with optimized buffering and pooling
func (p *Processor) ProcessStream(reader io.Reader, callback types.StreamCallback) (string, error) {
	var response strings.Builder
	response.Grow(4096) // Pre-allocate for better performance
	
	scanner := bufio.NewScanner(reader)
	
	// Get buffer from pool
	buf := p.pool.Get().([]byte)
	defer p.pool.Put(buf)
	
	scanner.Buffer(buf, MaxBufferSize)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return p.cleanResponse(response.String()), nil
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue // Skip invalid JSON
		}

		if chunk.Type == "chat:completion" && chunk.Data.DeltaContent != "" {
			if callback != nil {
				callback(chunk.Data.DeltaContent)
			}
			response.WriteString(chunk.Data.DeltaContent)
		}

		if chunk.Data.Done {
			return p.cleanResponse(response.String()), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read stream: %w", err)
	}

	return p.cleanResponse(response.String()), nil
}

// cleanResponse normalizes and cleans the response text
func (p *Processor) cleanResponse(text string) string {
	if text == "" {
		return ""
	}

	// Normalize line endings
	cleaned := strings.ReplaceAll(text, "\r\n", "\n")
	cleaned = strings.ReplaceAll(cleaned, "\r", "\n")
	cleaned = strings.TrimSpace(cleaned)

	// Balance triple backticks
	fenceCount := strings.Count(cleaned, "```")
	if fenceCount%2 == 1 {
		cleaned += "\n```"
	}

	// Limit consecutive newlines
	for strings.Contains(cleaned, "\n\n\n\n") {
		cleaned = strings.ReplaceAll(cleaned, "\n\n\n\n", "\n\n\n")
	}

	return cleaned
}
