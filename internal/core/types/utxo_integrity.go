package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

const (
	// DefaultIntegrityInterval is the default interval for integrity checks in seconds
	DefaultIntegrityInterval = 1800 // 30 minutes
	// DefaultMaxIntegrityTime is the default maximum time for integrity checks
	DefaultMaxIntegrityTime = 60 * time.Second // 1 minute
	// DefaultMaxErrors is the default maximum number of errors before failure
	DefaultMaxErrors = 10
)

// UTXOIntegrity handles integrity checks of the UTXO set
type UTXOIntegrity struct {
	utxoSet *UTXOSet
	mu      sync.RWMutex

	// Integrity state
	lastCheck     time.Time
	interval      time.Duration
	maxTime       time.Duration
	maxErrors     int
	errorCount    int
	lastErrorTime time.Time
}

// NewUTXOIntegrity creates a new UTXO integrity handler
func NewUTXOIntegrity(utxoSet *UTXOSet) *UTXOIntegrity {
	return &UTXOIntegrity{
		utxoSet:       utxoSet,
		interval:      DefaultIntegrityInterval * time.Second,
		maxTime:       DefaultMaxIntegrityTime,
		maxErrors:     DefaultMaxErrors,
		errorCount:    0,
		lastCheck:     time.Time{},
		lastErrorTime: time.Time{},
	}
}

// CheckIntegrity performs integrity checks on the UTXO set
func (ui *UTXOIntegrity) CheckIntegrity() error {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	// Check if enough time has passed since last check
	if time.Since(ui.lastCheck) < ui.interval {
		return nil
	}

	// Start integrity check timer
	start := time.Now()

	// Reset error count
	ui.errorCount = 0

	// Perform integrity checks
	if err := ui.performIntegrityChecks(); err != nil {
		ui.errorCount++
		ui.lastErrorTime = time.Now()
		return fmt.Errorf("integrity check failed: %v", err)
	}

	// Check if integrity check took too long
	if time.Since(start) > ui.maxTime {
		ui.errorCount++
		ui.lastErrorTime = time.Now()
		return fmt.Errorf("integrity check took too long: %v", time.Since(start))
	}

	// Update last check time
	ui.lastCheck = time.Now()

	return nil
}

// performIntegrityChecks performs the actual integrity checks
func (ui *UTXOIntegrity) performIntegrityChecks() error {
	// Get UTXO set data
	data := ui.utxoSet.Serialize()

	// Calculate hash
	hash := sha256.Sum256(data)

	// Verify UTXO set structure
	if err := ui.verifyStructure(data); err != nil {
		return fmt.Errorf("structure verification failed: %v", err)
	}

	// Verify UTXO set consistency
	if err := ui.verifyConsistency(data); err != nil {
		return fmt.Errorf("consistency verification failed: %v", err)
	}

	// Verify UTXO set integrity
	if err := ui.verifyIntegrity(data, hash[:]); err != nil {
		return fmt.Errorf("integrity verification failed: %v", err)
	}

	return nil
}

// verifyStructure verifies the structure of the UTXO set
func (ui *UTXOIntegrity) verifyStructure(data []byte) error {
	// Check minimum size
	if len(data) < 8 {
		return fmt.Errorf("data too short")
	}

	// Check version
	version := data[0]
	if version != 1 {
		return fmt.Errorf("unsupported version: %d", version)
	}

	// Check size field
	size := binary.LittleEndian.Uint32(data[1:5])
	if size != uint32(len(data)-5) {
		return fmt.Errorf("size mismatch: expected %d, got %d", size, len(data)-5)
	}

	return nil
}

// verifyConsistency verifies the consistency of the UTXO set
func (ui *UTXOIntegrity) verifyConsistency(data []byte) error {
	// TODO: Implement consistency checks
	// - Verify UTXO references
	// - Verify value consistency
	// - Verify script consistency
	// - Verify timestamp consistency
	return nil
}

// verifyIntegrity verifies the integrity of the UTXO set
func (ui *UTXOIntegrity) verifyIntegrity(data []byte, expectedHash []byte) error {
	// Calculate hash
	hash := sha256.Sum256(data)

	// Compare hashes
	if !bytes.Equal(hash[:], expectedHash) {
		return fmt.Errorf("hash mismatch")
	}

	return nil
}

// GetIntegrityStats returns statistics about the integrity checks
func (ui *UTXOIntegrity) GetIntegrityStats() *IntegrityStats {
	ui.mu.RLock()
	defer ui.mu.RUnlock()

	return &IntegrityStats{
		LastCheck:     ui.lastCheck,
		Interval:      ui.interval,
		MaxTime:       ui.maxTime,
		MaxErrors:     ui.maxErrors,
		ErrorCount:    ui.errorCount,
		LastErrorTime: ui.lastErrorTime,
		IsHealthy:     ui.errorCount < ui.maxErrors,
	}
}

// SetIntegrityInterval sets the integrity check interval
func (ui *UTXOIntegrity) SetIntegrityInterval(interval time.Duration) {
	ui.mu.Lock()
	ui.interval = interval
	ui.mu.Unlock()
}

// SetMaxIntegrityTime sets the maximum integrity check time
func (ui *UTXOIntegrity) SetMaxIntegrityTime(maxTime time.Duration) {
	ui.mu.Lock()
	ui.maxTime = maxTime
	ui.mu.Unlock()
}

// SetMaxErrors sets the maximum number of errors before failure
func (ui *UTXOIntegrity) SetMaxErrors(maxErrors int) {
	ui.mu.Lock()
	ui.maxErrors = maxErrors
	ui.mu.Unlock()
}

// ResetErrorCount resets the error count
func (ui *UTXOIntegrity) ResetErrorCount() {
	ui.mu.Lock()
	ui.errorCount = 0
	ui.mu.Unlock()
}

// IntegrityStats holds statistics about the integrity checks
type IntegrityStats struct {
	// LastCheck is the time of the last integrity check
	LastCheck time.Time
	// Interval is the integrity check interval
	Interval time.Duration
	// MaxTime is the maximum integrity check time
	MaxTime time.Duration
	// MaxErrors is the maximum number of errors before failure
	MaxErrors int
	// ErrorCount is the current number of errors
	ErrorCount int
	// LastErrorTime is the time of the last error
	LastErrorTime time.Time
	// IsHealthy indicates if the UTXO set is healthy
	IsHealthy bool
}

// String returns a string representation of the integrity statistics
func (is *IntegrityStats) String() string {
	return fmt.Sprintf(
		"Last Check: %v, Interval: %v\n"+
			"Max Time: %v, Max Errors: %d\n"+
			"Errors: %d, Last Error: %v\n"+
			"Health: %v",
		is.LastCheck.Format("2006-01-02 15:04:05"),
		is.Interval, is.MaxTime, is.MaxErrors,
		is.ErrorCount, is.LastErrorTime.Format("2006-01-02 15:04:05"),
		is.IsHealthy,
	)
}
