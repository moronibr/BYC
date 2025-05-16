package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

// Block represents a block in the blockchain
type Block struct {
	Number    uint64
	Hash      []byte
	PrevHash  []byte
	Data      []byte
	Timestamp int64
	// Add other block fields as needed
}

// Storage defines the interface for blockchain storage
type Storage interface {
	GetBlockByHash(hash []byte) (*Block, error)
	GetBlockByNumber(number uint64) (*Block, error)
	GetLatestBlock() (*Block, error)
	StoreBlock(block *Block) error
	// Add other storage methods as needed
}

type Blockchain struct {
	mu         sync.RWMutex
	blockCache map[string]*Block
	storage    Storage
}

// NewBlockchain creates a new blockchain instance
func NewBlockchain(storage Storage) *Blockchain {
	return &Blockchain{
		blockCache: make(map[string]*Block),
		storage:    storage,
	}
}

// GetBlockByHash retrieves a block by its hash
func (bc *Blockchain) GetBlockByHash(hash []byte) (*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Check if block exists in cache
	if block, exists := bc.blockCache[string(hash)]; exists {
		return block, nil
	}

	// If not in cache, try to get from storage
	block, err := bc.storage.GetBlockByHash(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by hash: %w", err)
	}

	// Cache the block for future use
	bc.blockCache[string(hash)] = block
	return block, nil
}

// GetBlockByNumber retrieves a block by its number
func (bc *Blockchain) GetBlockByNumber(number uint64) (*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Check if block exists in cache
	if block, exists := bc.blockCache[fmt.Sprintf("%d", number)]; exists {
		return block, nil
	}

	// If not in cache, try to get from storage
	block, err := bc.storage.GetBlockByNumber(number)
	if err != nil {
		return nil, fmt.Errorf("failed to get block by number: %w", err)
	}

	// Cache the block for future use
	bc.blockCache[fmt.Sprintf("%d", number)] = block
	return block, nil
}

// GetLatestBlock returns the most recent block in the chain
func (bc *Blockchain) GetLatestBlock() (*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Try to get from storage
	block, err := bc.storage.GetLatestBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %w", err)
	}

	// Cache the block for future use
	bc.blockCache[string(block.Hash)] = block
	return block, nil
}

// GetBlockRange retrieves a range of blocks from start to end (inclusive)
func (bc *Blockchain) GetBlockRange(start, end uint64) ([]*Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if start > end {
		return nil, fmt.Errorf("invalid range: start (%d) > end (%d)", start, end)
	}

	blocks := make([]*Block, 0, end-start+1)
	for i := start; i <= end; i++ {
		block, err := bc.GetBlockByNumber(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get block %d: %w", i, err)
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

// BlockHeader represents the header of a block
type BlockHeader struct {
	Number    uint64
	Hash      []byte
	PrevHash  []byte
	Timestamp int64
	// Add other header fields as needed
}

// GetBlockHeaders retrieves headers for a range of blocks
func (bc *Blockchain) GetBlockHeaders(start, end uint64) ([]*BlockHeader, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if start > end {
		return nil, fmt.Errorf("invalid range: start (%d) > end (%d)", start, end)
	}

	headers := make([]*BlockHeader, 0, end-start+1)
	for i := start; i <= end; i++ {
		block, err := bc.GetBlockByNumber(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get block %d: %w", i, err)
		}

		header := &BlockHeader{
			Number:    i,
			Hash:      block.Hash,
			PrevHash:  block.PrevHash,
			Timestamp: 0, // TODO: Add timestamp to Block struct
		}
		headers = append(headers, header)
	}

	return headers, nil
}

// GetBlockHashes retrieves hashes for a range of blocks
func (bc *Blockchain) GetBlockHashes(start, end uint64) ([][]byte, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if start > end {
		return nil, fmt.Errorf("invalid range: start (%d) > end (%d)", start, end)
	}

	hashes := make([][]byte, 0, end-start+1)
	for i := start; i <= end; i++ {
		block, err := bc.GetBlockByNumber(i)
		if err != nil {
			return nil, fmt.Errorf("failed to get block %d: %w", i, err)
		}
		hashes = append(hashes, block.Hash)
	}

	return hashes, nil
}

// AddBlock adds a new block to the chain
func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Validate the block
	if err := bc.validateBlock(block); err != nil {
		return fmt.Errorf("invalid block: %w", err)
	}

	// Store the block
	if err := bc.storage.StoreBlock(block); err != nil {
		return fmt.Errorf("failed to store block: %w", err)
	}

	// Update cache
	bc.blockCache[string(block.Hash)] = block
	bc.blockCache[fmt.Sprintf("%d", block.Number)] = block

	return nil
}

// ValidateChain validates the entire blockchain
func (bc *Blockchain) ValidateChain() error {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	latestBlock, err := bc.GetLatestBlock()
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}

	currentBlock := latestBlock
	for currentBlock != nil {
		if err := bc.validateBlock(currentBlock); err != nil {
			return fmt.Errorf("invalid block at height %d: %w", currentBlock.Number, err)
		}

		if currentBlock.Number == 0 {
			break // Reached genesis block
		}

		prevBlock, err := bc.GetBlockByHash(currentBlock.PrevHash)
		if err != nil {
			return fmt.Errorf("failed to get previous block: %w", err)
		}

		currentBlock = prevBlock
	}

	return nil
}

// validateBlock performs basic validation on a block
func (bc *Blockchain) validateBlock(block *Block) error {
	if block == nil {
		return fmt.Errorf("block is nil")
	}

	if block.Number == 0 {
		// Genesis block validation
		if len(block.PrevHash) != 0 {
			return fmt.Errorf("genesis block should have empty previous hash")
		}
	} else {
		// Regular block validation
		if len(block.PrevHash) == 0 {
			return fmt.Errorf("block must have a previous hash")
		}

		prevBlock, err := bc.GetBlockByHash(block.PrevHash)
		if err != nil {
			return fmt.Errorf("failed to get previous block: %w", err)
		}

		if prevBlock.Number+1 != block.Number {
			return fmt.Errorf("invalid block number: expected %d, got %d", prevBlock.Number+1, block.Number)
		}
	}

	return nil
}

// CreateGenesisBlock creates the first block in the chain
func (bc *Blockchain) CreateGenesisBlock(data []byte) (*Block, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Check if genesis block already exists
	if _, err := bc.GetBlockByNumber(0); err == nil {
		return nil, fmt.Errorf("genesis block already exists")
	}

	genesisBlock := &Block{
		Number:    0,
		Hash:      []byte("genesis"), // TODO: Implement proper hash calculation
		PrevHash:  []byte{},
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	if err := bc.AddBlock(genesisBlock); err != nil {
		return nil, fmt.Errorf("failed to add genesis block: %w", err)
	}

	return genesisBlock, nil
}

// CalculateBlockHash calculates the hash of a block
func (bc *Blockchain) CalculateBlockHash(block *Block) ([]byte, error) {
	if block == nil {
		return nil, fmt.Errorf("block is nil")
	}

	// Create a deterministic byte slice for hashing
	data := make([]byte, 0)
	data = append(data, block.PrevHash...)
	data = append(data, block.Data...)

	// Add number and timestamp
	numBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(numBytes, block.Number)
	data = append(data, numBytes...)

	timeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timeBytes, uint64(block.Timestamp))
	data = append(data, timeBytes...)

	// Calculate SHA-256 hash
	hash := sha256.Sum256(data)
	return hash[:], nil
}
