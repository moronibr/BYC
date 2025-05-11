package transaction

import (
	"container/heap"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/youngchain/internal/core/types"
)

var (
	ErrMempoolEmpty = errors.New("mempool is empty")
)

// MempoolEntry represents a transaction in the mempool
type MempoolEntry struct {
	Tx       *types.Transaction
	AddedAt  time.Time
	Fee      int64
	Priority float64
	index    int // Used by heap.Interface
}

// Mempool manages pending transactions
type Mempool struct {
	entries map[string]*MempoolEntry // Map of tx hash to entry
	queue   MempoolQueue             // Priority queue
	mu      sync.RWMutex
	maxSize int
}

// MempoolQueue implements heap.Interface
type MempoolQueue []*MempoolEntry

func (q MempoolQueue) Len() int { return len(q) }

func (q MempoolQueue) Less(i, j int) bool {
	return q[i].Priority > q[j].Priority // Higher priority first
}

func (q MempoolQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *MempoolQueue) Push(x interface{}) {
	n := len(*q)
	entry := x.(*MempoolEntry)
	entry.index = n
	*q = append(*q, entry)
}

func (q *MempoolQueue) Pop() interface{} {
	old := *q
	n := len(old)
	entry := old[n-1]
	old[n-1] = nil
	entry.index = -1
	*q = old[0 : n-1]
	return entry
}

// NewMempool creates a new mempool
func NewMempool(maxSize int) *Mempool {
	return &Mempool{
		entries: make(map[string]*MempoolEntry),
		queue:   make(MempoolQueue, 0),
		maxSize: maxSize,
	}
}

// Add adds a transaction to the mempool
func (mp *Mempool) Add(tx *types.Transaction, fee int64) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	// Check if mempool is full
	if len(mp.entries) >= mp.maxSize {
		// Remove lowest priority transaction
		if err := mp.removeLowestPriority(); err != nil {
			return err
		}
	}

	// Create entry
	entry := &MempoolEntry{
		Tx:       tx,
		AddedAt:  time.Now(),
		Fee:      fee,
		Priority: mp.calculatePriority(tx, fee),
	}

	// Add to map
	txHash := hex.EncodeToString(tx.Hash)
	mp.entries[txHash] = entry

	// Add to queue
	heap.Push(&mp.queue, entry)

	return nil
}

// Remove removes a transaction from the mempool
func (mp *Mempool) Remove(txHash string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if entry, exists := mp.entries[txHash]; exists {
		delete(mp.entries, txHash)
		heap.Remove(&mp.queue, entry.index)
	}
}

// Get returns a transaction from the mempool
func (mp *Mempool) Get(txHash string) (*types.Transaction, bool) {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	entry, exists := mp.entries[txHash]
	if !exists {
		return nil, false
	}
	return entry.Tx, true
}

// GetAll returns all transactions in the mempool
func (mp *Mempool) GetAll() []*types.Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	txs := make([]*types.Transaction, 0, len(mp.entries))
	for _, entry := range mp.entries {
		txs = append(txs, entry.Tx)
	}
	return txs
}

// IsInputSpent checks if an input is already spent in the mempool
func (mp *Mempool) IsInputSpent(input *types.TxInput) bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	for _, entry := range mp.entries {
		for _, txInput := range entry.Tx.Inputs {
			if hex.EncodeToString(txInput.PreviousTxHash) == hex.EncodeToString(input.PreviousTxHash) &&
				txInput.PreviousTxIndex == input.PreviousTxIndex {
				return true
			}
		}
	}
	return false
}

// removeLowestPriority removes the lowest priority transaction
func (mp *Mempool) removeLowestPriority() error {
	if len(mp.queue) == 0 {
		return ErrMempoolEmpty
	}

	entry := heap.Pop(&mp.queue).(*MempoolEntry)
	delete(mp.entries, hex.EncodeToString(entry.Tx.Hash))
	return nil
}

// calculatePriority calculates the priority of a transaction
func (mp *Mempool) calculatePriority(tx *types.Transaction, fee int64) float64 {
	// Base priority on fee and age
	age := time.Since(tx.Timestamp).Seconds()
	return float64(fee) * (1 + age/86400) // 86400 seconds in a day
}

// Supply limits
const (
	MaxEphraimSupply  = 15_000_000
	MaxManassehSupply = 15_000_000
	MaxJosephSupply   = 3_000_000
)
