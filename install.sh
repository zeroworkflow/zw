#!/bin/bash

# ZeroWorkflow Installation Script
# This script builds and installs ZeroWorkflow globally

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21+ first."
        print_status "Visit: https://golang.org/doc/install"
        exit 1
    fi
    
    GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
    print_status "Found Go version: $GO_VERSION"
}

# Build the application
build_app() {
    print_status "Building ZeroWorkflow..."
    
    if ! go mod tidy; then
        print_error "Failed to tidy Go modules"
        exit 1
    fi
    
    if ! go build -o zw src/main.go; then
        print_error "Failed to build ZeroWorkflow"
        exit 1
    fi
    
    print_success "Build completed successfully"
}

# Install globally
install_global() {
    print_status "Installing ZeroWorkflow globally..."
    
    # Determine installation directory
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        INSTALL_DIR="/usr/local/bin"
    else
        # Linux and others
        INSTALL_DIR="/usr/local/bin"
    fi
    
    # Check if we have write permissions
    if [[ ! -w "$INSTALL_DIR" ]]; then
        print_warning "Need sudo permissions to install to $INSTALL_DIR"
        if ! sudo cp zw "$INSTALL_DIR/zw"; then
            print_error "Failed to install ZeroWorkflow to $INSTALL_DIR"
            exit 1
        fi
        sudo chmod +x "$INSTALL_DIR/zw"
    else
        if ! cp zw "$INSTALL_DIR/zw"; then
            print_error "Failed to install ZeroWorkflow to $INSTALL_DIR"
            exit 1
        fi
        chmod +x "$INSTALL_DIR/zw"
    fi
    
    print_success "ZeroWorkflow installed to $INSTALL_DIR/zw"
}

# Setup .env file
setup_env() {
    if [[ ! -f ".env" ]]; then
        print_status "Creating .env file template..."
        cat > .env << 'EOF'
# ZeroWorkflow Configuration
# Add your AI token here
AI_TOKEN=your_ai_token_here
EOF
        print_warning "Please edit .env file and add your AI_TOKEN"
        print_status "You can also set AI_TOKEN as environment variable"
    else
        print_status ".env file already exists"
    fi
}

# Verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    if command -v zw &> /dev/null; then
        print_success "ZeroWorkflow is now available globally!"
        print_status "Try running: zw --help"
        
        # Test if .env is being read
        if [[ -f ".env" ]] && grep -q "AI_TOKEN=" .env && ! grep -q "your_ai_token_here" .env; then
            print_status "Testing AI connection..."
            if zw ask "test" &> /dev/null; then
                print_success "AI connection working!"
            else
                print_warning "AI connection failed. Please check your AI_TOKEN in .env"
            fi
        else
            print_warning "Please configure AI_TOKEN in .env file or environment variable"
        fi
    else
        print_error "Installation verification failed"
        exit 1
    fi
}

# Main installation process
main() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════╗"
    echo "║        ZeroWorkflow Installer        ║"
    echo "║   AI-powered developer tools suite   ║"
    echo "╚══════════════════════════════════════╝"
    echo -e "${NC}"
    
    check_go
    build_app
    setup_env
    install_global
    verify_installation
    
    echo
    print_success "Installation completed successfully!"
    echo
    print_status "Quick start:"
    echo "  1. Edit .env file and add your AI_TOKEN"
    echo "  2. Run: zw ask \"How to create a Go struct?\""
    echo "  3. For interactive mode: zw ask -i"
    echo
    print_status "For more information, visit: https://github.com/derxanax/ZeroWorkflow"
}

# Run main function
main "$@"
