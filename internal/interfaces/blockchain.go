package interfaces

import (
	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/types"
)

// BlockChain defines the interface for blockchain operations
type BlockChain interface {
	GetLastBlock() (*block.Block, error)
	GetBlockByHeight(height uint64) (*block.Block, error)
	GetBlock(hash []byte) (*block.Block, error)
	AddBlock(block *block.Block) error
	GetPendingTransactions() []*types.Transaction
}
