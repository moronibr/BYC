package network

import (
	"sync"

	"github.com/youngchain/internal/core/common"
)

// TransactionPool manages pending transactions
type TransactionPool struct {
	mu sync.RWMutex

	// Pending transactions
	pendingTxs []*common.Transaction

	// Maximum number of transactions in the pool
	maxSize int
}

// NewTransactionPool creates a new transaction pool
func NewTransactionPool() *TransactionPool {
	return &TransactionPool{
		pendingTxs: make([]*common.Transaction, 0),
		maxSize:    1000,
	}
}

// AddTransaction adds a transaction to the pool
func (tp *TransactionPool) AddTransaction(tx *common.Transaction) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if len(tp.pendingTxs) >= tp.maxSize {
		return nil // Pool is full, silently drop transaction
	}

	tp.pendingTxs = append(tp.pendingTxs, tx)
	return nil
}

// GetPendingTransactions returns all pending transactions
func (tp *TransactionPool) GetPendingTransactions() []*common.Transaction {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	// Return a copy to prevent modification of the internal slice
	txs := make([]*common.Transaction, len(tp.pendingTxs))
	copy(txs, tp.pendingTxs)
	return txs
}

// RemoveTransactions removes transactions from the pool
func (tp *TransactionPool) RemoveTransactions(txs []*common.Transaction) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Create a map of transaction hashes to remove
	toRemove := make(map[string]struct{})
	for _, tx := range txs {
		toRemove[string(tx.Hash())] = struct{}{}
	}

	// Remove transactions
	var newTxs []*common.Transaction
	for _, tx := range tp.pendingTxs {
		if _, exists := toRemove[string(tx.Hash())]; !exists {
			newTxs = append(newTxs, tx)
		}
	}
	tp.pendingTxs = newTxs
}
