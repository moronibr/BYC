package storage

import (
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

	a.store.utxos[string(utxo.TxHash)+string(utxo.OutputIndex)] = utxo
}

// RemoveUTXO removes a UTXO from the store
func (a *UTXOAdapter) RemoveUTXO(txHash []byte, index uint32) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.store.utxos, string(txHash)+string(index))
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
	// ... existing code ...
	return nil
}

// UpdateWithBlock updates the UTXO set with a new block
func (a *UTXOAdapter) UpdateWithBlock(block *block.Block) error {
	// ... existing code ...
	return nil
}
