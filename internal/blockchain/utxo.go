package blockchain

import (
	"strconv"
	"sync"
)

// UTXO represents an unspent transaction output
type UTXO struct {
	TxID        string
	OutputIndex int
	Amount      float64
	Address     string
}

// UTXOSet manages the set of unspent transaction outputs
type UTXOSet struct {
	utxos map[string]UTXO // key: txID:outputIndex
	mu    sync.RWMutex
}

// NewUTXOSet creates a new UTXO set
func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		utxos: make(map[string]UTXO),
	}
}

// Add adds a UTXO to the set
func (u *UTXOSet) Add(utxo UTXO) {
	u.mu.Lock()
	defer u.mu.Unlock()
	key := utxo.TxID + ":" + strconv.Itoa(utxo.OutputIndex)
	u.utxos[key] = utxo
}

// Remove removes a UTXO from the set
func (u *UTXOSet) Remove(txID string, outputIndex int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	key := txID + ":" + strconv.Itoa(outputIndex)
	delete(u.utxos, key)
}

// Get returns a UTXO by its transaction ID and output index
func (u *UTXOSet) Get(txID string, outputIndex int) (UTXO, bool) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	key := txID + ":" + strconv.Itoa(outputIndex)
	utxo, exists := u.utxos[key]
	return utxo, exists
}

// GetAll returns all UTXOs
func (u *UTXOSet) GetAll() []UTXO {
	u.mu.RLock()
	defer u.mu.RUnlock()
	utxos := make([]UTXO, 0, len(u.utxos))
	for _, utxo := range u.utxos {
		utxos = append(utxos, utxo)
	}
	return utxos
}
