#!/bin/bash

# ZeroWorkflow Installation Script
# Downloads and installs the latest release from GitHub

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
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

print_header() {
    echo -e "${PURPLE}"
    echo "╔══════════════════════════════════════╗"
    echo "║        ZeroWorkflow Installer        ║"
    echo "║   AI-powered developer tools suite   ║"
    echo "╚══════════════════════════════════════╝"
    echo -e "${NC}"
}

# Detect OS and architecture
detect_platform() {
    local os=""
    local arch=""
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        CYGWIN*|MINGW*|MSYS*) os="windows";;
        *)          
            print_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64";;
        arm64|aarch64)  arch="arm64";;
        *)              
            print_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    echo "${os}-${arch}"
}

# Get latest release version from GitHub API
get_latest_version() {
    local repo="zeroworkflow/zw"
    local api_url="https://api.github.com/repos/${repo}/releases/latest"
    
    if command -v curl >/dev/null 2>&1; then
        curl -s "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- "$api_url" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
}

# Download and install binary
install_binary() {
    local platform="$1"
    local version="$2"
    local repo="zeroworkflow/zw"
    
    # For now, only support Linux AMD64
    if [[ "$platform" != "linux-amd64" ]]; then
        print_error "Currently only Linux AMD64 is supported"
        print_status "Supported platform: linux-amd64"
        print_status "Your platform: $platform"
        exit 1
    fi
    
    local binary_name="zw-${platform}"
    local download_url="https://github.com/${repo}/releases/download/${version}/${binary_name}"
    local checksum_url="https://github.com/${repo}/releases/download/${version}/${binary_name}.sha256"
    
    print_status "Downloading ZeroWorkflow ${version} for ${platform}..."
    
    # Create temporary directory
    local temp_dir=$(mktemp -d)
    cd "$temp_dir"
    
    # Download binary
    if command -v curl >/dev/null 2>&1; then
        if ! curl -L -o "$binary_name" "$download_url"; then
            print_error "Failed to download binary"
            exit 1
        fi
        
        # Download checksum
        if ! curl -L -o "${binary_name}.sha256" "$checksum_url"; then
            print_warning "Failed to download checksum, skipping verification"
        else
            print_status "Verifying checksum..."
            if command -v sha256sum >/dev/null 2>&1; then
                sha256sum -c "${binary_name}.sha256"
            else
                print_warning "sha256sum not found, skipping verification"
            fi
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -O "$binary_name" "$download_url"; then
            print_error "Failed to download binary"
            exit 1
        fi
        
        # Download and verify checksum
        if wget -O "${binary_name}.sha256" "$checksum_url" 2>/dev/null; then
            print_status "Verifying checksum..."
            if command -v sha256sum >/dev/null 2>&1; then
                sha256sum -c "${binary_name}.sha256"
            else
                print_warning "sha256sum not found, skipping verification"
            fi
        else
            print_warning "Failed to download checksum, skipping verification"
        fi
    fi
    
    # Make binary executable
    chmod +x "$binary_name"
    
    # Install binary
    local install_dir="/usr/local/bin"
    local final_name="zw"
    
    if [[ -w "$install_dir" ]]; then
        cp "$binary_name" "$install_dir/$final_name"
    else
        print_warning "Need sudo permissions to install to $install_dir"
        sudo cp "$binary_name" "$install_dir/$final_name"
    fi
    
    # Clean up
    cd - >/dev/null
    rm -rf "$temp_dir"
    
    print_success "ZeroWorkflow installed to $install_dir/$final_name"
}

# Setup .env file
setup_env() {
    print_status "Creating .env file template..."
    cat > .env << 'EOF'
# ZeroWorkflow Configuration
# Get your AI token from https://chat.z.ai
AI_TOKEN=your_ai_token_here
EOF
    print_warning "Please edit .env file and add your AI_TOKEN"
    print_status "Or set AI_TOKEN as environment variable: export AI_TOKEN=\"your_token\""
}

# Verify installation
verify_installation() {
    print_status "Verifying installation..."
    
    if command -v zw >/dev/null 2>&1; then
        local installed_version=$(zw --version 2>/dev/null || echo "unknown")
        print_success "ZeroWorkflow is now available globally!"
        print_status "Installed version: $installed_version"
        print_status "Try running: zw --help"
    else
        print_error "Installation verification failed"
        exit 1
    fi
}

# Main installation process
main() {
    print_header
    
    # Detect platform
    local platform=$(detect_platform)
    print_status "Detected platform: $platform"
    
    # Get latest version
    print_status "Fetching latest release information..."
    local version=$(get_latest_version)
    if [[ -z "$version" ]]; then
        print_error "Failed to get latest version"
        exit 1
    fi
    print_status "Latest version: $version"
    
    # Install binary
    install_binary "$platform" "$version"
    
    # Setup environment
    setup_env
    
    # Verify installation
    verify_installation
    
    echo
    print_success "Installation completed successfully!"
    echo
    print_status "🚀 Quick start:"
    echo "  1. Get your AI token from https://chat.z.ai"
    echo "  2. Set it: export AI_TOKEN=\"your_token_here\""
    echo "  3. Ask AI: zw ask \"How to create a Go struct?\""
    echo "  4. Include files: zw ask \"Review this code\" --file main.go"
    echo "  5. Interactive mode: zw ask -i"
    echo
    print_status "📚 Documentation: https://github.com/zeroworkflow/zw"
    print_status "🐛 Issues: https://github.com/zeroworkflow/zw/issues"
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "ZeroWorkflow Installer"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --version, -v  Show version and exit"
        echo ""
        echo "This script automatically downloads and installs the latest"
        echo "ZeroWorkflow release from GitHub for your platform."
        exit 0
        ;;
    --version|-v)
        echo "ZeroWorkflow Installer v1.0.0"
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
