package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
)

// PruningManager manages block and UTXO pruning
type PruningManager struct {
	// Minimum number of blocks to keep
	minBlocks uint64
	// Maximum number of blocks to keep
	maxBlocks uint64
	// Current pruning height
	pruningHeight uint64
	// Last pruning time
	lastPruning time.Time
	// Pruning interval
	pruningInterval time.Duration
	mu              sync.RWMutex
}

// NewPruningManager creates a new pruning manager
func NewPruningManager(minBlocks, maxBlocks uint64, interval time.Duration) *PruningManager {
	return &PruningManager{
		minBlocks:       minBlocks,
		maxBlocks:       maxBlocks,
		pruningInterval: interval,
		lastPruning:     time.Now(),
	}
}

// ShouldPrune checks if pruning should be performed
func (pm *PruningManager) ShouldPrune(currentHeight uint64) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Check if enough time has passed since last pruning
	if time.Since(pm.lastPruning) < pm.pruningInterval {
		return false
	}

	// Check if we have enough blocks to prune
	return currentHeight > pm.minBlocks
}

// PruneBlocks prunes old blocks
func (pm *PruningManager) PruneBlocks(currentHeight uint64, store BlockStore) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Calculate pruning height
	pruneHeight := currentHeight - pm.minBlocks
	if pruneHeight <= pm.pruningHeight {
		return nil
	}

	// Prune blocks
	for height := pm.pruningHeight; height < pruneHeight; height++ {
		if err := store.DeleteBlock(height); err != nil {
			return fmt.Errorf("failed to prune block at height %d: %v", height, err)
		}
	}

	pm.pruningHeight = pruneHeight
	pm.lastPruning = time.Now()

	return nil
}

// PruneUTXOs prunes spent UTXOs
func (pm *PruningManager) PruneUTXOs(utxoSet *UTXOSet, currentHeight uint64) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Calculate pruning height
	pruneHeight := currentHeight - pm.minBlocks
	if pruneHeight <= pm.pruningHeight {
		return nil
	}

	// Prune spent UTXOs
	return utxoSet.PruneSpent(pruneHeight)
}

// GetPruningHeight returns the current pruning height
func (pm *PruningManager) GetPruningHeight() uint64 {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.pruningHeight
}

// SetPruningHeight sets the pruning height
func (pm *PruningManager) SetPruningHeight(height uint64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.pruningHeight = height
}

// BlockStore interface for block storage
type BlockStore interface {
	DeleteBlock(height uint64) error
	GetBlock(height uint64) (*block.Block, error)
	PutBlock(block *block.Block) error
}
