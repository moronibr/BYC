package block

import (
	"fmt"
)

const (
	// MaxBlockSize is the maximum size of a block in bytes
	MaxBlockSize = 1000000 // 1MB

	// MaxBlockWeight is the maximum weight of a block
	MaxBlockWeight = 4000000 // 4M weight units

	// WitnessScaleFactor is the scaling factor for witness data
	WitnessScaleFactor = 4

	// BaseSizeWeight is the weight of base transaction data
	BaseSizeWeight = 1

	// WitnessSizeWeight is the weight of witness data
	WitnessSizeWeight = WitnessScaleFactor
)

// BlockWeight represents the weight of a block
type BlockWeight struct {
	BaseSize    int
	WitnessSize int
	TotalSize   int
	Weight      int
}

// CalculateBlockWeight calculates the weight of a block
func (b *Block) CalculateBlockWeight() (*BlockWeight, error) {
	weight := &BlockWeight{}

	// Calculate base size (excluding witness data)
	for _, tx := range b.Transactions {
		baseSize := tx.GetTransactionSize()
		weight.BaseSize += baseSize
		weight.TotalSize += baseSize
	}

	// Calculate witness size
	for _, tx := range b.Transactions {
		if tx.Witness != nil {
			witnessSize := tx.Witness.Size()
			weight.WitnessSize += witnessSize
			weight.TotalSize += witnessSize
		}
	}

	// Calculate total weight
	weight.Weight = (weight.BaseSize * BaseSizeWeight) + (weight.WitnessSize * WitnessSizeWeight)

	// Validate block size
	if weight.TotalSize > MaxBlockSize {
		return nil, fmt.Errorf("block size exceeds maximum: %d > %d", weight.TotalSize, MaxBlockSize)
	}

	// Validate block weight
	if weight.Weight > MaxBlockWeight {
		return nil, fmt.Errorf("block weight exceeds maximum: %d > %d", weight.Weight, MaxBlockWeight)
	}

	return weight, nil
}

// ValidateBlockWeight validates that a block's weight is within limits
func (b *Block) ValidateBlockWeight() error {
	weight, err := b.CalculateBlockWeight()
	if err != nil {
		return err
	}

	if weight.TotalSize > MaxBlockSize {
		return fmt.Errorf("block size exceeds maximum: %d > %d", weight.TotalSize, MaxBlockSize)
	}

	if weight.Weight > MaxBlockWeight {
		return fmt.Errorf("block weight exceeds maximum: %d > %d", weight.Weight, MaxBlockWeight)
	}

	return nil
}

// GetBlockSize returns the size of a block in bytes
func (b *Block) GetBlockSize() int {
	size := 0

	// Add block header size
	size += 80 // Fixed header size

	// Add transaction sizes
	for _, tx := range b.Transactions {
		size += tx.GetTransactionSize()
	}

	return size
}

// GetBlockWeight returns the weight of a block
func (b *Block) GetBlockWeight() int {
	weight := 0

	// Add block header weight
	weight += 80 * BaseSizeWeight // Fixed header weight

	// Add transaction weights
	for _, tx := range b.Transactions {
		// Base transaction weight
		weight += tx.GetTransactionSize() * BaseSizeWeight

		// Witness weight
		if tx.Witness != nil {
			weight += tx.Witness.Size() * WitnessSizeWeight
		}
	}

	return weight
}

// IsBlockFull checks if a block is full
func (b *Block) IsBlockFull() bool {
	weight, err := b.CalculateBlockWeight()
	if err != nil {
		return true
	}

	return weight.TotalSize >= MaxBlockSize || weight.Weight >= MaxBlockWeight
}

// CanAddTransaction checks if a transaction can be added to the block
func (b *Block) CanAddTransaction(tx *Transaction) bool {
	// Calculate current block weight
	currentWeight, err := b.CalculateBlockWeight()
	if err != nil {
		return false
	}

	// Calculate transaction weight
	txWeight := tx.GetTransactionSize() * BaseSizeWeight
	if tx.Witness != nil {
		txWeight += tx.Witness.Size() * WitnessSizeWeight
	}

	// Check if adding the transaction would exceed limits
	return (currentWeight.TotalSize+tx.GetTransactionSize() <= MaxBlockSize) &&
		(currentWeight.Weight+txWeight <= MaxBlockWeight)
}
