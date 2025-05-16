package types

import (
	"fmt"
	"sync"
	"time"
)

const (
	// DefaultMonitorInterval is the default interval for monitoring
	DefaultMonitorInterval = 1 * time.Second
	// DefaultMonitorThreshold is the default threshold for monitoring
	DefaultMonitorThreshold = 1000
	// DefaultMonitorWindow is the default window for monitoring
	DefaultMonitorWindow = 60 * time.Second
)

// MonitorType represents the type of monitoring
type MonitorType byte

const (
	// MonitorTypeNone indicates no monitoring
	MonitorTypeNone MonitorType = iota
	// MonitorTypePerformance indicates performance monitoring
	MonitorTypePerformance
	// MonitorTypeHealth indicates health monitoring
	MonitorTypeHealth
	// MonitorTypeMetrics indicates metrics monitoring
	MonitorTypeMetrics
)

// MonitorState represents the state of monitoring
type MonitorState struct {
	// Type is the type of monitoring
	Type MonitorType
	// Active indicates whether monitoring is active
	Active bool
	// LastCheck is the time of the last check
	LastCheck time.Time
	// Errors is the number of errors
	Errors int
	// Warnings is the number of warnings
	Warnings int
}

// MonitorMetrics represents metrics for monitoring
type MonitorMetrics struct {
	// Size is the size of the UTXO set
	Size int64
	// Operations is the number of operations
	Operations int64
	// Errors is the number of errors
	Errors int64
	// Warnings is the number of warnings
	Warnings int64
	// Latency is the average latency
	Latency time.Duration
	// Throughput is the average throughput
	Throughput float64
}

// UTXOMonitor handles monitoring of the UTXO set
type UTXOMonitor struct {
	utxoSet *UTXOSet
	mu      sync.RWMutex

	// Monitoring state
	monitorType MonitorType
	interval    time.Duration
	threshold   int64
	window      time.Duration
	state       *MonitorState
	metrics     *MonitorMetrics
	stopChan    chan struct{}
	doneChan    chan struct{}
	errorChan   chan error
	metricsChan chan *MonitorMetrics
	checkChan   chan struct{}
	alertChan   chan string
}

// NewUTXOMonitor creates a new UTXO monitoring handler
func NewUTXOMonitor(utxoSet *UTXOSet) *UTXOMonitor {
	return &UTXOMonitor{
		utxoSet:     utxoSet,
		monitorType: MonitorTypeNone,
		interval:    DefaultMonitorInterval,
		threshold:   DefaultMonitorThreshold,
		window:      DefaultMonitorWindow,
		state:       &MonitorState{Type: MonitorTypeNone},
		metrics:     &MonitorMetrics{},
		stopChan:    make(chan struct{}),
		doneChan:    make(chan struct{}),
		errorChan:   make(chan error, 1),
		metricsChan: make(chan *MonitorMetrics, 1),
		checkChan:   make(chan struct{}, 1),
		alertChan:   make(chan string, 1),
	}
}

// Start starts the monitoring
func (um *UTXOMonitor) Start() error {
	um.mu.Lock()
	defer um.mu.Unlock()

	// Check if monitoring is enabled
	if um.monitorType == MonitorTypeNone {
		return fmt.Errorf("monitoring is not enabled")
	}

	// Start monitoring based on type
	switch um.monitorType {
	case MonitorTypePerformance:
		// Start performance monitoring
		go um.performanceMonitor()
	case MonitorTypeHealth:
		// Start health monitoring
		go um.healthMonitor()
	case MonitorTypeMetrics:
		// Start metrics monitoring
		go um.metricsMonitor()
	default:
		return fmt.Errorf("unsupported monitoring type: %d", um.monitorType)
	}

	return nil
}

// Stop stops the monitoring
func (um *UTXOMonitor) Stop() error {
	um.mu.Lock()
	defer um.mu.Unlock()

	// Check if monitoring is enabled
	if um.monitorType == MonitorTypeNone {
		return fmt.Errorf("monitoring is not enabled")
	}

	// Stop monitoring
	close(um.stopChan)

	// Wait for monitoring to stop
	<-um.doneChan

	return nil
}

// Check performs a monitoring check
func (um *UTXOMonitor) Check() error {
	um.mu.Lock()
	defer um.mu.Unlock()

	// Check if monitoring is enabled
	if um.monitorType == MonitorTypeNone {
		return fmt.Errorf("monitoring is not enabled")
	}

	// Perform check
	um.checkChan <- struct{}{}

	return nil
}

// GetMonitorStats returns statistics about the monitoring
func (um *UTXOMonitor) GetMonitorStats() *MonitorStats {
	um.mu.RLock()
	defer um.mu.RUnlock()

	stats := &MonitorStats{
		MonitorType: um.monitorType,
		Interval:    um.interval,
		Threshold:   um.threshold,
		Window:      um.window,
		State:       um.state,
		Metrics:     um.metrics,
	}

	return stats
}

// SetMonitorType sets the type of monitoring
func (um *UTXOMonitor) SetMonitorType(monitorType MonitorType) {
	um.mu.Lock()
	um.monitorType = monitorType
	um.mu.Unlock()
}

// SetInterval sets the interval for monitoring
func (um *UTXOMonitor) SetInterval(interval time.Duration) {
	um.mu.Lock()
	um.interval = interval
	um.mu.Unlock()
}

// SetThreshold sets the threshold for monitoring
func (um *UTXOMonitor) SetThreshold(threshold int64) {
	um.mu.Lock()
	um.threshold = threshold
	um.mu.Unlock()
}

// SetWindow sets the window for monitoring
func (um *UTXOMonitor) SetWindow(window time.Duration) {
	um.mu.Lock()
	um.window = window
	um.mu.Unlock()
}

