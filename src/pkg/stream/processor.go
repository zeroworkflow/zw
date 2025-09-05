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
	DefaultBufferSize = 64 * 1024
	MaxBufferSize     = 2 * 1024 * 1024
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
	// Use sync.Pool for string builder to reduce allocations
	builderPool := &sync.Pool{
		New: func() interface{} {
			builder := &strings.Builder{}
			builder.Grow(4096)
			return builder
		},
	}
	
	response := builderPool.Get().(*strings.Builder)
	defer func() {
		response.Reset()
		builderPool.Put(response)
	}()
	
	scanner := bufio.NewScanner(reader)
	
	// Get buffer from pool
	buf := p.pool.Get().([]byte)
	defer p.pool.Put(buf)
	
	scanner.Buffer(buf, MaxBufferSize)

	// Reuse chunk struct to avoid allocations
	var chunk StreamChunk
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Avoid string allocation for prefix check
		if len(line) < 6 || line[:6] != "data: " {
			continue
		}

		data := line[6:] // Avoid TrimPrefix allocation
		if data == "[DONE]" {
			return p.cleanResponse(response.String()), nil
		}

		// Reset chunk for reuse
		chunk = StreamChunk{}
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

// cleanResponse normalizes and cleans the response text with memory optimization
func (p *Processor) cleanResponse(text string) string {
	if text == "" {
		return ""
	}

	// Use builder for efficient string operations
	var builder strings.Builder
	builder.Grow(len(text))
	
	// Normalize line endings in single pass
	runes := []rune(text)
	fenceCount := 0
	
	for i, r := range runes {
		switch r {
		case '\r':
			if i+1 < len(runes) && runes[i+1] == '\n' {
				continue // Skip \r in \r\n
			}
			builder.WriteRune('\n')
		case '`':
			if i+2 < len(runes) && runes[i+1] == '`' && runes[i+2] == '`' {
				fenceCount++
			}
			builder.WriteRune(r)
		default:
			builder.WriteRune(r)
		}
	}
	
	cleaned := strings.TrimSpace(builder.String())
	
	// Balance triple backticks if needed
	if fenceCount%2 == 1 {
		cleaned += "\n```"
	}
	
	// Limit consecutive newlines efficiently
	return p.limitNewlines(cleaned)
}

// limitNewlines reduces consecutive newlines efficiently
func (p *Processor) limitNewlines(text string) string {
	var result strings.Builder
	result.Grow(len(text))
	
	newlineCount := 0
	for _, r := range text {
		if r == '\n' {
			newlineCount++
			if newlineCount <= 3 {
				result.WriteRune(r)
			}
		} else {
			newlineCount = 0
			result.WriteRune(r)
		}
	}
	
	return result.String()
}
