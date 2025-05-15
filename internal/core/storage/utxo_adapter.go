package storage

import (
	"fmt"
	"sync"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/types"
)

// UTXOAdapter adapts the UTXO set to the storage interface
type UTXOAdapter struct {
	store *TransactionStore
	mu    sync.RWMutex
}

// NewUTXOAdapter creates a new UTXO adapter
func NewUTXOAdapter(store *TransactionStore) *UTXOAdapter {
	return &UTXOAdapter{
		store: store,
	}
}

// GetUTXO gets a UTXO by its transaction hash and output index
func (a *UTXOAdapter) GetUTXO(txHash []byte, index uint32) *types.UTXO {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.GetUTXO(txHash, index)
}

// AddUTXO adds a UTXO to the store
func (a *UTXOAdapter) AddUTXO(utxo *types.UTXO) {
	a.mu.Lock()
	defer a.mu.Unlock()

	utxoKey := fmt.Sprintf("%x:%d", utxo.TxHash, utxo.OutputIndex)
	a.store.utxos[utxoKey] = utxo
}

// RemoveUTXO removes a UTXO from the store
func (a *UTXOAdapter) RemoveUTXO(txHash []byte, index uint32) {
	a.mu.Lock()
	defer a.mu.Unlock()

	utxoKey := fmt.Sprintf("%x:%d", txHash, index)
	delete(a.store.utxos, utxoKey)
}

// GetUTXOsByAddress gets all UTXOs for an address
func (a *UTXOAdapter) GetUTXOsByAddress(address string) []*types.UTXO {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.GetUTXOsByAddress(address)
}

// GetBalance gets the balance for an address and coin type
func (a *UTXOAdapter) GetBalance(address string, coinType coin.Type) int64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.GetBalance(address, coinType)
}

// SpendUTXO marks a UTXO as spent
func (a *UTXOAdapter) SpendUTXO(txHash []byte, index uint32) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.store.SpendUTXO(txHash, index)
}

// ValidateUTXO validates a UTXO
func (a *UTXOAdapter) ValidateUTXO(txHash []byte, index uint32) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.store.ValidateUTXO(txHash, index)
}

// GetAllUTXOs returns all UTXOs in the set
func (a *UTXOAdapter) GetAllUTXOs() []*types.UTXO {
	a.mu.RLock()
	defer a.mu.RUnlock()

	utxos := make([]*types.UTXO, 0, len(a.store.utxos))
	for _, utxo := range a.store.utxos {
		utxos = append(utxos, utxo)
	}
	return utxos
}

// UpdateWithBlock updates the UTXO set with a new block
func (a *UTXOAdapter) UpdateWithBlock(block *block.Block) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Process transactions in the block
	for _, tx := range block.GetTransactions() {
		// Remove spent UTXOs
		for _, input := range tx.GetInputs() {
			utxoKey := fmt.Sprintf("%x:%d", input.PreviousTxHash, input.OutputIndex)
			delete(a.store.utxos, utxoKey)
		}

		// Add new UTXOs
		for i, output := range tx.GetOutputs() {
			utxo := types.NewUTXO(
				tx.GetHash(),
				uint32(i),
				output.Value,
				output.ScriptPubKey,
				output.Address,
				tx.GetCoinType(),
			)
			utxoKey := fmt.Sprintf("%x:%d", tx.GetHash(), i)
			a.store.utxos[utxoKey] = utxo
		}
	}

	return nil
}
