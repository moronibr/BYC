package types

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// DefaultMetricsWindow is the default window size for metrics in seconds
	DefaultMetricsWindow = 300 // 5 minutes
	// DefaultMetricsInterval is the default interval for metrics collection in seconds
	DefaultMetricsInterval = 60 // 1 minute
)

// UTXOMonitoring handles monitoring of the UTXO set
type UTXOMonitoring struct {
	utxoSet *UTXOSet
	mu      sync.RWMutex

	// Metrics
	totalUTXOs         int64
	spentUTXOs         int64
	unspentUTXOs       int64
	totalValue         int64
	lastUpdateTime     time.Time
	operationCount     int64
	errorCount         int64
	compressionRatio   float64
	indexSize          int64
	storageSize        int64
	operationLatency   time.Duration
	lastPruneTime      time.Time
	lastBackupTime     time.Time
	lastValidationTime time.Time
}

// NewUTXOMonitoring creates a new UTXO monitoring handler
func NewUTXOMonitoring(utxoSet *UTXOSet) *UTXOMonitoring {
	return &UTXOMonitoring{
		utxoSet:          utxoSet,
		lastUpdateTime:   time.Now(),
		operationLatency: 0,
	}
}

// UpdateMetrics updates the monitoring metrics
func (um *UTXOMonitoring) UpdateMetrics() {
	um.mu.Lock()
	defer um.mu.Unlock()

	// Get all UTXOs
	utxos := um.utxoSet.GetAllUTXOs()

	// Reset counters
	atomic.StoreInt64(&um.totalUTXOs, 0)
	atomic.StoreInt64(&um.spentUTXOs, 0)
	atomic.StoreInt64(&um.unspentUTXOs, 0)
	atomic.StoreInt64(&um.totalValue, 0)

	// Update metrics
	for _, utxo := range utxos {
		atomic.AddInt64(&um.totalUTXOs, 1)
		atomic.AddInt64(&um.totalValue, utxo.Value)
		if utxo.Spent {
			atomic.AddInt64(&um.spentUTXOs, 1)
		} else {
			atomic.AddInt64(&um.unspentUTXOs, 1)
		}
	}

	// Update index size
	atomic.StoreInt64(&um.indexSize, int64(um.utxoSet.index.Size()))

	// Update storage size
	atomic.StoreInt64(&um.storageSize, int64(len(um.utxoSet.Serialize())))

	// Update last update time
	um.lastUpdateTime = time.Now()
}

// RecordOperation records an operation for monitoring
func (um *UTXOMonitoring) RecordOperation(operationType string, duration time.Duration, err error) {
	atomic.AddInt64(&um.operationCount, 1)
	if err != nil {
		atomic.AddInt64(&um.errorCount, 1)
	}

	// Update operation latency
	um.mu.Lock()
	um.operationLatency = (um.operationLatency + duration) / 2
	um.mu.Unlock()
}

// GetMetrics returns the current monitoring metrics
func (um *UTXOMonitoring) GetMetrics() *MonitoringMetrics {
	um.mu.RLock()
	defer um.mu.RUnlock()

	return &MonitoringMetrics{
		TotalUTXOs:         atomic.LoadInt64(&um.totalUTXOs),
		SpentUTXOs:         atomic.LoadInt64(&um.spentUTXOs),
		UnspentUTXOs:       atomic.LoadInt64(&um.unspentUTXOs),
		TotalValue:         atomic.LoadInt64(&um.totalValue),
		LastUpdateTime:     um.lastUpdateTime,
		OperationCount:     atomic.LoadInt64(&um.operationCount),
		ErrorCount:         atomic.LoadInt64(&um.errorCount),
		CompressionRatio:   um.compressionRatio,
		IndexSize:          atomic.LoadInt64(&um.indexSize),
		StorageSize:        atomic.LoadInt64(&um.storageSize),
		OperationLatency:   um.operationLatency,
		LastPruneTime:      um.lastPruneTime,
		LastBackupTime:     um.lastBackupTime,
		LastValidationTime: um.lastValidationTime,
	}
}

// UpdateLastPruneTime updates the last prune time
func (um *UTXOMonitoring) UpdateLastPruneTime() {
	um.mu.Lock()
	um.lastPruneTime = time.Now()
	um.mu.Unlock()
}

// UpdateLastBackupTime updates the last backup time
func (um *UTXOMonitoring) UpdateLastBackupTime() {
	um.mu.Lock()
	um.lastBackupTime = time.Now()
	um.mu.Unlock()
}

// UpdateLastValidationTime updates the last validation time
func (um *UTXOMonitoring) UpdateLastValidationTime() {
	um.mu.Lock()
	um.lastValidationTime = time.Now()
	um.mu.Unlock()
}

// UpdateCompressionRatio updates the compression ratio
func (um *UTXOMonitoring) UpdateCompressionRatio(ratio float64) {
	um.mu.Lock()
	um.compressionRatio = ratio
	um.mu.Unlock()
}

// MonitoringMetrics holds the monitoring metrics
type MonitoringMetrics struct {
	// TotalUTXOs is the total number of UTXOs
	TotalUTXOs int64
	// SpentUTXOs is the number of spent UTXOs
	SpentUTXOs int64
	// UnspentUTXOs is the number of unspent UTXOs
	UnspentUTXOs int64
	// TotalValue is the total value of all UTXOs
	TotalValue int64
	// LastUpdateTime is the time of the last metrics update
	LastUpdateTime time.Time
	// OperationCount is the total number of operations
	OperationCount int64
	// ErrorCount is the total number of errors
	ErrorCount int64
	// CompressionRatio is the current compression ratio
	CompressionRatio float64
	// IndexSize is the size of the index in bytes
	IndexSize int64
	// StorageSize is the size of the storage in bytes
	StorageSize int64
	// OperationLatency is the average operation latency
	OperationLatency time.Duration
	// LastPruneTime is the time of the last prune operation
	LastPruneTime time.Time
	// LastBackupTime is the time of the last backup operation
	LastBackupTime time.Time
	// LastValidationTime is the time of the last validation operation
	LastValidationTime time.Time
}

// String returns a string representation of the monitoring metrics
func (mm *MonitoringMetrics) String() string {
	return fmt.Sprintf(
		"UTXOs: %d total (%d spent, %d unspent), Value: %d\n"+
			"Operations: %d total (%d errors), Latency: %v\n"+
			"Storage: %d bytes (Index: %d bytes), Compression: %.2f\n"+
			"Last Update: %v, Last Prune: %v, Last Backup: %v, Last Validation: %v",
		mm.TotalUTXOs, mm.SpentUTXOs, mm.UnspentUTXOs, mm.TotalValue,
		mm.OperationCount, mm.ErrorCount, mm.OperationLatency,
		mm.StorageSize, mm.IndexSize, mm.CompressionRatio,
		mm.LastUpdateTime.Format("2006-01-02 15:04:05"),
		mm.LastPruneTime.Format("2006-01-02 15:04:05"),
		mm.LastBackupTime.Format("2006-01-02 15:04:05"),
		mm.LastValidationTime.Format("2006-01-02 15:04:05"),
	)
}
