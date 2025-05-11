package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/youngchain/internal/core/types"
)

// IndexType represents the type of index
type IndexType uint8

const (
	// Index types
	TransactionIndex IndexType = iota
	AddressIndexType
	BlockIndexType
)

// IndexManager manages blockchain indices
type IndexManager struct {
	indices map[IndexType]Index
	mu      sync.RWMutex
}

// Index represents a blockchain index
type Index interface {
	Add(key []byte, value interface{}) error
	Get(key []byte) (interface{}, bool)
	Delete(key []byte) error
	Serialize() []byte
	Deserialize(data []byte) error
}

// TransactionIndexData manages transaction indices
type TransactionIndexData struct {
	txs map[string]*types.Transaction
	mu  sync.RWMutex
}

// AddressIndexData manages address indices
type AddressIndexData struct {
	addresses map[string][]string // Map of address to transaction hashes
	mu        sync.RWMutex
}

// BlockIndexData manages block indices
type BlockIndexData struct {
	blocks map[uint64][]byte // Map of height to block hash
	mu     sync.RWMutex
}

// NewIndexManager creates a new index manager
func NewIndexManager() *IndexManager {
	return &IndexManager{
		indices: make(map[IndexType]Index),
	}
}

// AddIndex adds an index
func (im *IndexManager) AddIndex(indexType IndexType, index Index) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.indices[indexType] = index
}

// GetIndex gets an index
func (im *IndexManager) GetIndex(indexType IndexType) (Index, bool) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	index, exists := im.indices[indexType]
	return index, exists
}

// NewTransactionIndex creates a new transaction index
func NewTransactionIndex() *TransactionIndexData {
	return &TransactionIndexData{
		txs: make(map[string]*types.Transaction),
	}
}

// Add adds a transaction to the index
func (ti *TransactionIndexData) Add(key []byte, value interface{}) error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	tx, ok := value.(*types.Transaction)
	if !ok {
		return fmt.Errorf("invalid value type")
	}

	ti.txs[string(key)] = tx
	return nil
}

// Get gets a transaction from the index
func (ti *TransactionIndexData) Get(key []byte) (interface{}, bool) {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	tx, exists := ti.txs[string(key)]
	return tx, exists
}

// Delete deletes a transaction from the index
func (ti *TransactionIndexData) Delete(key []byte) error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	delete(ti.txs, string(key))
	return nil
}

// NewAddressIndex creates a new address index
func NewAddressIndex() *AddressIndexData {
	return &AddressIndexData{
		addresses: make(map[string][]string),
	}
}

// Add adds an address to the index
func (ai *AddressIndexData) Add(key []byte, value interface{}) error {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	txHash, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid value type")
	}

	addr := string(key)
	ai.addresses[addr] = append(ai.addresses[addr], txHash)
	return nil
}

// Get gets transactions for an address
func (ai *AddressIndexData) Get(key []byte) (interface{}, bool) {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	txs, exists := ai.addresses[string(key)]
	return txs, exists
}

// Delete deletes an address from the index
func (ai *AddressIndexData) Delete(key []byte) error {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	delete(ai.addresses, string(key))
	return nil
}

// NewBlockIndex creates a new block index
func NewBlockIndex() *BlockIndexData {
	return &BlockIndexData{
		blocks: make(map[uint64][]byte),
	}
}

// Add adds a block to the index
func (bi *BlockIndexData) Add(key []byte, value interface{}) error {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	height := binary.BigEndian.Uint64(key)
	hash, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid value type")
	}

	bi.blocks[height] = hash
	return nil
}

// Get gets a block hash from the index
func (bi *BlockIndexData) Get(key []byte) (interface{}, bool) {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	height := binary.BigEndian.Uint64(key)
	hash, exists := bi.blocks[height]
	return hash, exists
}

// Delete deletes a block from the index
func (bi *BlockIndexData) Delete(key []byte) error {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	height := binary.BigEndian.Uint64(key)
	delete(bi.blocks, height)
	return nil
}

// Serialize serializes the transaction index
func (ti *TransactionIndexData) Serialize() []byte {
	ti.mu.RLock()
	defer ti.mu.RUnlock()

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(len(ti.txs)))

	for hash, tx := range ti.txs {
		// Write hash length and hash
		binary.Write(buf, binary.BigEndian, uint32(len(hash)))
		buf.WriteString(hash)

		// Write transaction data
		txData := tx.Serialize()
		binary.Write(buf, binary.BigEndian, uint32(len(txData)))
		buf.Write(txData)
	}

	return buf.Bytes()
}

