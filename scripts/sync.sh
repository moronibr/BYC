#!/bin/bash

# BYC Chain Sync Script

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print with color
print_color() {
    echo -e "${2}${1}${NC}"
}

# Check if byc is installed
check_byc() {
    if ! command -v byc &> /dev/null; then
        print_color "Error: BYC is not installed. Please run install.sh first." "$RED"
        exit 1
    fi
}

# Create necessary directories
setup_directories() {
    print_color "Setting up directories..." "$GREEN"
    mkdir -p ~/.byc/{blocks,transactions,wallets}
}

# Download initial chain data
download_chain() {
    print_color "Downloading initial chain data..." "$GREEN"
    
    # Create a temporary directory for downloads
    TEMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TEMP_DIR"' EXIT
    
    # Download genesis blocks
    print_color "Downloading genesis blocks..." "$GREEN"
    curl -L -o "$TEMP_DIR/golden_genesis.json" "https://raw.githubusercontent.com/moroni/BYC/main/data/genesis/golden_genesis.json"
    curl -L -o "$TEMP_DIR/silver_genesis.json" "https://raw.githubusercontent.com/moroni/BYC/main/data/genesis/silver_genesis.json"
    
    # Verify genesis blocks
    if [ ! -f "$TEMP_DIR/golden_genesis.json" ] || [ ! -f "$TEMP_DIR/silver_genesis.json" ]; then
        print_color "Error: Failed to download genesis blocks" "$RED"
        exit 1
    fi
    
    # Copy genesis blocks to the blockchain directory
    cp "$TEMP_DIR/golden_genesis.json" ~/.byc/blocks/
    cp "$TEMP_DIR/silver_genesis.json" ~/.byc/blocks/
    
    # Download initial UTXO set
    print_color "Downloading initial UTXO set..." "$GREEN"
    curl -L -o "$TEMP_DIR/utxo_set.json" "https://raw.githubusercontent.com/moroni/BYC/main/data/utxo/utxo_set.json"
    
    if [ ! -f "$TEMP_DIR/utxo_set.json" ]; then
        print_color "Error: Failed to download UTXO set" "$RED"
        exit 1
    fi
    
    # Copy UTXO set to the blockchain directory
    cp "$TEMP_DIR/utxo_set.json" ~/.byc/blocks/
    
    # Download recent blocks (last 1000 blocks)
    print_color "Downloading recent blocks..." "$GREEN"
    curl -L -o "$TEMP_DIR/recent_blocks.tar.gz" "https://raw.githubusercontent.com/moroni/BYC/main/data/blocks/recent_blocks.tar.gz"
    
    if [ ! -f "$TEMP_DIR/recent_blocks.tar.gz" ]; then
        print_color "Error: Failed to download recent blocks" "$RED"
        exit 1
    fi
    
    # Extract recent blocks
    tar -xzf "$TEMP_DIR/recent_blocks.tar.gz" -C ~/.byc/blocks/
    
    print_color "Chain data downloaded successfully!" "$GREEN"
}

# Configure the node
configure_node() {
    print_color "Configuring node..." "$GREEN"
    cat > ~/.byc/config.json << EOF
{
  "api": {
    "address": "localhost:8000",
    "cors": {
      "allowed_origins": [
        "http://localhost:8000",
        "http://127.0.0.1:8000"
      ]
    },
    "rate_limit": {
      "requests_per_second": 100,
      "burst": 1000
    },
    "tls": {
      "enabled": true,
      "cert_file": "cert.pem",
      "key_file": "key.pem"
    }
  },
  "p2p": {
    "address": "localhost:3000",
    "bootstrap_peers": [
      "node1.byc.network:3000",
      "node2.byc.network:3000"
    ],
    "max_peers": 100,
    "ping_interval": 30000000000,
    "ping_timeout": 10000000000
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "stdout"
  },
  "blockchain": {
    "block_type": "golden",
    "difficulty": 4,
    "max_block_size": 1048576,
    "mining_reward": 50
  },
  "mining": {
    "enabled": true,
    "coin_type": "LEAH",
    "auto_start": true,
    "max_threads": 4,
    "target_blocks_per_minute": 1
  }
}
EOF
}

# Start the node
start_node() {
    print_color "Starting BYC node..." "$GREEN"
    byc node start --config ~/.byc/config.json
}

# Main process
main() {
    print_color "Starting BYC chain sync..." "$GREEN"
    
    # Check prerequisites
    check_byc
    
    # Setup
    setup_directories
    configure_node
    
    # Download and sync
    download_chain
    
    # Start node
    start_node
}

# Run the sync
main 