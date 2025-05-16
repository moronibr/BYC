package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

const (
	// DefaultIndexSize is the default size of the index
	DefaultIndexSize = 1000000
	// DefaultIndexLoadFactor is the default load factor for the index
	DefaultIndexLoadFactor = 0.75
)

// IndexType represents the type of index
type IndexType byte

const (
	// IndexTypeNone indicates no index
	IndexTypeNone IndexType = iota
	// IndexTypeHash indicates hash-based index
	IndexTypeHash
	// IndexTypeBTree indicates B-tree index
	IndexTypeBTree
)

// IndexEntry represents an entry in the index
type IndexEntry struct {
	// Key is the index key
	Key []byte
	// Value is the indexed value
	Value interface{}
	// Size is the size of the indexed value in bytes
	Size int64
}

// UTXOIndex provides fast lookups for UTXOs
type UTXOIndex struct {
	// Index by public key hash for balance lookups
	byPubKeyHash map[string][]*UTXO
	// Index by transaction ID and output index for input validation
	byTxID map[string]map[int]*UTXO
	mu     sync.RWMutex

	// Index state
	indexType   IndexType
	index       map[string]*IndexEntry
	size        int
	loadFactor  float64
	rehashSize  int
	rehashCount int
	rehashMutex sync.Mutex
}

// NewUTXOIndex creates a new UTXO index
func NewUTXOIndex() *UTXOIndex {
	return &UTXOIndex{
		byPubKeyHash: make(map[string][]*UTXO),
		byTxID:       make(map[string]map[int]*UTXO),
		indexType:    IndexTypeNone,
		index:        make(map[string]*IndexEntry, DefaultIndexSize),
		size:         DefaultIndexSize,
		loadFactor:   DefaultIndexLoadFactor,
	}
}

// AddUTXO adds a UTXO to the index
func (idx *UTXOIndex) AddUTXO(utxo *UTXO) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Add to public key hash index
	pubKeyHashStr := fmt.Sprintf("%x", utxo.PubKeyHash)
	idx.byPubKeyHash[pubKeyHashStr] = append(idx.byPubKeyHash[pubKeyHashStr], utxo)

	// Add to transaction ID index
	txIDStr := fmt.Sprintf("%x", utxo.TxID)
	if _, exists := idx.byTxID[txIDStr]; !exists {
		idx.byTxID[txIDStr] = make(map[int]*UTXO)
	}
	idx.byTxID[txIDStr][utxo.Vout] = utxo

	return nil
}

// RemoveUTXO removes a UTXO from the index
func (idx *UTXOIndex) RemoveUTXO(txID []byte, vout int) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	txIDStr := fmt.Sprintf("%x", txID)
	utxoMap, exists := idx.byTxID[txIDStr]
	if !exists {
		return fmt.Errorf("UTXO not found in index: %x:%d", txID, vout)
	}

	utxo, exists := utxoMap[vout]
	if !exists {
		return fmt.Errorf("UTXO not found in index: %x:%d", txID, vout)
	}

	// Remove from public key hash index
	pubKeyHashStr := fmt.Sprintf("%x", utxo.PubKeyHash)
	utxos := idx.byPubKeyHash[pubKeyHashStr]
	for i, u := range utxos {
		if bytes.Equal(u.TxID, txID) && u.Vout == vout {
			// Remove UTXO from slice
			utxos = append(utxos[:i], utxos[i+1:]...)
			if len(utxos) == 0 {
				delete(idx.byPubKeyHash, pubKeyHashStr)
			} else {
				idx.byPubKeyHash[pubKeyHashStr] = utxos
			}
			break
		}
	}

	// Remove from transaction ID index
	delete(utxoMap, vout)
	if len(utxoMap) == 0 {
		delete(idx.byTxID, txIDStr)
	}

	return nil
}

// GetUTXOsByPubKeyHash returns all UTXOs for a given public key hash
func (idx *UTXOIndex) GetUTXOsByPubKeyHash(pubKeyHash []byte) []*UTXO {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	pubKeyHashStr := fmt.Sprintf("%x", pubKeyHash)
	utxos := idx.byPubKeyHash[pubKeyHashStr]
	if utxos == nil {
		return []*UTXO{}
	}

	// Return a copy to prevent modification
	result := make([]*UTXO, len(utxos))
	copy(result, utxos)
	return result
}

// GetUTXOByTxID returns a UTXO for a given transaction ID and output index
func (idx *UTXOIndex) GetUTXOByTxID(txID []byte, vout int) (*UTXO, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	txIDStr := fmt.Sprintf("%x", txID)
	utxoMap, exists := idx.byTxID[txIDStr]
	if !exists {
		return nil, fmt.Errorf("UTXO not found in index: %x:%d", txID, vout)
	}

	utxo, exists := utxoMap[vout]
	if !exists {
		return nil, fmt.Errorf("UTXO not found in index: %x:%d", txID, vout)
	}

	return utxo, nil
}

// GetBalance returns the balance for a given public key hash
func (idx *UTXOIndex) GetBalance(pubKeyHash []byte) int64 {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var balance int64
	pubKeyHashStr := fmt.Sprintf("%x", pubKeyHash)
	for _, utxo := range idx.byPubKeyHash[pubKeyHashStr] {
		if !utxo.Spent {
			balance += utxo.Value
		}
	}
	return balance
}

