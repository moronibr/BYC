package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/common"
	"go.etcd.io/bbolt"
)

// Storage represents the blockchain storage
type Storage struct {
	db     *bbolt.DB
	mu     sync.RWMutex
	config *Config
}

// Config holds the storage configuration
type Config struct {
	Path string
}

// NewStorage creates a new blockchain storage
func NewStorage(config *Config) (*Storage, error) {
	db, err := bbolt.Open(config.Path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Create necessary buckets
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("blocks"))
		if err != nil {
			return fmt.Errorf("failed to create blocks bucket: %v", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte("transactions"))
		if err != nil {
			return fmt.Errorf("failed to create transactions bucket: %v", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte("state"))
		if err != nil {
			return fmt.Errorf("failed to create state bucket: %v", err)
		}

		return nil
	})

	if err != nil {
		db.Close()
		return nil, err
	}

	return &Storage{
		db:     db,
		config: config,
	}, nil
}

// StoreBlock stores a block in the database
func (s *Storage) StoreBlock(block *block.Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bbolt.Tx) error {
		// Store block
		blocksBucket := tx.Bucket([]byte("blocks"))
		blockData, err := json.Marshal(block)
		if err != nil {
			return err
		}

		// Use block height as key
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, block.Header.Height)

		if err := blocksBucket.Put(key, blockData); err != nil {
			return err
		}

		// Store transactions
		txBucket := tx.Bucket([]byte("transactions"))
		for _, tx := range block.Transactions {
			txData, err := json.Marshal(tx)
			if err != nil {
				return err
			}

			if err := txBucket.Put(tx.Hash(), txData); err != nil {
				return err
			}
		}

		return nil
	})
}

// GetBlock retrieves a block by height
func (s *Storage) GetBlock(height uint64) (*block.Block, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var blockData []byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		blocksBucket := tx.Bucket([]byte("blocks"))
		key := make([]byte, 8)
		binary.BigEndian.PutUint64(key, height)

		blockData = blocksBucket.Get(key)
		return nil
	})

	if err != nil {
		return nil, err
	}

	if blockData == nil {
		return nil, fmt.Errorf("block not found at height %d", height)
	}

	var block block.Block
	if err := json.Unmarshal(blockData, &block); err != nil {
		return nil, err
	}

	return &block, nil
}

// GetTransaction retrieves a transaction by hash
func (s *Storage) GetTransaction(hash []byte) (*common.Transaction, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var txData []byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		txBucket := tx.Bucket([]byte("transactions"))
		txData = txBucket.Get(hash)
		return nil
	})

	if err != nil {
		return nil, err
	}

	if txData == nil {
		return nil, fmt.Errorf("transaction not found")
	}

	var tx common.Transaction
	if err := json.Unmarshal(txData, &tx); err != nil {
		return nil, err
	}

	return &tx, nil
}

// UpdateState updates the blockchain state
func (s *Storage) UpdateState(key []byte, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bbolt.Tx) error {
		stateBucket := tx.Bucket([]byte("state"))
		return stateBucket.Put(key, value)
	})
}

// GetState retrieves a value from the blockchain state
func (s *Storage) GetState(key []byte) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var value []byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		stateBucket := tx.Bucket([]byte("state"))
		value = stateBucket.Get(key)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return value, nil
}

// Close closes the storage
func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Close()
}
