package block

import (
	"fmt"
	"sync"

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
	blocks    []*Block
	bestBlock *Block
	mu        sync.RWMutex
}

// NewBlockchain creates a new blockchain
func NewBlockchain() *Blockchain {
	return &Blockchain{
		blocks: make([]*Block, 0),
	}
}

// AddBlock adds a block to the blockchain
func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Validate block
	if err := block.Validate(); err != nil {
		return fmt.Errorf("invalid block: %v", err)
	}

	// Add block
	bc.blocks = append(bc.blocks, block)
	bc.bestBlock = block

	return nil
}

// GetBestBlock returns the best block
func (bc *Blockchain) GetBestBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.bestBlock
}

// GetBlockByHash returns a block by its hash
func (bc *Blockchain) GetBlockByHash(hash []byte) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for _, block := range bc.blocks {
		if block.Header.Hash == hash {
			return block
		}
	}

	return nil
}

// GetBlockByHeight returns a block by its height
func (bc *Blockchain) GetBlockByHeight(height uint64) *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for _, block := range bc.blocks {
		if block.Header.Height == height {
			return block
		}
	}

	return nil
}

// GetBlockCount returns the number of blocks
func (bc *Blockchain) GetBlockCount() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return len(bc.blocks)
}

// GetBlocks returns all blocks
func (bc *Blockchain) GetBlocks() []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.blocks
}

// GetBlocksByType returns blocks of a specific type
func (bc *Blockchain) GetBlocksByType(blockType BlockType) []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	blocks := make([]*Block, 0)
	for _, block := range bc.blocks {
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

	for _, block := range bc.blocks {
		for _, tx := range block.Transactions {
			if tx.Hash == hash {
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
	for _, block := range bc.blocks {
		for _, tx := range block.Transactions {
			for _, input := range tx.Inputs {
				if input.ScriptSig == address {
					transactions = append(transactions, tx)
					break
				}
			}
			for _, output := range tx.Outputs {
				if output.ScriptPubKey == address {
					transactions = append(transactions, tx)
					break
				}
			}
		}
	}

	return transactions
}
