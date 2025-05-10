package storage

import (
	"sync"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/utxo"
)

// UTXOAdapter adapts utxo.UTXOSet to a generic interface
// (if you need to implement an interface, do so here)
type UTXOAdapter struct {
	utxoSet *utxo.UTXOSet
	mu      sync.RWMutex
}

// NewUTXOAdapter creates a new UTXOAdapter
func NewUTXOAdapter(utxoSet *utxo.UTXOSet) *UTXOAdapter {
	return &UTXOAdapter{
		utxoSet: utxoSet,
	}
}

// GetUTXO retrieves a UTXO by its transaction hash and output index
func (a *UTXOAdapter) GetUTXO(txHash []byte, outputIndex uint32) (*utxo.UTXO, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.utxoSet.GetUTXO(txHash, outputIndex)
}

// AddUTXO adds a UTXO to the set
func (a *UTXOAdapter) AddUTXO(utxoObj *utxo.UTXO) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.utxoSet.AddUTXO(utxoObj)
}

// RemoveUTXO removes a UTXO from the set
func (a *UTXOAdapter) RemoveUTXO(txHash []byte, outputIndex uint32) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.utxoSet.RemoveUTXO(txHash, outputIndex)
}

// GetBalance gets the balance for an address
func (a *UTXOAdapter) GetBalance(address string, coinType interface{}) (uint64, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var balance uint64
	for _, utxo := range a.utxoSet.All() {
		if utxo.ScriptPub != nil && utxo.ScriptPub.MatchesAddress(address) {
			// TODO: Handle coinType if needed
			balance += utxo.Amount
		}
	}
	return balance, nil
}

// GetAllUTXOs returns all UTXOs in the set
func (a *UTXOAdapter) GetAllUTXOs() []*utxo.UTXO {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.utxoSet.All()
}

// UpdateWithBlock updates the UTXO set with a new block
func (a *UTXOAdapter) UpdateWithBlock(block *block.Block) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.utxoSet.UpdateWithBlock(block)
}
