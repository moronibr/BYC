package block

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/types"
)

// BlockType represents the type of a block
type BlockType string

const (
	GoldenBlock BlockType = "golden"
	SilverBlock BlockType = "silver"
)

// Blockchain represents the blockchain
type Blockchain struct {
	Chain      []*Block
	Difficulty uint32
	CoinType   coin.Type
	mu         sync.RWMutex
}

// Config represents blockchain configuration
type Config struct {
	Difficulty    uint32
	CoinType      coin.Type
	BlockInterval time.Duration
	MaxBlockSize  int
	BlockReward   int64
}

// NewBlockchain creates a new blockchain with the given configuration
func NewBlockchain(config *Config) *Blockchain {
	bc := &Blockchain{
		Chain:      make([]*Block, 0),
		Difficulty: config.Difficulty,
		CoinType:   config.CoinType,
	}

	// Create genesis block
	genesisBlock := NewBlock(nil, uint64(time.Now().Unix()))
	genesisBlock.Difficulty = config.Difficulty
	genesisBlock.CoinType = config.CoinType

	// Add genesis block to chain
	bc.Chain = append(bc.Chain, genesisBlock)

	return bc
}

// AddBlock adds a block to the blockchain
func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Validate block
	if err := block.Validate(); err != nil {
		return fmt.Errorf("invalid block: %v", err)
	}

	// Check if block is already in the chain
	for _, b := range bc.Chain {
		if bytes.Equal(b.Hash, block.Hash) {
			return fmt.Errorf("block already exists")
		}
	}

	// Add block to chain
	bc.Chain = append(bc.Chain, block)

	return nil
}

// GetBestBlock returns the latest block in the chain
func (bc *Blockchain) GetBestBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if len(bc.Chain) == 0 {
		return nil
	}
	return bc.Chain[len(bc.Chain)-1]
}

// GetBlockByHash returns a block by its hash
func (bc *Blockchain) GetBlockByHash(hash []byte) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for _, block := range bc.Chain {
		if bytes.Equal(block.Hash, hash) {
			return block
		}
	}
	return nil
}

// GetBlockByHeight returns a block by its height
func (bc *Blockchain) GetBlockByHeight(height uint64) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if height >= uint64(len(bc.Chain)) {
		return nil
	}
	return bc.Chain[height]
}

// GetBlockCount returns the number of blocks in the chain
func (bc *Blockchain) GetBlockCount() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return len(bc.Chain)
}

// validateBlock validates a block before adding it to the chain
func (bc *Blockchain) validateBlock(block *Block) error {
	// Check if block is nil
	if block == nil {
		return fmt.Errorf("block is nil")
	}

	// Check if block has valid previous hash
	if len(bc.Chain) > 0 {
		lastBlock := bc.Chain[len(bc.Chain)-1]
		if string(block.PreviousHash) != string(lastBlock.Hash) {
			return fmt.Errorf("invalid previous hash")
		}
	}

	// Check if block has valid timestamp
	if len(bc.Chain) > 0 {
		lastBlock := bc.Chain[len(bc.Chain)-1]
		if block.Timestamp <= lastBlock.Timestamp {
			return fmt.Errorf("invalid timestamp")
		}
	}

	// Check if block has valid difficulty
	if block.Difficulty != bc.Difficulty {
		return fmt.Errorf("invalid difficulty")
	}

	// Check if block has valid coin type
	if block.CoinType != bc.CoinType {
		return fmt.Errorf("invalid coin type")
	}

	// Validate block
	if err := block.Validate(); err != nil {
		return fmt.Errorf("invalid block: %v", err)
	}

	return nil
}

// GetBlockHash returns the hash of a block at the given height
func (bc *Blockchain) GetBlockHash(height uint64) []byte {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if height >= uint64(len(bc.Chain)) {
		return nil
	}
	return bc.Chain[height].Hash
}

// GetBlockHeader returns the header of a block at the given height
func (bc *Blockchain) GetBlockHeader(height uint64) *types.BlockHeader {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if height >= uint64(len(bc.Chain)) {
		return nil
	}
	return bc.Chain[height].Header
}

// GetBlockTransactions returns the transactions of a block at the given height
func (bc *Blockchain) GetBlockTransactions(height uint64) []*types.Transaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if height >= uint64(len(bc.Chain)) {
		return nil
	}
	return bc.Chain[height].Transactions
}

// GetBlockSize returns the size of a block at the given height
func (bc *Blockchain) GetBlockSize(height uint64) int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if height >= uint64(len(bc.Chain)) {
		return 0
	}
	return bc.Chain[height].Size
}

// GetBlockTime returns the timestamp of a block at the given height
func (bc *Blockchain) GetBlockTime(height uint64) uint64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if height >= uint64(len(bc.Chain)) {
		return 0
	}
	return bc.Chain[height].Timestamp
}

// GetBlockDifficulty returns the difficulty of a block at the given height
func (bc *Blockchain) GetBlockDifficulty(height uint64) uint32 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if height >= uint64(len(bc.Chain)) {
		return 0
	}
	return bc.Chain[height].Difficulty
}

// GetBlockCoinType returns the coin type of a block at the given height
func (bc *Blockchain) GetBlockCoinType(height uint64) coin.Type {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if height >= uint64(len(bc.Chain)) {
		return ""
	}
	return bc.Chain[height].CoinType
}

// GetBlocks returns all blocks
func (bc *Blockchain) GetBlocks() []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.Chain
}

// GetBlocksByType returns blocks of a specific type
func (bc *Blockchain) GetBlocksByType(blockType types.BlockType) []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	blocks := make([]*Block, 0)
	for _, block := range bc.Chain {
		if block.Header.Type == blockType {
			blocks = append(blocks, block)
		}
	}

	return blocks
}

// GetTransactionByHash returns a transaction by its hash
func (bc *Blockchain) GetTransactionByHash(hash []byte) *types.Transaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for _, block := range bc.Chain {
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.Hash, hash) {
				return tx
			}
		}
	}

	return nil
}

// GetTransactionsByAddress returns transactions for a specific address
func (bc *Blockchain) GetTransactionsByAddress(address string) []*types.Transaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	transactions := make([]*types.Transaction, 0)
	for _, block := range bc.Chain {
		for _, tx := range block.Transactions {
			for _, input := range tx.Inputs {
				if input.Address == address {
					transactions = append(transactions, tx)
					break
				}
			}
			for _, output := range tx.Outputs {
				if output.Address == address {
					transactions = append(transactions, tx)
					break
				}
			}
		}
	}

	return transactions
}

// GetAllBlocks returns all blocks in the chain
func (bc *Blockchain) GetAllBlocks() []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	blocks := make([]*Block, len(bc.Chain))
	copy(blocks, bc.Chain)
	return blocks
}
