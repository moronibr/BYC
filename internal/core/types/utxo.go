package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

// UTXO represents an unspent transaction output
type UTXO struct {
	Value      int64
	PubKeyHash []byte
	TxID       []byte
	Vout       int
	Spent      bool
	Timestamp  int64
}

// UTXOSetInterface defines the interface for UTXO set operations
type UTXOSetInterface interface {
	// GetUTXO returns the UTXO for a given transaction ID and output index
	GetUTXO(txID []byte, vout int) (*UTXO, error)
	// AddUTXO adds a new UTXO to the set
	AddUTXO(utxo *UTXO) error
	// SpendUTXO marks a UTXO as spent
	SpendUTXO(txID []byte, vout int) error
	// GetBalance returns the balance for a given public key hash
	GetBalance(pubKeyHash []byte) (int64, error)
}

// UTXOSet implements the UTXOSetInterface
type UTXOSet struct {
	utxos map[string]*UTXO // key: txID:vout
	index *UTXOIndex
	mu    sync.RWMutex
}

// NewUTXOSet creates a new UTXO set
func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		utxos: make(map[string]*UTXO),
		index: NewUTXOIndex(),
	}
}

// GetUTXO returns the UTXO for a given transaction ID and output index
func (us *UTXOSet) GetUTXO(txID []byte, vout int) (*UTXO, error) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	// Try to get from index first
	utxo, err := us.index.GetUTXOByTxID(txID, vout)
	if err == nil {
		return utxo, nil
	}

	// Fall back to map lookup
	key := fmt.Sprintf("%x:%d", txID, vout)
	utxo, exists := us.utxos[key]
	if !exists {
		return nil, fmt.Errorf("UTXO not found: %s", key)
	}
	if utxo.Spent {
		return nil, fmt.Errorf("UTXO already spent: %s", key)
	}
	return utxo, nil
}

// AddUTXO adds a new UTXO to the set
func (us *UTXOSet) AddUTXO(utxo *UTXO) error {
	us.mu.Lock()
	defer us.mu.Unlock()

	key := fmt.Sprintf("%x:%d", utxo.TxID, utxo.Vout)
	if _, exists := us.utxos[key]; exists {
		return fmt.Errorf("UTXO already exists: %s", key)
	}

	// Add to map
	us.utxos[key] = utxo

	// Add to index
	if err := us.index.AddUTXO(utxo); err != nil {
		// Remove from map if index update fails
		delete(us.utxos, key)
		return fmt.Errorf("failed to add UTXO to index: %v", err)
	}

	return nil
}

// SpendUTXO marks a UTXO as spent
func (us *UTXOSet) SpendUTXO(txID []byte, vout int) error {
	us.mu.Lock()
	defer us.mu.Unlock()

	key := fmt.Sprintf("%x:%d", txID, vout)
	utxo, exists := us.utxos[key]
	if !exists {
		return fmt.Errorf("UTXO not found: %s", key)
	}
	if utxo.Spent {
		return fmt.Errorf("UTXO already spent: %s", key)
	}

	// Mark as spent
	utxo.Spent = true

	// Remove from index
	if err := us.index.RemoveUTXO(txID, vout); err != nil {
		// Revert spent status if index update fails
		utxo.Spent = false
		return fmt.Errorf("failed to remove UTXO from index: %v", err)
	}

	return nil
}

// GetBalance returns the balance for a given public key hash
func (us *UTXOSet) GetBalance(pubKeyHash []byte) (int64, error) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	// Use index for faster balance calculation
	return us.index.GetBalance(pubKeyHash), nil
}

// GetUTXOsByPubKeyHash returns all UTXOs for a given public key hash
func (us *UTXOSet) GetUTXOsByPubKeyHash(pubKeyHash []byte) []*UTXO {
	us.mu.RLock()
	defer us.mu.RUnlock()

	// Use index for faster lookup
	return us.index.GetUTXOsByPubKeyHash(pubKeyHash)
}

// GetAllUTXOs returns all UTXOs in the set
func (us *UTXOSet) GetAllUTXOs() []*UTXO {
	us.mu.RLock()
	defer us.mu.RUnlock()

	utxos := make([]*UTXO, 0, len(us.utxos))
	for _, utxo := range us.utxos {
		utxos = append(utxos, utxo)
	}
	return utxos
}

// RemoveUTXO removes a UTXO from the set
func (us *UTXOSet) RemoveUTXO(txID []byte, vout int) error {
	us.mu.Lock()
	defer us.mu.Unlock()

	key := fmt.Sprintf("%x:%d", txID, vout)
	if _, exists := us.utxos[key]; !exists {
		return fmt.Errorf("UTXO not found: %s", key)
	}

	// Remove from map
	delete(us.utxos, key)

	// Remove from index
	if err := us.index.RemoveUTXO(txID, vout); err != nil {
		// Add back to map if index update fails
		us.utxos[key] = us.utxos[key]
		return fmt.Errorf("failed to remove UTXO from index: %v", err)
	}

	return nil
}

// Serialize serializes the UTXO set
func (us *UTXOSet) Serialize() []byte {
	us.mu.RLock()
	defer us.mu.RUnlock()

	var result bytes.Buffer

	// Write number of UTXOs
	binary.Write(&result, binary.LittleEndian, int64(len(us.utxos)))

	// Write each UTXO
	for _, utxo := range us.utxos {
		// Write value
		binary.Write(&result, binary.LittleEndian, utxo.Value)

		// Write public key hash
		binary.Write(&result, binary.LittleEndian, int64(len(utxo.PubKeyHash)))
		result.Write(utxo.PubKeyHash)

		// Write transaction ID
		binary.Write(&result, binary.LittleEndian, int64(len(utxo.TxID)))
		result.Write(utxo.TxID)

		// Write output index
		binary.Write(&result, binary.LittleEndian, int64(utxo.Vout))

		// Write spent flag
		binary.Write(&result, binary.LittleEndian, utxo.Spent)

		// Write timestamp
		binary.Write(&result, binary.LittleEndian, utxo.Timestamp)
	}

	return result.Bytes()
}

// DeserializeUTXOSet deserializes a UTXO set
func DeserializeUTXOSet(data []byte) (*UTXOSet, error) {
	us := NewUTXOSet()
	reader := bytes.NewReader(data)

	// Read number of UTXOs
	var count int64
	binary.Read(reader, binary.LittleEndian, &count)

	// Read each UTXO
	for i := int64(0); i < count; i++ {
		utxo := &UTXO{}

		// Read value
		binary.Read(reader, binary.LittleEndian, &utxo.Value)

		// Read public key hash
		var pubKeyHashLen int64
		binary.Read(reader, binary.LittleEndian, &pubKeyHashLen)
		utxo.PubKeyHash = make([]byte, pubKeyHashLen)
		reader.Read(utxo.PubKeyHash)

		// Read transaction ID
		var txIDLen int64
		binary.Read(reader, binary.LittleEndian, &txIDLen)
		utxo.TxID = make([]byte, txIDLen)
		reader.Read(utxo.TxID)

		// Read output index
		var vout int64
		binary.Read(reader, binary.LittleEndian, &vout)
		utxo.Vout = int(vout)

		// Read spent flag
		binary.Read(reader, binary.LittleEndian, &utxo.Spent)

		// Read timestamp
		binary.Read(reader, binary.LittleEndian, &utxo.Timestamp)

		// Add UTXO to set
		key := fmt.Sprintf("%x:%d", utxo.TxID, utxo.Vout)
		us.utxos[key] = utxo
	}

	return us, nil
}
