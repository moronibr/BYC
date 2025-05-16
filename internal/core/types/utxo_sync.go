package types

import (
	"fmt"
	"sync"
	"time"
)

const (
	// DefaultSyncInterval is the default interval for synchronization
	DefaultSyncInterval = 100 * time.Millisecond
	// DefaultSyncTimeout is the default timeout for synchronization
	DefaultSyncTimeout = 5 * time.Second
	// DefaultSyncRetries is the default number of retries for synchronization
	DefaultSyncRetries = 3
)

// SyncType represents the type of synchronization
type SyncType byte

const (
	// SyncTypeNone indicates no synchronization
	SyncTypeNone SyncType = iota
	// SyncTypeLock indicates lock-based synchronization
	SyncTypeLock
	// SyncTypeChannel indicates channel-based synchronization
	SyncTypeChannel
)

// SyncState represents the state of synchronization
type SyncState struct {
	// Type is the type of synchronization
	Type SyncType
	// Locked indicates whether the UTXO set is locked
	Locked bool
	// LastSync is the time of the last synchronization
	LastSync time.Time
	// Retries is the number of retries
	Retries int
	// Error is the last error that occurred
	Error error
}

// UTXOSync handles synchronization of the UTXO set
type UTXOSync struct {
	utxoSet *UTXOSet
	mu      sync.RWMutex

	// Synchronization state
	syncType   SyncType
	interval   time.Duration
	timeout    time.Duration
	retries    int
	state      *SyncState
	stopChan   chan struct{}
	doneChan   chan struct{}
	errorChan  chan error
	updateChan chan struct{}
	lockChan   chan struct{}
	unlockChan chan struct{}
}

// NewUTXOSync creates a new UTXO synchronization handler
func NewUTXOSync(utxoSet *UTXOSet) *UTXOSync {
	return &UTXOSync{
		utxoSet:    utxoSet,
		syncType:   SyncTypeNone,
		interval:   DefaultSyncInterval,
		timeout:    DefaultSyncTimeout,
		retries:    DefaultSyncRetries,
		state:      &SyncState{Type: SyncTypeNone},
		stopChan:   make(chan struct{}),
		doneChan:   make(chan struct{}),
		errorChan:  make(chan error, 1),
		updateChan: make(chan struct{}, 1),
		lockChan:   make(chan struct{}, 1),
		unlockChan: make(chan struct{}, 1),
	}
}

// Start starts the synchronization
func (us *UTXOSync) Start() error {
	us.mu.Lock()
	defer us.mu.Unlock()

	// Check if synchronization is enabled
	if us.syncType == SyncTypeNone {
		return fmt.Errorf("synchronization is not enabled")
	}

	// Start synchronization based on type
	switch us.syncType {
	case SyncTypeLock:
		// Start lock-based synchronization
		go us.lockSync()
	case SyncTypeChannel:
		// Start channel-based synchronization
		go us.channelSync()
	default:
		return fmt.Errorf("unsupported synchronization type: %d", us.syncType)
	}

	return nil
}

// Stop stops the synchronization
func (us *UTXOSync) Stop() error {
	us.mu.Lock()
	defer us.mu.Unlock()

	// Check if synchronization is enabled
	if us.syncType == SyncTypeNone {
		return fmt.Errorf("synchronization is not enabled")
	}

	// Stop synchronization
	close(us.stopChan)

	// Wait for synchronization to stop
	<-us.doneChan

	return nil
}

// Lock locks the UTXO set
func (us *UTXOSync) Lock() error {
	us.mu.Lock()
	defer us.mu.Unlock()

	// Check if synchronization is enabled
	if us.syncType == SyncTypeNone {
		return fmt.Errorf("synchronization is not enabled")
	}

	// Lock UTXO set
	us.lockChan <- struct{}{}

	return nil
}

// Unlock unlocks the UTXO set
func (us *UTXOSync) Unlock() error {
	us.mu.Lock()
	defer us.mu.Unlock()

	// Check if synchronization is enabled
	if us.syncType == SyncTypeNone {
		return fmt.Errorf("synchronization is not enabled")
	}

	// Unlock UTXO set
	us.unlockChan <- struct{}{}

	return nil
}

