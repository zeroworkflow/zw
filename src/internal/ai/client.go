package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"zero-workflow/src/internal/config"
)

// Client represents an AI client
type Client struct {
	config     *config.Config
	httpClient *http.Client
}

// NewClient creates a new AI client
func NewClient(cfg *config.Config) (*Client, error) {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}, nil
}

// GenerateText generates text using AI
func (c *Client) GenerateText(prompt string) (string, error) {
	token, err := config.GetToken()
	if err != nil {
		return "", fmt.Errorf("failed to get AI token: %w", err)
	}

	requestBody := map[string]interface{}{
		"model": c.config.Model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.APIBaseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse SSE format response
	content, err := parseSSEResponse(string(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to parse SSE response: %w. Raw response: %s", err, string(bodyBytes))
	}

	return content, nil
}

// parseSSEResponse parses Server-Sent Events format response
func parseSSEResponse(response string) (string, error) {
	lines := strings.Split(response, "\n")
	var content strings.Builder
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and [DONE] markers
		if line == "" || line == "data: [DONE]" {
			continue
		}
		
		// Parse data: lines
		if strings.HasPrefix(line, "data: ") {
			jsonStr := strings.TrimPrefix(line, "data: ")
			
			// Parse the JSON structure
			var data struct {
				Type string `json:"type"`
				Data struct {
					Data struct {
						Error struct {
							Detail string `json:"detail"`
							Code   int    `json:"code"`
						} `json:"error"`
						Done bool `json:"done"`
					} `json:"data"`
					Error struct {
						Detail string `json:"detail"`
						Code   int    `json:"code"`
					} `json:"error"`
					Done bool `json:"done"`
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					} `json:"choices"`
				} `json:"data"`
			}
			
			if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
				continue // Skip malformed JSON
			}
			
			// Check for errors
			if data.Data.Error.Code != 0 {
				return "", fmt.Errorf("API error %d: %s", data.Data.Error.Code, data.Data.Error.Detail)
			}
			if data.Data.Data.Error.Code != 0 {
				return "", fmt.Errorf("API error %d: %s", data.Data.Data.Error.Code, data.Data.Data.Error.Detail)
			}
			
			// Extract content from choices
			for _, choice := range data.Data.Choices {
				if choice.Message.Content != "" {
					content.WriteString(choice.Message.Content)
				}
				if choice.Delta.Content != "" {
					content.WriteString(choice.Delta.Content)
				}
			}
		}
	}
	
	result := content.String()
	if result == "" {
		return "", fmt.Errorf("no content found in response")
	}
	
	return result, nil
}
