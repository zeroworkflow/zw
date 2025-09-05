# ZeroWorkflow Makefile

.PHONY: build install clean test help dev run

# Variables
BINARY_NAME=zw
MAIN_PATH=src/main.go
BUILD_DIR=.
INSTALL_DIR=/usr/local/bin

# Default target
all: build

# Build the application
build:
	@echo "Building ZeroWorkflow..."
	@go mod tidy
	@go build -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build completed: ./$(BINARY_NAME)"

# Install globally
install: build
	@echo "Installing ZeroWorkflow globally..."
	@sudo cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@sudo chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "ZeroWorkflow installed to $(INSTALL_DIR)/$(BINARY_NAME)"

# Quick install using script
install-script:
	@./install.sh

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@go clean
	@echo "Clean completed"

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Development build with race detection
dev:
	@echo "Building for development..."
	@go build -race -o $(BINARY_NAME) $(MAIN_PATH)

# Run the application
run: build
	@./$(BINARY_NAME)

# Run with arguments
run-ask:
	@./$(BINARY_NAME) ask "$(ARGS)"

# Run in interactive mode
run-interactive: build
	@./$(BINARY_NAME) ask -i

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Show help
help:
	@echo "ZeroWorkflow Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  install        - Install globally (requires sudo)"
	@echo "  install-script - Install using install.sh script"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  dev            - Build with race detection"
	@echo "  run            - Build and run"
	@echo "  run-ask        - Run with question (use ARGS='your question')"
	@echo "  run-interactive- Run in interactive mode"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make install"
	@echo "  make run-ask ARGS='How to create a Go struct?'"
	@echo "  make run-interactive"
