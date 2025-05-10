package block

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
)

// Checkpoint represents a checkpoint in the blockchain
type Checkpoint struct {
	Height    uint64
	BlockHash []byte
	Timestamp int64
}

// CheckpointManager manages blockchain checkpoints
type CheckpointManager struct {
	checkpoints map[uint64]*Checkpoint
	mu          sync.RWMutex
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager() *CheckpointManager {
	return &CheckpointManager{
		checkpoints: make(map[uint64]*Checkpoint),
	}
}

// AddCheckpoint adds a checkpoint
func (cm *CheckpointManager) AddCheckpoint(height uint64, blockHash []byte, timestamp int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.checkpoints[height] = &Checkpoint{
		Height:    height,
		BlockHash: blockHash,
		Timestamp: timestamp,
	}
}

// GetCheckpoint gets a checkpoint at a specific height
func (cm *CheckpointManager) GetCheckpoint(height uint64) (*Checkpoint, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	checkpoint, exists := cm.checkpoints[height]
	return checkpoint, exists
}

// ValidateBlock validates a block against checkpoints
func (cm *CheckpointManager) ValidateBlock(block *Block) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Check if there's a checkpoint at this height
	if checkpoint, exists := cm.checkpoints[block.Header.Height]; exists {
		// Verify block hash matches checkpoint
		if !bytes.Equal(block.Header.Hash, checkpoint.BlockHash) {
			return fmt.Errorf("block hash does not match checkpoint at height %d", block.Header.Height)
		}

		// Verify block timestamp is not before checkpoint
		if block.Header.Timestamp.Unix() < checkpoint.Timestamp {
			return fmt.Errorf("block timestamp is before checkpoint at height %d", block.Header.Height)
		}
	}

	return nil
}

// ValidateChain validates a chain of blocks against checkpoints
func (cm *CheckpointManager) ValidateChain(blocks []*Block) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Sort blocks by height
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Header.Height < blocks[j].Header.Height
	})

	// Validate each block
	for _, block := range blocks {
		if err := cm.ValidateBlock(block); err != nil {
			return err
		}
	}

	// Validate chain continuity
	for i := 1; i < len(blocks); i++ {
		if blocks[i].Header.Height != blocks[i-1].Header.Height+1 {
			return fmt.Errorf("chain discontinuity at height %d", blocks[i].Header.Height)
		}
		if !bytes.Equal(blocks[i].Header.PrevBlockHash, blocks[i-1].Header.Hash) {
			return fmt.Errorf("invalid previous block hash at height %d", blocks[i].Header.Height)
		}
	}

	return nil
}

// GetLatestCheckpoint gets the latest checkpoint
func (cm *CheckpointManager) GetLatestCheckpoint() *Checkpoint {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var latest *Checkpoint
	for _, checkpoint := range cm.checkpoints {
		if latest == nil || checkpoint.Height > latest.Height {
			latest = checkpoint
		}
	}
	return latest
}

// IsCheckpointed checks if a block height is checkpointed
func (cm *CheckpointManager) IsCheckpointed(height uint64) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	_, exists := cm.checkpoints[height]
	return exists
}

// GetCheckpointCount returns the number of checkpoints
func (cm *CheckpointManager) GetCheckpointCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.checkpoints)
}

// String returns a string representation of a checkpoint
func (c *Checkpoint) String() string {
	return fmt.Sprintf("Checkpoint{Height: %d, Hash: %s, Timestamp: %d}",
		c.Height,
		hex.EncodeToString(c.BlockHash),
		c.Timestamp,
	)
}
