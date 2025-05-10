package mining

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/common"
	"github.com/youngchain/internal/core/storage"
	"github.com/youngchain/internal/core/transaction"
	"github.com/youngchain/internal/core/types"
)

const (
	// Mining parameters
	TargetBits                   = 24
	MaxNonce                     = math.MaxInt64
	BlockReward                  = 50 // Initial block reward in base units
	HalvingInterval              = 210000
	DifficultyAdjustmentInterval = 2016
	TargetTimePerBlock           = 10 * time.Minute
)

// Miner represents a blockchain miner
type Miner struct {
	// Current block being mined
	currentBlock *block.Block
	// Target difficulty
	target *big.Int
	// Transaction pool
	txPool *transaction.TxPool
	// Block store
	blockStore storage.BlockStore
	// Mining address
	miningAddress string
	// Mining status
	isMining bool
	// Mining channel
	stopChan chan struct{}
}

// NewMiner creates a new miner
func NewMiner(txPool *transaction.TxPool, blockStore storage.BlockStore, miningAddress string) *Miner {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-TargetBits))

	return &Miner{
		target:        target,
		txPool:        txPool,
		blockStore:    blockStore,
		miningAddress: miningAddress,
		stopChan:      make(chan struct{}),
	}
}

// StartMining starts the mining process
func (m *Miner) StartMining() {
	if m.isMining {
		return
	}

	m.isMining = true
	go m.mine()
}

// StopMining stops the mining process
func (m *Miner) StopMining() {
	if !m.isMining {
		return
	}

	m.isMining = false
	close(m.stopChan)
}

// mine performs the mining process
func (m *Miner) mine() {
	for m.isMining {
		// Create new block
		block, err := m.createBlock()
		if err != nil {
			fmt.Printf("Error creating block: %v\n", err)
			continue
		}

		// Mine the block
		nonce, err := m.mineBlock(block)
		if err != nil {
			fmt.Printf("Error mining block: %v\n", err)
			continue
		}

		// Set the nonce
		block.Header.Nonce = nonce

		// Save the block
		if err := m.blockStore.PutBlock(block); err != nil {
			fmt.Printf("Error saving block: %v\n", err)
			continue
		}

		// Update UTXO set
		if err := m.updateUTXOSet(block); err != nil {
			fmt.Printf("Error updating UTXO set: %v\n", err)
			continue
		}
	}
}

// createBlock creates a new block to mine
func (m *Miner) createBlock() (*block.Block, error) {
	// Get the last block
	lastBlock, err := m.blockStore.GetLastBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get last block: %v", err)
	}

	// Get transactions from pool
	txs := m.txPool.GetBest(1000) // Get up to 1000 transactions

	// Create coinbase transaction
	coinbaseTx, err := m.createCoinbaseTx(lastBlock.Header.Height + 1)
	if err != nil {
		return nil, fmt.Errorf("failed to create coinbase transaction: %v", err)
	}

	// Add coinbase transaction to the beginning
	txs = append([]*types.Transaction{coinbaseTx}, txs...)

	// Convert []*types.Transaction to []*common.Transaction
	commonTxs := make([]*common.Transaction, len(txs))
	for i, tx := range txs {
		commonTxs[i] = &common.Transaction{
			Hash:      tx.Hash,
			Version:   tx.Version,
			Timestamp: tx.Timestamp,
			From:      tx.From,
			To:        tx.To,
			Amount:    tx.Amount,
			Data:      tx.Data,
		}
	}

	// Create block header
	header := &common.Header{
		Version:       1,
		PrevBlockHash: lastBlock.Header.Hash,
		MerkleRoot:    calculateMerkleRoot(txs),
		Timestamp:     time.Now(),
		Difficulty:    TargetBits,
		Height:        lastBlock.Header.Height + 1,
	}

	// Create block
	return &block.Block{
		Header:       header,
		Transactions: commonTxs,
		Size:         0, // Will be calculated when needed
		Weight:       0, // Will be calculated when needed
		IsValid:      true,
	}, nil
}

// mineBlock mines a block
func (m *Miner) mineBlock(block *block.Block) (uint32, error) {
	var hashInt big.Int
	var hash [32]byte
	nonce := uint32(0)

	for nonce < math.MaxUint32 {
		select {
		case <-m.stopChan:
			return 0, fmt.Errorf("mining stopped")
		default:
			// Update nonce
			block.Header.Nonce = nonce

			// Calculate hash
			hash = sha256.Sum256(block.CalculateHash())
			hashInt.SetBytes(hash[:])

			// Check if hash is less than target
			if hashInt.Cmp(m.target) == -1 {
				return nonce, nil
			}

			nonce++
		}
	}

	return 0, fmt.Errorf("max nonce reached")
}

// createCoinbaseTx creates a coinbase transaction
func (m *Miner) createCoinbaseTx(height uint64) (*types.Transaction, error) {
	// Calculate block reward
	reward := calculateBlockReward(height)

	// Create transaction
	return &types.Transaction{
		Version:   1,
		Timestamp: time.Now(),
		From:      []byte{}, // Coinbase has no inputs
		To:        []byte(m.miningAddress),
		Amount:    reward,
		Data:      []byte("coinbase"),
	}, nil
}

// calculateBlockReward calculates the block reward for a given height
func calculateBlockReward(height uint64) uint64 {
	halvings := height / HalvingInterval
	if halvings >= 64 {
		return 0
	}
	return BlockReward >> halvings
}

// calculateMerkleRoot calculates the Merkle root of transactions
func calculateMerkleRoot(txs []*types.Transaction) []byte {
	if len(txs) == 0 {
		return nil
	}

	// Create leaf nodes
	leaves := make([][]byte, len(txs))
	for i, tx := range txs {
		leaves[i] = tx.CalculateHash()
	}

	// Build tree
	for len(leaves) > 1 {
		var newLeaves [][]byte
		for i := 0; i < len(leaves); i += 2 {
			if i+1 == len(leaves) {
				newLeaves = append(newLeaves, leaves[i])
				continue
			}
			concat := append(leaves[i], leaves[i+1]...)
			hash := sha256.Sum256(concat)
			newLeaves = append(newLeaves, hash[:])
		}
		leaves = newLeaves
	}

	return leaves[0]
}

// updateUTXOSet updates the UTXO set with new block
func (m *Miner) updateUTXOSet(block *block.Block) error {
	// TODO: Implement UTXO set update
	return nil
}

// adjustDifficulty adjusts the mining difficulty
func (m *Miner) adjustDifficulty() error {
	// Get last block
	lastBlock, err := m.blockStore.GetLastBlock()
	if err != nil {
		return fmt.Errorf("failed to get last block: %v", err)
	}

	// Check if we need to adjust difficulty
	if lastBlock.Header.Height%DifficultyAdjustmentInterval != 0 {
		return nil
	}

	// TODO: Implement GetBlockByHeight or equivalent in BlockStore for difficulty adjustment
	// prevAdjustmentBlock, err := m.blockStore.GetBlockByHeight(lastBlock.Header.Height - DifficultyAdjustmentInterval)
	// if err != nil {
	// 	return fmt.Errorf("failed to get previous adjustment block: %v", err)
	// }

	// For now, skip actual adjustment logic
	return nil
}
