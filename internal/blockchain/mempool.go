package blockchain

import (
	"fmt"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
)

// Mempool represents the transaction mempool
type Mempool struct {
	transactions map[string]*TransactionInfo
	mu           sync.RWMutex
	maxSize      int
}

// TransactionInfo holds transaction metadata
type TransactionInfo struct {
	Transaction *block.Transaction
	AddedAt     time.Time
	Priority    float64
}

// NewMempool creates a new mempool
func NewMempool(maxSize int) *Mempool {
	return &Mempool{
		transactions: make(map[string]*TransactionInfo),
		maxSize:      maxSize,
	}
}

// AddTransaction adds a transaction to the mempool
func (m *Mempool) AddTransaction(tx *block.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if mempool is full
	if len(m.transactions) >= m.maxSize {
		return fmt.Errorf("mempool is full")
	}

	// Calculate transaction priority
	priority := m.calculatePriority(tx)

	// Add transaction to mempool
	m.transactions[string(tx.Hash)] = &TransactionInfo{
		Transaction: tx,
		AddedAt:     time.Now(),
		Priority:    priority,
	}

	return nil
}

// RemoveTransaction removes a transaction from the mempool
func (m *Mempool) RemoveTransaction(txHash []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.transactions, string(txHash))
}

// GetTransaction retrieves a transaction from the mempool
func (m *Mempool) GetTransaction(txHash []byte) (*block.Transaction, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, exists := m.transactions[string(txHash)]
	if !exists {
		return nil, false
	}

	return info.Transaction, true
}

// GetTransactions returns all transactions in the mempool
func (m *Mempool) GetTransactions() []*block.Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	transactions := make([]*block.Transaction, 0, len(m.transactions))
	for _, info := range m.transactions {
		transactions = append(transactions, info.Transaction)
	}

	return transactions
}

// GetPendingTransactions returns pending transactions sorted by priority
func (m *Mempool) GetPendingTransactions() []*block.Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a slice of transaction info
	infos := make([]*TransactionInfo, 0, len(m.transactions))
	for _, info := range m.transactions {
		infos = append(infos, info)
	}

	// Sort by priority
	sortTransactionsByPriority(infos)

	// Extract transactions
	transactions := make([]*block.Transaction, len(infos))
	for i, info := range infos {
		transactions[i] = info.Transaction
	}

	return transactions
}

// calculatePriority calculates the priority of a transaction
func (m *Mempool) calculatePriority(tx *block.Transaction) float64 {
	// Priority is based on:
	// 1. Transaction amount
	// 2. Transaction age
	// 3. Gas price (if applicable)

	// For now, use a simple priority calculation
	return float64(tx.Amount)
}

// sortTransactionsByPriority sorts transactions by priority
func sortTransactionsByPriority(infos []*TransactionInfo) {
	// Implement sorting logic
	// This is a placeholder for the actual sorting implementation
}

// Cleanup removes old transactions from the mempool
func (m *Mempool) Cleanup(maxAge time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for hash, info := range m.transactions {
		if now.Sub(info.AddedAt) > maxAge {
			delete(m.transactions, hash)
		}
	}
}

// Size returns the current size of the mempool
func (m *Mempool) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.transactions)
}

// IsFull returns whether the mempool is full
func (m *Mempool) IsFull() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.transactions) >= m.maxSize
}
