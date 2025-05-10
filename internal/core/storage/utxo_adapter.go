package storage

import (
	"github.com/youngchain/internal/core/utxo"
)

// UTXOAdapter adapts utxo.UTXOSet to storage.UTXOSet
type UTXOAdapter struct {
	utxoSet *utxo.UTXOSet
}

// NewUTXOAdapter creates a new UTXO adapter
func NewUTXOAdapter(utxoSet *utxo.UTXOSet) *UTXOAdapter {
	return &UTXOAdapter{
		utxoSet: utxoSet,
	}
}

// GetUTXO gets a UTXO from the set
func (a *UTXOAdapter) GetUTXO(txHash []byte, outputIndex uint32) (*UTXO, bool) {
	utxo, exists := a.utxoSet.GetUTXO(txHash, outputIndex)
	if !exists {
		return nil, false
	}

	return &UTXO{
		TxHash:      utxo.TxHash,
		OutputIndex: utxo.OutIndex,
		Value:       utxo.Amount,
		Script:      utxo.ScriptPub.Serialize(),
		Height:      utxo.BlockHeight,
	}, true
}

// AddUTXO adds a UTXO to the set
func (a *UTXOAdapter) AddUTXO(utxo *UTXO) error {
	utxoObj := &utxo.UTXO{
		TxHash:      utxo.TxHash,
		OutIndex:    utxo.OutputIndex,
		Amount:      utxo.Value,
		ScriptPub:   nil, // TODO: Convert script to *script.Script
		BlockHeight: utxo.Height,
		IsCoinbase:  false,
		IsSegWit:    false,
	}
	return a.utxoSet.AddUTXO(utxoObj)
}

// SpendUTXO marks a UTXO as spent
func (a *UTXOAdapter) SpendUTXO(txHash []byte, outputIndex uint32, height uint64) {
	a.utxoSet.RemoveUTXO(txHash, outputIndex)
}

// PruneSpent prunes spent UTXOs up to a certain height
func (a *UTXOAdapter) PruneSpent(height uint64) error {
	// No-op for now as utxo.UTXOSet doesn't have pruning
	return nil
}

// GetUTXOCount returns the number of UTXOs
func (a *UTXOAdapter) GetUTXOCount() int {
	return a.utxoSet.Size()
}

// GetSpentCount returns the number of spent UTXOs
func (a *UTXOAdapter) GetSpentCount() int {
	// No-op for now as utxo.UTXOSet doesn't track spent UTXOs
	return 0
}

// Serialize serializes the UTXO set
func (a *UTXOAdapter) Serialize() []byte {
	// No-op for now as utxo.UTXOSet doesn't have serialization
	return nil
}

// Deserialize deserializes the UTXO set
func (a *UTXOAdapter) Deserialize(data []byte) error {
	// No-op for now as utxo.UTXOSet doesn't have deserialization
	return nil
}
