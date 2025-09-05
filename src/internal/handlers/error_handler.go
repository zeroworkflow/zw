package handlers

import (
	"fmt"
	"os"
)

// ErrorHandler provides centralized error handling
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// HandleFatalError handles fatal errors and exits
func (h *ErrorHandler) HandleFatalError(err error, context string) {
	fmt.Fprintf(os.Stderr, "Error in %s: %v\n", context, err)
	os.Exit(1)
}

// HandleError handles non-fatal errors
func (h *ErrorHandler) HandleError(err error, context string) {
	fmt.Fprintf(os.Stderr, "Warning in %s: %v\n", context, err)
}
