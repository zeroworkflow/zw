package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	baseURL   string
	authToken string
	userAgent string
	httpClient *http.Client
}

// SendMessageStream sends messages and streams deltas via onDelta callback while also returning the final message.
func (c *Client) SendMessageStream(chatID string, messages []Message, onDelta func(string)) (string, error) {
	requestID := uuid.New().String()

	payload := map[string]interface{}{
		"stream":  true,
		"model":   "0727-360B-API",
		"messages": messages,
		"params": map[string]interface{}{
			"temperature":  0.8,
			"top_p":       0.95,
			"max_tokens":  4000,
		},
		"tool_servers": []string{},
		"features": map[string]interface{}{
			"image_generation":  false,
			"code_interpreter":  false,
			"web_search":       false,
			"auto_web_search":  false,
			"preview_mode":     true,
			"flags":           []string{},
			"features": []map[string]interface{}{
				{"type": "mcp", "server": "vibe-coding", "status": "hidden"},
				{"type": "mcp", "server": "ppt-maker", "status": "hidden"},
				{"type": "mcp", "server": "image-search", "status": "hidden"},
			},
			"enable_thinking": false,
		},
		"variables": map[string]string{
			"{{USER_NAME}}":        "Developer",
			"{{USER_LOCATION}}":    "Russia",
			"{{CURRENT_DATETIME}}": time.Now().Format("2006-01-02 15:04:05"),
			"{{CURRENT_DATE}}":     time.Now().Format("2006-01-02"),
			"{{CURRENT_TIME}}":     time.Now().Format("15:04:05"),
			"{{CURRENT_WEEKDAY}}":  time.Now().Weekday().String(),
			"{{CURRENT_TIMEZONE}}": "Europe/Moscow",
			"{{USER_LANGUAGE}}":    "ru-RU",
		},
		"model_item": map[string]interface{}{
			"id":       "0727-360B-API",
			"name":     "GLM-4.5",
			"owned_by": "openai",
		},
		"chat_id": chatID,
		"id":      requestID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", "https://chat.z.ai")
	req.Header.Set("Referer", "https://chat.z.ai/c/"+chatID)
	req.Header.Set("X-FE-Version", "prod-fe-1.0.57")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return c.parseStreamResponseStream(resp.Body, onDelta)
}

// parseStreamResponseStream parses SSE stream, calls onDelta for each token, and returns the final cleaned response.
func (c *Client) parseStreamResponseStream(reader io.Reader, onDelta func(string)) (string, error) {
	var fullResponse strings.Builder
	bufReader := bufio.NewReader(reader)

	for {
		line, err := bufReader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read response: %w", err)
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return c.cleanResponse(fullResponse.String()), nil
			}

			var chunk StreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				// Skip invalid JSON
			} else {
				if chunk.Type == "chat:completion" && chunk.Data.DeltaContent != "" {
					if onDelta != nil {
						onDelta(chunk.Data.DeltaContent)
					}
					fullResponse.WriteString(chunk.Data.DeltaContent)
				}
				if chunk.Data.Done {
					return c.cleanResponse(fullResponse.String()), nil
				}
			}
		}

		if err == io.EOF {
			break
		}
	}

	return c.cleanResponse(fullResponse.String()), nil
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Title     string `json:"title"`
	Chat      interface{} `json:"chat"`
	UpdatedAt int64  `json:"updated_at"`
	CreatedAt int64  `json:"created_at"`
}

type StreamChunk struct {
	Type string `json:"type"`
	Data struct {
		DeltaContent string `json:"delta_content,omitempty"`
		Phase        string `json:"phase"`
		Done         bool   `json:"done,omitempty"`
		Usage        interface{} `json:"usage,omitempty"`
	} `json:"data"`
}

