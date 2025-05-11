package transaction

import (
	"sync"
	"time"

	"github.com/youngchain/internal/core/types"
)

const (
	// Base fee per byte in satoshis
	baseFeePerByte = 1

	// Priority thresholds for fee calculation
	highPriorityThreshold   = 0.5
	mediumPriorityThreshold = 0.3
	lowPriorityThreshold    = 0.1

	// Fee multipliers for different priorities
	highPriorityMultiplier   = 2.0
	mediumPriorityMultiplier = 1.5
	lowPriorityMultiplier    = 1.0

	// Time-based fee adjustment
	urgentTimeThreshold   = 10 * time.Minute
	standardTimeThreshold = 1 * time.Hour
	relaxedTimeThreshold  = 24 * time.Hour

	// Size-based fee adjustment
	smallTxSizeThreshold  = 250  // bytes
	mediumTxSizeThreshold = 1000 // bytes
	largeTxSizeThreshold  = 2000 // bytes

	// UTXO age multipliers
	youngUTXOMultiplier  = 1.2
	mediumUTXOMultiplier = 1.0
	oldUTXOMultiplier    = 0.8

	// UTXO age thresholds
	youngUTXOThreshold  = 24 * time.Hour
	mediumUTXOThreshold = 7 * 24 * time.Hour
)

// EnhancedFeeCalculator calculates transaction fees using multiple factors
type EnhancedFeeCalculator struct {
	networkLoad    float64
	avgBlockTime   time.Duration
	avgBlockSize   int
	mempoolSize    int
	recentFees     []uint64
	recentFeesLock sync.RWMutex
	maxRecentFees  int
	baseFeeRate    uint64 // Base fee rate in satoshis per byte
}

// NewEnhancedFeeCalculator creates a new enhanced fee calculator
func NewEnhancedFeeCalculator() *EnhancedFeeCalculator {
	return &EnhancedFeeCalculator{
		networkLoad:   0.5,
		avgBlockTime:  10 * time.Minute,
		avgBlockSize:  1000000, // 1MB
		mempoolSize:   0,
		maxRecentFees: 100,
		baseFeeRate:   baseFeePerByte,
	}
}

// SetNetworkLoad updates the network load factor
func (fc *EnhancedFeeCalculator) SetNetworkLoad(load float64) {
	if load < 0 {
		load = 0
	} else if load > 1 {
		load = 1
	}
	fc.networkLoad = load
}

// UpdateNetworkMetrics updates network-related metrics
func (fc *EnhancedFeeCalculator) UpdateNetworkMetrics(avgBlockTime time.Duration, avgBlockSize, mempoolSize int) {
	fc.avgBlockTime = avgBlockTime
	fc.avgBlockSize = avgBlockSize
	fc.mempoolSize = mempoolSize
}

// AddRecentFee adds a recent fee to the calculator's history
func (fc *EnhancedFeeCalculator) AddRecentFee(fee uint64) {
	fc.recentFeesLock.Lock()
	defer fc.recentFeesLock.Unlock()

	fc.recentFees = append(fc.recentFees, fee)
	if len(fc.recentFees) > fc.maxRecentFees {
		fc.recentFees = fc.recentFees[1:]
	}
}

// CalculateEnhancedFee calculates the fee for a transaction using multiple factors
func (fc *EnhancedFeeCalculator) CalculateEnhancedFee(tx *types.Transaction) uint64 {
	// Calculate base size-based fee
	size := uint64(fc.CalculateTransactionSize(tx))
	baseFee := size * fc.baseFeeRate

	// Get all multipliers
	priorityMultiplier := fc.getPriorityMultiplier(fc.calculatePriority(tx))
	networkLoadMultiplier := fc.getNetworkLoadMultiplier()
	mempoolMultiplier := fc.getMempoolMultiplier()
	sizeMultiplier := fc.getSizeMultiplier(size)
	timeMultiplier := fc.getTimeMultiplier(tx)
	utxoAgeMultiplier := fc.getUTXOAgeMultiplier(tx)

	// Combine all multipliers
	totalMultiplier := float64(priorityMultiplier) * networkLoadMultiplier * mempoolMultiplier * sizeMultiplier * timeMultiplier * utxoAgeMultiplier

	return uint64(float64(baseFee) * totalMultiplier)
}

// getNetworkLoadMultiplier calculates fee multiplier based on network load
func (fc *EnhancedFeeCalculator) getNetworkLoadMultiplier() float64 {
	// Consider both network load and block time
	blockTimeFactor := float64(fc.avgBlockTime) / float64(10*time.Minute)
	loadFactor := fc.networkLoad

	return 1.0 + (blockTimeFactor * loadFactor)
}

