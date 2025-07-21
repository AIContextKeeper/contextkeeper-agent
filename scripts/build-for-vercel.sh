#!/bin/bash

# Build script for Vercel deployment

set -e

VERSION=${VERSION:-"1.0.0"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# Output directory for Vercel public folder
OUTPUT_DIR="../ContextKeeper/public/downloads"
mkdir -p ${OUTPUT_DIR}

# Platforms to build for
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

echo "Building ContextKeeper Agent v${VERSION} for Vercel"
echo "Output directory: ${OUTPUT_DIR}"
echo ""

# Build for each platform
for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r os arch <<< "$platform"
    
    output_name="contextkeeper-agent-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    echo "Building for ${os}/${arch}..."
    
    GOOS=$os GOARCH=$arch go build \
        -ldflags="${LDFLAGS}" \
        -o "${OUTPUT_DIR}/${output_name}" \
        ./cmd/agent
    
    echo "✓ Built ${output_name}"
done

# Create checksums for security
echo ""
echo "Creating checksums..."
cd ${OUTPUT_DIR}
sha256sum contextkeeper-agent-* > checksums.txt
echo "✓ Created checksums.txt"

# Create latest symlinks for easy download URLs
echo ""
echo "Creating latest download links..."
ln -sf contextkeeper-agent-darwin-arm64 contextkeeper-agent-latest-macos-arm64
ln -sf contextkeeper-agent-darwin-amd64 contextkeeper-agent-latest-macos-amd64
ln -sf contextkeeper-agent-linux-amd64 contextkeeper-agent-latest-linux-amd64
ln -sf contextkeeper-agent-windows-amd64.exe contextkeeper-agent-latest-windows-amd64.exe

echo ""
echo "Build complete! Binaries available at:"
echo "https://contextkeeper.dev/downloads/contextkeeper-agent-latest-macos-arm64"
echo "https://contextkeeper.dev/downloads/contextkeeper-agent-latest-linux-amd64"
echo "https://contextkeeper.dev/downloads/contextkeeper-agent-latest-windows-amd64.exe"