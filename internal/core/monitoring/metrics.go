package monitoring

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MetricsCollector manages system metrics
type MetricsCollector struct {
	mu sync.RWMutex

	// System metrics
	startTime     time.Time
	uptime        time.Duration
	blockHeight   uint64
	txCount       uint64
	peerCount     int
	activePeers   int
	memoryUsage   uint64
	cpuUsage      float64
	networkIO     NetworkMetrics
	storageUsage  StorageMetrics
	errorCount    map[string]uint64
	lastBlockTime time.Time
}

// NetworkMetrics tracks network-related metrics
type NetworkMetrics struct {
	BytesReceived uint64
	BytesSent     uint64
	Connections   int
	Latency       time.Duration
}

// StorageMetrics tracks storage-related metrics
type StorageMetrics struct {
	TotalSize        uint64
	UsedSize         uint64
	FreeSize         uint64
	BlockCount       uint64
	TransactionCount uint64
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime:  time.Now(),
		errorCount: make(map[string]uint64),
	}
}

// UpdateMetrics updates the metrics with current values
func (mc *MetricsCollector) UpdateMetrics(metrics map[string]interface{}) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.uptime = time.Since(mc.startTime)

	if v, ok := metrics["blockHeight"].(uint64); ok {
		mc.blockHeight = v
	}
	if v, ok := metrics["txCount"].(uint64); ok {
		mc.txCount = v
	}
	if v, ok := metrics["peerCount"].(int); ok {
		mc.peerCount = v
	}
	if v, ok := metrics["activePeers"].(int); ok {
		mc.activePeers = v
	}
	if v, ok := metrics["memoryUsage"].(uint64); ok {
		mc.memoryUsage = v
	}
	if v, ok := metrics["cpuUsage"].(float64); ok {
		mc.cpuUsage = v
	}
	if v, ok := metrics["networkIO"].(NetworkMetrics); ok {
		mc.networkIO = v
	}
	if v, ok := metrics["storageUsage"].(StorageMetrics); ok {
		mc.storageUsage = v
	}
	if v, ok := metrics["lastBlockTime"].(time.Time); ok {
		mc.lastBlockTime = v
	}
}

// IncrementError increments the error count for a specific error type
func (mc *MetricsCollector) IncrementError(errorType string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.errorCount[errorType]++
}

// GetMetrics returns the current metrics as a JSON string
func (mc *MetricsCollector) GetMetrics() (string, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics := map[string]interface{}{
		"uptime":        mc.uptime.String(),
		"blockHeight":   mc.blockHeight,
		"txCount":       mc.txCount,
		"peerCount":     mc.peerCount,
		"activePeers":   mc.activePeers,
		"memoryUsage":   mc.memoryUsage,
		"cpuUsage":      mc.cpuUsage,
		"networkIO":     mc.networkIO,
		"storageUsage":  mc.storageUsage,
		"errorCount":    mc.errorCount,
		"lastBlockTime": mc.lastBlockTime,
	}

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal metrics: %v", err)
	}

	return string(data), nil
}

// GetStatus returns the current system status
func (mc *MetricsCollector) GetStatus() (string, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	status := map[string]interface{}{
		"status":        "running",
		"uptime":        mc.uptime.String(),
		"blockHeight":   mc.blockHeight,
		"peerCount":     mc.peerCount,
		"activePeers":   mc.activePeers,
		"lastBlockTime": mc.lastBlockTime,
	}

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal status: %v", err)
	}

	return string(data), nil
}

// GetPeers returns information about connected peers
func (mc *MetricsCollector) GetPeers() (string, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	peers := map[string]interface{}{
		"totalPeers":    mc.peerCount,
		"activePeers":   mc.activePeers,
		"networkIO":     mc.networkIO,
		"lastBlockTime": mc.lastBlockTime,
	}

	data, err := json.MarshalIndent(peers, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal peers: %v", err)
	}

	return string(data), nil
}
