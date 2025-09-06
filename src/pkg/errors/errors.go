package errors

import (
	"fmt"
	"strings"
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
	ErrorTypeGit         ErrorType = "git"
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

// GitError represents git operation errors
type GitError struct {
	*BaseError
	Command string
	Args    []string
}

func NewGitError(command string, args []string, message string, cause error) *GitError {
	return &GitError{
		BaseError: &BaseError{
			Type:      ErrorTypeGit,
			Message:   message,
			Timestamp: time.Now(),
			Cause:     cause,
		},
		Command: command,
		Args:    args,
	}
}

// SanitizeForLog removes sensitive information from error messages for logging
func SanitizeForLog(err error) string {
	msg := err.Error()
	
	// Remove Bearer tokens
	msg = strings.ReplaceAll(msg, "Bearer ", "Bearer [REDACTED]")
	
	// Remove Authorization headers
	if strings.Contains(msg, "Authorization:") {
		lines := strings.Split(msg, "\n")
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), "authorization:") {
				lines[i] = "Authorization: [REDACTED]"
			}
		}
		msg = strings.Join(lines, "\n")
	}
	
	return msg
}

// ValidateGitCommand validates git command arguments for security
func ValidateGitCommand(command string, args []string) error {
	// Whitelist of allowed git commands
	allowedCommands := map[string]bool{
		"diff":   true,
		"status": true,
		"push":   true,
		"config": true,
	}
	
	if !allowedCommands[command] {
		return NewValidationError("git_command", command, "git command not allowed")
	}
	
	// Validate arguments based on command
	switch command {
	case "diff":
		return validateDiffArgs(args)
	case "push":
		return validatePushArgs(args)
	case "config":
		return validateConfigArgs(args)
	}
	
	return nil
}

func validateDiffArgs(args []string) error {
	allowedArgs := map[string]bool{
		"--staged": true,
		"--cached": true,
	}
	
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") && !allowedArgs[arg] {
			return NewValidationError("git_diff_arg", arg, "git diff argument not allowed")
		}
		// Prevent command injection
		if strings.Contains(arg, ";") || strings.Contains(arg, "|") || strings.Contains(arg, "&") {
			return NewValidationError("git_diff_arg", arg, "potentially dangerous characters in git argument")
		}
	}
	return nil
}

func validatePushArgs(args []string) error {
	if len(args) != 2 {
		return NewValidationError("git_push_args", args, "git push requires exactly 2 arguments: remote and branch")
	}
	
	remote := args[0]
	branch := args[1]
	
	// Validate remote name (only alphanumeric, dash, underscore)
	for _, r := range remote {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return NewValidationError("git_remote", remote, "invalid characters in remote name")
		}
	}
	
	// Validate branch name
	for _, r := range branch {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '/' || r == '.') {
			return NewValidationError("git_branch", branch, "invalid characters in branch name")
		}
	}
	
	return nil
}

func validateConfigArgs(args []string) error {
	if len(args) != 1 {
		return NewValidationError("git_config_args", args, "git config requires exactly 1 argument")
	}
	
	key := args[0]
	allowedKeys := map[string]bool{
		"user.name":  true,
		"user.email": true,
	}
	
	if !allowedKeys[key] {
		return NewValidationError("git_config_key", key, "git config key not allowed")
	}
	
	return nil
}