func NewClient() (*Client, error) {
	token, err := GetToken()
	if err != nil {
		return nil, fmt.Errorf("AI token not found. Please run 'zw ai login' to configure it: %w", err)
	}

	return &Client{
		baseURL:   "https://chat.z.ai/api",
		authToken: token,
		userAgent: "Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0",
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

func (c *Client) createNewChat(firstMessage string) (string, error) {
	messageID := uuid.New().String()
	timestamp := time.Now().Unix()

	payload := map[string]interface{}{
		"chat": map[string]interface{}{
			"id":     "",
			"title":  "ZeroWorkflow Chat",
			"models": []string{"0727-360B-API"},
			"params": map[string]interface{}{},
			"history": map[string]interface{}{
				"messages": map[string]interface{}{
					messageID: map[string]interface{}{
						"id":         messageID,
						"parentId":   nil,
						"childrenIds": []string{},
						"role":       "user",
						"content":    firstMessage,
						"timestamp":  timestamp,
						"models":     []string{"0727-360B-API"},
					},
				},
				"currentId": messageID,
			},
			"messages": []map[string]interface{}{
				{
					"id":         messageID,
					"parentId":   nil,
					"childrenIds": []string{},
					"role":       "user",
					"content":    firstMessage,
					"timestamp":  timestamp,
					"models":     []string{"0727-360B-API"},
				},
			},
			"tags":  []string{},
			"flags": []string{},
			"features": []map[string]interface{}{
				{"type": "mcp", "server": "vibe-coding", "status": "hidden"},
				{"type": "mcp", "server": "ppt-maker", "status": "hidden"},
				{"type": "mcp", "server": "image-search", "status": "hidden"},
			},
			"enable_thinking": false,
			"timestamp":       timestamp * 1000,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/v1/chats/new", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://chat.z.ai")
	req.Header.Set("Referer", "https://chat.z.ai/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return chatResp.ID, nil
}

func (c *Client) SendMessage(chatID string, messages []Message) (string, error) {
	requestID := uuid.New().String()

	payload := map[string]interface{}{
		"stream":  true,
		"model":   "0727-360B-API",
		"messages": messages,
		"params": map[string]interface{}{
			"temperature":  0.8,
			"top_p":       0.95,
			"max_tokens":  4000,
		},
		"tool_servers": []string{},
		"features": map[string]interface{}{
			"image_generation":  false,
			"code_interpreter":  false,
			"web_search":       false,
			"auto_web_search":  false,
			"preview_mode":     true,
			"flags":           []string{},
			"features": []map[string]interface{}{
				{"type": "mcp", "server": "vibe-coding", "status": "hidden"},
				{"type": "mcp", "server": "ppt-maker", "status": "hidden"},
				{"type": "mcp", "server": "image-search", "status": "hidden"},
			},
			"enable_thinking": false,
		},
		"variables": map[string]string{
			"{{USER_NAME}}":        "Developer",
			"{{USER_LOCATION}}":    "Russia",
			"{{CURRENT_DATETIME}}": time.Now().Format("2006-01-02 15:04:05"),
			"{{CURRENT_DATE}}":     time.Now().Format("2006-01-02"),
			"{{CURRENT_TIME}}":     time.Now().Format("15:04:05"),
			"{{CURRENT_WEEKDAY}}":  time.Now().Weekday().String(),
			"{{CURRENT_TIMEZONE}}": "Europe/Moscow",
			"{{USER_LANGUAGE}}":    "ru-RU",
		},
		"model_item": map[string]interface{}{
			"id":       "0727-360B-API",
			"name":     "GLM-4.5",
			"owned_by": "openai",
		},
		"chat_id": chatID,
		"id":      requestID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", "https://chat.z.ai")
	req.Header.Set("Referer", "https://chat.z.ai/c/"+chatID)
	req.Header.Set("X-FE-Version", "prod-fe-1.0.57")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return c.parseStreamResponse(resp.Body)
}

func (c *Client) parseStreamResponse(reader io.Reader) (string, error) {
	var fullResponse strings.Builder
	buffer := make([]byte, 4096)

	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read response: %w", err)
		}

		if n == 0 {
			break
		}

		lines := strings.Split(string(buffer[:n]), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					return c.cleanResponse(fullResponse.String()), nil
				}

				var chunk StreamChunk
				if err := json.Unmarshal([]byte(data), &chunk); err != nil {
					continue // Skip invalid JSON
				}

				if chunk.Type == "chat:completion" && chunk.Data.DeltaContent != "" {
					fullResponse.WriteString(chunk.Data.DeltaContent)
				}

				if chunk.Data.Done {
					return c.cleanResponse(fullResponse.String()), nil
				}
			}
		}
	}

	return c.cleanResponse(fullResponse.String()), nil
}

func (c *Client) cleanResponse(text string) string {
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

func (c *Client) Chat(message string) (string, error) {
	// Create new chat for each request (simple implementation)
	chatID, err := c.createNewChat(message)
	if err != nil {
		return "", fmt.Errorf("failed to create chat: %w", err)
	}

	systemPrompt := Message{
		Role: "system",
		Content: `Ты ZeroWorkflow AI - помощник разработчика. 
Отвечай кратко и по делу на русском языке.
Используй markdown для форматирования.
Для блоков кода используй тройные бэктики с указанием языка: ` + "```язык\nкод\n```",
	}

	userMessage := Message{
		Role:    "user",
		Content: message,
	}

	messages := []Message{systemPrompt, userMessage}
	return c.SendMessage(chatID, messages)
}

func (c *Client) ChatStream(message string, onDelta func(string)) (string, error) {
	chatID, err := c.createNewChat(message)
	if err != nil {
		return "", fmt.Errorf("failed to create chat: %w", err)
	}

	systemPrompt := Message{
		Role: "system",
		Content: `Ты ZeroWorkflow AI - помощник разработчика. 
Отвечай кратко и по делу на русском языке.
Используй markdown для форматирования.
Для блоков кода используй тройные бэктики с указанием языка: ` + "```язык\nкод\n```",
	}

	userMessage := Message{
		Role:    "user",
		Content: message,
	}

	messages := []Message{systemPrompt, userMessage}
	return c.SendMessageStream(chatID, messages, onDelta)
}
