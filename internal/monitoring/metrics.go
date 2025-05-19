package monitoring

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/byc/internal/blockchain"
	"github.com/byc/internal/network"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics represents the metrics collection system
type Metrics struct {
	// Prometheus metrics
	blockHeight      prometheus.Gauge
	blockTime        prometheus.Histogram
	transactionCount prometheus.Counter
	networkLatency   prometheus.Histogram
	peerCount        prometheus.Gauge
	hashRate         prometheus.Gauge
	memoryUsage      prometheus.Gauge
	cpuUsage         prometheus.Gauge
	errorCount       prometheus.Counter
	blockSize        prometheus.Histogram
	propagationTime  prometheus.Histogram

	// Internal state
	blockchain *blockchain.Blockchain
	node       *network.Node
	startTime  time.Time
	mu         sync.RWMutex

	// Custom metrics
	lastBlockTime    time.Time
	lastBlockHash    []byte
	blockPropagation map[string]time.Duration // block hash -> propagation time
	transactionPool  map[string]time.Time     // tx hash -> arrival time
	networkMessages  map[string]int64         // message type -> count
	peerLatencies    map[string]time.Duration // peer address -> latency
	miningStats      *MiningStats
	resourceStats    *ResourceStats
}

// MiningStats tracks mining performance
type MiningStats struct {
	TotalHashes    int64
	LastHashRate   float64
	LastBlockFound time.Time
	Difficulty     int
	CurrentNonce   int64
	SharesAccepted int64
	SharesRejected int64
	mu             sync.RWMutex
}

// ResourceStats tracks system resource usage
type ResourceStats struct {
	MemoryAlloc   uint64
	MemoryTotal   uint64
	NumGoroutines int
	NumCPU        int
	LastUpdate    time.Time
	mu            sync.RWMutex
}

// NewMetrics creates a new metrics collection system
func NewMetrics(bc *blockchain.Blockchain, node *network.Node) *Metrics {
	m := &Metrics{
		blockchain:       bc,
		node:             node,
		startTime:        time.Now(),
		blockPropagation: make(map[string]time.Duration),
		transactionPool:  make(map[string]time.Time),
		networkMessages:  make(map[string]int64),
		peerLatencies:    make(map[string]time.Duration),
		miningStats:      &MiningStats{},
		resourceStats:    &ResourceStats{},
	}

	// Initialize Prometheus metrics
	m.blockHeight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "byc_block_height",
		Help: "Current blockchain height",
	})

	m.blockTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "byc_block_time",
		Help:    "Time between blocks",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10),
	})

	m.transactionCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "byc_transaction_count",
		Help: "Total number of transactions processed",
	})

	m.networkLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "byc_network_latency",
		Help:    "Network message latency",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
	})

	m.peerCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "byc_peer_count",
		Help: "Number of connected peers",
	})

	m.hashRate = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "byc_hash_rate",
		Help: "Current mining hash rate",
	})

	m.memoryUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "byc_memory_usage",
		Help: "Memory usage in bytes",
	})

	m.cpuUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "byc_cpu_usage",
		Help: "CPU usage percentage",
	})

	m.errorCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "byc_error_count",
		Help: "Total number of errors",
	})

	m.blockSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "byc_block_size",
		Help:    "Block size in bytes",
		Buckets: prometheus.ExponentialBuckets(1024, 2, 10),
	})

	m.propagationTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "byc_block_propagation_time",
		Help:    "Block propagation time in seconds",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
	})

	return m
}

// Start starts the metrics collection
func (m *Metrics) Start() {
	// Start periodic updates
	go m.updateMetrics()
	go m.updateResourceStats()
}

// updateMetrics updates all metrics periodically
func (m *Metrics) updateMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()

		// Update blockchain metrics
		m.updateBlockchainMetrics()

		// Update network metrics
		m.updateNetworkMetrics()

		// Update mining metrics
		m.updateMiningMetrics()

		// Update transaction metrics
		m.updateTransactionMetrics()

		m.mu.Unlock()
	}
}

// updateResourceStats updates system resource usage metrics
func (m *Metrics) updateResourceStats() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		m.resourceStats.mu.Lock()
		m.resourceStats.MemoryAlloc = mem.Alloc
		m.resourceStats.MemoryTotal = mem.TotalAlloc
		m.resourceStats.NumGoroutines = runtime.NumGoroutine()
		m.resourceStats.NumCPU = runtime.NumCPU()
		m.resourceStats.LastUpdate = time.Now()
		m.resourceStats.mu.Unlock()

		// Update Prometheus metrics
		m.memoryUsage.Set(float64(mem.Alloc))
		m.cpuUsage.Set(float64(runtime.NumCPU()))
	}
}

