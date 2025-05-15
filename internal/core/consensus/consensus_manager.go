package consensus

import (
	"fmt"
	"sync"
	"time"
)

// ConsensusManager manages the consensus process
type ConsensusManager struct {
	mu sync.RWMutex

	// Validators
	blockValidator     *BlockValidator
	txValidator        *TransactionValidator
	difficultyAdjuster *DifficultyAdjuster

	// State
	currentBlock       *Block
	bestBlock          *Block
	utxoSet            UTXOSet
	orphanBlocks       map[[32]byte]*Block
	orphanTransactions map[[32]byte]*Transaction

	// Configuration
	maxOrphanBlocks int
	maxOrphanTxs    int
	reorgThreshold  int
}

// NewConsensusManager creates a new consensus manager
func NewConsensusManager(
	blockValidator *BlockValidator,
	txValidator *TransactionValidator,
	difficultyAdjuster *DifficultyAdjuster,
	utxoSet UTXOSet,
	maxOrphanBlocks,
	maxOrphanTxs,
	reorgThreshold int,
) *ConsensusManager {
	return &ConsensusManager{
		blockValidator:     blockValidator,
		txValidator:        txValidator,
		difficultyAdjuster: difficultyAdjuster,
		utxoSet:            utxoSet,
		orphanBlocks:       make(map[[32]byte]*Block),
		orphanTransactions: make(map[[32]byte]*Transaction),
		maxOrphanBlocks:    maxOrphanBlocks,
		maxOrphanTxs:       maxOrphanTxs,
		reorgThreshold:     reorgThreshold,
	}
}

// ProcessBlock processes a new block
func (cm *ConsensusManager) ProcessBlock(block *Block) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Validate block
	if err := cm.blockValidator.ValidateBlock(block, cm.bestBlock); err != nil {
		return fmt.Errorf("invalid block: %v", err)
	}

	// Check if block is orphan
	if cm.bestBlock != nil && block.PrevHash != cm.bestBlock.Hash {
		return cm.handleOrphanBlock(block)
	}

	// Process block transactions
	if err := cm.processBlockTransactions(block); err != nil {
		return fmt.Errorf("failed to process block transactions: %v", err)
	}

	// Update best block if needed
	if cm.bestBlock == nil || block.Height > cm.bestBlock.Height {
		cm.bestBlock = block
		cm.currentBlock = block
	}

	// Process orphan blocks that may now be valid
	if err := cm.processOrphanBlocks(); err != nil {
		return fmt.Errorf("failed to process orphan blocks: %v", err)
	}

	return nil
}

// ProcessTransaction processes a new transaction
func (cm *ConsensusManager) ProcessTransaction(tx *Transaction) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Validate transaction
	if err := cm.txValidator.ValidateTransaction(tx, time.Now().Unix()); err != nil {
		return fmt.Errorf("invalid transaction: %v", err)
	}

	// Check if transaction is orphan
	if err := cm.checkOrphanTransaction(tx); err != nil {
		return cm.handleOrphanTransaction(tx)
	}

	// Add transaction to current block
	if cm.currentBlock != nil {
		cm.currentBlock.Transactions = append(cm.currentBlock.Transactions, tx)
	}

	return nil
}

// handleOrphanBlock handles an orphan block
func (cm *ConsensusManager) handleOrphanBlock(block *Block) error {
	// Check if we have too many orphan blocks
	if len(cm.orphanBlocks) >= cm.maxOrphanBlocks {
		// Remove oldest orphan block
		var oldestBlock *Block
		var oldestTime int64
		for _, orphan := range cm.orphanBlocks {
			if oldestBlock == nil || orphan.Timestamp < oldestTime {
				oldestBlock = orphan
				oldestTime = orphan.Timestamp
			}
		}
		if oldestBlock != nil {
			delete(cm.orphanBlocks, oldestBlock.Hash)
		}
	}

	// Add block to orphan blocks
	cm.orphanBlocks[block.Hash] = block
	return fmt.Errorf("orphan block")
}

// handleOrphanTransaction handles an orphan transaction
func (cm *ConsensusManager) handleOrphanTransaction(tx *Transaction) error {
	// Check if we have too many orphan transactions
	if len(cm.orphanTransactions) >= cm.maxOrphanTxs {
		// Remove oldest orphan transaction
		var oldestTx *Transaction
		var oldestTime int64
		for _, orphan := range cm.orphanTransactions {
			if oldestTx == nil || orphan.Timestamp < oldestTime {
				oldestTx = orphan
				oldestTime = orphan.Timestamp
			}
		}
		if oldestTx != nil {
			delete(cm.orphanTransactions, oldestTx.Hash)
		}
	}

	// Add transaction to orphan transactions
	cm.orphanTransactions[tx.Hash] = tx
	return fmt.Errorf("orphan transaction")
}

// processOrphanBlocks processes orphan blocks that may now be valid
func (cm *ConsensusManager) processOrphanBlocks() error {
	for hash, block := range cm.orphanBlocks {
		if cm.bestBlock != nil && block.PrevHash == cm.bestBlock.Hash {
			if err := cm.ProcessBlock(block); err == nil {
				delete(cm.orphanBlocks, hash)
			}
		}
	}
	return nil
}

// checkOrphanTransaction checks if a transaction is orphan
func (cm *ConsensusManager) checkOrphanTransaction(tx *Transaction) error {
	for _, input := range tx.Inputs {
		if _, err := cm.utxoSet.GetUTXO(input.PrevHash, input.PrevIndex); err != nil {
			return fmt.Errorf("orphan transaction: %v", err)
		}
	}
	return nil
}

// processBlockTransactions processes all transactions in a block
func (cm *ConsensusManager) processBlockTransactions(block *Block) error {
	// Process coinbase transaction first
	if err := cm.processTransaction(block.Transactions[0]); err != nil {
		return fmt.Errorf("failed to process coinbase transaction: %v", err)
	}

	// Process other transactions
	for i, tx := range block.Transactions[1:] {
		if err := cm.processTransaction(tx); err != nil {
			return fmt.Errorf("failed to process transaction at index %d: %v", i+1, err)
		}
	}

	return nil
}

// processTransaction processes a single transaction
func (cm *ConsensusManager) processTransaction(tx *Transaction) error {
	// Skip validation for coinbase transactions
	if !tx.IsCoinbase() {
		// Spend inputs
		for _, input := range tx.Inputs {
			if err := cm.utxoSet.SpendUTXO(input.PrevHash, input.PrevIndex); err != nil {
				return fmt.Errorf("failed to spend UTXO: %v", err)
			}
		}
	}

	// Add outputs
	for i, output := range tx.Outputs {
		utxo := &UTXO{
			Value:     output.Value,
			Script:    output.Script,
			BlockTime: time.Now().Unix(),
		}
		if err := cm.utxoSet.AddUTXO(tx.Hash, uint32(i), utxo); err != nil {
			return fmt.Errorf("failed to add UTXO: %v", err)
		}
	}

	return nil
}

// GetBestBlock returns the best block
func (cm *ConsensusManager) GetBestBlock() *Block {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.bestBlock
}

// GetCurrentBlock returns the current block
func (cm *ConsensusManager) GetCurrentBlock() *Block {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.currentBlock
}

// GetOrphanBlocks returns all orphan blocks
func (cm *ConsensusManager) GetOrphanBlocks() map[[32]byte]*Block {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.orphanBlocks
}

// GetOrphanTransactions returns all orphan transactions
func (cm *ConsensusManager) GetOrphanTransactions() map[[32]byte]*Transaction {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.orphanTransactions
}
