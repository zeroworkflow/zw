#!/bin/bash
set -e

echo "🚀 Installing ZeroWorkflow..."

# Configuration
BINARY_NAME="zw"
INSTALL_DIR="$HOME/.local/bin"
REPO_OWNER="zeroworkflow"
REPO_NAME="zw"
CONFIG_DIR="$HOME/.config/zw"

# GitHub API token for higher rate limits (optional)
if [ -n "$GITHUB_TOKEN" ]; then
    GITHUB_AUTH="-H \"Authorization: token $GITHUB_TOKEN\""
else
    GITHUB_AUTH=""
fi

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

# Get latest release info
echo -e "${BLUE}🔍 Finding latest release...${NC}"
API_RESPONSE=$(curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest")
HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest")

# Check for GitHub API errors
if [ "$HTTP_CODE" = "403" ]; then
    if echo "$API_RESPONSE" | grep -q "rate limit"; then
        echo -e "${RED}❌ GitHub API rate limit exceeded${NC}"
        echo -e "${YELLOW}💡 Too many requests. Please try again later or use a GitHub token${NC}"
        echo -e "${BLUE}ℹ️  Set GITHUB_TOKEN environment variable to increase rate limit${NC}"
    else
        echo -e "${RED}❌ GitHub API access forbidden (403)${NC}"
    fi
    exit 1
elif [ "$HTTP_CODE" = "404" ]; then
    echo -e "${RED}❌ Repository not found (404)${NC}"
    exit 1
elif [ "$HTTP_CODE" != "200" ]; then
    echo -e "${RED}❌ GitHub API error (HTTP $HTTP_CODE)${NC}"
    exit 1
fi

LATEST_TAG=$(echo "$API_RESPONSE" | grep -Po '"tag_name": "\K.*?(?=")')

if [ -z "$LATEST_TAG" ]; then
    echo -e "${RED}❌ Failed to parse release tag from API response${NC}"
    echo -e "${YELLOW}💡 API Response: $API_RESPONSE${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Found latest version: $LATEST_TAG${NC}"

# Construct download URL
BINARY_FILE="$BINARY_NAME-$OS-$ARCH"
DOWNLOAD_URL="https://github.com/$REPO_OWNER/$REPO_NAME/releases/download/$LATEST_TAG/$BINARY_FILE"

echo -e "${BLUE}📥 Downloading binary...${NC}"
echo -e "  URL: $DOWNLOAD_URL"

# Create temporary directory
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

# Download binary with better error handling
echo -e "${BLUE}⬇️  Downloading $BINARY_FILE...${NC}"
HTTP_CODE=$(curl -L -w "%{http_code}" -o "$BINARY_NAME" "$DOWNLOAD_URL")

if [ "$HTTP_CODE" = "404" ]; then
    echo -e "${RED}❌ Binary not found (404)${NC}"
    echo -e "${YELLOW}💡 Available binaries for $LATEST_TAG:${NC}"
    curl -s "https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/tags/$LATEST_TAG" | grep -Po '"name": "\K[^"]*(?=")' | grep "^$BINARY_NAME" || echo "  No matching binaries found"
    exit 1
elif [ "$HTTP_CODE" = "403" ]; then
    echo -e "${RED}❌ Download forbidden - too many requests${NC}"
    echo -e "${YELLOW}💡 GitHub rate limit exceeded. Try again later${NC}"
    exit 1
elif [ "$HTTP_CODE" != "200" ]; then
    echo -e "${RED}❌ Download failed (HTTP $HTTP_CODE)${NC}"
    exit 1
fi

# Verify downloaded file
if [ ! -f "$BINARY_NAME" ] || [ ! -s "$BINARY_NAME" ]; then
    echo -e "${RED}❌ Downloaded file is empty or missing${NC}"
    exit 1
fi

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

echo -e "${GREEN}✅ Binary installed to: $INSTALL_DIR/$BINARY_NAME${NC}"

# Check if ~/.local/bin is in PATH
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    echo -e "${YELLOW}⚠️  $HOME/.local/bin is not in your PATH${NC}"
    echo -e "${BLUE}💡 Add this to your shell profile (~/.bashrc or ~/.zshrc):${NC}"
    echo -e "  ${YELLOW}export PATH=\"\$HOME/.local/bin:\$PATH\"${NC}"
    echo
fi

echo
echo -e "${GREEN}🎉 ZeroWorkflow $LATEST_TAG installed successfully!${NC}"
echo
echo -e "${BLUE}📚 Quick Start:${NC}"
echo -e "  ${YELLOW}zw ask \"How to create a Go struct?\"${NC}"
echo -e "  ${YELLOW}zw ask -i${NC}  # Interactive mode"
echo -e "  ${YELLOW}zw commit${NC}  # AI-powered commit messages"
echo
echo -e "${BLUE}🔧 Configuration:${NC}"
echo -e "  Edit: ${YELLOW}$CONFIG_DIR/.env${NC}"
echo -e "  Free tokens : github.com/zeroworkflow/zw-keys"
echo
echo -e "${BLUE}📖 Documentation:${NC}"
echo -e "  ${YELLOW}https://github.com/$REPO_OWNER/$REPO_NAME${NC}"
echo