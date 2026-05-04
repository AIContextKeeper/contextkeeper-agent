#!/bin/bash

# Installation script for ContextKeeper Agent

set -e

REPO="AIContextKeeper/contextkeeper-agent"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.contextkeeper"

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

# Get latest release URL
LATEST_URL="https://api.github.com/repos/$REPO/releases/latest"
DOWNLOAD_URL=$(curl -s "$LATEST_URL" | grep "browser_download_url.*$OS-$ARCH" | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find download URL for $OS-$ARCH"
    exit 1
fi

echo "Download URL: $DOWNLOAD_URL"

# Download and extract binary
TEMP_ARCHIVE="/tmp/contextkeeper-agent.tar.gz"
TEMP_DIR="/tmp/contextkeeper-agent-extract"

curl -L -o "$TEMP_ARCHIVE" "$DOWNLOAD_URL"

mkdir -p "$TEMP_DIR"
tar -xzf "$TEMP_ARCHIVE" -C "$TEMP_DIR"

# Find the binary (may be nested in a subdirectory)
BINARY=$(find "$TEMP_DIR" -type f -name "contextkeeper-agent" | head -1)
if [ -z "$BINARY" ]; then
    echo "Error: Could not find binary in archive"
    exit 1
fi

chmod +x "$BINARY"
mv "$BINARY" "$INSTALL_DIR/contextkeeper-agent"

rm -rf "$TEMP_ARCHIVE" "$TEMP_DIR"

echo "✓ Installed contextkeeper-agent to $INSTALL_DIR"

# Create config directory
mkdir -p "$CONFIG_DIR"
echo "✓ Created config directory: $CONFIG_DIR"

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
echo "Installation complete!"
echo ""
echo "Usage:"
echo "  contextkeeper-agent              # Start the agent"
echo "  contextkeeper-agent --usage      # Show usage information"
echo "  contextkeeper-agent --version    # Show version"
echo "  contextkeeper-agent --help       # Show help"
echo ""
echo "The agent will start monitoring AI tool outputs automatically."
echo "Visit https://contextkeeper.dev to view your sessions in the web dashboard."