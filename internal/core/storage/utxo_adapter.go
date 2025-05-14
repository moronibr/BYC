package storage

import (
	"sync"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/coin"
	"github.com/youngchain/internal/core/types"
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
func (a *UTXOAdapter) GetUTXO(txHash []byte, outputIndex uint32) (*types.UTXO, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	utxo, err := a.utxoSet.GetUTXO(txHash, outputIndex)
	if err != nil {
		return nil, false
	}
	return &types.UTXO{
		TxHash:    utxo.TxHash,
		TxIndex:   utxo.OutIndex,
		Value:     utxo.Amount,
		Address:   "",                // TODO: Extract from script
		CoinType:  coin.CoinType(""), // TODO: Get from somewhere
		IsSpent:   false,
		BlockHash: nil, // TODO: Get from somewhere
	}, true
}

// AddUTXO adds a UTXO to the set
func (a *UTXOAdapter) AddUTXO(utxoTyped *types.UTXO) {
	a.mu.Lock()
	defer a.mu.Unlock()

	utxoObj := &utxo.UTXO{
		TxHash:      utxoTyped.TxHash,
		OutIndex:    utxoTyped.TxIndex,
		Amount:      utxoTyped.Value,
		ScriptPub:   nil, // TODO: Convert from address
		BlockHeight: 0,   // TODO: Get from somewhere
		IsCoinbase:  false,
		IsSegWit:    false,
	}
	a.utxoSet.AddUTXO(utxoObj)
}

// RemoveUTXO removes a UTXO from the set
func (a *UTXOAdapter) RemoveUTXO(txHash []byte, outputIndex uint32) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.utxoSet.RemoveUTXO(txHash, outputIndex)
}

// GetBalance gets the balance for an address
func (a *UTXOAdapter) GetBalance(address string, coinType coin.CoinType) uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	balance, _ := a.utxoSet.GetBalance([]byte(address))
	return balance
}

// GetAllUTXOs returns all UTXOs in the set
func (a *UTXOAdapter) GetAllUTXOs() []*types.UTXO {
	a.mu.RLock()
	defer a.mu.RUnlock()

	utxos := a.utxoSet.All()
	result := make([]*types.UTXO, len(utxos))
	for i, u := range utxos {
		result[i] = &types.UTXO{
			TxHash:    u.TxHash,
			TxIndex:   u.OutIndex,
			Value:     u.Amount,
			Address:   "",                // TODO: Extract from script
			CoinType:  coin.CoinType(""), // TODO: Get from somewhere
			IsSpent:   false,
			BlockHash: nil, // TODO: Get from somewhere
		}
	}
	return result
}

// UpdateWithBlock updates the UTXO set with a new block
func (a *UTXOAdapter) UpdateWithBlock(block *block.Block) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.utxoSet.UpdateWithBlock(block)
}
