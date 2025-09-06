package zai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"zero-workflow/src/internal/config"
	"zero-workflow/src/pkg/errors"
	httplib "zero-workflow/src/pkg/http"
	"zero-workflow/src/pkg/stream"
	"zero-workflow/src/pkg/types"

	"github.com/google/uuid"
)

const (
	bufferSize = 4096
	apiVersion = "prod-fe-1.0.57"
)

// Client implements the AI client for Z.ai
type Client struct {
	config          *config.Config
	aiParams        *config.AIParams
	userCtx         *config.UserContext
	authToken       string
	httpClient      *httplib.SecureHTTPClient
	streamProcessor *stream.Processor
}

// ChatResponse represents the chat creation response
type ChatResponse struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Title     string      `json:"title"`
	Chat      interface{} `json:"chat"`
	UpdatedAt int64       `json:"updated_at"`
	CreatedAt int64       `json:"created_at"`
}

// NewClient creates a new Z.ai client
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	cfg := config.DefaultConfig()
	return &Client{
		config:          cfg,
		aiParams:        config.DefaultAIParams(),
		userCtx:         config.DefaultUserContext(),
		authToken:       token,
		httpClient:      httplib.NewSecureHTTPClient(cfg.Timeout),
		streamProcessor: stream.NewProcessor(),
	}, nil
}

// Chat implements the client interface
func (c *Client) Chat(ctx context.Context, message string) (string, error) {
	return c.ChatStream(ctx, message, nil)
}

// ChatStream implements the client interface
func (c *Client) ChatStream(ctx context.Context, message string, callback types.StreamCallback) (string, error) {
	systemPrompt := types.Message{
		Role: "system",
		Content: `Ты ZeroWorkflow AI - помощник разработчика. 
Отвечай кратко и по делу на русском языке.
Используй markdown для форматирования.
Для блоков кода используй тройные бэктики с указанием языка: ` + "```язык\nкод\n```",
	}

	userMessage := types.Message{
		Role:    "user",
		Content: message,
	}

	messages := []types.Message{systemPrompt, userMessage}
	return c.ChatStreamWithMessages(ctx, messages, callback)
}

// ChatWithMessages implements the client interface
func (c *Client) ChatWithMessages(ctx context.Context, messages []types.Message) (string, error) {
	return c.ChatStreamWithMessages(ctx, messages, nil)
}

// ChatStreamWithMessages implements the client interface
func (c *Client) ChatStreamWithMessages(ctx context.Context, messages []types.Message, callback types.StreamCallback) (string, error) {
	chatID, err := c.createNewChat(ctx, messages[len(messages)-1].Content)
	if err != nil {
		return "", fmt.Errorf("failed to create chat: %w", err)
	}

	return c.sendMessageStream(ctx, chatID, messages, callback)
}