// updateBlockchainMetrics updates blockchain-related metrics
func (m *Metrics) updateBlockchainMetrics() {
	// Update block height
	height := len(m.blockchain.GoldenBlocks)
	m.blockHeight.Set(float64(height))

	// Update block time
	if m.lastBlockTime.IsZero() {
		m.lastBlockTime = time.Now()
	} else {
		blockTime := time.Since(m.lastBlockTime).Seconds()
		m.blockTime.Observe(blockTime)
	}

	// Update block size
	for _, block := range m.blockchain.GoldenBlocks {
		blockSize := len(block.Hash) + len(block.PrevHash)
		for _, tx := range block.Transactions {
			blockSize += len(tx.ID)
		}
		m.blockSize.Observe(float64(blockSize))
	}
}

// updateNetworkMetrics updates network-related metrics
func (m *Metrics) updateNetworkMetrics() {
	// Update peer count
	peerCount := len(m.node.Peers)
	m.peerCount.Set(float64(peerCount))

	// Update network latency
	for _, latency := range m.peerLatencies {
		m.networkLatency.Observe(latency.Seconds())
	}
}

// updateMiningMetrics updates mining-related metrics
func (m *Metrics) updateMiningMetrics() {
	m.miningStats.mu.RLock()
	defer m.miningStats.mu.RUnlock()

	// Update hash rate
	m.hashRate.Set(m.miningStats.LastHashRate)
}

// updateTransactionMetrics updates transaction-related metrics
func (m *Metrics) updateTransactionMetrics() {
	// Update transaction count
	txCount := 0
	for _, block := range m.blockchain.GoldenBlocks {
		txCount += len(block.Transactions)
	}
	m.transactionCount.Add(float64(txCount))
}

// RecordBlockPropagation records block propagation time
func (m *Metrics) RecordBlockPropagation(blockHash []byte, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hashStr := fmt.Sprintf("%x", blockHash)
	m.blockPropagation[hashStr] = duration
	m.propagationTime.Observe(duration.Seconds())
}

// RecordTransaction records a new transaction
func (m *Metrics) RecordTransaction(txHash []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hashStr := fmt.Sprintf("%x", txHash)
	m.transactionPool[hashStr] = time.Now()
}

// RecordNetworkMessage records a network message
func (m *Metrics) RecordNetworkMessage(msgType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.networkMessages[msgType]++
}

// RecordPeerLatency records peer latency
func (m *Metrics) RecordPeerLatency(peerAddr string, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.peerLatencies[peerAddr] = latency
}

// RecordMiningStats records mining statistics
func (m *Metrics) RecordMiningStats(stats *MiningStats) {
	m.miningStats.mu.Lock()
	defer m.miningStats.mu.Unlock()

	m.miningStats = stats
}

// ServeHTTP implements the http.Handler interface for Prometheus metrics
func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Collect all metrics
	metrics := make(map[string]interface{})

	// Add Prometheus metrics
	metrics["block_height"] = m.blockHeight
	metrics["block_time"] = m.blockTime
	metrics["transaction_count"] = m.transactionCount
	metrics["network_latency"] = m.networkLatency
	metrics["peer_count"] = m.peerCount
	metrics["hash_rate"] = m.hashRate
	metrics["memory_usage"] = m.memoryUsage
	metrics["cpu_usage"] = m.cpuUsage
	metrics["error_count"] = m.errorCount
	metrics["block_size"] = m.blockSize
	metrics["propagation_time"] = m.propagationTime

	// Add custom metrics
	m.mu.RLock()
	metrics["block_propagation"] = m.blockPropagation
	metrics["transaction_pool"] = m.transactionPool
	metrics["network_messages"] = m.networkMessages
	metrics["peer_latencies"] = m.peerLatencies
	m.mu.RUnlock()

	// Add mining stats
	m.miningStats.mu.RLock()
	metrics["mining_stats"] = m.miningStats
	m.miningStats.mu.RUnlock()

	// Add resource stats
	m.resourceStats.mu.RLock()
	metrics["resource_stats"] = m.resourceStats
	m.resourceStats.mu.RUnlock()

	// Return metrics as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
