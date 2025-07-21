#!/bin/bash

# Build script for ContextKeeper Agent

set -e

# Build information
VERSION=${VERSION:-"dev"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

# Output directory
OUTPUT_DIR="bin"
mkdir -p ${OUTPUT_DIR}

# Platforms to build for
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

echo "Building ContextKeeper Agent v${VERSION}"
echo "Commit: ${COMMIT}"
echo "Date: ${DATE}"
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

echo ""
echo "Build complete! Binaries available in ${OUTPUT_DIR}/"
ls -la ${OUTPUT_DIR}/

# Create release archives
echo ""
echo "Creating release archives..."

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r os arch <<< "$platform"
    
    binary_name="contextkeeper-agent-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        binary_name="${binary_name}.exe"
    fi
    
    archive_name="contextkeeper-agent-${VERSION}-${os}-${arch}"
    if [ "$os" = "windows" ]; then
        # Create ZIP for Windows
        cd ${OUTPUT_DIR}
        zip "${archive_name}.zip" "${binary_name}"
        cd ..
        echo "✓ Created ${archive_name}.zip"
    else
        # Create tar.gz for Unix-like systems
        cd ${OUTPUT_DIR}
        tar -czf "${archive_name}.tar.gz" "${binary_name}"
        cd ..
        echo "✓ Created ${archive_name}.tar.gz"
    fi
done

echo ""
echo "Release archives created:"
ls -la ${OUTPUT_DIR}/*.{tar.gz,zip} 2>/dev/null || true