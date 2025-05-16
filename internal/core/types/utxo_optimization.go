package types

import (
	"fmt"
	"sync"
	"time"
)

const (
	// DefaultOptimizationInterval is the default interval for optimization in seconds
	DefaultOptimizationInterval = 3600 // 1 hour
	// DefaultCompactionThreshold is the default threshold for compaction
	DefaultCompactionThreshold = 0.7 // 70% fragmentation
	// DefaultPruneThreshold is the default threshold for pruning
	DefaultPruneThreshold = 0.3 // 30% unused space
	// DefaultMaxOptimizationTime is the default maximum time for optimization
	DefaultMaxOptimizationTime = 300 * time.Second // 5 minutes
)

// UTXOOptimization handles optimization of the UTXO set
type UTXOOptimization struct {
	utxoSet *UTXOSet
	mu      sync.RWMutex

	// Optimization state
	lastOptimization time.Time
	interval         time.Duration
	compactionThresh float64
	pruneThresh      float64
	maxTime          time.Duration
}

// NewUTXOOptimization creates a new UTXO optimization handler
func NewUTXOOptimization(utxoSet *UTXOSet) *UTXOOptimization {
	return &UTXOOptimization{
		utxoSet:          utxoSet,
		interval:         DefaultOptimizationInterval * time.Second,
		compactionThresh: DefaultCompactionThreshold,
		pruneThresh:      DefaultPruneThreshold,
		maxTime:          DefaultMaxOptimizationTime,
	}
}

// Optimize optimizes the UTXO set
func (uo *UTXOOptimization) Optimize() error {
	uo.mu.Lock()
	defer uo.mu.Unlock()

	// Check if enough time has passed since last optimization
	if time.Since(uo.lastOptimization) < uo.interval {
		return nil
	}

	// Start optimization timer
	start := time.Now()

	// Get optimization stats
	stats := uo.GetOptimizationStats()

	// Check if optimization is needed
	if stats.Fragmentation < uo.compactionThresh && stats.UnusedSpace < uo.pruneThresh {
		return nil
	}

	// Perform optimization
	if err := uo.performOptimization(); err != nil {
		return fmt.Errorf("optimization failed: %v", err)
	}

	// Check if optimization took too long
	if time.Since(start) > uo.maxTime {
		return fmt.Errorf("optimization took too long: %v", time.Since(start))
	}

	// Update last optimization time
	uo.lastOptimization = time.Now()

	return nil
}

// performOptimization performs the actual optimization
func (uo *UTXOOptimization) performOptimization() error {
	// Get current UTXO set
	current := uo.utxoSet

	// Create new optimized UTXO set
	optimized := NewUTXOSet()

	// Copy UTXOs to optimized set
	for _, utxo := range current.GetAll() {
		optimized.Add(utxo)
	}

	// Replace current set with optimized set
	uo.utxoSet = optimized

	return nil
}

// GetOptimizationStats returns statistics about the optimization
func (uo *UTXOOptimization) GetOptimizationStats() *OptimizationStats {
	uo.mu.RLock()
	defer uo.mu.RUnlock()

	stats := &OptimizationStats{
		LastOptimization: uo.lastOptimization,
		Interval:         uo.interval,
		CompactionThresh: uo.compactionThresh,
		PruneThresh:      uo.pruneThresh,
		MaxTime:          uo.maxTime,
	}

	// Calculate fragmentation
	total := uo.utxoSet.Size()
	used := uo.utxoSet.Count()
	if total > 0 {
		stats.Fragmentation = 1 - float64(used)/float64(total)
	}

	// Calculate unused space
	stats.UnusedSpace = stats.Fragmentation

	return stats
}

// SetOptimizationInterval sets the optimization interval
func (uo *UTXOOptimization) SetOptimizationInterval(interval time.Duration) {
	uo.mu.Lock()
	uo.interval = interval
	uo.mu.Unlock()
}

// SetCompactionThreshold sets the compaction threshold
func (uo *UTXOOptimization) SetCompactionThreshold(threshold float64) {
	uo.mu.Lock()
	uo.compactionThresh = threshold
	uo.mu.Unlock()
}

// SetPruneThreshold sets the prune threshold
func (uo *UTXOOptimization) SetPruneThreshold(threshold float64) {
	uo.mu.Lock()
	uo.pruneThresh = threshold
	uo.mu.Unlock()
}

// SetMaxOptimizationTime sets the maximum optimization time
func (uo *UTXOOptimization) SetMaxOptimizationTime(maxTime time.Duration) {
	uo.mu.Lock()
	uo.maxTime = maxTime
	uo.mu.Unlock()
}

// OptimizationStats holds statistics about the optimization
type OptimizationStats struct {
	// LastOptimization is the time of the last optimization
	LastOptimization time.Time
	// Interval is the optimization interval
	Interval time.Duration
	// CompactionThresh is the compaction threshold
	CompactionThresh float64
	// PruneThresh is the prune threshold
	PruneThresh float64
	// MaxTime is the maximum optimization time
	MaxTime time.Duration
	// Fragmentation is the current fragmentation level
	Fragmentation float64
	// UnusedSpace is the current unused space level
	UnusedSpace float64
}

// String returns a string representation of the optimization statistics
func (os *OptimizationStats) String() string {
	return fmt.Sprintf(
		"Last: %v, Interval: %v\n"+
			"Compaction: %.2f, Prune: %.2f, Max Time: %v\n"+
			"Fragmentation: %.2f, Unused Space: %.2f",
		os.LastOptimization.Format("2006-01-02 15:04:05"),
		os.Interval, os.CompactionThresh, os.PruneThresh,
		os.MaxTime, os.Fragmentation, os.UnusedSpace,
	)
}