// Deserialize deserializes the transaction index
func (ti *TransactionIndexData) Deserialize(data []byte) error {
	ti.mu.Lock()
	defer ti.mu.Unlock()

	buf := bytes.NewReader(data)

	// Read number of transactions
	var numTxs uint32
	if err := binary.Read(buf, binary.BigEndian, &numTxs); err != nil {
		return err
	}

	ti.txs = make(map[string]*types.Transaction)

	for i := uint32(0); i < numTxs; i++ {
		// Read hash length and hash
		var hashLen uint32
		if err := binary.Read(buf, binary.BigEndian, &hashLen); err != nil {
			return err
		}
		hash := make([]byte, hashLen)
		if _, err := buf.Read(hash); err != nil {
			return err
		}

		// Read transaction data length and data
		var txDataLen uint32
		if err := binary.Read(buf, binary.BigEndian, &txDataLen); err != nil {
			return err
		}
		txData := make([]byte, txDataLen)
		if _, err := buf.Read(txData); err != nil {
			return err
		}

		// Deserialize transaction
		tx := &types.Transaction{}
		if err := tx.Deserialize(bytes.NewBuffer(txData)); err != nil {
			return err
		}

		ti.txs[string(hash)] = tx
	}

	return nil
}

// Serialize serializes the address index
func (ai *AddressIndexData) Serialize() []byte {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	var buf bytes.Buffer

	// Write address count
	binary.Write(&buf, binary.LittleEndian, uint32(len(ai.addresses)))

	// Write addresses
	for addr, txs := range ai.addresses {
		// Write address length and address
		binary.Write(&buf, binary.LittleEndian, uint32(len(addr)))
		buf.WriteString(addr)

		// Write transaction count
		binary.Write(&buf, binary.LittleEndian, uint32(len(txs)))

		// Write transactions
		for _, tx := range txs {
			binary.Write(&buf, binary.LittleEndian, uint32(len(tx)))
			buf.WriteString(tx)
		}
	}

	return buf.Bytes()
}

// Deserialize deserializes the address index
func (ai *AddressIndexData) Deserialize(data []byte) error {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	buf := bytes.NewReader(data)

	// Read address count
	var count uint32
	if err := binary.Read(buf, binary.LittleEndian, &count); err != nil {
		return err
	}

	// Read addresses
	for i := uint32(0); i < count; i++ {
		// Read address
		var addrLen uint32
		if err := binary.Read(buf, binary.LittleEndian, &addrLen); err != nil {
			return err
		}
		addr := make([]byte, addrLen)
		if _, err := buf.Read(addr); err != nil {
			return err
		}

		// Read transaction count
		var txCount uint32
		if err := binary.Read(buf, binary.LittleEndian, &txCount); err != nil {
			return err
		}

		// Read transactions
		txs := make([]string, txCount)
		for j := uint32(0); j < txCount; j++ {
			var txLen uint32
			if err := binary.Read(buf, binary.LittleEndian, &txLen); err != nil {
				return err
			}
			tx := make([]byte, txLen)
			if _, err := buf.Read(tx); err != nil {
				return err
			}
			txs[j] = string(tx)
		}

		ai.addresses[string(addr)] = txs
	}

	return nil
}

// Serialize serializes the block index
func (bi *BlockIndexData) Serialize() []byte {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	var buf bytes.Buffer

	// Write block count
	binary.Write(&buf, binary.LittleEndian, uint32(len(bi.blocks)))

	// Write blocks
	for height, hash := range bi.blocks {
		// Write height
		binary.Write(&buf, binary.LittleEndian, height)

		// Write hash
		buf.Write(hash)
	}

	return buf.Bytes()
}

// Deserialize deserializes the block index
func (bi *BlockIndexData) Deserialize(data []byte) error {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	buf := bytes.NewReader(data)

	// Read block count
	var count uint32
	if err := binary.Read(buf, binary.LittleEndian, &count); err != nil {
		return err
	}

	// Read blocks
	for i := uint32(0); i < count; i++ {
		// Read height
		var height uint64
		if err := binary.Read(buf, binary.LittleEndian, &height); err != nil {
			return err
		}

		// Read hash
		hash := make([]byte, 32)
		if _, err := buf.Read(hash); err != nil {
			return err
		}

		bi.blocks[height] = hash
	}

	return nil
}
