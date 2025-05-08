package interfaces

import (
	"github.com/youngchain/internal/core/block"
)

// Consensus defines the interface for consensus operations
type Consensus interface {
	// ValidateBlock validates a block
	ValidateBlock(block *block.Block) error

	// GetDifficulty returns the current difficulty
	GetDifficulty() uint32

	// AdjustDifficulty adjusts the difficulty based on the latest block
	AdjustDifficulty(block *block.Block) error
}
