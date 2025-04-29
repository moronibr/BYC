#!/bin/bash

# Installation script for BYC

# Detect OS
OS=$(uname -s)
ARCH=$(uname -m)

# Set version
VERSION="0.1.0"

# Determine binary name
if [ "$OS" = "Linux" ]; then
    BINARY_NAME="bycnode_linux_amd64_${VERSION}"
    MINER_NAME="bycminer_linux_amd64_${VERSION}"
elif [ "$OS" = "Darwin" ]; then
    BINARY_NAME="bycnode_darwin_amd64_${VERSION}"
    MINER_NAME="bycminer_darwin_amd64_${VERSION}"
else
    echo "Unsupported operating system: $OS"
    exit 1
fi

# Create installation directory
INSTALL_DIR="$HOME/.byc"
mkdir -p "$INSTALL_DIR/bin"

# Download binaries (replace with actual download URLs)
echo "Downloading BYC binaries..."
# Uncomment these lines when you have actual download URLs
# curl -L "https://github.com/yourusername/byc/releases/download/v${VERSION}/${BINARY_NAME}" -o "$INSTALL_DIR/bin/bycnode"
# curl -L "https://github.com/yourusername/byc/releases/download/v${VERSION}/${MINER_NAME}" -o "$INSTALL_DIR/bin/bycminer"

# Make binaries executable
chmod +x "$INSTALL_DIR/bin/bycnode"
chmod +x "$INSTALL_DIR/bin/bycminer"

# Create symlinks in /usr/local/bin
echo "Creating symlinks..."
sudo ln -sf "$INSTALL_DIR/bin/bycnode" /usr/local/bin/bycnode
sudo ln -sf "$INSTALL_DIR/bin/bycminer" /usr/local/bin/bycminer

echo "Installation complete!"
echo "You can now run BYC using:"
echo "  bycnode -type full -port 8333"
echo "  bycminer -node localhost:8333 -coin leah" 