// createNewChat creates a new chat session
func (c *Client) createNewChat(ctx context.Context, firstMessage string) (string, error) {
	messageID := uuid.New().String()
	timestamp := time.Now().Unix()

	payload := map[string]interface{}{
		"chat": map[string]interface{}{
			"id":     "",
			"title":  "ZeroWorkflow Chat",
			"models": []string{c.config.Model},
			"params": map[string]interface{}{},
			"history": map[string]interface{}{
				"messages": map[string]interface{}{
					messageID: map[string]interface{}{
						"id":          messageID,
						"parentId":    nil,
						"childrenIds": []string{},
						"role":        "user",
						"content":     firstMessage,
						"timestamp":   timestamp,
						"models":      []string{c.config.Model},
					},
				},
				"currentId": messageID,
			},
			"messages": []map[string]interface{}{
				{
					"id":          messageID,
					"parentId":    nil,
					"childrenIds": []string{},
					"role":        "user",
					"content":     firstMessage,
					"timestamp":   timestamp,
					"models":      []string{c.config.Model},
				},
			},
			"tags":            []string{},
			"flags":           []string{},
			"features":        c.buildFeatures(),
			"enable_thinking": false,
			"timestamp":       timestamp * 1000,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", errors.NewValidationError("payload", payload, "failed to marshal chat payload")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.APIBaseURL+"/v1/chats/new", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", errors.NewNetworkError("failed to create request", "", c.config.APIBaseURL+"/v1/chats/new", 0, err)
	}

	c.setHeaders(req, "")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://chat.z.ai/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Sanitize error to prevent token leakage
		sanitizedErr := fmt.Errorf("network error: %s", errors.SanitizeForLog(err))
		return "", errors.NewNetworkError("failed to send request", "", c.config.APIBaseURL+"/v1/chats/new", 0, sanitizedErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Sanitize response body to prevent token leakage
		sanitizedBody := errors.SanitizeForLog(fmt.Errorf(string(body)))
		return "", errors.NewNetworkError("API request failed", "", c.config.APIBaseURL+"/v1/chats/new", resp.StatusCode, fmt.Errorf(sanitizedBody))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", errors.NewValidationError("response", nil, "failed to decode chat response")
	}

	return chatResp.ID, nil
}

// sendMessageStream sends messages and streams the response
func (c *Client) sendMessageStream(ctx context.Context, chatID string, messages []types.Message, callback types.StreamCallback) (string, error) {
	requestID := uuid.New().String()

	payload := map[string]interface{}{
		"stream":       true,
		"model":        c.config.Model,
		"messages":     messages,
		"params":       c.buildParams(),
		"tool_servers": []string{},
		"features":     c.buildFeaturesMap(),
		"variables":    c.buildVariables(),
		"model_item":   c.buildModelItem(),
		"chat_id":      chatID,
		"id":           requestID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", errors.NewValidationError("payload", payload, "failed to marshal message payload")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.APIBaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", errors.NewNetworkError("failed to create request", requestID, c.config.APIBaseURL+"/chat/completions", 0, err)
	}

	c.setHeaders(req, chatID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Sanitize error to prevent token leakage
		sanitizedErr := fmt.Errorf("network error: %s", errors.SanitizeForLog(err))
		return "", errors.NewNetworkError("failed to send request", requestID, c.config.APIBaseURL+"/chat/completions", 0, sanitizedErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Sanitize response body to prevent token leakage
		sanitizedBody := errors.SanitizeForLog(fmt.Errorf(string(body)))
		return "", errors.NewNetworkError("API request failed", requestID, c.config.APIBaseURL+"/chat/completions", resp.StatusCode, fmt.Errorf(sanitizedBody))
	}

	return c.streamProcessor.ProcessStream(resp.Body, callback)
}

// Helper methods for building request components

func (c *Client) setHeaders(req *http.Request, chatID string) {
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.config.UserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Origin", "https://chat.z.ai")
	req.Header.Set("X-FE-Version", apiVersion)

	if chatID != "" {
		req.Header.Set("Referer", "https://chat.z.ai/c/"+chatID)
	}
}

func (c *Client) buildParams() map[string]interface{} {
	return map[string]interface{}{
		"temperature": c.aiParams.Temperature,
		"top_p":       c.aiParams.TopP,
		"max_tokens":  c.aiParams.MaxTokens,
	}
}

func (c *Client) buildVariables() map[string]string {
	now := time.Now()
	return map[string]string{
		"{{USER_NAME}}":        c.userCtx.Name,
		"{{USER_LOCATION}}":    c.userCtx.Location,
		"{{CURRENT_DATETIME}}": now.Format("2006-01-02 15:04:05"),
		"{{CURRENT_DATE}}":     now.Format("2006-01-02"),
		"{{CURRENT_TIME}}":     now.Format("15:04:05"),
		"{{CURRENT_WEEKDAY}}":  now.Weekday().String(),
		"{{CURRENT_TIMEZONE}}": c.userCtx.Timezone,
		"{{USER_LANGUAGE}}":    c.userCtx.Language,
	}
}

func (c *Client) buildFeatures() []map[string]interface{} {
	return []map[string]interface{}{
		{"type": "mcp", "server": "vibe-coding", "status": "hidden"},
		{"type": "mcp", "server": "ppt-maker", "status": "hidden"},
		{"type": "mcp", "server": "image-search", "status": "hidden"},
	}
}

func (c *Client) buildFeaturesMap() map[string]interface{} {
	return map[string]interface{}{
		"image_generation": false,
		"code_interpreter": false,
		"web_search":       false,
		"auto_web_search":  false,
		"preview_mode":     true,
		"flags":            []string{},
		"features":         c.buildFeatures(),
		"enable_thinking":  false,
	}
}

func (c *Client) buildModelItem() map[string]interface{} {
	return map[string]interface{}{
		"id":       c.config.Model,
		"name":     "GLM-4.5",
		"owned_by": "openai",
	}
}
