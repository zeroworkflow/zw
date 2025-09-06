#!/bin/bash
set -e

echo "🚀 Installing ZeroWorkflow..."

# Configuration
BINARY_NAME="zw"
INSTALL_DIR="$HOME/.local/bin"
REPO_URL="https://github.com/zeroworkflow/zw"
CONFIG_DIR="$HOME/.config/zw"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
    *) echo -e "${RED}❌ Unsupported architecture: $ARCH${NC}"; exit 1 ;;
esac

echo -e "${BLUE}📋 System Info:${NC}"
echo -e "  OS: $OS"
echo -e "  Architecture: $ARCH"
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go 1.21+ first.${NC}"
    echo -e "${YELLOW}💡 Visit: https://golang.org/dl/${NC}"
    exit 1
fi

GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
echo -e "${GREEN}✅ Go $GO_VERSION detected${NC}"

# Create temporary directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

echo -e "${BLUE}📥 Cloning repository...${NC}"
git clone "$REPO_URL" zeroworkflow
cd zeroworkflow

echo -e "${BLUE}🔨 Building ZeroWorkflow...${NC}"
go mod tidy
go build -o "$BINARY_NAME" src/main.go

echo -e "${BLUE}📦 Installing binary...${NC}"
mkdir -p "$INSTALL_DIR"
cp "$BINARY_NAME" "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Create config directory and .env template
echo -e "${BLUE}⚙️  Setting up configuration...${NC}"
mkdir -p "$CONFIG_DIR"

if [ ! -f "$CONFIG_DIR/.env" ]; then
    cat > "$CONFIG_DIR/.env" << EOF
# ZeroWorkflow Configuration
# Get your token from: https://chat.z.ai
AI_TOKEN=your_token_here

# Optional: Custom API settings
# ZW_API_URL=https://chat.z.ai/api
# ZW_MODEL=0727-360B-API
# ZW_USER_AGENT=Mozilla/5.0 (X11; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0
EOF
    echo -e "${GREEN}✅ Created config template: $CONFIG_DIR/.env${NC}"
    echo -e "${YELLOW}⚠️  Please edit $CONFIG_DIR/.env and add your AI_TOKEN${NC}"
else
    echo -e "${GREEN}✅ Config file already exists: $CONFIG_DIR/.env${NC}"
fi

# Cleanup
cd /
rm -rf "$TMP_DIR"

echo
echo -e "${GREEN}🎉 ZeroWorkflow installed successfully!${NC}"
echo
echo -e "${BLUE}📚 Quick Start:${NC}"
echo -e "  ${YELLOW}zw ask \"How to create a Go struct?\"${NC}"
echo -e "  ${YELLOW}zw ask -i${NC}  # Interactive mode"
echo -e "  ${YELLOW}zw commit${NC}  # AI-powered commit messages"
echo
echo -e "${BLUE}🔧 Configuration:${NC}"
echo -e "  Edit: ${YELLOW}$CONFIG_DIR/.env${NC}"
echo -e "  Add your AI token from: ${YELLOW}https://chat.z.ai${NC}"
echo
echo -e "${BLUE}📖 Documentation:${NC}"
echo -e "  ${YELLOW}$REPO_URL${NC}"
echo