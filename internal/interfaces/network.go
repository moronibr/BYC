package interfaces

import (
	"github.com/youngchain/internal/core/block"
)

// Network defines the interface for network operations
type Network interface {
	// GetPeerHeights returns the block heights of all connected peers
	GetPeerHeights() []uint64

	// RequestBlocks requests blocks from peers in the given height range
	RequestBlocks(startHeight, endHeight uint64) ([]*block.Block, error)

	// BroadcastBlock broadcasts a block to all peers
	BroadcastBlock(block *block.Block) error

	// BroadcastTransaction broadcasts a transaction to all peers
	BroadcastTransaction(tx interface{}) error
}
