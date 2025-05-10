package difficulty

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/youngchain/internal/core/block"
)

const (
	// TargetTimePerBlock is the target time between blocks in seconds
	TargetTimePerBlock = 600 // 10 minutes

	// DifficultyAdjustmentInterval is the number of blocks between difficulty adjustments
	DifficultyAdjustmentInterval = 2016

	// MinDifficulty is the minimum difficulty
	MinDifficulty = 0x1d00ffff

	// MaxDifficulty is the maximum difficulty
	MaxDifficulty = 0x00000000ffff0000000000000000000000000000000000000000000000000000
)

// CalculateNextDifficulty calculates the next difficulty based on the last blocks
func CalculateNextDifficulty(blocks []*block.Block) (uint64, error) {
	if len(blocks) < DifficultyAdjustmentInterval {
		return uint64(MinDifficulty), nil
	}

	// Get the last block in the interval
	lastBlock := blocks[len(blocks)-1]

	// Get the first block in the interval
	firstBlock := blocks[0]

	// Calculate the time difference
	timeDiff := lastBlock.Header.Timestamp.Sub(firstBlock.Header.Timestamp).Seconds()

	// Calculate the difficulty adjustment
	adjustment := float64(TargetTimePerBlock*DifficultyAdjustmentInterval) / timeDiff

	// Limit the adjustment
	if adjustment < 0.25 {
		adjustment = 0.25
	} else if adjustment > 4 {
		adjustment = 4
	}

	// Calculate the new difficulty
	newDifficulty := float64(firstBlock.Header.Difficulty) * adjustment

	// Ensure the difficulty is within bounds
	if newDifficulty < float64(MinDifficulty) {
		newDifficulty = float64(MinDifficulty)
	} else if newDifficulty > float64(MaxDifficulty) {
		newDifficulty = float64(MaxDifficulty)
	}

	return uint64(newDifficulty), nil
}

// BitsToDifficulty converts compact bits to difficulty
func BitsToDifficulty(bits uint32) uint64 {
	// Extract the exponent and mantissa
	exponent := bits >> 24
	mantissa := bits & 0x00ffffff

	// Calculate the difficulty
	if exponent <= 3 {
		return uint64(mantissa >> (8 * (3 - exponent)))
	}
	return uint64(mantissa) << (8 * (exponent - 3))
}

// DifficultyToBits converts difficulty to compact bits
func DifficultyToBits(difficulty uint32) uint32 {
	// Find the highest set bit
	highestBit := uint32(0)
	for i := uint32(0); i < 256; i++ {
		if difficulty&(1<<i) != 0 {
			highestBit = i
		}
	}

	// Calculate the exponent and mantissa
	exponent := (highestBit / 8) + 3
	mantissa := difficulty >> (8 * (exponent - 3))

	// Combine into bits
	return (exponent << 24) | uint32(mantissa)
}

// ValidateDifficulty validates that a block's difficulty is correct
func ValidateDifficulty(block *block.Block, prevBlocks []*block.Block) error {
	// Calculate the expected difficulty
	expectedDifficulty, err := CalculateNextDifficulty(prevBlocks)
	if err != nil {
		return err
	}

	// Check if the block's difficulty matches the expected difficulty
	if block.Header.Difficulty != uint32(expectedDifficulty) {
		return fmt.Errorf("invalid difficulty: got %d, want %d", block.Header.Difficulty, expectedDifficulty)
	}

	return nil
}

// GetNextWorkRequired calculates the next required work
func GetNextWorkRequired(currentBlock *block.Block, prevBlocks []*block.Block) (uint64, error) {
	// If we're not at a difficulty adjustment interval, use the current difficulty
	if len(prevBlocks) < DifficultyAdjustmentInterval {
		return uint64(currentBlock.Header.Difficulty), nil
	}

	// Calculate the next difficulty
	return CalculateNextDifficulty(prevBlocks)
}

// CheckProofOfWork checks if a block's proof of work is valid
func CheckProofOfWork(block *block.Block) bool {
	// Convert difficulty to target
	target := BitsToDifficulty(DifficultyToBits(uint32(block.Header.Difficulty)))

	// Calculate the block hash
	hash := block.CalculateHash()

	// Convert hash to uint64
	hashInt := binary.BigEndian.Uint64(hash)

	// Check if the hash is less than the target
	return hashInt <= target
}

// GetASERTDifficulty calculates the difficulty using ASERT algorithm
func GetASERTDifficulty(block *block.Block, prevBlocks []*block.Block) (uint64, error) {
	if len(prevBlocks) == 0 {
		return uint64(MinDifficulty), nil
	}

	// Get the anchor block (last block in the interval)
	anchorBlock := prevBlocks[len(prevBlocks)-1]

	// Calculate the time difference
	timeDiff := block.Header.Timestamp.Sub(anchorBlock.Header.Timestamp).Seconds()

	// Calculate the difficulty adjustment
	adjustment := math.Pow(2, timeDiff/float64(TargetTimePerBlock))

	// Calculate the new difficulty
	newDifficulty := float64(anchorBlock.Header.Difficulty) * adjustment

	// Ensure the difficulty is within bounds
	if newDifficulty < float64(MinDifficulty) {
		newDifficulty = float64(MinDifficulty)
	} else if newDifficulty > float64(MaxDifficulty) {
		newDifficulty = float64(MaxDifficulty)
	}

	return uint64(newDifficulty), nil
}
