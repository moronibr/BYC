package block

import (
	"errors"
	"sync"
	"time"

	"github.com/youngchain/internal/core/common"
)

// ReplayProtection manages transaction replay protection
type ReplayProtection struct {
	// Map of transaction hashes to their first seen timestamp
	seenTxs map[string]time.Time
	// Maximum time a transaction can be replayed
	maxReplayTime time.Duration
	mu            sync.RWMutex
}

// NewReplayProtection creates a new replay protection system
func NewReplayProtection(maxReplayTime time.Duration) *ReplayProtection {
	return &ReplayProtection{
		seenTxs:       make(map[string]time.Time),
		maxReplayTime: maxReplayTime,
	}
}

// IsReplay checks if a transaction is a replay
func (rp *ReplayProtection) IsReplay(txHash []byte) bool {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	// Convert hash to string for map key
	hashStr := string(txHash)

	// Check if we've seen this transaction before
	if firstSeen, exists := rp.seenTxs[hashStr]; exists {
		// If the transaction is older than maxReplayTime, it's not a replay
		if time.Since(firstSeen) > rp.maxReplayTime {
			delete(rp.seenTxs, hashStr)
			return false
		}
		return true
	}

	return false
}

// AddTransaction adds a transaction to the replay protection system
func (rp *ReplayProtection) AddTransaction(txHash []byte) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	hashStr := string(txHash)
	rp.seenTxs[hashStr] = time.Now()
}

// Cleanup removes old transactions from the replay protection system
func (rp *ReplayProtection) Cleanup() {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	now := time.Now()
	for hash, firstSeen := range rp.seenTxs {
		if now.Sub(firstSeen) > rp.maxReplayTime {
			delete(rp.seenTxs, hash)
		}
	}
}

// GetSeenCount returns the number of transactions in the replay protection system
func (rp *ReplayProtection) GetSeenCount() int {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	return len(rp.seenTxs)
}

// ValidateTransaction validates a transaction against replay protection
func (rp *ReplayProtection) ValidateTransaction(tx *common.Transaction) error {
	// Check if transaction is a replay
	if rp.IsReplay(tx.Hash) {
		return errors.New("transaction is a replay")
	}

	// Add transaction to replay protection
	rp.AddTransaction(tx.Hash)

	return nil
}
