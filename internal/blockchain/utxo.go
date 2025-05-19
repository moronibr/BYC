package blockchain

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
)

// UTXO represents an unspent transaction output
type UTXO struct {
	TxID          []byte
	OutputIndex   int
	Amount        float64
	Address       string
	PublicKeyHash []byte
	CoinType      CoinType
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
	key := hex.EncodeToString(utxo.TxID) + ":" + strconv.Itoa(utxo.OutputIndex)
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

// GetUTXOs returns all UTXOs for a given address
func (utxoSet *UTXOSet) GetUTXOs(address string) ([]UTXO, error) {
	utxoSet.mu.RLock()
	defer utxoSet.mu.RUnlock()

	var utxos []UTXO
	for _, utxo := range utxoSet.utxos {
		if utxo.Address == address {
			utxos = append(utxos, utxo)
		}
	}
	return utxos, nil
}

// Update updates the UTXO set with a new transaction
func (utxoSet *UTXOSet) Update(tx *Transaction) error {
	utxoSet.mu.Lock()
	defer utxoSet.mu.Unlock()

	// Remove spent UTXOs
	for _, input := range tx.Inputs {
		key := fmt.Sprintf("%x:%d", input.TxID, input.OutputIndex)
		delete(utxoSet.utxos, key)
	}

	// Add new UTXOs
	for i, output := range tx.Outputs {
		utxo := UTXO{
			TxID:          tx.ID,
			OutputIndex:   i,
			Amount:        output.Value,
			Address:       output.Address,
			PublicKeyHash: output.PublicKeyHash,
			CoinType:      output.CoinType,
		}
		key := fmt.Sprintf("%x:%d", tx.ID, i)
		utxoSet.utxos[key] = utxo
	}

	return nil
}

// GetUTXO retrieves a UTXO by its transaction ID and output index
func (utxoSet *UTXOSet) GetUTXO(txID []byte, outputIndex int) UTXO {
	utxoSet.mu.RLock()
	defer utxoSet.mu.RUnlock()

	key := fmt.Sprintf("%x:%d", txID, outputIndex)
	utxo, exists := utxoSet.utxos[key]
	if !exists {
		return UTXO{}
	}

	return utxo
}
