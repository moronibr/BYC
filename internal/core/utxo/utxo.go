package utxo

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/youngchain/internal/core/script"
)

const (
	// MaxUTXOSetSize is the maximum number of UTXOs that can be stored
	MaxUTXOSetSize = 1000000

	// MaxUTXOValue is the maximum value a single UTXO can have
	MaxUTXOValue = 2100000000000000 // 21 million BTC in satoshis
)

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash      []byte
	OutIndex    uint32
	Amount      uint64
	ScriptPub   *script.Script
	BlockHeight uint64
	IsCoinbase  bool
	IsSegWit    bool
}

// UTXOSet manages the set of unspent transaction outputs
type UTXOSet struct {
	utxos map[string]*UTXO // key: txHash:outIndex
	mu    sync.RWMutex
}

// NewUTXOSet creates a new UTXO set
func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		utxos: make(map[string]*UTXO),
	}
}

// AddUTXO adds a UTXO to the set
func (us *UTXOSet) AddUTXO(utxo *UTXO) error {
	us.mu.Lock()
	defer us.mu.Unlock()

	// Check UTXO set size
	if len(us.utxos) >= MaxUTXOSetSize {
		return fmt.Errorf("utxo set is full")
	}

	// Check UTXO value
	if utxo.Amount > MaxUTXOValue {
		return fmt.Errorf("utxo value exceeds maximum")
	}

	// Create key
	key := utxoKey(utxo.TxHash, utxo.OutIndex)

	// Add UTXO
	us.utxos[key] = utxo

	return nil
}

// RemoveUTXO removes a UTXO from the set
func (us *UTXOSet) RemoveUTXO(txHash []byte, outIndex uint32) {
	us.mu.Lock()
	defer us.mu.Unlock()

	key := utxoKey(txHash, outIndex)
	delete(us.utxos, key)
}

// GetUTXO retrieves a UTXO from the set
func (us *UTXOSet) GetUTXO(txHash []byte, outIndex uint32) (*UTXO, bool) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	key := utxoKey(txHash, outIndex)
	utxo, exists := us.utxos[key]
	return utxo, exists
}

// GetUTXOsByAddress retrieves all UTXOs for an address
func (us *UTXOSet) GetUTXOsByAddress(address string) []*UTXO {
	us.mu.RLock()
	defer us.mu.RUnlock()

	var utxos []*UTXO
	for _, utxo := range us.utxos {
		if utxo.ScriptPub.MatchesAddress(address) {
			utxos = append(utxos, utxo)
		}
	}
	return utxos
}

// GetBalance returns the total balance for an address
func (us *UTXOSet) GetBalance(address string) uint64 {
	us.mu.RLock()
	defer us.mu.RUnlock()

	var balance uint64
	for _, utxo := range us.utxos {
		if utxo.ScriptPub.MatchesAddress(address) {
			balance += utxo.Amount
		}
	}
	return balance
}

// Size returns the number of UTXOs in the set
func (us *UTXOSet) Size() int {
	us.mu.RLock()
	defer us.mu.RUnlock()

	return len(us.utxos)
}

// utxoKey creates a key for a UTXO
func utxoKey(txHash []byte, outIndex uint32) string {
	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, outIndex)
	return hex.EncodeToString(txHash) + ":" + hex.EncodeToString(indexBytes)
}

// ValidateUTXO validates a UTXO
func (utxo *UTXO) Validate() error {
	// Check amount
	if utxo.Amount > MaxUTXOValue {
		return fmt.Errorf("utxo value exceeds maximum")
	}

	// Check script
	if utxo.ScriptPub == nil {
		return fmt.Errorf("missing script")
	}

	// Validate script
	if err := utxo.ScriptPub.Validate(); err != nil {
		return fmt.Errorf("invalid script: %v", err)
	}

	return nil
}

// IsMature checks if a coinbase UTXO is mature
func (utxo *UTXO) IsMature(currentHeight uint64) bool {
	if !utxo.IsCoinbase {
		return true
	}
	return currentHeight-utxo.BlockHeight >= 100 // 100 blocks for coinbase maturity
}

// Clone creates a deep copy of the UTXO
func (utxo *UTXO) Clone() *UTXO {
	return &UTXO{
		TxHash:      append([]byte{}, utxo.TxHash...),
		OutIndex:    utxo.OutIndex,
		Amount:      utxo.Amount,
		ScriptPub:   utxo.ScriptPub.Clone(),
		BlockHeight: utxo.BlockHeight,
		IsCoinbase:  utxo.IsCoinbase,
		IsSegWit:    utxo.IsSegWit,
	}
}
