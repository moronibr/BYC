#!/bin/bash

# Generate Recent Blocks Script

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print with color
print_color() {
    echo -e "${2}${1}${NC}"
}

# Create sample blocks
generate_blocks() {
    print_color "Generating sample blocks..." "$GREEN"
    
    # Create blocks directory if it doesn't exist
    mkdir -p data/blocks/recent
    
    # Generate 1000 sample blocks
    for i in $(seq 1 1000); do
        # Generate a random hash
        hash=$(openssl rand -hex 32)
        prev_hash=$(openssl rand -hex 32)
        
        # Create block file
        cat > "data/blocks/recent/block_$i.json" << EOF
{
    "hash": "$hash",
    "timestamp": $((1231006505 + i * 600)),
    "transactions": [
        {
            "id": "$(openssl rand -hex 32)",
            "inputs": [],
            "outputs": [
                {
                    "value": 50.0,
                    "coin_type": "LEAH",
                    "public_key_hash": "$(openssl rand -hex 20)",
                    "address": "Miner$i"
                }
            ],
            "timestamp": "$(date -d "@$((1231006505 + i * 600))" -Iseconds)Z",
            "block_type": "GOLDEN"
        }
    ],
    "nonce": $i,
    "prev_hash": "$prev_hash",
    "block_type": "GOLDEN",
    "difficulty": 4
}
EOF
    done
    
    print_color "Generated 1000 sample blocks" "$GREEN"
}

# Create archive
create_archive() {
    print_color "Creating blocks archive..." "$GREEN"
    
    # Create tar.gz archive
    tar -czf data/blocks/recent_blocks.tar.gz -C data/blocks recent
    
    print_color "Created archive: data/blocks/recent_blocks.tar.gz" "$GREEN"
}

# Cleanup
cleanup() {
    print_color "Cleaning up..." "$GREEN"
    rm -rf data/blocks/recent
}

# Main process
main() {
    print_color "Starting block generation..." "$GREEN"
    
    # Generate blocks
    generate_blocks
    
    # Create archive
    create_archive
    
    # Cleanup
    cleanup
    
    print_color "Block generation completed!" "$GREEN"
}

# Run the script
main 