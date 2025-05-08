package network

import (
	"sync"

	"github.com/youngchain/internal/core/types"
)

// TransactionPool manages pending transactions
type TransactionPool struct {
	mu sync.RWMutex

	// Pending transactions
	pendingTxs []*types.Transaction

	// Maximum number of transactions in the pool
	maxSize int
}

// NewTransactionPool creates a new transaction pool
func NewTransactionPool() *TransactionPool {
	return &TransactionPool{
		pendingTxs: make([]*types.Transaction, 0),
		maxSize:    1000,
	}
}

// AddTransaction adds a transaction to the pool
func (tp *TransactionPool) AddTransaction(tx *types.Transaction) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if len(tp.pendingTxs) >= tp.maxSize {
		return nil // Pool is full, silently drop transaction
	}

	tp.pendingTxs = append(tp.pendingTxs, tx)
	return nil
}

// GetPendingTransactions returns all pending transactions
func (tp *TransactionPool) GetPendingTransactions() []*types.Transaction {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	// Return a copy to prevent modification of the internal slice
	txs := make([]*types.Transaction, len(tp.pendingTxs))
	copy(txs, tp.pendingTxs)
	return txs
}

// RemoveTransactions removes transactions from the pool
func (tp *TransactionPool) RemoveTransactions(txs []*types.Transaction) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Create a map for quick lookup
	toRemove := make(map[string]struct{})
	for _, tx := range txs {
		toRemove[string(tx.Hash)] = struct{}{}
	}

	// Remove transactions
	newPendingTxs := make([]*types.Transaction, 0)
	for _, tx := range tp.pendingTxs {
		if _, exists := toRemove[string(tx.Hash)]; !exists {
			newPendingTxs = append(newPendingTxs, tx)
		}
	}

	tp.pendingTxs = newPendingTxs
}
