#!/bin/bash

# Exit on error
set -e

# Configuration
NODE_COUNT=3
DATA_DIR="./data"
CONFIG_DIR="./config"
LOG_DIR="./logs"
BINARY_PATH="./bin/youngchain"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Print with color
print_color() {
    color=$1
    message=$2
    echo -e "${color}${message}${NC}"
}

# Check if binary exists
if [ ! -f "$BINARY_PATH" ]; then
    print_color "$RED" "Error: Binary not found at $BINARY_PATH"
    print_color "$YELLOW" "Building binary..."
    go build -o "$BINARY_PATH" ./cmd/youngchain
fi

# Create necessary directories
print_color "$YELLOW" "Creating directories..."
mkdir -p "$DATA_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$LOG_DIR"

# Generate node configurations
print_color "$YELLOW" "Generating node configurations..."
for i in $(seq 0 $((NODE_COUNT-1))); do
    node_dir="$DATA_DIR/node_$i"
    config_file="$CONFIG_DIR/node_$i.yaml"
    log_file="$LOG_DIR/node_$i.log"

    # Create node directory
    mkdir -p "$node_dir"

    # Generate configuration
    cat > "$config_file" << EOF
network:
  listen_port: $((8333+i))
  max_peers: 10
  bootstrap_nodes:
    - "127.0.0.1:8333"

mining:
  enabled: true
  address: "node_$i"
  target_bits: 24

consensus:
  target_bits: 24
  max_nonce: 1000000

storage:
  data_dir: "$node_dir"
EOF

    print_color "$GREEN" "Generated configuration for node $i"
done

# Start nodes
print_color "$YELLOW" "Starting nodes..."
for i in $(seq 0 $((NODE_COUNT-1))); do
    config_file="$CONFIG_DIR/node_$i.yaml"
    log_file="$LOG_DIR/node_$i.log"

    # Start node in background
    "$BINARY_PATH" --config "$config_file" > "$log_file" 2>&1 &
    pid=$!

    # Save PID
    echo $pid > "$DATA_DIR/node_$i.pid"

    print_color "$GREEN" "Started node $i (PID: $pid)"
done

# Wait for nodes to start
print_color "$YELLOW" "Waiting for nodes to start..."
sleep 5

# Check if nodes are running
print_color "$YELLOW" "Checking node status..."
for i in $(seq 0 $((NODE_COUNT-1))); do
    pid_file="$DATA_DIR/node_$i.pid"
    if [ -f "$pid_file" ]; then
        pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null; then
            print_color "$GREEN" "Node $i is running (PID: $pid)"
        else
            print_color "$RED" "Node $i is not running"
        fi
    else
        print_color "$RED" "PID file not found for node $i"
    fi
done

print_color "$GREEN" "Deployment completed successfully"

# Function to stop nodes
stop_nodes() {
    print_color "$YELLOW" "Stopping nodes..."
    for i in $(seq 0 $((NODE_COUNT-1))); do
        pid_file="$DATA_DIR/node_$i.pid"
        if [ -f "$pid_file" ]; then
            pid=$(cat "$pid_file")
            if ps -p $pid > /dev/null; then
                kill $pid
                print_color "$GREEN" "Stopped node $i (PID: $pid)"
            fi
        fi
    done
}

# Handle script termination
trap stop_nodes EXIT

# Keep script running
print_color "$YELLOW" "Press Ctrl+C to stop nodes"
while true; do
    sleep 1
done 