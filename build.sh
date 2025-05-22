#!/bin/bash

# Build script for BYC CLI application

# Create build directory
mkdir -p build

# Build for Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o build/byc.exe ./cmd/byc

# Build for Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o build/byc ./cmd/byc

# Build for macOS
echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -o build/byc-mac ./cmd/byc

# Create release packages
echo "Creating release packages..."

# Windows package
mkdir -p release/windows
cp build/byc.exe release/windows/
cp README.md release/windows/
cp LICENSE release/windows/
cd release/windows
zip -r ../../byc-windows.zip *
cd ../..

# Linux package
mkdir -p release/linux
cp build/byc release/linux/
cp README.md release/linux/
cp LICENSE release/linux/
cd release/linux
tar -czf ../../byc-linux.tar.gz *
cd ../..

# macOS package
mkdir -p release/macos
cp build/byc-mac release/macos/
cp README.md release/macos/
cp LICENSE release/macos/
cd release/macos
tar -czf ../../byc-macos.tar.gz *
cd ../..

# Cleanup
rm -rf release

echo "Build complete! Release packages are in the root directory:"
echo "- byc-windows.zip"
echo "- byc-linux.tar.gz"
echo "- byc-macos.tar.gz" 