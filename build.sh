#!/bin/bash
# Build script for file-deduplicator
# Usage: ./build.sh [version]

VERSION=${1:-"3.0.0"}
DIST_DIR="dist"

echo "Building file-deduplicator v${VERSION}..."

# Clean dist directory
rm -rf ${DIST_DIR}
mkdir -p ${DIST_DIR}

# Build for each platform
echo "Building for darwin/amd64..."
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=${VERSION}" -o ${DIST_DIR}/file-deduplicator-darwin-amd64

echo "Building for darwin/arm64..."
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=${VERSION}" -o ${DIST_DIR}/file-deduplicator-darwin-arm64

echo "Building for linux/amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=${VERSION}" -o ${DIST_DIR}/file-deduplicator-linux-amd64

echo "Building for linux/arm64..."
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=${VERSION}" -o ${DIST_DIR}/file-deduplicator-linux-arm64

echo "Building for windows/amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=${VERSION}" -o ${DIST_DIR}/file-deduplicator-windows-amd64.exe

# Generate checksums
echo "Generating checksums..."
cd ${DIST_DIR}
sha256sum file-deduplicator-* > checksums.txt
cd ..

echo "Build complete!"
echo ""
echo "Binaries in ${DIST_DIR}/:"
ls -lh ${DIST_DIR}/
