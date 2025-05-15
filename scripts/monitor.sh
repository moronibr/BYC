#!/bin/bash

# Configuration
CONFIG_FILE="/data/config.yaml"
LOG_DIR="/data/logs"
METRICS_DIR="/data/metrics"
ALERT_THRESHOLDS_FILE="/data/alert_thresholds.json"
BACKUP_DIR="/backup"
RETENTION_DAYS=30

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "success") echo -e "${GREEN}[SUCCESS]${NC} $message" ;;
        "warning") echo -e "${YELLOW}[WARNING]${NC} $message" ;;
        "error") echo -e "${RED}[ERROR]${NC} $message" ;;
    esac
}

# Function to check node status
check_node_status() {
    local status=$(curl -s http://localhost:8332/status)
    if [ $? -eq 0 ]; then
        print_status "success" "Node is running"
        echo "$status" | jq .
    else
        print_status "error" "Node is not responding"
        return 1
    fi
}

# Function to collect metrics
collect_metrics() {
    local timestamp=$(date +%s)
    local metrics_file="$METRICS_DIR/metrics_$timestamp.json"
    
    # Collect system metrics
    local metrics=$(curl -s http://localhost:8332/metrics)
    if [ $? -eq 0 ]; then
        echo "$metrics" > "$metrics_file"
        print_status "success" "Metrics collected and saved to $metrics_file"
    else
        print_status "error" "Failed to collect metrics"
        return 1
    fi
}

# Function to check alert thresholds
check_alerts() {
    local metrics_file=$(ls -t $METRICS_DIR/metrics_*.json | head -n1)
    if [ -z "$metrics_file" ]; then
        print_status "error" "No metrics file found"
        return 1
    fi

    local metrics=$(cat "$metrics_file")
    local thresholds=$(cat "$ALERT_THRESHOLDS_FILE")

    # Check block time
    local block_time=$(echo "$metrics" | jq -r '.block_time')
    local block_time_threshold=$(echo "$thresholds" | jq -r '.block_time')
    if (( $(echo "$block_time > $block_time_threshold" | bc -l) )); then
        print_status "warning" "Block time ($block_time) exceeds threshold ($block_time_threshold)"
    fi

    # Check mempool size
    local mempool_size=$(echo "$metrics" | jq -r '.mempool_size')
    local mempool_threshold=$(echo "$thresholds" | jq -r '.mempool_size')
    if (( mempool_size > mempool_threshold )); then
        print_status "warning" "Mempool size ($mempool_size) exceeds threshold ($mempool_threshold)"
    fi

    # Check peer count
    local peer_count=$(echo "$metrics" | jq -r '.peer_count')
    local peer_threshold=$(echo "$thresholds" | jq -r '.peer_count')
    if (( peer_count < peer_threshold )); then
        print_status "warning" "Peer count ($peer_count) below threshold ($peer_threshold)"
    fi

    # Check resource usage
    local memory_usage=$(echo "$metrics" | jq -r '.memory_usage')
    local cpu_usage=$(echo "$metrics" | jq -r '.cpu_usage')
    local disk_usage=$(echo "$metrics" | jq -r '.disk_usage')
    
    local memory_threshold=$(echo "$thresholds" | jq -r '.memory_usage')
    local cpu_threshold=$(echo "$thresholds" | jq -r '.cpu_usage')
    local disk_threshold=$(echo "$thresholds" | jq -r '.disk_usage')

    if (( $(echo "$memory_usage > $memory_threshold" | bc -l) )); then
        print_status "warning" "Memory usage ($memory_usage%) exceeds threshold ($memory_threshold%)"
    fi

    if (( $(echo "$cpu_usage > $cpu_threshold" | bc -l) )); then
        print_status "warning" "CPU usage ($cpu_usage%) exceeds threshold ($cpu_threshold%)"
    fi

    if (( $(echo "$disk_usage > $disk_threshold" | bc -l) )); then
        print_status "warning" "Disk usage ($disk_usage%) exceeds threshold ($disk_threshold%)"
    fi
}

# Function to rotate logs
rotate_logs() {
    local log_file="$LOG_DIR/byc.log"
    if [ -f "$log_file" ]; then
        local size=$(stat -f%z "$log_file")
        local max_size=$((100 * 1024 * 1024)) # 100MB
        
        if (( size > max_size )); then
            mv "$log_file" "$log_file.$(date +%Y%m%d%H%M%S)"
            touch "$log_file"
            print_status "success" "Log file rotated"
        fi
    fi
}

# Function to clean up old files
cleanup_old_files() {
    # Clean up old metrics files
    find "$METRICS_DIR" -name "metrics_*.json" -mtime +$RETENTION_DAYS -delete
    print_status "success" "Cleaned up metrics files older than $RETENTION_DAYS days"

    # Clean up old log files
    find "$LOG_DIR" -name "byc.log.*" -mtime +$RETENTION_DAYS -delete
    print_status "success" "Cleaned up log files older than $RETENTION_DAYS days"
}

# Function to create backup
create_backup() {
    local timestamp=$(date +%Y%m%d%H%M%S)
    local backup_file="$BACKUP_DIR/byc_backup_$timestamp.tar.gz"
    
    tar -czf "$backup_file" /data
    if [ $? -eq 0 ]; then
        print_status "success" "Backup created: $backup_file"
    else
        print_status "error" "Failed to create backup"
        return 1
    fi
}

# Function to check disk space
check_disk_space() {
    local usage=$(df -h /data | awk 'NR==2 {print $5}' | sed 's/%//')
    local threshold=80
    
    if (( usage > threshold )); then
        print_status "warning" "Disk usage ($usage%) exceeds threshold ($threshold%)"
        return 1
    fi
}

# Function to check memory usage
check_memory_usage() {
    local usage=$(free | awk '/Mem:/ {print $3/$2 * 100.0}')
    local threshold=80
    
    if (( $(echo "$usage > $threshold" | bc -l) )); then
        print_status "warning" "Memory usage ($usage%) exceeds threshold ($threshold%)"
        return 1
    fi
}

# Function to check CPU usage
check_cpu_usage() {
    local usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}')
    local threshold=80
    
    if (( $(echo "$usage > $threshold" | bc -l) )); then
        print_status "warning" "CPU usage ($usage%) exceeds threshold ($threshold%)"
        return 1
    fi
}

# Function to check network connectivity
check_network() {
    local peers=$(curl -s http://localhost:8332/peers | jq '.peers | length')
    if (( peers < 5 )); then
        print_status "warning" "Low peer count: $peers"
        return 1
    fi
}

# Main monitoring loop
monitor_loop() {
    while true; do
        echo "=== $(date) ==="
        
        # Check node status
        check_node_status
        
        # Collect metrics
        collect_metrics
        
        # Check alerts
        check_alerts
        
        # Rotate logs if needed
        rotate_logs
        
        # Check system resources
        check_disk_space
        check_memory_usage
        check_cpu_usage
        
        # Check network
        check_network
        
        # Clean up old files
        cleanup_old_files
        
        # Create backup if needed
        if [ $(date +%H) -eq 0 ]; then
            create_backup
        fi
        
        sleep 300 # Sleep for 5 minutes
    done
}

# Parse command line arguments
case "$1" in
    "start")
        monitor_loop
        ;;
    "status")
        check_node_status
        ;;
    "metrics")
        collect_metrics
        ;;
    "alerts")
        check_alerts
        ;;
    "backup")
        create_backup
        ;;
    "cleanup")
        cleanup_old_files
        ;;
    *)
        echo "Usage: $0 {start|status|metrics|alerts|backup|cleanup}"
        exit 1
        ;;
esac 