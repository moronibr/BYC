package storage

import (
	"fmt"
	"sync"

	"github.com/youngchain/internal/core/block"
	"go.etcd.io/bbolt"
)

// StorageInterface defines the interface for database operations
type StorageInterface interface {
	// Put stores a key-value pair in the database
	Put(key, value []byte) error

	// Get retrieves a value by key from the database
	Get(key []byte) ([]byte, error)

	// Update executes a function within a read-write transaction
	Update(fn func(*bbolt.Tx) error) error

	// View executes a function within a read-only transaction
	View(fn func(*bbolt.Tx) error) error
}

// BoltDB implements the StorageInterface using bbolt
type BoltDB struct {
	db *bbolt.DB
}

// NewBoltDB creates a new BoltDB instance
func NewBoltDB(db *bbolt.DB) *BoltDB {
	return &BoltDB{db: db}
}

// Put implements StorageInterface.Put
func (b *BoltDB) Put(key, value []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("transactions"))
		if err != nil {
			return err
		}
		return bucket.Put(key, value)
	})
}

// Get implements StorageInterface.Get
func (b *BoltDB) Get(key []byte) ([]byte, error) {
	var value []byte
	err := b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("transactions"))
		if bucket == nil {
			return nil
		}
		value = bucket.Get(key)
		return nil
	})
	return value, err
}

// Update implements StorageInterface.Update
func (b *BoltDB) Update(fn func(*bbolt.Tx) error) error {
	return b.db.Update(fn)
}

// View implements StorageInterface.View
func (b *BoltDB) View(fn func(*bbolt.Tx) error) error {
	return b.db.View(fn)
}

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

	db.blocks[string(block.Header.Hash)] = block
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
