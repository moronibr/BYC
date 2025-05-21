package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Storage handles persistent storage of blocks and transactions
type Storage struct {
	baseDir  string
	blockDir string
	txDir    string
	mu       sync.RWMutex
}

// NewStorage creates a new storage instance
func NewStorage(baseDir string) (*Storage, error) {
	// Create base directory
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %v", err)
	}

	// Create block and transaction directories
	blockDir := filepath.Join(baseDir, "blocks")
	txDir := filepath.Join(baseDir, "transactions")

	if err := os.MkdirAll(blockDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create block directory: %v", err)
	}
	if err := os.MkdirAll(txDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create transaction directory: %v", err)
	}

	return &Storage{
		baseDir:  baseDir,
		blockDir: blockDir,
		txDir:    txDir,
	}, nil
}

// SaveBlock saves a block to storage
func (s *Storage) SaveBlock(blockHash string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.blockDir, blockHash)
	return os.WriteFile(path, data, 0644)
}

// GetBlock retrieves a block from storage
func (s *Storage) GetBlock(blockHash string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.blockDir, blockHash)
	return os.ReadFile(path)
}

// DeleteBlock deletes a block from storage
func (s *Storage) DeleteBlock(blockHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.blockDir, blockHash)
	return os.Remove(path)
}

// SaveTransaction saves a transaction to storage
func (s *Storage) SaveTransaction(txID string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.txDir, txID)
	return os.WriteFile(path, data, 0644)
}

// GetTransaction retrieves a transaction from storage
func (s *Storage) GetTransaction(txID string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.txDir, txID)
	return os.ReadFile(path)
}

// DeleteTransaction deletes a transaction from storage
func (s *Storage) DeleteTransaction(txID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.txDir, txID)
	return os.Remove(path)
}

// ListBlocks lists all blocks in storage
func (s *Storage) ListBlocks() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.blockDir)
	if err != nil {
		return nil, err
	}

	var blocks []string
	for _, entry := range entries {
		if !entry.IsDir() {
			blocks = append(blocks, entry.Name())
		}
	}

	return blocks, nil
}

// ListTransactions lists all transactions in storage
func (s *Storage) ListTransactions() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.txDir)
	if err != nil {
		return nil, err
	}

	var transactions []string
	for _, entry := range entries {
		if !entry.IsDir() {
			transactions = append(transactions, entry.Name())
		}
	}

	return transactions, nil
}

// Close closes the storage
func (s *Storage) Close() error {
	return nil
}
