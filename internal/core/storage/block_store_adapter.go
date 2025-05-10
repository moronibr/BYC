package storage

import (
	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/storage"
)

// BlockStoreAdapter adapts storage.DB to storage.BlockStore interface
type BlockStoreAdapter struct {
	db *storage.DB
}

// NewBlockStoreAdapter creates a new BlockStoreAdapter
func NewBlockStoreAdapter(db *storage.DB) *BlockStoreAdapter {
	return &BlockStoreAdapter{
		db: db,
	}
}

// GetBlock retrieves a block by its height
func (a *BlockStoreAdapter) GetBlock(height uint64) (*block.Block, error) {
	return a.db.GetBlockByHeight(height)
}

// PutBlock stores a block
func (a *BlockStoreAdapter) PutBlock(block *block.Block) error {
	return a.db.SaveBlock(block)
}

// GetLastBlock retrieves the last block in the chain
func (a *BlockStoreAdapter) GetLastBlock() (*block.Block, error) {
	// TODO: Implement this method
	return nil, nil
}

// DeleteBlock removes a block from storage
func (a *BlockStoreAdapter) DeleteBlock(height uint64) error {
	// TODO: Implement this method
	return nil
}
