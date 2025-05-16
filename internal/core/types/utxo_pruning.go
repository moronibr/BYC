package types

import (
	"fmt"
	"sync"
	"time"
)

const (
	// DefaultPruneAge is the default age in days after which spent UTXOs are pruned
	DefaultPruneAge = 30
	// DefaultPruneBatchSize is the default number of UTXOs to prune in a single batch
	DefaultPruneBatchSize = 1000
)

// UTXOPruningConfig holds configuration for UTXO pruning
type UTXOPruningConfig struct {
	// PruneAge is the age in days after which spent UTXOs are pruned
	PruneAge int
	// BatchSize is the number of UTXOs to prune in a single batch
	BatchSize int
	// MinUTXOs is the minimum number of UTXOs to keep
	MinUTXOs int
}

// DefaultPruningConfig returns the default pruning configuration
func DefaultPruningConfig() *UTXOPruningConfig {
	return &UTXOPruningConfig{
		PruneAge:  DefaultPruneAge,
		BatchSize: DefaultPruneBatchSize,
		MinUTXOs:  1000,
	}
}

// UTXOPruning handles pruning of spent UTXOs
type UTXOPruning struct {
	utxoSet *UTXOSet
	config  *UTXOPruningConfig
	mu      sync.RWMutex
}

// NewUTXOPruning creates a new UTXO pruning handler
func NewUTXOPruning(utxoSet *UTXOSet, config *UTXOPruningConfig) *UTXOPruning {
	if config == nil {
		config = DefaultPruningConfig()
	}
	return &UTXOPruning{
		utxoSet: utxoSet,
		config:  config,
	}
}

// Prune removes spent UTXOs that are older than the configured age
func (up *UTXOPruning) Prune() error {
	up.mu.Lock()
	defer up.mu.Unlock()

	// Get current time
	now := time.Now().Unix()

	// Calculate cutoff time
	cutoff := now - int64(up.config.PruneAge*24*60*60)

	// Get all UTXOs
	utxos := up.utxoSet.GetAllUTXOs()

	// Count spent UTXOs
	var spentCount int
	for _, utxo := range utxos {
		if utxo.Spent && utxo.Timestamp < cutoff {
			spentCount++
		}
	}

	// Check if we have enough UTXOs to prune
	if len(utxos)-spentCount < up.config.MinUTXOs {
		return fmt.Errorf("not enough UTXOs to prune (min: %d, available: %d)", up.config.MinUTXOs, len(utxos)-spentCount)
	}

	// Prune in batches
	for i := 0; i < spentCount; i += up.config.BatchSize {
		end := i + up.config.BatchSize
		if end > spentCount {
			end = spentCount
		}

		// Get batch of UTXOs to prune
		batch := make([]*UTXO, 0, end-i)
		for _, utxo := range utxos {
			if utxo.Spent && utxo.Timestamp < cutoff {
				batch = append(batch, utxo)
				if len(batch) == end-i {
					break
				}
			}
		}

		// Remove UTXOs from set
		for _, utxo := range batch {
			if err := up.utxoSet.RemoveUTXO(utxo.TxID, utxo.Vout); err != nil {
				return fmt.Errorf("failed to remove UTXO: %v", err)
			}
		}
	}

	return nil
}

// GetPruningStats returns statistics about the UTXO set
func (up *UTXOPruning) GetPruningStats() *PruningStats {
	up.mu.RLock()
	defer up.mu.RUnlock()

	// Get all UTXOs
	utxos := up.utxoSet.GetAllUTXOs()

	// Calculate statistics
	stats := &PruningStats{
		TotalUTXOs:    len(utxos),
		SpentUTXOs:    0,
		UnspentUTXOs:  0,
		PrunableUTXOs: 0,
	}

	// Get current time
	now := time.Now().Unix()

	// Calculate cutoff time
	cutoff := now - int64(up.config.PruneAge*24*60*60)

	// Count UTXOs
	for _, utxo := range utxos {
		if utxo.Spent {
			stats.SpentUTXOs++
			if utxo.Timestamp < cutoff {
				stats.PrunableUTXOs++
			}
		} else {
			stats.UnspentUTXOs++
		}
	}

	return stats
}

// PruningStats holds statistics about the UTXO set
type PruningStats struct {
	// TotalUTXOs is the total number of UTXOs
	TotalUTXOs int
	// SpentUTXOs is the number of spent UTXOs
	SpentUTXOs int
	// UnspentUTXOs is the number of unspent UTXOs
	UnspentUTXOs int
	// PrunableUTXOs is the number of UTXOs that can be pruned
	PrunableUTXOs int
}

// String returns a string representation of the pruning statistics
func (ps *PruningStats) String() string {
	return fmt.Sprintf("Total UTXOs: %d, Spent: %d, Unspent: %d, Prunable: %d",
		ps.TotalUTXOs, ps.SpentUTXOs, ps.UnspentUTXOs, ps.PrunableUTXOs)
}