// UpdateFromUTXOSet updates the index from a UTXO set
func (idx *UTXOIndex) UpdateFromUTXOSet(utxoSet *UTXOSet) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Clear existing indexes
	idx.byPubKeyHash = make(map[string][]*UTXO)
	idx.byTxID = make(map[string]map[int]*UTXO)

	// Add all UTXOs to indexes
	for _, utxo := range utxoSet.utxos {
		if err := idx.AddUTXO(utxo); err != nil {
			return fmt.Errorf("failed to add UTXO to index: %v", err)
		}
	}

	return nil
}

// Size returns the number of UTXOs in the index
func (idx *UTXOIndex) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var count int
	for _, utxos := range idx.byPubKeyHash {
		count += len(utxos)
	}
	return count
}

// Clear clears the index
func (idx *UTXOIndex) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.byPubKeyHash = make(map[string][]*UTXO)
	idx.byTxID = make(map[string]map[int]*UTXO)
}

// Get retrieves a value from the index
func (ui *UTXOIndex) Get(key []byte) (interface{}, bool) {
	ui.mu.RLock()
	defer ui.mu.RUnlock()

	// Check if index is enabled
	if ui.indexType == IndexTypeNone {
		return nil, false
	}

	// Get entry
	entry, exists := ui.index[string(key)]
	if !exists {
		return nil, false
	}

	return entry.Value, true
}

// Set adds a value to the index
func (ui *UTXOIndex) Set(key []byte, value interface{}, size int64) error {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	// Check if index is enabled
	if ui.indexType == IndexTypeNone {
		return fmt.Errorf("index is not enabled")
	}

	// Create entry
	entry := &IndexEntry{
		Key:   key,
		Value: value,
		Size:  size,
	}

	// Add entry to index
	ui.index[string(key)] = entry

	// Check if rehash is needed
	if float64(len(ui.index))/float64(ui.size) >= ui.loadFactor {
		ui.rehashMutex.Lock()
		defer ui.rehashMutex.Unlock()

		// Double size
		newSize := ui.size * 2

		// Create new index
		newIndex := make(map[string]*IndexEntry, newSize)

		// Copy entries
		for k, v := range ui.index {
			newIndex[k] = v
		}

		// Update index
		ui.index = newIndex
		ui.size = newSize
		ui.rehashCount++
	}

	return nil
}

// Remove removes a value from the index
func (ui *UTXOIndex) Remove(key []byte) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	// Check if index is enabled
	if ui.indexType == IndexTypeNone {
		return
	}

	// Remove entry
	delete(ui.index, string(key))
}

// GetIndexStats returns statistics about the index
func (ui *UTXOIndex) GetIndexStats() *IndexStats {
	ui.mu.RLock()
	defer ui.mu.RUnlock()

	stats := &IndexStats{
		IndexType:   ui.indexType,
		Size:        ui.size,
		LoadFactor:  ui.loadFactor,
		EntryCount:  len(ui.index),
		RehashCount: ui.rehashCount,
	}

	// Calculate load
	stats.Load = float64(stats.EntryCount) / float64(stats.Size)

	return stats
}

// SetIndexType sets the type of index
func (ui *UTXOIndex) SetIndexType(indexType IndexType) {
	ui.mu.Lock()
	ui.indexType = indexType
	ui.mu.Unlock()
}

// SetSize sets the size of the index
func (ui *UTXOIndex) SetSize(size int) {
	ui.mu.Lock()
	ui.size = size
	ui.mu.Unlock()
}

// SetLoadFactor sets the load factor for the index
func (ui *UTXOIndex) SetLoadFactor(loadFactor float64) {
	ui.mu.Lock()
	ui.loadFactor = loadFactor
	ui.mu.Unlock()
}

// IndexStats holds statistics about the index
type IndexStats struct {
	// IndexType is the type of index
	IndexType IndexType
	// Size is the size of the index
	Size int
	// LoadFactor is the load factor for the index
	LoadFactor float64
	// EntryCount is the number of entries in the index
	EntryCount int
	// RehashCount is the number of times the index has been rehashed
	RehashCount int
	// Load is the current load of the index
	Load float64
}

// String returns a string representation of the index statistics
func (is *IndexStats) String() string {
	return fmt.Sprintf(
		"Index Type: %d\n"+
			"Size: %d, Entries: %d\n"+
			"Load Factor: %.2f, Load: %.2f\n"+
			"Rehash Count: %d",
		is.IndexType,
		is.Size, is.EntryCount,
		is.LoadFactor, is.Load,
		is.RehashCount,
	)
}

// IndexKey generates an index key from a transaction hash and output index
func IndexKey(txHash []byte, outputIndex uint32) []byte {
	key := make([]byte, len(txHash)+4)
	copy(key, txHash)
	binary.LittleEndian.PutUint32(key[len(txHash):], outputIndex)
	return key
}

// ParseIndexKey parses an index key into a transaction hash and output index
func ParseIndexKey(key []byte) ([]byte, uint32) {
	if len(key) < 4 {
		return nil, 0
	}
	txHash := make([]byte, len(key)-4)
	copy(txHash, key[:len(key)-4])
	outputIndex := binary.LittleEndian.Uint32(key[len(key)-4:])
	return txHash, outputIndex
}
