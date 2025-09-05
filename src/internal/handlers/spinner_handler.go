package handlers

import (
	"zero-workflow/src/internal/ui"
)

// SpinnerHandler manages spinner lifecycle
type SpinnerHandler struct {
	spinner *ui.RightSpinner
}

// NewSpinnerHandler creates a new spinner handler
func NewSpinnerHandler(text string) *SpinnerHandler {
	return &SpinnerHandler{
		spinner: ui.NewRightSpinner(text),
	}
}

// Start starts the spinner
func (h *SpinnerHandler) Start() {
	h.spinner.Start()
}

// Stop stops the spinner
func (h *SpinnerHandler) Stop() {
	h.spinner.Stop()
}

// WithSpinner executes function with spinner
func (h *SpinnerHandler) WithSpinner(fn func() error) error {
	h.Start()
	defer h.Stop()
	return fn()
}
