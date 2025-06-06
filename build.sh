#!/bin/bash

# BYC Build Script

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print with color
print_color() {
    echo -e "${2}${1}${NC}"
}

# Build for a specific platform
build_for_platform() {
    local os=$1
    local arch=$2
    local output_dir="build/${os}_${arch}"
    
    print_color "Building for ${os}/${arch}..." "$GREEN"
    
    # Create output directory
    mkdir -p "$output_dir"
    
    # Set environment variables for cross-compilation
    export GOOS=$os
    export GOARCH=$arch
    
    # Build the binary
    go build -o "$output_dir/byc" ./cmd/byc-node
    
    if [ $? -ne 0 ]; then
        print_color "Error: Build failed for ${os}/${arch}" "$RED"
        return 1
    fi
    
    # Copy scripts and config
    cp install.sh "$output_dir/"
    cp scripts/sync.sh "$output_dir/"
    cp config/config.json "$output_dir/"
    
    # Create release package
    local package_name="byc_${os}_${arch}.tar.gz"
    tar -czf "build/$package_name" -C "$output_dir" .
    
    print_color "Created package: $package_name" "$GREEN"
}

# Main build process
main() {
    print_color "Starting BYC build..." "$GREEN"
    
    # Create build directory
    mkdir -p build
    
    # Build for different platforms
    build_for_platform "windows" "amd64"
    build_for_platform "linux" "amd64"
    build_for_platform "darwin" "amd64"
    
    print_color "Build completed!" "$GREEN"
    print_color "Packages are available in the build directory" "$YELLOW"
}

# Run the build
main 