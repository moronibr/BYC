#!/bin/bash

# BYC Installation Script

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print with color
print_color() {
    echo -e "${2}${1}${NC}"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_color "Error: Go is not installed. Please install Go first." "$RED"
        print_color "Visit: https://golang.org/doc/install" "$YELLOW"
        exit 1
    fi
}

# Check if Make is installed
check_make() {
    if ! command -v make &> /dev/null; then
        print_color "Error: Make is not installed. Please install Make first." "$RED"
        print_color "On Ubuntu/Debian: sudo apt-get install make" "$YELLOW"
        print_color "On macOS: xcode-select --install" "$YELLOW"
        exit 1
    fi
}

# Main installation process
main() {
    print_color "Starting BYC installation..." "$GREEN"
    
    # Check prerequisites
    check_go
    check_make
    
    # Build the project
    print_color "Building BYC..." "$GREEN"
    make build
    
    if [ $? -ne 0 ]; then
        print_color "Error: Build failed" "$RED"
        exit 1
    fi
    
    # Install the binary
    print_color "Installing BYC..." "$GREEN"
    make install
    
    if [ $? -ne 0 ]; then
        print_color "Error: Installation failed" "$RED"
        exit 1
    fi
    
    print_color "BYC has been successfully installed!" "$GREEN"
    print_color "You can now run 'byc' from anywhere" "$GREEN"
    print_color "Try 'byc --help' to see available commands" "$YELLOW"
}

# Run the installation
main 