// Update updates the UTXO set
func (us *UTXOSync) Update() error {
	us.mu.Lock()
	defer us.mu.Unlock()

	// Check if synchronization is enabled
	if us.syncType == SyncTypeNone {
		return fmt.Errorf("synchronization is not enabled")
	}

	// Update UTXO set
	us.updateChan <- struct{}{}

	return nil
}

// GetSyncStats returns statistics about the synchronization
func (us *UTXOSync) GetSyncStats() *SyncStats {
	us.mu.RLock()
	defer us.mu.RUnlock()

	stats := &SyncStats{
		SyncType:  us.syncType,
		Interval:  us.interval,
		Timeout:   us.timeout,
		Retries:   us.retries,
		State:     us.state,
		ErrorRate: 0,
	}

	// Calculate error rate
	if us.state.Retries > 0 {
		stats.ErrorRate = float64(us.state.Retries) / float64(us.retries)
	}

	return stats
}

// SetSyncType sets the type of synchronization
func (us *UTXOSync) SetSyncType(syncType SyncType) {
	us.mu.Lock()
	us.syncType = syncType
	us.mu.Unlock()
}

// SetInterval sets the interval for synchronization
func (us *UTXOSync) SetInterval(interval time.Duration) {
	us.mu.Lock()
	us.interval = interval
	us.mu.Unlock()
}

// SetTimeout sets the timeout for synchronization
func (us *UTXOSync) SetTimeout(timeout time.Duration) {
	us.mu.Lock()
	us.timeout = timeout
	us.mu.Unlock()
}

// SetRetries sets the number of retries for synchronization
func (us *UTXOSync) SetRetries(retries int) {
	us.mu.Lock()
	us.retries = retries
	us.mu.Unlock()
}

// lockSync handles lock-based synchronization
func (us *UTXOSync) lockSync() {
	ticker := time.NewTicker(us.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Lock UTXO set
			us.mu.Lock()
			us.state.Locked = true
			us.state.LastSync = time.Now()
			us.mu.Unlock()

			// Wait for update
			select {
			case <-us.updateChan:
				// Update UTXO set
				us.mu.Lock()
				us.state.LastSync = time.Now()
				us.mu.Unlock()
			case <-time.After(us.timeout):
				// Timeout
				us.mu.Lock()
				us.state.Error = fmt.Errorf("synchronization timeout")
				us.state.Retries++
				us.mu.Unlock()
			}

			// Unlock UTXO set
			us.mu.Lock()
			us.state.Locked = false
			us.mu.Unlock()

		case <-us.stopChan:
			// Stop synchronization
			close(us.doneChan)
			return
		}
	}
}

// channelSync handles channel-based synchronization
func (us *UTXOSync) channelSync() {
	for {
		select {
		case <-us.lockChan:
			// Lock UTXO set
			us.mu.Lock()
			us.state.Locked = true
			us.state.LastSync = time.Now()
			us.mu.Unlock()

		case <-us.unlockChan:
			// Unlock UTXO set
			us.mu.Lock()
			us.state.Locked = false
			us.mu.Unlock()

		case <-us.updateChan:
			// Update UTXO set
			us.mu.Lock()
			us.state.LastSync = time.Now()
			us.mu.Unlock()

		case <-us.stopChan:
			// Stop synchronization
			close(us.doneChan)
			return
		}
	}
}

// SyncStats holds statistics about the synchronization
type SyncStats struct {
	// SyncType is the type of synchronization
	SyncType SyncType
	// Interval is the interval for synchronization
	Interval time.Duration
	// Timeout is the timeout for synchronization
	Timeout time.Duration
	// Retries is the number of retries for synchronization
	Retries int
	// State is the state of synchronization
	State *SyncState
	// ErrorRate is the rate of synchronization errors
	ErrorRate float64
}

// String returns a string representation of the synchronization statistics
func (ss *SyncStats) String() string {
	return fmt.Sprintf(
		"Synchronization Type: %d\n"+
			"Interval: %v, Timeout: %v\n"+
			"Retries: %d, Error Rate: %.2f\n"+
			"Locked: %v, Last Sync: %v\n"+
			"Error: %v",
		ss.SyncType,
		ss.Interval, ss.Timeout,
		ss.Retries, ss.ErrorRate,
		ss.State.Locked, ss.State.LastSync,
		ss.State.Error,
	)
}
