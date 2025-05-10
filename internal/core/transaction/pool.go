package transaction

import (
	"container/heap"
	"fmt"
	"sync"

	"github.com/youngchain/internal/core/storage"
	"github.com/youngchain/internal/core/types"
)

// TxPool manages the transaction pool
type TxPool struct {
	// Maximum number of transactions in the pool
	maxSize int
	// Minimum fee rate (satoshis per byte)
	minFeeRate uint64
	// Transactions by hash
	txs map[string]*types.Transaction
	// Priority queue for fee-based ordering
	queue *txPriorityQueue
	// UTXO set for validation
	utxoSet *storage.UTXOSet
	mu      sync.RWMutex
}

// txPriorityQueue implements heap.Interface for transaction prioritization
type txPriorityQueue []*types.Transaction

// NewTxPool creates a new transaction pool
func NewTxPool(maxSize int, minFeeRate uint64, utxoSet *storage.UTXOSet) *TxPool {
	pq := &txPriorityQueue{}
	heap.Init(pq)
	return &TxPool{
		maxSize:    maxSize,
		minFeeRate: minFeeRate,
		txs:        make(map[string]*types.Transaction),
		queue:      pq,
		utxoSet:    utxoSet,
	}
}

// Add adds a transaction to the pool
func (p *TxPool) Add(tx *types.Transaction) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if pool is full
	if len(p.txs) >= p.maxSize {
		// Remove lowest fee transaction if new one has higher fee
		if p.queue.Len() > 0 {
			lowestFeeTx := heap.Pop(p.queue).(*types.Transaction)
			if tx.Size() > 0 && lowestFeeTx.Size() > 0 {
				newFeeRate := tx.Fee / uint64(tx.Size())
				lowestFeeRate := lowestFeeTx.Fee / uint64(lowestFeeTx.Size())
				if newFeeRate > lowestFeeRate {
					delete(p.txs, string(lowestFeeTx.Hash))
				} else {
					heap.Push(p.queue, lowestFeeTx)
					return fmt.Errorf("transaction pool is full and new transaction has lower fee rate")
				}
			} else {
				heap.Push(p.queue, lowestFeeTx)
				return fmt.Errorf("transaction pool is full")
			}
		} else {
			return fmt.Errorf("transaction pool is full")
		}
	}

	// Validate transaction
	if err := p.validateTx(tx); err != nil {
		return fmt.Errorf("invalid transaction: %v", err)
	}

	// Add to pool
	p.txs[string(tx.Hash)] = tx
	heap.Push(p.queue, tx)

	return nil
}

// Remove removes a transaction from the pool
func (p *TxPool) Remove(txHash []byte) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.txs[string(txHash)]; exists {
		delete(p.txs, string(txHash))
		// Remove from priority queue
		for i, t := range *p.queue {
			if string(t.Hash) == string(txHash) {
				heap.Remove(p.queue, i)
				break
			}
		}
	}
}

// Get returns a transaction from the pool
func (p *TxPool) Get(txHash []byte) (*types.Transaction, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	tx, exists := p.txs[string(txHash)]
	return tx, exists
}

// GetBest returns the best transactions for mining
func (p *TxPool) GetBest(maxSize int) []*types.Transaction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	txs := make([]*types.Transaction, 0, maxSize)
	tempQueue := make([]*types.Transaction, p.queue.Len())
	copy(tempQueue, *p.queue)
	tempPQ := txPriorityQueue(tempQueue)

	for i := 0; i < maxSize && len(tempQueue) > 0; i++ {
		tx := heap.Pop(&tempPQ).(*types.Transaction)
		txs = append(txs, tx)
	}

	return txs
}

// validateTx validates a transaction
func (p *TxPool) validateTx(tx *types.Transaction) error {
	// Check fee rate
	if tx.Size() > 0 {
		feeRate := tx.Fee / uint64(tx.Size())
		if feeRate < p.minFeeRate {
			return fmt.Errorf("fee rate too low")
		}
	}

	// Check inputs
	var totalInput uint64
	for _, input := range tx.Inputs {
		utxo, exists := p.utxoSet.GetUTXO(input.PreviousTxHash, input.PreviousTxIndex)
		if !exists {
			return fmt.Errorf("input not found in UTXO set")
		}
		totalInput += utxo.Value
	}

	// Check outputs
	var totalOutput uint64
	for _, output := range tx.Outputs {
		totalOutput += output.Value
	}

	// Check if inputs cover outputs plus fee
	if totalInput < totalOutput+tx.Fee {
		return fmt.Errorf("insufficient input amount")
	}

	return nil
}

// Len returns the number of transactions in the pool
func (p *TxPool) Len() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.txs)
}

// Clear clears the transaction pool
func (p *TxPool) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.txs = make(map[string]*types.Transaction)
	p.queue = &txPriorityQueue{}
	heap.Init(p.queue)
}

// Implement heap.Interface for txPriorityQueue
func (pq txPriorityQueue) Len() int { return len(pq) }

func (pq txPriorityQueue) Less(i, j int) bool {
	if pq[i].Size() == 0 || pq[j].Size() == 0 {
		return false
	}
	return pq[i].Fee/uint64(pq[i].Size()) > pq[j].Fee/uint64(pq[j].Size())
}

func (pq txPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *txPriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*types.Transaction))
}

func (pq *txPriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	x := old[n-1]
	*pq = old[0 : n-1]
	return x
}
