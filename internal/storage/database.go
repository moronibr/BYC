package storage

import (
	"fmt"
	"sync"

	"github.com/youngchain/internal/core/block"
)

// Database represents the blockchain database
type Database struct {
	mu sync.RWMutex

	// Chain state
	bestBlockHash []byte
	bestHeight    uint64

	// Block storage
	blocks map[string]*block.Block
}

// NewDatabase creates a new database instance
func NewDatabase() *Database {
	return &Database{
		blocks: make(map[string]*block.Block),
	}
}

// GetChainState returns the current chain state
func (db *Database) GetChainState() (uint64, []byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.bestHeight, db.bestBlockHash, nil
}

// UpdateChainState updates the chain state
func (db *Database) UpdateChainState(hash []byte, height uint64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.bestBlockHash = hash
	db.bestHeight = height
	return nil
}

// SaveBlock saves a block to the database
func (db *Database) SaveBlock(block *block.Block) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.blocks[string(block.Hash)] = block
	return nil
}

// GetBlock retrieves a block from the database
func (db *Database) GetBlock(hash []byte) (*block.Block, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	block, exists := db.blocks[string(hash)]
	if !exists {
		return nil, fmt.Errorf("block not found")
	}

	return block, nil
}

// GetBlockByHeight retrieves a block by its height
func (db *Database) GetBlockByHeight(height uint64) (*block.Block, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Iterate through blocks to find one with matching height
	for _, block := range db.blocks {
		if block.Header.Height == height {
			return block, nil
		}
	}

	return nil, fmt.Errorf("block not found at height %d", height)
}
