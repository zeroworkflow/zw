package errors

import (
	"fmt"
)

// AIError represents an AI-related error
type AIError struct {
	Type      string
	Message   string
	RequestID string
	Code      int
}

func (e *AIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("AI error [%s]: %s (request: %s)", e.Type, e.Message, e.RequestID)
	}
	return fmt.Sprintf("AI error [%s]: %s", e.Type, e.Message)
}

// APIError represents an API-related error
type APIError struct {
	StatusCode int
	Message    string
	RequestID  string
	Endpoint   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d at %s: %s (request: %s)", 
		e.StatusCode, e.Endpoint, e.Message, e.RequestID)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s (value: %v)", 
		e.Field, e.Message, e.Value)
}

// StreamError represents a streaming-related error
type StreamError struct {
	Phase   string
	Message string
}

func (e *StreamError) Error() string {
	return fmt.Sprintf("stream error in %s: %s", e.Phase, e.Message)
}

// NewAIError creates a new AI error
func NewAIError(errorType, message, requestID string) *AIError {
	return &AIError{
		Type:      errorType,
		Message:   message,
		RequestID: requestID,
	}
}

// NewAPIError creates a new API error
func NewAPIError(statusCode int, message, requestID, endpoint string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		RequestID:  requestID,
		Endpoint:   endpoint,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// NewStreamError creates a new stream error
func NewStreamError(phase, message string) *StreamError {
	return &StreamError{
		Phase:   phase,
		Message: message,
	}
}
