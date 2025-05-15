#!/bin/bash

# Exit on error
set -e

# Configuration
DATA_DIR="./data"
LOG_DIR="./logs"
METRICS_FILE="./metrics.json"
ALERT_THRESHOLDS_FILE="./alert_thresholds.json"

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

# Load alert thresholds
if [ -f "$ALERT_THRESHOLDS_FILE" ]; then
    source "$ALERT_THRESHOLDS_FILE"
else
    # Default thresholds
    BLOCK_TIME_THRESHOLD=30
    MEMPOOL_SIZE_THRESHOLD=10000
    PEER_COUNT_THRESHOLD=5
    MEMORY_USAGE_THRESHOLD=80
    CPU_USAGE_THRESHOLD=90
    DISK_USAGE_THRESHOLD=90
fi

# Function to get node metrics
get_node_metrics() {
    node_id=$1
    metrics_file="$DATA_DIR/node_$node_id/metrics.json"
    
    if [ -f "$metrics_file" ]; then
        cat "$metrics_file"
    else
        echo "{}"
    fi
}

# Function to check alerts
check_alerts() {
    node_id=$1
    metrics=$(get_node_metrics $node_id)
    
    # Block time alert
    block_time=$(echo "$metrics" | jq -r '.block_time')
    if [ "$block_time" != "null" ] && [ "$block_time" -gt "$BLOCK_TIME_THRESHOLD" ]; then
        print_color "$RED" "Alert: Node $node_id - Block time is high: ${block_time}s"
    fi
    
    # Mempool size alert
    mempool_size=$(echo "$metrics" | jq -r '.mempool_size')
    if [ "$mempool_size" != "null" ] && [ "$mempool_size" -gt "$MEMPOOL_SIZE_THRESHOLD" ]; then
        print_color "$RED" "Alert: Node $node_id - Mempool size is high: ${mempool_size} transactions"
    fi
    
    # Peer count alert
    peer_count=$(echo "$metrics" | jq -r '.peer_count')
    if [ "$peer_count" != "null" ] && [ "$peer_count" -lt "$PEER_COUNT_THRESHOLD" ]; then
        print_color "$RED" "Alert: Node $node_id - Low number of peers: ${peer_count}"
    fi
    
    # Memory usage alert
    memory_usage=$(echo "$metrics" | jq -r '.memory_usage')
    if [ "$memory_usage" != "null" ] && [ "$memory_usage" -gt "$MEMORY_USAGE_THRESHOLD" ]; then
        print_color "$RED" "Alert: Node $node_id - High memory usage: ${memory_usage}%"
    fi
    
    # CPU usage alert
    cpu_usage=$(echo "$metrics" | jq -r '.cpu_usage')
    if [ "$cpu_usage" != "null" ] && [ "$cpu_usage" -gt "$CPU_USAGE_THRESHOLD" ]; then
        print_color "$RED" "Alert: Node $node_id - High CPU usage: ${cpu_usage}%"
    fi
    
    # Disk usage alert
    disk_usage=$(echo "$metrics" | jq -r '.disk_usage')
    if [ "$disk_usage" != "null" ] && [ "$disk_usage" -gt "$DISK_USAGE_THRESHOLD" ]; then
        print_color "$RED" "Alert: Node $node_id - High disk usage: ${disk_usage}%"
    fi
}

# Function to display node status
display_node_status() {
    node_id=$1
    pid_file="$DATA_DIR/node_$node_id.pid"
    
    if [ -f "$pid_file" ]; then
        pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null; then
            print_color "$GREEN" "Node $node_id is running (PID: $pid)"
            
            # Get and display metrics
            metrics=$(get_node_metrics $node_id)
            
            # Block height
            block_height=$(echo "$metrics" | jq -r '.block_height')
            if [ "$block_height" != "null" ]; then
                echo "  Block height: $block_height"
            fi
            
            # Block time
            block_time=$(echo "$metrics" | jq -r '.block_time')
            if [ "$block_time" != "null" ]; then
                echo "  Block time: ${block_time}s"
            fi
            
            # Transaction count
            tx_count=$(echo "$metrics" | jq -r '.tx_count')
            if [ "$tx_count" != "null" ]; then
                echo "  Transaction count: $tx_count"
            fi
            
            # Mempool size
            mempool_size=$(echo "$metrics" | jq -r '.mempool_size')
            if [ "$mempool_size" != "null" ]; then
                echo "  Mempool size: $mempool_size"
            fi
            
            # Peer count
            peer_count=$(echo "$metrics" | jq -r '.peer_count')
            if [ "$peer_count" != "null" ]; then
                echo "  Peer count: $peer_count"
            fi
            
            # Hash rate
            hash_rate=$(echo "$metrics" | jq -r '.hash_rate')
            if [ "$hash_rate" != "null" ]; then
                echo "  Hash rate: ${hash_rate} H/s"
            fi
            
            # Difficulty
            difficulty=$(echo "$metrics" | jq -r '.difficulty')
            if [ "$difficulty" != "null" ]; then
                echo "  Difficulty: $difficulty"
            fi
            
            # System metrics
            memory_usage=$(echo "$metrics" | jq -r '.memory_usage')
            if [ "$memory_usage" != "null" ]; then
                echo "  Memory usage: ${memory_usage}%"
            fi
            
            cpu_usage=$(echo "$metrics" | jq -r '.cpu_usage')
            if [ "$cpu_usage" != "null" ]; then
                echo "  CPU usage: ${cpu_usage}%"
            fi
            
            disk_usage=$(echo "$metrics" | jq -r '.disk_usage')
            if [ "$disk_usage" != "null" ]; then
                echo "  Disk usage: ${disk_usage}%"
            fi
            
            # Check for alerts
            check_alerts $node_id
        else
            print_color "$RED" "Node $node_id is not running"
        fi
    else
        print_color "$RED" "PID file not found for node $node_id"
    fi
}

# Main monitoring loop
print_color "$YELLOW" "Starting monitoring..."
while true; do
    clear
    print_color "$GREEN" "=== Blockchain Node Monitoring ==="
    echo
    
    # Get list of nodes
    for node_dir in "$DATA_DIR"/node_*; do
        if [ -d "$node_dir" ]; then
            node_id=$(basename "$node_dir" | sed 's/node_//')
            display_node_status $node_id
            echo
        fi
    done
    
    # Wait before next update
    sleep 5
done 