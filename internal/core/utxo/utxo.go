package utxo

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/common"
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

// UTXOSet represents a set of unspent transaction outputs
type UTXOSet struct {
	// Map of transaction hash to map of output index to UTXO
	utxos map[string]map[uint32]*UTXO
	// Mutex for thread safety
	mu sync.RWMutex
}

// UTXOSetInterface defines the interface for UTXO set operations
type UTXOSetInterface interface {
	// GetUTXO returns a UTXO by its transaction hash and output index
	GetUTXO(txHash []byte, outIndex uint32) (*UTXO, error)
	// AddUTXO adds a new UTXO to the set
	AddUTXO(utxo *UTXO) error
	// RemoveUTXO removes a UTXO from the set
	RemoveUTXO(txHash []byte, outIndex uint32) error
	// GetBalance returns the balance of an address
	GetBalance(address []byte) (uint64, error)
	// UpdateWithBlock updates the UTXO set with a new block
	UpdateWithBlock(block *block.Block) error
}

// NewUTXOSet creates a new UTXO set
func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		utxos: make(map[string]map[uint32]*UTXO),
	}
}

// GetUTXO returns a UTXO by its transaction hash and output index
func (u *UTXOSet) GetUTXO(txHash []byte, outIndex uint32) (*UTXO, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	txHashStr := hex.EncodeToString(txHash)
	if utxos, ok := u.utxos[txHashStr]; ok {
		if utxo, ok := utxos[outIndex]; ok {
			return utxo, nil
		}
	}
	return nil, fmt.Errorf("UTXO not found")
}

// AddUTXO adds a new UTXO to the set
func (u *UTXOSet) AddUTXO(utxo *UTXO) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	txHashStr := hex.EncodeToString(utxo.TxHash)
	if _, ok := u.utxos[txHashStr]; !ok {
		u.utxos[txHashStr] = make(map[uint32]*UTXO)
	}
	u.utxos[txHashStr][utxo.OutIndex] = utxo
	return nil
}

// RemoveUTXO removes a UTXO from the set
func (u *UTXOSet) RemoveUTXO(txHash []byte, outIndex uint32) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	txHashStr := hex.EncodeToString(txHash)
	if utxos, ok := u.utxos[txHashStr]; ok {
		if _, ok := utxos[outIndex]; ok {
			delete(utxos, outIndex)
			if len(utxos) == 0 {
				delete(u.utxos, txHashStr)
			}
			return nil
		}
	}
	return fmt.Errorf("UTXO not found")
}

// GetBalance returns the balance of an address
func (u *UTXOSet) GetBalance(address []byte) (uint64, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	var balance uint64
	for _, utxos := range u.utxos {
		for range utxos {
			// TODO: Implement address extraction from script and compare to address
			// if bytes.Equal(utxo.ScriptPub.GetAddress(), address) {
			// 	balance += utxo.Amount
			// }
		}
	}
	return balance, nil
}

// UpdateWithBlock updates the UTXO set with a new block
func (u *UTXOSet) UpdateWithBlock(block *block.Block) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	for _, tx := range block.Transactions {
		// Remove spent UTXOs
		for _, input := range tx.Inputs() {
			err := u.RemoveUTXO(input.PreviousTxHash, input.PreviousTxIndex)
			if err != nil {
				return fmt.Errorf("error removing spent UTXO: %w", err)
			}
		}

		// Add new UTXOs
		for i, output := range tx.Outputs() {
			scriptObj := script.NewScript()
			scriptObj.AddData(output.ScriptPubKey)
			utxo := &UTXO{
				TxHash:      tx.Hash(),
				OutIndex:    uint32(i),
				Amount:      output.Value,
				ScriptPub:   scriptObj,
				BlockHeight: block.Header.Height,
				IsCoinbase:  tx.IsCoinbase(),
				IsSegWit:    len(tx.Witness()) > 0,
			}
			err := u.AddUTXO(utxo)
			if err != nil {
				return fmt.Errorf("error adding new UTXO: %w", err)
			}
		}
	}
	return nil
}

// GetUTXOsByAddress retrieves all UTXOs for an address
func (u *UTXOSet) GetUTXOsByAddress(address string) []*UTXO {
	var result []*UTXO
	for _, utxos := range u.utxos {
		for _, utxo := range utxos {
			if utxo.ScriptPub.MatchesAddress(address) {
				result = append(result, utxo)
			}
		}
	}
	return result
}

// Size returns the number of UTXOs in the set
func (us *UTXOSet) Size() int {
	us.mu.RLock()
	defer us.mu.RUnlock()

	var count int
	for _, utxos := range us.utxos {
		count += len(utxos)
	}
	return count
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

func (u *UTXO) GetAddress() string {
	// TODO: Implement proper address extraction from script
	return ""
}

func (u *UTXOSet) Update(tx *common.Transaction) error {
	// Handle inputs (spending UTXOs)
	inputs := tx.Inputs()
	for _, input := range inputs {
		// Skip coinbase inputs
		if input.PreviousTxHash == nil {
			continue
		}

		// Create key for the UTXO being spent
		key := fmt.Sprintf("%x:%d", input.PreviousTxHash, input.PreviousTxIndex)
		delete(u.utxos, key)
	}

	// Handle outputs (creating new UTXOs)
	outputs := tx.Outputs()
	for i, output := range outputs {
		// Create new UTXO
		script := script.NewScript()
		script.AddData(output.ScriptPubKey)

		utxo := &UTXO{
			TxHash:      tx.Hash(),
			OutIndex:    uint32(i),
			Amount:      output.Value,
			ScriptPub:   script,
			BlockHeight: 0, // TODO: Get from block
			IsCoinbase:  tx.IsCoinbase(),
			IsSegWit:    len(tx.Witness()) > 0, // Check if transaction has witness data
		}

		// Add to UTXO set
		key := fmt.Sprintf("%x:%d", utxo.TxHash, utxo.OutIndex)
		if _, ok := u.utxos[key]; !ok {
			u.utxos[key] = make(map[uint32]*UTXO)
		}
		u.utxos[key][utxo.OutIndex] = utxo
	}

	return nil
}

// All returns a flat slice of all UTXOs in the set
func (u *UTXOSet) All() []*UTXO {
	u.mu.RLock()
	defer u.mu.RUnlock()

	var utxos []*UTXO
	for _, utxoMap := range u.utxos {
		for _, utxo := range utxoMap {
			utxos = append(utxos, utxo)
		}
	}
	return utxos
}
