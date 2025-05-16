package sync

import (
	"fmt"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/utxo"
	"github.com/youngchain/internal/logger"
)

const (
	// MaxBlockSize is the maximum size of a block in bytes
	MaxBlockSize = 1000000 // 1MB

	// MaxFutureBlockTime is the maximum time a block can be in the future
	MaxFutureBlockTime = 7200 // 2 hours

	// BlockSyncTimeout is the timeout for block synchronization
	BlockSyncTimeout = 30 * time.Second

	// MaxBlocksPerRequest is the maximum number of blocks to request at once
	MaxBlocksPerRequest = 500
)

// BlockSync represents a block synchronizer
type BlockSync struct {
	mu sync.RWMutex

	// Blockchain
	blockchain *block.Blockchain

	// UTXO set
	utxoSet utxo.UTXOSetInterface

	// Known blocks
	knownBlocks map[string]*block.Block

	// Block queue
	blockQueue chan *block.Block

	// Stop channel
	stopChan chan struct{}

	// Sync status
	isSyncing bool

	// Last sync time
	lastSyncTime time.Time

	// Sync height
	syncHeight uint64
}

// NewBlockSync creates a new block synchronizer
func NewBlockSync(blockchain *block.Blockchain, utxoSet utxo.UTXOSetInterface) *BlockSync {
	return &BlockSync{
		blockchain:   blockchain,
		utxoSet:      utxoSet,
		knownBlocks:  make(map[string]*block.Block),
		blockQueue:   make(chan *block.Block, 1000),
		stopChan:     make(chan struct{}),
		lastSyncTime: time.Now(),
	}
}

// Start starts the block synchronizer
func (bs *BlockSync) Start() {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if bs.isSyncing {
		return
	}

	bs.isSyncing = true
	go bs.syncLoop()
}

// Stop stops the block synchronizer
func (bs *BlockSync) Stop() {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if !bs.isSyncing {
		return
	}

	bs.isSyncing = false
	close(bs.stopChan)
}

// HandleBlock handles an incoming block
func (bs *BlockSync) HandleBlock(block *block.Block) error {
	// Validate block
	if err := bs.validateBlock(block); err != nil {
		return fmt.Errorf("invalid block: %v", err)
	}

	// Check if block is already known
	bs.mu.RLock()
	_, exists := bs.knownBlocks[string(block.Hash)]
	bs.mu.RUnlock()
	if exists {
		return nil
	}

	// Add block to queue
	select {
	case bs.blockQueue <- block:
		return nil
	default:
		return fmt.Errorf("block queue is full")
	}
}

// syncLoop is the main synchronization loop
func (bs *BlockSync) syncLoop() {
	for {
		select {
		case <-bs.stopChan:
			return
		case block := <-bs.blockQueue:
			if err := bs.processBlock(block); err != nil {
				logger.Error("Failed to process block", logger.Error(err))
				continue
			}
		}
	}
}

// processBlock processes a block from the queue
func (bs *BlockSync) processBlock(block *block.Block) error {
	// Get current best block
	bestBlock := bs.blockchain.GetBestBlock()
	if bestBlock == nil {
		return fmt.Errorf("no best block found")
	}

	// Check if block is already in the chain
	if block.Height <= bestBlock.Height {
		return nil
	}

	// Check if block is too far in the future
	if block.Height > bestBlock.Height+MaxBlocksPerRequest {
		return fmt.Errorf("block height too far in the future")
	}

	// Add block to chain
	if err := bs.blockchain.AddBlock(block); err != nil {
		return fmt.Errorf("failed to add block: %v", err)
	}

	// Update UTXO set
	if err := bs.updateUTXOSet(block); err != nil {
		return fmt.Errorf("failed to update UTXO set: %v", err)
	}

	// Update known blocks
	bs.mu.Lock()
	bs.knownBlocks[string(block.Hash)] = block
	bs.mu.Unlock()

	// Update sync height
	bs.syncHeight = block.Height

	return nil
}

// validateBlock validates a block
func (bs *BlockSync) validateBlock(block *block.Block) error {
	// Check block size
	if block.GetBlockSize() > MaxBlockSize {
		return fmt.Errorf("block size exceeds maximum")
	}

	// Check block time
	if block.Timestamp > time.Now().Unix()+MaxFutureBlockTime {
		return fmt.Errorf("block timestamp too far in the future")
	}

	// Validate block
	if err := block.Validate(); err != nil {
		return fmt.Errorf("invalid block: %v", err)
	}

	return nil
}

// updateUTXOSet updates the UTXO set with a new block
func (bs *BlockSync) updateUTXOSet(block *block.Block) error {
	// Process each transaction
	for _, tx := range block.Transactions {
		// Remove spent UTXOs
		for _, input := range tx.Inputs {
			if err := bs.utxoSet.RemoveUTXO(input.PreviousTxHash, input.PreviousTxIndex); err != nil {
				return fmt.Errorf("failed to remove UTXO: %v", err)
			}
		}

		// Add new UTXOs
		for i, output := range tx.Outputs {
			utxo := &utxo.UTXO{
				TxHash:      tx.Hash,
				OutIndex:    uint32(i),
				Amount:      output.Value,
				ScriptPub:   output.ScriptPubKey,
				BlockHeight: block.Height,
				IsCoinbase:  tx.IsCoinbase(),
			}
			if err := bs.utxoSet.AddUTXO(utxo); err != nil {
				return fmt.Errorf("failed to add UTXO: %v", err)
			}
		}
	}

	return nil
}

// GetSyncStatus returns the current synchronization status
func (bs *BlockSync) GetSyncStatus() (bool, uint64, time.Time) {
	bs.mu.RLock()
	defer bs.mu.RUnlock()

	return bs.isSyncing, bs.syncHeight, bs.lastSyncTime
}

// RequestBlocks requests blocks from peers
func (bs *BlockSync) RequestBlocks(startHeight uint64) error {
	// Get current best block
	bestBlock := bs.blockchain.GetBestBlock()
	if bestBlock == nil {
		return fmt.Errorf("no best block found")
	}

	// Calculate end height
	endHeight := startHeight + MaxBlocksPerRequest
	if endHeight > bestBlock.Height {
		endHeight = bestBlock.Height
	}

	// TODO: Implement peer communication to request blocks
	// This would typically involve:
	// 1. Selecting peers to request from
	// 2. Sending getblocks message
	// 3. Handling block responses
	// 4. Validating and processing received blocks

	return nil
}
