package mining

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/common"
	"github.com/youngchain/internal/core/storage"
	"github.com/youngchain/internal/core/transaction"
	"github.com/youngchain/internal/core/types"
	"github.com/youngchain/internal/core/utxo"
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
	// UTXO set
	utxoSet utxo.UTXOSetInterface
	// Mining address
	miningAddress string
	// Mining status
	isMining bool
	// Mining channel
	stopChan chan struct{}
}

// NewMiner creates a new miner
func NewMiner(txPool *transaction.TxPool, blockStore storage.BlockStore, utxoSet utxo.UTXOSetInterface, miningAddress string) *Miner {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-TargetBits))

	return &Miner{
		target:        target,
		txPool:        txPool,
		blockStore:    blockStore,
		utxoSet:       utxoSet,
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

// mine performs the mining process
func (m *Miner) mine() {
	for m.isMining {
		// Create new block
		block, err := m.createBlock()
		if err != nil {
			fmt.Printf("Error creating block: %v\n", err)
			continue
		}

		// Set current block
		m.currentBlock = block

		// Mine the block
		nonce, err := m.mineBlock(block)
		if err != nil {
			fmt.Printf("Error mining block: %v\n", err)
			m.currentBlock = nil
			continue
		}

		// Set the nonce
		block.Header.Nonce = nonce

		// Save the block
		if err := m.blockStore.PutBlock(block); err != nil {
			fmt.Printf("Error saving block: %v\n", err)
			m.currentBlock = nil
			continue
		}

		// Update UTXO set
		if err := m.updateUTXOSet(block); err != nil {
			fmt.Printf("Error updating UTXO set: %v\n", err)
			m.currentBlock = nil
			continue
		}

		// Clear current block after successful mining
		m.currentBlock = nil
	}
}

// GetCurrentBlock returns the current block being mined
func (m *Miner) GetCurrentBlock() *block.Block {
	return m.currentBlock
}

// StopMining stops the mining proscess
func (m *Miner) StopMining() {
	if !m.isMining {
		return
	}

	m.isMining = false
	m.currentBlock = nil
	close(m.stopChan)
}

// createCoinbaseTx creates a new coinbase transaction
func createCoinbaseTx(minerAddress []byte, blockHeight uint64) *common.Transaction {
	// Create a new transaction with empty data
	tx := &types.Transaction{
		Version:   1,
		Timestamp: time.Now(),
		Data:      []byte(fmt.Sprintf("Coinbase transaction for block %d", blockHeight)),
		Inputs:    make([]*types.TxInput, 1),
		Outputs:   make([]*types.TxOutput, 1),
		Witness:   make([][]byte, 0),
		LockTime:  0,
		Fee:       0,
		Hash:      make([]byte, 0),
		CoinType:  coin.Leah,
	}

	// Set up coinbase input
	tx.Inputs[0] = &types.TxInput{
		PreviousTxHash:  make([]byte, 0),
		PreviousTxIndex: 0xffffffff,
		ScriptSig:       []byte(fmt.Sprintf("Coinbase transaction for block %d", blockHeight)),
		Sequence:        0xffffffff,
		Address:         "",
	}

	// Set up mining reward output
	tx.Outputs[0] = &types.TxOutput{
		Value:        BlockReward,
		ScriptPubKey: []byte(fmt.Sprintf("Mining reward for block %d", blockHeight)),
		Address:      string(minerAddress),
	}

	// Calculate hash
	tx.CalculateHash()

	// Create common transaction wrapper
	commonTx := &common.Transaction{
		Tx: tx,
	}

	return commonTx
}

// createBlock creates a new block to mine
func (m *Miner) createBlock() (*block.Block, error) {
	// Get the last block
	lastBlock, err := m.blockStore.GetLastBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get last block: %v", err)
	}

	// Create coinbase transaction first
	coinbaseTx := createCoinbaseTx([]byte(m.miningAddress), lastBlock.Header.Height+1)
	if coinbaseTx == nil || coinbaseTx.Tx == nil {
		return nil, fmt.Errorf("failed to create coinbase transaction")
	}

	// Initialize transactions slice with coinbase
	txs := []*types.Transaction{coinbaseTx.Tx}

	// Get transactions from pool
	poolTxs := m.txPool.GetBest(1000) // Get up to 1000 transactions
	if len(poolTxs) > 0 {
		// Filter out nil transactions
		for _, tx := range poolTxs {
			if tx != nil {
				txs = append(txs, tx)
			}
		}
	}

	// Convert []*types.Transaction to []*common.Transaction
	commonTxs := make([]*common.Transaction, 0, len(txs))
	for _, tx := range txs {
		if tx == nil {
			continue
		}

		// Create a new common transaction with the existing transaction
		commonTx := &common.Transaction{
			Tx: tx,
		}

		// Ensure the transaction is properly initialized
		if commonTx.Tx.Inputs == nil {
			commonTx.Tx.Inputs = make([]*types.TxInput, 0)
		}
		if commonTx.Tx.Outputs == nil {
			commonTx.Tx.Outputs = make([]*types.TxOutput, 0)
		}
		if commonTx.Tx.Witness == nil {
			commonTx.Tx.Witness = make([][]byte, 0)
		}
		if commonTx.Tx.Data == nil {
			commonTx.Tx.Data = make([]byte, 0)
		}

		// Calculate hash if not already set
		if commonTx.Tx.Hash == nil {
			commonTx.Tx.CalculateHash()
		}

		commonTxs = append(commonTxs, commonTx)
	}

	// Ensure we have at least the coinbase transaction
	if len(commonTxs) == 0 {
		return nil, fmt.Errorf("no valid transactions for block")
	}

	// Calculate Merkle root
	merkleRoot := calculateMerkleRoot(txs)
	if merkleRoot == nil {
		return nil, fmt.Errorf("failed to calculate Merkle root")
	}

	// Create block header
	header := &common.Header{
		Version:       1,
		PrevBlockHash: lastBlock.Header.Hash,
		MerkleRoot:    merkleRoot,
		Timestamp:     time.Now(),
		Difficulty:    TargetBits,
		Height:        lastBlock.Header.Height + 1,
	}

	// Create block
	newBlock := &block.Block{
		Header:       header,
		Transactions: commonTxs,
		BlockSize:    0, // Will be calculated when needed
		Weight:       0, // Will be calculated when needed
		IsValid:      true,
	}

	// Calculate block hash
	newBlock.Header.Hash = newBlock.CalculateHash()
	if newBlock.Header.Hash == nil {
		return nil, fmt.Errorf("failed to calculate block hash")
	}

	return newBlock, nil
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
	leaves := make([][]byte, 0, len(txs))
	for _, tx := range txs {
		if tx == nil {
			continue
		}

		// Ensure transaction hash is calculated
		if tx.Hash == nil {
			tx.CalculateHash()
		}

		// Skip if hash is still nil
		if tx.Hash == nil {
			continue
		}

		// Create a copy of the hash to avoid modifying the original
		hashCopy := make([]byte, len(tx.Hash))
		copy(hashCopy, tx.Hash)
		leaves = append(leaves, hashCopy)
	}

	// If no valid leaves, return nil
	if len(leaves) == 0 {
		return nil
	}

	// Build tree
	for len(leaves) > 1 {
		var newLeaves [][]byte
		for i := 0; i < len(leaves); i += 2 {
			if i+1 == len(leaves) {
				// If odd number of leaves, duplicate the last one
				lastLeaf := make([]byte, len(leaves[i]))
				copy(lastLeaf, leaves[i])
				newLeaves = append(newLeaves, lastLeaf)
				continue
			}

			// Concatenate and hash pairs of leaves
			concat := make([]byte, len(leaves[i])+len(leaves[i+1]))
			copy(concat, leaves[i])
			copy(concat[len(leaves[i]):], leaves[i+1])
			hash := sha256.Sum256(concat)
			newLeaves = append(newLeaves, hash[:])
		}
		leaves = newLeaves
	}

	// Return the root hash
	return leaves[0]
}

// updateUTXOSet updates the UTXO set with new block
func (m *Miner) updateUTXOSet(block *block.Block) error {
	for _, tx := range block.Transactions {
		underlyingTx := tx.GetTransaction()
		// Remove spent UTXOs for each input
		for _, input := range underlyingTx.Inputs {
			m.utxoSet.RemoveUTXO(input.PreviousTxHash, input.PreviousTxIndex)
		}
		// Add new UTXOs for each output
		for idx, output := range underlyingTx.Outputs {
			utxoObj := &utxo.UTXO{
				TxHash:      underlyingTx.Hash,
				OutIndex:    uint32(idx),
				Amount:      output.Value,
				ScriptPub:   nil, // TODO: Convert output.ScriptPubKey to *script.Script
				BlockHeight: block.Header.Height,
				IsCoinbase:  false, // Set true for coinbase tx if needed
				IsSegWit:    false, // Set if SegWit
			}
			m.utxoSet.AddUTXO(utxoObj)
		}
	}
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
