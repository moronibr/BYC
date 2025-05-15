package consensus

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

// DifficultyAdjuster manages block difficulty adjustment
type DifficultyAdjuster struct {
	// Configuration
	targetBlockTime  time.Duration
	adjustmentWindow int
	minDifficulty    uint32
	maxDifficulty    uint32
	adjustmentFactor float64
}

// NewDifficultyAdjuster creates a new difficulty adjuster
func NewDifficultyAdjuster(targetBlockTime time.Duration, adjustmentWindow int, minDifficulty, maxDifficulty uint32) *DifficultyAdjuster {
	return &DifficultyAdjuster{
		targetBlockTime:  targetBlockTime,
		adjustmentWindow: adjustmentWindow,
		minDifficulty:    minDifficulty,
		maxDifficulty:    maxDifficulty,
		adjustmentFactor: 0.25, // Allow 25% adjustment per window
	}
}

// CalculateNextDifficulty calculates the next block difficulty
func (da *DifficultyAdjuster) CalculateNextDifficulty(blockTimes []time.Time, currentDifficulty uint32) (uint32, error) {
	if len(blockTimes) < 2 {
		return currentDifficulty, nil
	}

	// Calculate average block time
	var totalTime time.Duration
	for i := 1; i < len(blockTimes); i++ {
		totalTime += blockTimes[i].Sub(blockTimes[i-1])
	}
	avgBlockTime := totalTime / time.Duration(len(blockTimes)-1)

	// Calculate difficulty adjustment
	adjustment := float64(da.targetBlockTime) / float64(avgBlockTime)
	if adjustment > 1+da.adjustmentFactor {
		adjustment = 1 + da.adjustmentFactor
	} else if adjustment < 1-da.adjustmentFactor {
		adjustment = 1 - da.adjustmentFactor
	}

	// Calculate new difficulty
	newDifficulty := uint32(float64(currentDifficulty) * adjustment)

	// Apply limits
	if newDifficulty < da.minDifficulty {
		newDifficulty = da.minDifficulty
	} else if newDifficulty > da.maxDifficulty {
		newDifficulty = da.maxDifficulty
	}

	return newDifficulty, nil
}

// CalculateTarget calculates the target hash for a given difficulty
func (da *DifficultyAdjuster) CalculateTarget(difficulty uint32) [32]byte {
	// Convert difficulty to target
	target := make([]byte, 32)
	binary.BigEndian.PutUint32(target[28:], difficulty)

	// Calculate target value
	targetValue := math.MaxUint32 / float64(difficulty)
	targetBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(targetBytes, uint64(targetValue))

	// Copy target bytes
	copy(target[24:], targetBytes)

	return [32]byte(target)
}

// ValidateDifficulty validates if a block's difficulty is correct
func (da *DifficultyAdjuster) ValidateDifficulty(blockTime time.Time, prevBlockTimes []time.Time, currentDifficulty uint32) error {
	// Calculate expected difficulty
	expectedDifficulty, err := da.CalculateNextDifficulty(append(prevBlockTimes, blockTime), currentDifficulty)
	if err != nil {
		return fmt.Errorf("failed to calculate expected difficulty: %v", err)
	}

	// Check if difficulty is within acceptable range
	minAllowed := uint32(float64(expectedDifficulty) * (1 - da.adjustmentFactor))
	maxAllowed := uint32(float64(expectedDifficulty) * (1 + da.adjustmentFactor))

	if currentDifficulty < minAllowed || currentDifficulty > maxAllowed {
		return fmt.Errorf("difficulty %d is outside acceptable range [%d, %d]",
			currentDifficulty, minAllowed, maxAllowed)
	}

	return nil
}

// GetTargetBlockTime returns the target block time
func (da *DifficultyAdjuster) GetTargetBlockTime() time.Duration {
	return da.targetBlockTime
}

// GetAdjustmentWindow returns the adjustment window size
func (da *DifficultyAdjuster) GetAdjustmentWindow() int {
	return da.adjustmentWindow
}

// GetMinDifficulty returns the minimum difficulty
func (da *DifficultyAdjuster) GetMinDifficulty() uint32 {
	return da.minDifficulty
}

// GetMaxDifficulty returns the maximum difficulty
func (da *DifficultyAdjuster) GetMaxDifficulty() uint32 {
	return da.maxDifficulty
}
