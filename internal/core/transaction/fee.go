package transaction

import (
	"math"
	"time"

	"github.com/youngchain/internal/core/common"
)

// FeeCalculator handles transaction fee calculations
type FeeCalculator struct {
	baseFee            int64
	feePerByte         int64
	minFee             int64
	maxFee             int64
	priorityMultiplier float64
}

// NewFeeCalculator creates a new fee calculator
func NewFeeCalculator(baseFee, feePerByte, minFee, maxFee int64) *FeeCalculator {
	return &FeeCalculator{
		baseFee:            baseFee,
		feePerByte:         feePerByte,
		minFee:             minFee,
		maxFee:             maxFee,
		priorityMultiplier: 1.0,
	}
}

// CalculateFee calculates the fee for a transaction
func (fc *FeeCalculator) CalculateFee(tx *common.Transaction) int64 {
	// Calculate base fee
	fee := fc.baseFee

	// Add per-byte fee
	size := int64(tx.Size())
	fee += size * fc.feePerByte

	// Apply priority multiplier based on transaction age
	age := time.Since(tx.Timestamp()).Hours()
	if age > 24 {
		fee = int64(float64(fee) * fc.priorityMultiplier)
	}

	// Ensure fee is within bounds
	fee = int64(math.Max(float64(fee), float64(fc.minFee)))
	fee = int64(math.Min(float64(fee), float64(fc.maxFee)))

	return fee
}

// EstimateFee estimates the fee for a transaction with given parameters
func (fc *FeeCalculator) EstimateFee(inputCount, outputCount int) int64 {
	// Estimate transaction size
	estimatedSize := fc.estimateSize(inputCount, outputCount)

	// Calculate base fee
	fee := fc.baseFee

	// Add per-byte fee
	fee += estimatedSize * fc.feePerByte

	// Ensure fee is within bounds
	fee = int64(math.Max(float64(fee), float64(fc.minFee)))
	fee = int64(math.Min(float64(fee), float64(fc.maxFee)))

	return fee
}

// estimateSize estimates the size of a transaction
func (fc *FeeCalculator) estimateSize(inputCount, outputCount int) int64 {
	// Base transaction size
	size := int64(4) // Version

	// Input count
	size += int64(1)                // VarInt for input count
	size += int64(inputCount) * 148 // Average input size

	// Output count
	size += int64(1)                // VarInt for output count
	size += int64(outputCount) * 34 // Average output size

	// Locktime
	size += int64(4)

	return size
}

// SetPriorityMultiplier sets the priority multiplier for older transactions
func (fc *FeeCalculator) SetPriorityMultiplier(multiplier float64) {
	if multiplier >= 1.0 {
		fc.priorityMultiplier = multiplier
	}
}

// GetBaseFee returns the current base fee
func (fc *FeeCalculator) GetBaseFee() int64 {
	return fc.baseFee
}

// SetBaseFee sets the base fee
func (fc *FeeCalculator) SetBaseFee(fee int64) {
	if fee >= fc.minFee {
		fc.baseFee = fee
	}
}

// GetFeePerByte returns the current per-byte fee
func (fc *FeeCalculator) GetFeePerByte() int64 {
	return fc.feePerByte
}

// SetFeePerByte sets the per-byte fee
func (fc *FeeCalculator) SetFeePerByte(fee int64) {
	if fee >= 0 {
		fc.feePerByte = fee
	}
}
