package blockchain

import (
	"fmt"
	"sync"
)

// UTXO represents an unspent transaction output
type UTXO struct {
	TxID      string
	Index     int
	Amount    float64
	Address   string
	CoinType  CoinType
	Spent     bool
	Timestamp int64
}

// UTXOSet manages the set of unspent transaction outputs
type UTXOSet struct {
	utxos map[string]UTXO
	mu    sync.RWMutex
}

// NewUTXOSet creates a new UTXO set
func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		utxos: make(map[string]UTXO),
	}
}

// Add adds a new UTXO to the set
func (us *UTXOSet) Add(utxo UTXO) {
	us.mu.Lock()
	defer us.mu.Unlock()
	us.utxos[utxo.TxID] = utxo
}

// Remove removes a UTXO from the set
func (us *UTXOSet) Remove(txID string) {
	us.mu.Lock()
	defer us.mu.Unlock()
	delete(us.utxos, txID)
}

// Get retrieves a UTXO from the set
func (us *UTXOSet) Get(txID string) (UTXO, bool) {
	us.mu.RLock()
	defer us.mu.RUnlock()
	utxo, exists := us.utxos[txID]
	return utxo, exists
}

// HasUTXO checks if a UTXO exists in the set
func (us *UTXOSet) HasUTXO(txID string) bool {
	us.mu.RLock()
	defer us.mu.RUnlock()
	_, exists := us.utxos[txID]
	return exists
}

// Update updates a UTXO in the set
func (us *UTXOSet) Update(utxo UTXO) {
	us.mu.Lock()
	defer us.mu.Unlock()
	us.utxos[utxo.TxID] = utxo
}

// GetBalance returns the balance for an address
func (us *UTXOSet) GetBalance(address string, coinType CoinType) float64 {
	us.mu.RLock()
	defer us.mu.RUnlock()

	var balance float64
	for _, utxo := range us.utxos {
		if utxo.Address == address && utxo.CoinType == coinType && !utxo.Spent {
			balance += utxo.Amount
		}
	}
	return balance
}

// GetUTXOsForAddress returns all UTXOs for an address
func (us *UTXOSet) GetUTXOsForAddress(address string, coinType CoinType) []UTXO {
	us.mu.RLock()
	defer us.mu.RUnlock()

	var utxos []UTXO
	for _, utxo := range us.utxos {
		if utxo.Address == address && utxo.CoinType == coinType && !utxo.Spent {
			utxos = append(utxos, utxo)
		}
	}
	return utxos
}

// MarkSpent marks a UTXO as spent
func (us *UTXOSet) MarkSpent(txID string) {
	us.mu.Lock()
	defer us.mu.Unlock()
	if utxo, exists := us.utxos[txID]; exists {
		utxo.Spent = true
		us.utxos[txID] = utxo
	}
}

// GetTotalSupply returns the total supply of a coin type
func (us *UTXOSet) GetTotalSupply(coinType CoinType) float64 {
	us.mu.RLock()
	defer us.mu.RUnlock()

	var supply float64
	for _, utxo := range us.utxos {
		if utxo.CoinType == coinType && !utxo.Spent {
			supply += utxo.Amount
		}
	}
	return supply
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