// performanceMonitor handles performance monitoring
func (um *UTXOMonitor) performanceMonitor() {
	ticker := time.NewTicker(um.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check performance
			um.mu.Lock()
			um.state.LastCheck = time.Now()
			um.mu.Unlock()

			// Update metrics
			metrics := &MonitorMetrics{
				Size:       um.utxoSet.Size(),
				Operations: um.metrics.Operations,
				Errors:     um.metrics.Errors,
				Warnings:   um.metrics.Warnings,
				Latency:    um.metrics.Latency,
				Throughput: float64(um.metrics.Operations) / um.window.Seconds(),
			}

			// Send metrics
			um.metricsChan <- metrics

		case <-um.checkChan:
			// Check performance
			um.mu.Lock()
			um.state.LastCheck = time.Now()
			um.mu.Unlock()

			// Update metrics
			metrics := &MonitorMetrics{
				Size:       um.utxoSet.Size(),
				Operations: um.metrics.Operations,
				Errors:     um.metrics.Errors,
				Warnings:   um.metrics.Warnings,
				Latency:    um.metrics.Latency,
				Throughput: float64(um.metrics.Operations) / um.window.Seconds(),
			}

			// Send metrics
			um.metricsChan <- metrics

		case <-um.stopChan:
			// Stop monitoring
			close(um.doneChan)
			return
		}
	}
}

// healthMonitor handles health monitoring
func (um *UTXOMonitor) healthMonitor() {
	ticker := time.NewTicker(um.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check health
			um.mu.Lock()
			um.state.LastCheck = time.Now()
			um.mu.Unlock()

			// Check size
			if um.utxoSet.Size() > um.threshold {
				um.alertChan <- fmt.Sprintf("UTXO set size exceeds threshold: %d > %d", um.utxoSet.Size(), um.threshold)
			}

			// Check errors
			if um.metrics.Errors > 0 {
				um.alertChan <- fmt.Sprintf("UTXO set has errors: %d", um.metrics.Errors)
			}

			// Check warnings
			if um.metrics.Warnings > 0 {
				um.alertChan <- fmt.Sprintf("UTXO set has warnings: %d", um.metrics.Warnings)
			}

		case <-um.checkChan:
			// Check health
			um.mu.Lock()
			um.state.LastCheck = time.Now()
			um.mu.Unlock()

			// Check size
			if um.utxoSet.Size() > um.threshold {
				um.alertChan <- fmt.Sprintf("UTXO set size exceeds threshold: %d > %d", um.utxoSet.Size(), um.threshold)
			}

			// Check errors
			if um.metrics.Errors > 0 {
				um.alertChan <- fmt.Sprintf("UTXO set has errors: %d", um.metrics.Errors)
			}

			// Check warnings
			if um.metrics.Warnings > 0 {
				um.alertChan <- fmt.Sprintf("UTXO set has warnings: %d", um.metrics.Warnings)
			}

		case <-um.stopChan:
			// Stop monitoring
			close(um.doneChan)
			return
		}
	}
}

// metricsMonitor handles metrics monitoring
func (um *UTXOMonitor) metricsMonitor() {
	ticker := time.NewTicker(um.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check metrics
			um.mu.Lock()
			um.state.LastCheck = time.Now()
			um.mu.Unlock()

			// Update metrics
			metrics := &MonitorMetrics{
				Size:       um.utxoSet.Size(),
				Operations: um.metrics.Operations,
				Errors:     um.metrics.Errors,
				Warnings:   um.metrics.Warnings,
				Latency:    um.metrics.Latency,
				Throughput: float64(um.metrics.Operations) / um.window.Seconds(),
			}

			// Send metrics
			um.metricsChan <- metrics

		case <-um.checkChan:
			// Check metrics
			um.mu.Lock()
			um.state.LastCheck = time.Now()
			um.mu.Unlock()

			// Update metrics
			metrics := &MonitorMetrics{
				Size:       um.utxoSet.Size(),
				Operations: um.metrics.Operations,
				Errors:     um.metrics.Errors,
				Warnings:   um.metrics.Warnings,
				Latency:    um.metrics.Latency,
				Throughput: float64(um.metrics.Operations) / um.window.Seconds(),
			}

			// Send metrics
			um.metricsChan <- metrics

		case <-um.stopChan:
			// Stop monitoring
			close(um.doneChan)
			return
		}
	}
}

// MonitorStats holds statistics about the monitoring
type MonitorStats struct {
	// MonitorType is the type of monitoring
	MonitorType MonitorType
	// Interval is the interval for monitoring
	Interval time.Duration
	// Threshold is the threshold for monitoring
	Threshold int64
	// Window is the window for monitoring
	Window time.Duration
	// State is the state of monitoring
	State *MonitorState
	// Metrics is the metrics for monitoring
	Metrics *MonitorMetrics
}

// String returns a string representation of the monitoring statistics
func (ms *MonitorStats) String() string {
	return fmt.Sprintf(
		"Monitoring Type: %d\n"+
			"Interval: %v, Threshold: %d, Window: %v\n"+
			"Active: %v, Last Check: %v\n"+
			"Errors: %d, Warnings: %d\n"+
			"Size: %d, Operations: %d\n"+
			"Latency: %v, Throughput: %.2f",
		ms.MonitorType,
		ms.Interval, ms.Threshold, ms.Window,
		ms.State.Active, ms.State.LastCheck,
		ms.State.Errors, ms.State.Warnings,
		ms.Metrics.Size, ms.Metrics.Operations,
		ms.Metrics.Latency, ms.Metrics.Throughput,
	)
}
