#!/bin/bash

# Build script for BYC binaries
# This script builds the BYC node and miner for different operating systems

# Set version
VERSION="0.1.0"

# Create release directory
mkdir -p release

# Build for Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o release/bycnode_linux_amd64_${VERSION} ./cmd/youngchain
GOOS=linux GOARCH=amd64 go build -o release/bycminer_linux_amd64_${VERSION} ./cmd/bycminer

# Build for Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o release/bycnode_windows_amd64_${VERSION}.exe ./cmd/youngchain
GOOS=windows GOARCH=amd64 go build -o release/bycminer_windows_amd64_${VERSION}.exe ./cmd/bycminer

# Build for macOS
echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -o release/bycnode_darwin_amd64_${VERSION} ./cmd/youngchain
GOOS=darwin GOARCH=amd64 go build -o release/bycminer_darwin_amd64_${VERSION} ./cmd/bycminer

# Create zip archives
echo "Creating zip archives..."
cd release
zip bycnode_linux_amd64_${VERSION}.zip bycnode_linux_amd64_${VERSION}
zip bycminer_linux_amd64_${VERSION}.zip bycminer_linux_amd64_${VERSION}
zip bycnode_windows_amd64_${VERSION}.zip bycnode_windows_amd64_${VERSION}.exe
zip bycminer_windows_amd64_${VERSION}.zip bycminer_windows_amd64_${VERSION}.exe
zip bycnode_darwin_amd64_${VERSION}.zip bycnode_darwin_amd64_${VERSION}
zip bycminer_darwin_amd64_${VERSION}.zip bycminer_darwin_amd64_${VERSION}
cd ..

echo "Build complete! Binaries are in the release directory." 