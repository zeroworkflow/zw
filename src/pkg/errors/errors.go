package errors

import (
	"fmt"
	"time"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeNetwork     ErrorType = "network"
	ErrorTypeAPI         ErrorType = "api"
	ErrorTypeValidation  ErrorType = "validation"
	ErrorTypeStream      ErrorType = "stream"
	ErrorTypeConfig      ErrorType = "config"
	ErrorTypeFile        ErrorType = "file"
	ErrorTypeAuth        ErrorType = "auth"
)

// BaseError represents a base error with context
type BaseError struct {
	Type      ErrorType
	Message   string
	RequestID string
	Timestamp time.Time
	Cause     error
}

func (e *BaseError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("[%s] %s (request: %s, time: %s)", 
			e.Type, e.Message, e.RequestID, e.Timestamp.Format(time.RFC3339))
	}
	return fmt.Sprintf("[%s] %s (time: %s)", 
		e.Type, e.Message, e.Timestamp.Format(time.RFC3339))
}

func (e *BaseError) Unwrap() error {
	return e.Cause
}

// NetworkError represents network-related errors
type NetworkError struct {
	*BaseError
	URL        string
	StatusCode int
}

func NewNetworkError(message, requestID, url string, statusCode int, cause error) *NetworkError {
	return &NetworkError{
		BaseError: &BaseError{
			Type:      ErrorTypeNetwork,
			Message:   message,
			RequestID: requestID,
			Timestamp: time.Now(),
			Cause:     cause,
		},
		URL:        url,
		StatusCode: statusCode,
	}
}

// ValidationError represents validation errors
type ValidationError struct {
	*BaseError
	Field string
	Value interface{}
}

func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		BaseError: &BaseError{
			Type:      ErrorTypeValidation,
			Message:   message,
			Timestamp: time.Now(),
		},
		Field: field,
		Value: value,
	}
}

// StreamError represents streaming errors
type StreamError struct {
	*BaseError
	Phase string
}

func NewStreamError(phase, message string, cause error) *StreamError {
	return &StreamError{
		BaseError: &BaseError{
			Type:      ErrorTypeStream,
			Message:   message,
			Timestamp: time.Now(),
			Cause:     cause,
		},
		Phase: phase,
	}
}

// ConfigError represents configuration errors
type ConfigError struct {
	*BaseError
	ConfigKey string
}

func NewConfigError(configKey, message string, cause error) *ConfigError {
	return &ConfigError{
		BaseError: &BaseError{
			Type:      ErrorTypeConfig,
			Message:   message,
			Timestamp: time.Now(),
			Cause:     cause,
		},
		ConfigKey: configKey,
	}
}

// FileError represents file operation errors
type FileError struct {
	*BaseError
	FilePath string
}

func NewFileError(filePath, message string, cause error) *FileError {
	return &FileError{
		BaseError: &BaseError{
			Type:      ErrorTypeFile,
			Message:   message,
			Timestamp: time.Now(),
			Cause:     cause,
		},
		FilePath: filePath,
	}
}

// AuthError represents authentication errors
type AuthError struct {
	*BaseError
}

func NewAuthError(message string, cause error) *AuthError {
	return &AuthError{
		BaseError: &BaseError{
			Type:      ErrorTypeAuth,
			Message:   message,
			Timestamp: time.Now(),
			Cause:     cause,
		},
	}
}
