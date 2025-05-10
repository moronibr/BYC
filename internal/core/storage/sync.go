package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

// SyncMode represents the synchronization mode
type SyncMode uint8

const (
	// Sync modes
	FullSync SyncMode = iota
	FastSync
	LightSync
)

// SyncManager manages blockchain synchronization
type SyncManager struct {
	mode          SyncMode
	currentHeight uint64
	targetHeight  uint64
	checkpoint    uint64
	utxoSet       *UTXOSet
	blockStore    BlockStore
	mu            sync.RWMutex
}

// UTXOSet manages the set of unspent transaction outputs
type UTXOSet struct {
	utxos map[string]*UTXO
	spent map[string]uint64 // Map of spent UTXOs to their spending height
	mu    sync.RWMutex
}

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash      []byte
	OutputIndex uint32
	Value       uint64
	Script      []byte
	Height      uint64
}

// NewSyncManager creates a new sync manager
func NewSyncManager(mode SyncMode, blockStore BlockStore) *SyncManager {
	return &SyncManager{
		mode:       mode,
		blockStore: blockStore,
		utxoSet:    NewUTXOSet(),
	}
}

// NewUTXOSet creates a new UTXO set
func NewUTXOSet() *UTXOSet {
	return &UTXOSet{
		utxos: make(map[string]*UTXO),
		spent: make(map[string]uint64),
	}
}

// StartFastSync starts fast synchronization
func (sm *SyncManager) StartFastSync(checkpoint uint64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.mode = FastSync
	sm.checkpoint = checkpoint

	// Download block headers
	if err := sm.downloadHeaders(); err != nil {
		return err
	}

	// Download UTXO set
	if err := sm.downloadUTXOSet(); err != nil {
		return err
	}

	return nil
}

// downloadHeaders downloads block headers
func (sm *SyncManager) downloadHeaders() error {
	// TODO: Implement header download
	return nil
}

// downloadUTXOSet downloads the UTXO set
func (sm *SyncManager) downloadUTXOSet() error {
	// TODO: Implement UTXO set download
	return nil
}

// AddUTXO adds a UTXO to the set
func (us *UTXOSet) AddUTXO(utxo *UTXO) {
	us.mu.Lock()
	defer us.mu.Unlock()

	key := fmt.Sprintf("%x:%d", utxo.TxHash, utxo.OutputIndex)
	us.utxos[key] = utxo
}

// SpendUTXO marks a UTXO as spent
func (us *UTXOSet) SpendUTXO(txHash []byte, outputIndex uint32, height uint64) {
	us.mu.Lock()
	defer us.mu.Unlock()

	key := fmt.Sprintf("%x:%d", txHash, outputIndex)
	if utxo, exists := us.utxos[key]; exists {
		us.spent[key] = height
		delete(us.utxos, key)
	}
}

// GetUTXO gets a UTXO
func (us *UTXOSet) GetUTXO(txHash []byte, outputIndex uint32) (*UTXO, bool) {
	us.mu.RLock()
	defer us.mu.RUnlock()

	key := fmt.Sprintf("%x:%d", txHash, outputIndex)
	utxo, exists := us.utxos[key]
	return utxo, exists
}

// PruneSpent prunes spent UTXOs up to a certain height
func (us *UTXOSet) PruneSpent(height uint64) error {
	us.mu.Lock()
	defer us.mu.Unlock()

	for key, spentHeight := range us.spent {
		if spentHeight <= height {
			delete(us.spent, key)
		}
	}

	return nil
}

// GetUTXOCount returns the number of UTXOs
func (us *UTXOSet) GetUTXOCount() int {
	us.mu.RLock()
	defer us.mu.RUnlock()

	return len(us.utxos)
}

// GetSpentCount returns the number of spent UTXOs
func (us *UTXOSet) GetSpentCount() int {
	us.mu.RLock()
	defer us.mu.RUnlock()

	return len(us.spent)
}

// Serialize serializes the UTXO set
func (us *UTXOSet) Serialize() []byte {
	us.mu.RLock()
	defer us.mu.RUnlock()

	var buf bytes.Buffer

	// Write UTXO count
	binary.Write(&buf, binary.LittleEndian, uint32(len(us.utxos)))

	// Write UTXOs
	for _, utxo := range us.utxos {
		// Write transaction hash
		buf.Write(utxo.TxHash)

		// Write output index
		binary.Write(&buf, binary.LittleEndian, utxo.OutputIndex)

		// Write value
		binary.Write(&buf, binary.LittleEndian, utxo.Value)

		// Write script length and script
		binary.Write(&buf, binary.LittleEndian, uint32(len(utxo.Script)))
		buf.Write(utxo.Script)

		// Write height
		binary.Write(&buf, binary.LittleEndian, utxo.Height)
	}

	return buf.Bytes()
}

// Deserialize deserializes the UTXO set
func (us *UTXOSet) Deserialize(data []byte) error {
	us.mu.Lock()
	defer us.mu.Unlock()

	buf := bytes.NewReader(data)

	// Read UTXO count
	var count uint32
	if err := binary.Read(buf, binary.LittleEndian, &count); err != nil {
		return err
	}

	// Read UTXOs
	for i := uint32(0); i < count; i++ {
		utxo := &UTXO{}

		// Read transaction hash
		utxo.TxHash = make([]byte, 32)
		if _, err := buf.Read(utxo.TxHash); err != nil {
			return err
		}

		// Read output index
		if err := binary.Read(buf, binary.LittleEndian, &utxo.OutputIndex); err != nil {
			return err
		}

		// Read value
		if err := binary.Read(buf, binary.LittleEndian, &utxo.Value); err != nil {
			return err
		}

		// Read script
		var scriptLen uint32
		if err := binary.Read(buf, binary.LittleEndian, &scriptLen); err != nil {
			return err
		}
		utxo.Script = make([]byte, scriptLen)
		if _, err := buf.Read(utxo.Script); err != nil {
			return err
		}

		// Read height
		if err := binary.Read(buf, binary.LittleEndian, &utxo.Height); err != nil {
			return err
		}

		// Add UTXO to set
		key := fmt.Sprintf("%x:%d", utxo.TxHash, utxo.OutputIndex)
		us.utxos[key] = utxo
	}

	return nil
}
