package transaction

import (
	"fmt"
	"sync"

	"github.com/youngchain/internal/core/types"
)

// UTXOSetInterface defines the interface for UTXO set operations
type UTXOSetInterface interface {
	GetUTXO(txHash []byte, index uint32) *types.UTXO
	SpendUTXO(txHash []byte, index uint32) error
	AddUTXO(txHash []byte, index uint32, utxo *types.UTXO) error
}

// TxPool represents the transaction pool
type TxPool struct {
	transactions map[string]*types.Transaction
	maxSize      int
	minFee       uint64
	utxoSet      UTXOSetInterface
	mu           sync.RWMutex
}

// NewTxPool creates a new transaction pool
func NewTxPool(maxSize int, minFee uint64, utxoSet UTXOSetInterface) *TxPool {
	return &TxPool{
		transactions: make(map[string]*types.Transaction),
		maxSize:      maxSize,
		minFee:       minFee,
		utxoSet:      utxoSet,
	}
}

// AddTransaction adds a transaction to the pool
func (p *TxPool) AddTransaction(tx *types.Transaction) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check pool size
	if len(p.transactions) >= p.maxSize {
		return fmt.Errorf("transaction pool is full")
	}

	// Validate transaction
	if err := p.validateTransaction(tx); err != nil {
		return err
	}

	// Add transaction
	p.transactions[string(tx.Hash())] = tx

	return nil
}

// GetTransaction gets a transaction from the pool
func (p *TxPool) GetTransaction(hash []byte) *types.Transaction {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.transactions[string(hash)]
}

// RemoveTransaction removes a transaction from the pool
func (p *TxPool) RemoveTransaction(hash []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.transactions, string(hash))
}

// GetBest returns the best transactions from the pool
func (p *TxPool) GetBest(maxCount int) []*types.Transaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Sort transactions by fee
	transactions := make([]*types.Transaction, 0, len(p.transactions))
	for _, tx := range p.transactions {
		transactions = append(transactions, tx)
	}

	// Sort by fee (descending)
	for i := 0; i < len(transactions); i++ {
		for j := i + 1; j < len(transactions); j++ {
			if transactions[i].Fee < transactions[j].Fee {
				transactions[i], transactions[j] = transactions[j], transactions[i]
			}
		}
	}

	// Return top transactions
	if len(transactions) > maxCount {
		return transactions[:maxCount]
	}
	return transactions
}

// validateTransaction validates a transaction
func (p *TxPool) validateTransaction(tx *types.Transaction) error {
	// Check fee
	if tx.Fee < p.minFee {
		return fmt.Errorf("transaction fee too low")
	}

	// Check inputs
	for _, input := range tx.Inputs {
		utxo := p.utxoSet.GetUTXO(input.PreviousTxHash, input.PreviousTxIndex)
		if utxo == nil {
			return fmt.Errorf("input not found in UTXO set")
		}
	}

	return nil
}

// GetSize returns the number of transactions in the pool
func (p *TxPool) GetSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.transactions)
}

// Clear clears the transaction pool
func (p *TxPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.transactions = make(map[string]*types.Transaction)
}
