package adapters

import (
	"github.com/youngchain/internal/core/script"
	"github.com/youngchain/internal/core/storage"
	"github.com/youngchain/internal/core/utxo"
)

// StorageUTXOAdapter adapts storage.UTXOSet to utxo.UTXOSetInterface
type StorageUTXOAdapter struct {
	utxoSet *storage.UTXOSet
}

// NewStorageUTXOAdapter creates a new StorageUTXOAdapter
func NewStorageUTXOAdapter(utxoSet *storage.UTXOSet) *StorageUTXOAdapter {
	return &StorageUTXOAdapter{
		utxoSet: utxoSet,
	}
}

// GetUTXO retrieves a UTXO by its transaction hash and output index
func (a *StorageUTXOAdapter) GetUTXO(txHash []byte, outputIndex uint32) (*utxo.UTXO, bool) {
	storageUTXO, exists := a.utxoSet.GetUTXO(txHash, outputIndex)
	if !exists {
		return nil, false
	}

	s := script.NewScript()
	if err := s.AddData(storageUTXO.Script); err != nil {
		return nil, false
	}

	return &utxo.UTXO{
		TxHash:      storageUTXO.TxHash,
		OutIndex:    storageUTXO.OutputIndex,
		Amount:      storageUTXO.Value,
		ScriptPub:   s,
		BlockHeight: storageUTXO.Height,
		IsCoinbase:  false, // TODO: Get this from somewhere
		IsSegWit:    false, // TODO: Get this from somewhere
	}, true
}

// AddUTXO adds a UTXO to the set
func (a *StorageUTXOAdapter) AddUTXO(utxo *utxo.UTXO) error {
	a.utxoSet.AddUTXO(&storage.UTXO{
		TxHash:      utxo.TxHash,
		OutputIndex: utxo.OutIndex,
		Value:       utxo.Amount,
		Script:      utxo.ScriptPub.Serialize(),
		Height:      utxo.BlockHeight,
	})
	return nil
}

// RemoveUTXO removes a UTXO from the set
func (a *StorageUTXOAdapter) RemoveUTXO(txHash []byte, outputIndex uint32) {
	a.utxoSet.SpendUTXO(txHash, outputIndex, 0) // TODO: Get current height
}

// Size returns the number of UTXOs in the set
func (a *StorageUTXOAdapter) Size() int {
	return a.utxoSet.GetUTXOCount()
}
