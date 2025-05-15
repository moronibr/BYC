#!/bin/bash

# Configuration
NODES=10
TRANSACTIONS_PER_NODE=1000
CONCURRENT_REQUESTS=50
TEST_DURATION=3600  # 1 hour
LOG_DIR="stress_test_logs"
METRICS_FILE="stress_test_metrics.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Create log directory
mkdir -p $LOG_DIR

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to generate random transaction
generate_transaction() {
    local from_node=$1
    local to_node=$2
    local amount=$((RANDOM % 1000 + 1))
    
    echo "{
        \"jsonrpc\": \"2.0\",
        \"method\": \"sendtransaction\",
        \"params\": {
            \"from\": \"node$from_node\",
            \"to\": \"node$to_node\",
            \"amount\": $amount,
            \"coin_type\": \"golden\"
        },
        \"id\": 1
    }"
}

# Function to send transaction to node
send_transaction() {
    local node=$1
    local tx=$2
    local response=$(curl -s -X POST -H "Content-Type: application/json" -d "$tx" http://localhost:$((8332 + node))/rpc)
    echo $response
}

# Function to collect metrics
collect_metrics() {
    local node=$1
    local metrics=$(curl -s http://localhost:$((8332 + node))/metrics)
    echo $metrics > "$LOG_DIR/node${node}_metrics.json"
}

# Function to analyze results
analyze_results() {
    local success_count=0
    local failure_count=0
    local total_latency=0
    local count=0

    for log in $LOG_DIR/node*_transaction.log; do
        while IFS= read -r line; do
            if [[ $line == *"success"* ]]; then
                success_count=$((success_count + 1))
                latency=$(echo $line | grep -oP 'latency: \K[0-9]+')
                total_latency=$((total_latency + latency))
                count=$((count + 1))
            elif [[ $line == *"error"* ]]; then
                failure_count=$((failure_count + 1))
            fi
        done < "$log"
    done

    local avg_latency=0
    if [ $count -gt 0 ]; then
        avg_latency=$((total_latency / count))
    fi

    echo "{
        \"total_transactions\": $((success_count + failure_count)),
        \"successful_transactions\": $success_count,
        \"failed_transactions\": $failure_count,
        \"success_rate\": $((success_count * 100 / (success_count + failure_count))),
        \"average_latency_ms\": $avg_latency
    }" > $METRICS_FILE
}

# Main stress test
print_status "Starting stress test with $NODES nodes"
print_status "Each node will send $TRANSACTIONS_PER_NODE transactions"
print_status "Test duration: $TEST_DURATION seconds"

# Start time
start_time=$(date +%s)
end_time=$((start_time + TEST_DURATION))

# Main test loop
while [ $(date +%s) -lt $end_time ]; do
    for node in $(seq 1 $NODES); do
        for i in $(seq 1 $CONCURRENT_REQUESTS); do
            to_node=$((RANDOM % NODES + 1))
            tx=$(generate_transaction $node $to_node)
            
            # Send transaction and log result
            start=$(date +%s%N)
            response=$(send_transaction $node "$tx")
            end=$(date +%s%N)
            latency=$((end - start))
            latency=$((latency / 1000000))  # Convert to milliseconds
            
            if [[ $response == *"error"* ]]; then
                echo "$(date '+%Y-%m-%d %H:%M:%S') - error: $response - latency: $latency ms" >> "$LOG_DIR/node${node}_transaction.log"
            else
                echo "$(date '+%Y-%m-%d %H:%M:%S') - success: $response - latency: $latency ms" >> "$LOG_DIR/node${node}_transaction.log"
            fi
            
            # Collect metrics every 100 transactions
            if [ $((i % 100)) -eq 0 ]; then
                collect_metrics $node
            fi
        done
    done
    
    # Sleep for a short time to prevent overwhelming the system
    sleep 1
done

# Analyze results
print_status "Test completed. Analyzing results..."
analyze_results

# Print summary
print_status "Stress test completed. Results saved to $METRICS_FILE"
print_status "Check $LOG_DIR for detailed logs"

# Cleanup
print_status "Cleaning up..."
rm -rf $LOG_DIR

print_status "Done!" 