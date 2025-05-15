package testing

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/consensus"
	"github.com/youngchain/internal/core/transaction"
	"github.com/youngchain/internal/core/types"
)

// PerformanceTest represents a performance test
type PerformanceTest struct {
	// Test configuration
	config *PerformanceTestConfig

	// Test metrics
	metrics *PerformanceMetrics

	// Test mutex
	mu sync.Mutex
}

// PerformanceTestConfig holds the configuration for a performance test
type PerformanceTestConfig struct {
	// Number of transactions to generate
	TransactionCount int

	// Number of blocks to generate
	BlockCount int

	// Number of concurrent miners
	MinerCount int

	// Test duration
	Duration time.Duration

	// Chain type to test
	ChainType block.BlockType
}

// PerformanceMetrics holds the metrics for a performance test
type PerformanceMetrics struct {
	// Transaction metrics
	TransactionsPerSecond float64
	AverageTxLatency      time.Duration
	MaxTxLatency          time.Duration
	MinTxLatency          time.Duration

	// Block metrics
	BlocksPerSecond     float64
	AverageBlockLatency time.Duration
	MaxBlockLatency     time.Duration
	MinBlockLatency     time.Duration

	// Mining metrics
	AverageHashRate float64
	MaxHashRate     float64
	MinHashRate     float64

	// Memory metrics
	AverageMemoryUsage int64
	MaxMemoryUsage     int64
	MinMemoryUsage     int64

	// Network metrics
	AverageNetworkLatency time.Duration
	MaxNetworkLatency     time.Duration
	MinNetworkLatency     time.Duration
}

// NewPerformanceTest creates a new performance test
func NewPerformanceTest(config *PerformanceTestConfig) *PerformanceTest {
	if config == nil {
		config = &PerformanceTestConfig{
			TransactionCount: 1000,
			BlockCount:       100,
			MinerCount:       4,
			Duration:         1 * time.Hour,
			ChainType:        block.GoldenBlock,
		}
	}

	return &PerformanceTest{
		config:  config,
		metrics: &PerformanceMetrics{},
	}
}

// RunTest runs the performance test
func (pt *PerformanceTest) RunTest(t *testing.T) {
	// Initialize test components
	consensus := consensus.NewConsensus(nil)
	txPool := transaction.NewTxPool()
	blockChain := block.NewBlockchain()

	// Start test timer
	startTime := time.Now()
	endTime := startTime.Add(pt.config.Duration)

	// Create channels for metrics
	txMetrics := make(chan *TransactionMetric, pt.config.TransactionCount)
	blockMetrics := make(chan *BlockMetric, pt.config.BlockCount)
	miningMetrics := make(chan *MiningMetric, pt.config.MinerCount)
	networkMetrics := make(chan *NetworkMetric, pt.config.TransactionCount)

	// Start metric collectors
	go pt.collectTransactionMetrics(txMetrics)
	go pt.collectBlockMetrics(blockMetrics)
	go pt.collectMiningMetrics(miningMetrics)
	go pt.collectNetworkMetrics(networkMetrics)

	// Start miners
	var wg sync.WaitGroup
	for i := 0; i < pt.config.MinerCount; i++ {
		wg.Add(1)
		go pt.runMiner(i, consensus, txPool, blockChain, miningMetrics, &wg)
	}

	// Generate transactions
	for i := 0; i < pt.config.TransactionCount; i++ {
		if time.Now().After(endTime) {
			break
		}

		tx := pt.generateTransaction()
		txStart := time.Now()

		// Submit transaction
		if err := txPool.AddTransaction(tx); err != nil {
			t.Errorf("Failed to add transaction: %v", err)
			continue
		}

		// Record transaction metric
		txMetrics <- &TransactionMetric{
			Latency: time.Since(txStart),
		}

		// Simulate network latency
		networkMetrics <- &NetworkMetric{
			Latency: time.Duration(rand.Int63n(100)) * time.Millisecond,
		}
	}

	// Wait for miners to finish
	wg.Wait()

	// Close metric channels
	close(txMetrics)
	close(blockMetrics)
	close(miningMetrics)
	close(networkMetrics)

	// Print test results
	pt.printResults(t)
}

