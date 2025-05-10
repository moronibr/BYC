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
		baseSize := tx.Size()
		weight.BaseSize += baseSize
		weight.TotalSize += baseSize
	}

	// Calculate witness size
	for _, tx := range b.Transactions {
		witnesses := tx.Witness()
		if witnesses != nil {
			witnessSize := 0
			for _, w := range witnesses {
				witnessSize += len(w)
			}
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

// IsBlockFull checks if a block is full
func (b *Block) IsBlockFull() bool {
	weight, err := b.CalculateBlockWeight()
	if err != nil {
		return true
	}

	return weight.TotalSize >= MaxBlockSize || weight.Weight >= MaxBlockWeight
}
