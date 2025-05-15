package storage

import (
	"fmt"
	"sync"

	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/types"
)

// TransactionStore manages transactions
type TransactionStore struct {
	transactions map[string]*types.Transaction
	utxos        map[string]*types.UTXO
	mu           sync.RWMutex
}

// NewTransactionStore creates a new transaction store
func NewTransactionStore() *TransactionStore {
	return &TransactionStore{
		transactions: make(map[string]*types.Transaction),
		utxos:        make(map[string]*types.UTXO),
	}
}

// AddTransaction adds a transaction to the store
func (ts *TransactionStore) AddTransaction(tx *types.Transaction) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// Add transaction
	ts.transactions[string(tx.GetHash())] = tx

	// Update UTXOs
	for i, output := range tx.GetOutputs() {
		utxo := types.NewUTXO(
			tx.GetHash(),
			uint32(i),
			output.Value,
			output.ScriptPubKey,
			output.Address,
			tx.GetCoinType(),
		)
		ts.utxos[string(tx.GetHash())+string(i)] = utxo
	}

	return nil
}

// GetTransaction gets a transaction by its hash
func (ts *TransactionStore) GetTransaction(hash []byte) *types.Transaction {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	return ts.transactions[string(hash)]
}

// GetUTXO gets a UTXO by its transaction hash and output index
func (ts *TransactionStore) GetUTXO(txHash []byte, index uint32) *types.UTXO {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	return ts.utxos[string(txHash)+string(index)]
}

// GetUTXOsByAddress gets all UTXOs for an address
func (ts *TransactionStore) GetUTXOsByAddress(address string) []*types.UTXO {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	utxos := make([]*types.UTXO, 0)
	for _, utxo := range ts.utxos {
		if utxo.Address == address && !utxo.Spent {
			utxos = append(utxos, utxo)
		}
	}

	return utxos
}

// GetBalance gets the balance for an address and coin type
func (ts *TransactionStore) GetBalance(address string, coinType coin.Type) int64 {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	var balance int64
	for _, utxo := range ts.utxos {
		if utxo.Address == address && utxo.CoinType == coinType && !utxo.Spent {
			balance += utxo.Value
		}
	}

	return balance
}

// SpendUTXO marks a UTXO as spent
func (ts *TransactionStore) SpendUTXO(txHash []byte, index uint32) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	utxo := ts.utxos[string(txHash)+string(index)]
	if utxo == nil {
		return fmt.Errorf("UTXO not found")
	}

	if utxo.Spent {
		return fmt.Errorf("UTXO already spent")
	}

	utxo.Spent = true
	return nil
}

// ValidateUTXO validates a UTXO
func (ts *TransactionStore) ValidateUTXO(txHash []byte, index uint32) bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	utxo := ts.utxos[string(txHash)+string(index)]
	return utxo != nil && !utxo.Spent
}