// getMempoolMultiplier calculates fee multiplier based on mempool size
func (fc *EnhancedFeeCalculator) getMempoolMultiplier() float64 {
	// Calculate how full the mempool is relative to average block size
	mempoolFullness := float64(fc.mempoolSize) / float64(fc.avgBlockSize)

	switch {
	case mempoolFullness >= 2.0:
		return 2.0
	case mempoolFullness >= 1.0:
		return 1.5
	case mempoolFullness >= 0.5:
		return 1.2
	default:
		return 1.0
	}
}

// getSizeMultiplier calculates fee multiplier based on transaction size
func (fc *EnhancedFeeCalculator) getSizeMultiplier(size uint64) float64 {
	switch {
	case size <= smallTxSizeThreshold:
		return 1.0
	case size <= mediumTxSizeThreshold:
		return 1.2
	case size <= largeTxSizeThreshold:
		return 1.5
	default:
		return 2.0
	}
}

// getTimeMultiplier calculates fee multiplier based on transaction time requirements
func (fc *EnhancedFeeCalculator) getTimeMultiplier(tx *types.Transaction) float64 {
	now := time.Now()
	timeUntilValid := time.Duration(tx.LockTime-uint32(now.Unix())) * time.Second

	switch {
	case timeUntilValid <= urgentTimeThreshold:
		return highPriorityMultiplier
	case timeUntilValid <= standardTimeThreshold:
		return mediumPriorityMultiplier
	case timeUntilValid <= relaxedTimeThreshold:
		return lowPriorityMultiplier
	default:
		return 1.0
	}
}

// getUTXOAgeMultiplier calculates fee multiplier based on UTXO age
func (fc *EnhancedFeeCalculator) getUTXOAgeMultiplier(tx *types.Transaction) float64 {
	now := time.Now()
	var oldestUTXOAge time.Duration

	// Find the oldest UTXO age
	for _, input := range tx.Inputs {
		utxoAge := now.Sub(time.Unix(int64(input.Sequence), 0))
		if utxoAge > oldestUTXOAge {
			oldestUTXOAge = utxoAge
		}
	}

	switch {
	case oldestUTXOAge <= youngUTXOThreshold:
		return youngUTXOMultiplier
	case oldestUTXOAge <= mediumUTXOThreshold:
		return mediumUTXOMultiplier
	default:
		return oldUTXOMultiplier
	}
}

// CalculateTransactionSize estimates the transaction size in bytes
func (fc *EnhancedFeeCalculator) CalculateTransactionSize(tx *types.Transaction) int {
	size := 0

	// Version (4 bytes)
	size += 4

	// Input count (varint)
	size += 1

	// Inputs
	for _, input := range tx.Inputs {
		// Previous tx hash (32 bytes)
		size += 32
		// Previous tx index (4 bytes)
		size += 4
		// Script length (varint)
		size += 1
		// Script
		size += len(input.ScriptSig)
		// Sequence (4 bytes)
		size += 4
	}

	// Output count (varint)
	size += 1

	// Outputs
	for _, output := range tx.Outputs {
		// Value (8 bytes)
		size += 8
		// Script length (varint)
		size += 1
		// Script
		size += len(output.ScriptPubKey)
	}

	// Lock time (4 bytes)
	size += 4

	return size
}

// calculatePriority calculates the transaction priority
func (fc *EnhancedFeeCalculator) calculatePriority(tx *types.Transaction) float64 {
	// Priority is based on the time until the transaction becomes valid
	// (i.e., the time until the lock time is reached)
	now := time.Now().Unix()
	if tx.LockTime <= uint32(now) {
		return 1.0 // Already valid
	}

	// Calculate time until valid in hours
	timeUntilValid := float64(tx.LockTime-uint32(now)) / 3600.0

	// Priority decreases as time until valid increases
	// 1 hour = 1.0, 2 hours = 0.5, 4 hours = 0.25, etc.
	return 1.0 / timeUntilValid
}

// getPriorityMultiplier returns the fee multiplier based on priority
func (fc *EnhancedFeeCalculator) getPriorityMultiplier(priority float64) uint64 {
	switch {
	case priority >= highPriorityThreshold:
		return uint64(highPriorityMultiplier*100) / 100
	case priority >= mediumPriorityThreshold:
		return uint64(mediumPriorityMultiplier*100) / 100
	case priority >= lowPriorityThreshold:
		return uint64(lowPriorityMultiplier*100) / 100
	default:
		return 1
	}
}
