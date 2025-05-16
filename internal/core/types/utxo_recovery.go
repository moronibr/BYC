package types

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

const (
	// DefaultRecoveryWindow is the default window size for recovery in seconds
	DefaultRecoveryWindow = 3600 // 1 hour
	// DefaultCheckpointInterval is the default interval for checkpoints in seconds
	DefaultCheckpointInterval = 300 // 5 minutes
	// DefaultMaxCheckpoints is the default maximum number of checkpoints to keep
	DefaultMaxCheckpoints = 12 // 1 hour of checkpoints
)

// UTXORecovery handles recovery of the UTXO set
type UTXORecovery struct {
	utxoSet *UTXOSet
	mu      sync.RWMutex

	// Recovery state
	checkpoints     []*Checkpoint
	lastCheckpoint  time.Time
	recoveryWindow  time.Duration
	checkpointCount int
}

// NewUTXORecovery creates a new UTXO recovery handler
func NewUTXORecovery(utxoSet *UTXOSet) *UTXORecovery {
	return &UTXORecovery{
		utxoSet:         utxoSet,
		checkpoints:     make([]*Checkpoint, 0),
		recoveryWindow:  DefaultRecoveryWindow * time.Second,
		checkpointCount: DefaultMaxCheckpoints,
	}
}

// Checkpoint represents a recovery checkpoint
type Checkpoint struct {
	// Timestamp is the time when the checkpoint was created
	Timestamp time.Time
	// Data is the serialized UTXO set
	Data []byte
	// Hash is the hash of the UTXO set
	Hash []byte
	// Size is the size of the UTXO set
	Size int64
}

// CreateCheckpoint creates a new recovery checkpoint
func (ur *UTXORecovery) CreateCheckpoint() error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	// Check if enough time has passed since last checkpoint
	if time.Since(ur.lastCheckpoint) < DefaultCheckpointInterval*time.Second {
		return nil
	}

	// Get serialized UTXO set
	data := ur.utxoSet.Serialize()

	// Calculate hash
	hash := sha256.Sum256(data)

	// Create checkpoint
	checkpoint := &Checkpoint{
		Timestamp: time.Now(),
		Data:      data,
		Hash:      hash[:],
		Size:      int64(len(data)),
	}

	// Add checkpoint
	ur.checkpoints = append(ur.checkpoints, checkpoint)

	// Remove old checkpoints
	if len(ur.checkpoints) > ur.checkpointCount {
		ur.checkpoints = ur.checkpoints[1:]
	}

	// Update last checkpoint time
	ur.lastCheckpoint = checkpoint.Timestamp

	return nil
}

// Recover recovers the UTXO set from a checkpoint
func (ur *UTXORecovery) Recover(timestamp time.Time) error {
	ur.mu.Lock()
	defer ur.mu.Unlock()

	// Find closest checkpoint
	var closest *Checkpoint
	var minDiff time.Duration
	for _, cp := range ur.checkpoints {
		diff := time.Since(cp.Timestamp)
		if diff < 0 {
			diff = -diff
		}
		if closest == nil || diff < minDiff {
			closest = cp
			minDiff = diff
		}
	}

	if closest == nil {
		return fmt.Errorf("no checkpoints available")
	}

	// Verify checkpoint
	hash := sha256.Sum256(closest.Data)
	if !bytes.Equal(hash[:], closest.Hash) {
		return fmt.Errorf("checkpoint verification failed")
	}

	// Deserialize UTXO set
	utxoSet, err := DeserializeUTXOSet(closest.Data)
	if err != nil {
		return fmt.Errorf("failed to deserialize UTXO set: %v", err)
	}

	// Update UTXO set
	ur.utxoSet = utxoSet

	return nil
}

// GetCheckpoints returns all available checkpoints
func (ur *UTXORecovery) GetCheckpoints() []*Checkpoint {
	ur.mu.RLock()
	defer ur.mu.RUnlock()

	// Create copy of checkpoints
	checkpoints := make([]*Checkpoint, len(ur.checkpoints))
	copy(checkpoints, ur.checkpoints)

	return checkpoints
}

// GetRecoveryStats returns statistics about the recovery system
func (ur *UTXORecovery) GetRecoveryStats() *RecoveryStats {
	ur.mu.RLock()
	defer ur.mu.RUnlock()

	stats := &RecoveryStats{
		CheckpointCount: len(ur.checkpoints),
		LastCheckpoint:  ur.lastCheckpoint,
		RecoveryWindow:  ur.recoveryWindow,
		TotalSize:       0,
		AverageSize:     0,
	}

	// Calculate size statistics
	if len(ur.checkpoints) > 0 {
		for _, cp := range ur.checkpoints {
			stats.TotalSize += cp.Size
		}
		stats.AverageSize = stats.TotalSize / int64(len(ur.checkpoints))
	}

	return stats
}

// SetRecoveryWindow sets the recovery window
func (ur *UTXORecovery) SetRecoveryWindow(window time.Duration) {
	ur.mu.Lock()
	ur.recoveryWindow = window
	ur.mu.Unlock()
}

// SetCheckpointCount sets the maximum number of checkpoints to keep
func (ur *UTXORecovery) SetCheckpointCount(count int) {
	ur.mu.Lock()
	ur.checkpointCount = count
	ur.mu.Unlock()
}

// RecoveryStats holds statistics about the recovery system
type RecoveryStats struct {
	// CheckpointCount is the number of available checkpoints
	CheckpointCount int
	// LastCheckpoint is the time of the last checkpoint
	LastCheckpoint time.Time
	// RecoveryWindow is the size of the recovery window
	RecoveryWindow time.Duration
	// TotalSize is the total size of all checkpoints in bytes
	TotalSize int64
	// AverageSize is the average size of checkpoints in bytes
	AverageSize int64
}

// String returns a string representation of the recovery statistics
func (rs *RecoveryStats) String() string {
	return fmt.Sprintf(
		"Checkpoints: %d, Last: %v, Window: %v\n"+
			"Size: %d bytes total, %d bytes average",
		rs.CheckpointCount, rs.LastCheckpoint.Format("2006-01-02 15:04:05"),
		rs.RecoveryWindow, rs.TotalSize, rs.AverageSize,
	)
}