// runMiner runs a miner for the test
func (pt *PerformanceTest) runMiner(id int, consensus *consensus.Consensus, txPool *transaction.TxPool, blockChain *block.Blockchain, metrics chan<- *MiningMetric, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		// Check if test is complete
		if time.Now().After(time.Now().Add(pt.config.Duration)) {
			return
		}

		// Create block
		block, err := pt.createBlock(txPool, blockChain)
		if err != nil {
			continue
		}

		// Mine block
		startTime := time.Now()
		if err := pt.mineBlock(block, consensus); err != nil {
			continue
		}

		// Record mining metric
		metrics <- &MiningMetric{
			HashRate: float64(block.Header.Difficulty) / time.Since(startTime).Seconds(),
		}

		// Add block to chain
		if err := blockChain.AddBlock(block); err != nil {
			continue
		}
	}
}

// generateTransaction generates a test transaction
func (pt *PerformanceTest) generateTransaction() *types.Transaction {
	// Create a random transaction
	tx := &types.Transaction{
		Version:   1,
		Timestamp: time.Now(),
		Inputs:    make([]*types.TxInput, 1),
		Outputs:   make([]*types.TxOutput, 1),
		Fee:       uint64(rand.Int63n(1000)),
	}

	// Set random input
	tx.Inputs[0] = &types.TxInput{
		PreviousTxHash:  make([]byte, 32),
		PreviousTxIndex: uint32(rand.Int63n(100)),
		ScriptSig:       make([]byte, 100),
		Sequence:        0xffffffff,
	}

	// Set random output
	tx.Outputs[0] = &types.TxOutput{
		Value:        uint64(rand.Int63n(1000000)),
		ScriptPubKey: make([]byte, 100),
	}

	return tx
}

// createBlock creates a test block
func (pt *PerformanceTest) createBlock(txPool *transaction.TxPool, blockChain *block.Blockchain) (*block.Block, error) {
	// Get transactions from pool
	txs := txPool.GetBest(1000)
	if len(txs) == 0 {
		return nil, fmt.Errorf("no transactions available")
	}

	// Create block
	block := block.NewBlock(blockChain.GetBestBlock().Header.Hash, blockChain.GetBestBlock().Header.Height+1)
	block.Header.Type = pt.config.ChainType

	// Add transactions
	for _, tx := range txs {
		if err := block.AddTransaction(tx); err != nil {
			continue
		}
	}

	return block, nil
}

// mineBlock mines a test block
func (pt *PerformanceTest) mineBlock(block *block.Block, consensus *consensus.Consensus) error {
	// Set initial nonce
	block.Header.Nonce = 0

	// Mine block
	for {
		// Calculate hash
		hash := block.CalculateHash()

		// Check if hash meets difficulty
		if consensus.ValidateBlock(block) == nil {
			block.Header.Hash = hash
			return nil
		}

		// Increment nonce
		block.Header.Nonce++
	}
}

// collectTransactionMetrics collects transaction metrics
func (pt *PerformanceTest) collectTransactionMetrics(metrics <-chan *TransactionMetric) {
	var totalLatency time.Duration
	var count int
	var maxLatency, minLatency time.Duration

	for metric := range metrics {
		totalLatency += metric.Latency
		count++

		if count == 1 {
			maxLatency = metric.Latency
			minLatency = metric.Latency
		} else {
			if metric.Latency > maxLatency {
				maxLatency = metric.Latency
			}
			if metric.Latency < minLatency {
				minLatency = metric.Latency
			}
		}
	}

	pt.mu.Lock()
	pt.metrics.TransactionsPerSecond = float64(count) / time.Since(time.Now()).Seconds()
	pt.metrics.AverageTxLatency = totalLatency / time.Duration(count)
	pt.metrics.MaxTxLatency = maxLatency
	pt.metrics.MinTxLatency = minLatency
	pt.mu.Unlock()
}

