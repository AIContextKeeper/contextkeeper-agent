#!/bin/bash

# Enhanced installation script with analytics

REPO="AIContextKeeper/contextkeeper-agent"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.contextkeeper"

# Analytics endpoint
ANALYTICS_URL="https://contextkeeper.dev/api/analytics/install"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "ContextKeeper Agent Installer"
echo "============================="
echo "OS: $OS"
echo "Architecture: $ARCH"
echo ""

# Send install analytics (non-blocking)
send_analytics() {
    curl -s -X POST "$ANALYTICS_URL" \
        -H "Content-Type: application/json" \
        -d "{\"event\":\"install_started\",\"os\":\"$OS\",\"arch\":\"$ARCH\",\"timestamp\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}" \
        >/dev/null 2>&1 &
}

# Send analytics
send_analytics

# Check if running as root for system-wide install
if [ "$EUID" -eq 0 ]; then
    echo "Installing system-wide to $INSTALL_DIR"
else
    INSTALL_DIR="$HOME/.local/bin"
    echo "Installing to user directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
fi

# Download latest release
echo "Downloading latest release..."

LATEST_URL="https://api.github.com/repos/$REPO/releases/latest"
DOWNLOAD_URL=$(curl -s "$LATEST_URL" | grep "browser_download_url.*$OS-$ARCH" | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find download URL for $OS-$ARCH"
    exit 1
fi

echo "Download URL: $DOWNLOAD_URL"

# Download and extract
TEMP_DIR="/tmp/contextkeeper-install"
mkdir -p "$TEMP_DIR"

if [[ "$DOWNLOAD_URL" == *.tar.gz ]]; then
    curl -L -o "$TEMP_DIR/agent.tar.gz" "$DOWNLOAD_URL"
    tar -xzf "$TEMP_DIR/agent.tar.gz" -C "$TEMP_DIR"
    BINARY_NAME="contextkeeper-agent-$OS-$ARCH"
else
    curl -L -o "$TEMP_DIR/agent.zip" "$DOWNLOAD_URL"
    unzip -q "$TEMP_DIR/agent.zip" -d "$TEMP_DIR"
    BINARY_NAME="contextkeeper-agent-$OS-$ARCH.exe"
fi

# Install binary
chmod +x "$TEMP_DIR/$BINARY_NAME"
mv "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/contextkeeper-agent"

echo "✅ Installed contextkeeper-agent to $INSTALL_DIR"

# Create config directory
mkdir -p "$CONFIG_DIR"
echo "✅ Created config directory: $CONFIG_DIR"

# Cleanup
rm -rf "$TEMP_DIR"

# Send success analytics
curl -s -X POST "$ANALYTICS_URL" \
    -H "Content-Type: application/json" \
    -d "{\"event\":\"install_completed\",\"os\":\"$OS\",\"arch\":\"$ARCH\",\"timestamp\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}" \
    >/dev/null 2>&1 &

# Add to PATH if not already there
case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *) 
        if [ "$INSTALL_DIR" = "$HOME/.local/bin" ]; then
            echo ""
            echo "Add $INSTALL_DIR to your PATH by adding this line to your shell profile:"
            echo "export PATH=\"$INSTALL_DIR:\$PATH\""
        fi
        ;;
esac

echo ""
echo "🎉 Installation complete!"
echo ""
echo "Usage:"
echo "  contextkeeper-agent              # Start the agent"
echo "  contextkeeper-agent --usage      # Show usage information"
echo "  contextkeeper-agent --dashboard  # Open dashboard"
echo ""
echo "The agent will start monitoring AI tool outputs automatically."
echo "Visit https://contextkeeper.dev/app to view your sessions."
echo ""
echo "💡 Want unlimited sessions? Upgrade to Pro at https://contextkeeper.dev/pricing"