package transaction

import (
	"math"
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
)

// FeeCalculator calculates transaction fees
type FeeCalculator struct {
	networkLoad float64 // 0.0 to 1.0 representing network congestion
}

// NewFeeCalculator creates a new fee calculator
func NewFeeCalculator() *FeeCalculator {
	return &FeeCalculator{
		networkLoad: 0.5, // Default to medium load
	}
}

// SetNetworkLoad updates the network load factor
func (fc *FeeCalculator) SetNetworkLoad(load float64) {
	if load < 0 {
		load = 0
	} else if load > 1 {
		load = 1
	}
	fc.networkLoad = load
}

// CalculateFee calculates the fee for a transaction
func (fc *FeeCalculator) CalculateFee(tx *Transaction) uint64 {
	// Calculate transaction size in bytes
	size := fc.calculateTransactionSize(tx)

	// Calculate base fee
	baseFee := uint64(size * baseFeePerByte)

	// Apply priority multiplier based on network load
	multiplier := fc.getPriorityMultiplier()

	// Calculate final fee
	fee := uint64(math.Ceil(float64(baseFee) * multiplier))

	return fee
}

// calculateTransactionSize estimates the transaction size in bytes
func (fc *FeeCalculator) calculateTransactionSize(tx *Transaction) int {
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

// getPriorityMultiplier returns the fee multiplier based on network load
func (fc *FeeCalculator) getPriorityMultiplier() float64 {
	switch {
	case fc.networkLoad >= highPriorityThreshold:
		return highPriorityMultiplier
	case fc.networkLoad >= mediumPriorityThreshold:
		return mediumPriorityMultiplier
	default:
		return lowPriorityMultiplier
	}
}