// collectBlockMetrics collects block metrics
func (pt *PerformanceTest) collectBlockMetrics(metrics <-chan *BlockMetric) {
	var totalLatency time.Duration
	var count int
	var maxLatency, minLatency time.Duration

	for metric := range metrics {
		totalLatency += metric.Latency
		count++

		if count == 1 {
			maxLatency = metric.Latency
			minLatency = metric.Latency
		} else {
			if metric.Latency > maxLatency {
				maxLatency = metric.Latency
			}
			if metric.Latency < minLatency {
				minLatency = metric.Latency
			}
		}
	}

	pt.mu.Lock()
	pt.metrics.BlocksPerSecond = float64(count) / time.Since(time.Now()).Seconds()
	pt.metrics.AverageBlockLatency = totalLatency / time.Duration(count)
	pt.metrics.MaxBlockLatency = maxLatency
	pt.metrics.MinBlockLatency = minLatency
	pt.mu.Unlock()
}

// collectMiningMetrics collects mining metrics
func (pt *PerformanceTest) collectMiningMetrics(metrics <-chan *MiningMetric) {
	var totalHashRate float64
	var count int
	var maxHashRate, minHashRate float64

	for metric := range metrics {
		totalHashRate += metric.HashRate
		count++

		if count == 1 {
			maxHashRate = metric.HashRate
			minHashRate = metric.HashRate
		} else {
			if metric.HashRate > maxHashRate {
				maxHashRate = metric.HashRate
			}
			if metric.HashRate < minHashRate {
				minHashRate = metric.HashRate
			}
		}
	}

	pt.mu.Lock()
	pt.metrics.AverageHashRate = totalHashRate / float64(count)
	pt.metrics.MaxHashRate = maxHashRate
	pt.metrics.MinHashRate = minHashRate
	pt.mu.Unlock()
}

// collectNetworkMetrics collects network metrics
func (pt *PerformanceTest) collectNetworkMetrics(metrics <-chan *NetworkMetric) {
	var totalLatency time.Duration
	var count int
	var maxLatency, minLatency time.Duration

	for metric := range metrics {
		totalLatency += metric.Latency
		count++

		if count == 1 {
			maxLatency = metric.Latency
			minLatency = metric.Latency
		} else {
			if metric.Latency > maxLatency {
				maxLatency = metric.Latency
			}
			if metric.Latency < minLatency {
				minLatency = metric.Latency
			}
		}
	}

	pt.mu.Lock()
	pt.metrics.AverageNetworkLatency = totalLatency / time.Duration(count)
	pt.metrics.MaxNetworkLatency = maxLatency
	pt.metrics.MinNetworkLatency = minLatency
	pt.mu.Unlock()
}

// printResults prints the test results
func (pt *PerformanceTest) printResults(t *testing.T) {
	t.Logf("Performance Test Results:")
	t.Logf("Transactions:")
	t.Logf("  Per Second: %.2f", pt.metrics.TransactionsPerSecond)
	t.Logf("  Average Latency: %v", pt.metrics.AverageTxLatency)
	t.Logf("  Max Latency: %v", pt.metrics.MaxTxLatency)
	t.Logf("  Min Latency: %v", pt.metrics.MinTxLatency)
	t.Logf("Blocks:")
	t.Logf("  Per Second: %.2f", pt.metrics.BlocksPerSecond)
	t.Logf("  Average Latency: %v", pt.metrics.AverageBlockLatency)
	t.Logf("  Max Latency: %v", pt.metrics.MaxBlockLatency)
	t.Logf("  Min Latency: %v", pt.metrics.MinBlockLatency)
	t.Logf("Mining:")
	t.Logf("  Average Hash Rate: %.2f H/s", pt.metrics.AverageHashRate)
	t.Logf("  Max Hash Rate: %.2f H/s", pt.metrics.MaxHashRate)
	t.Logf("  Min Hash Rate: %.2f H/s", pt.metrics.MinHashRate)
	t.Logf("Network:")
	t.Logf("  Average Latency: %v", pt.metrics.AverageNetworkLatency)
	t.Logf("  Max Latency: %v", pt.metrics.MaxNetworkLatency)
	t.Logf("  Min Latency: %v", pt.metrics.MinNetworkLatency)
}

// Metric types
type TransactionMetric struct {
	Latency time.Duration
}

type BlockMetric struct {
	Latency time.Duration
}

type MiningMetric struct {
	HashRate float64
}

type NetworkMetric struct {
	Latency time.Duration
}
