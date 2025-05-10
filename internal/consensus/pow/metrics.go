// Package pow implements a robust Proof of Work consensus mechanism with adaptive mining capabilities.
package pow

import (
	"sync/atomic"
	"time"
)

// Metrics tracks various performance and operational metrics for the Proof of Work system.
type Metrics struct {
	// Mining metrics
	blocksMined        uint64
	hashRate           uint64
	avgMiningTime      float64
	totalMiningTime    time.Duration
	miningAttempts     uint64
	miningFailures     uint64
	lastMiningDuration time.Duration

	// Resource metrics
	cpuUtilization     float64
	memoryUtilization  float64
	workerCount        int32
	optimalWorkerCount int32

	// Circuit breaker metrics
	circuitBreakerTrips    uint64
	circuitBreakerResets   uint64
	circuitBreakerState    CircuitBreakerState
	lastCircuitBreakerTrip time.Time

	// Difficulty metrics
	currentDifficulty     int
	difficultyAdjustments uint64
	lastDifficultyChange  time.Time

	// Performance metrics
	hashOperations    uint64
	nonceIterations   uint64
	validationChecks  uint64
	validationSuccess uint64
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		circuitBreakerState: CircuitBreakerClosed,
	}
}

// RecordBlockMined records a successfully mined block.
func (m *Metrics) RecordBlockMined(duration time.Duration) {
	atomic.AddUint64(&m.blocksMined, 1)
	m.lastMiningDuration = duration
	m.totalMiningTime += duration
	m.avgMiningTime = float64(m.totalMiningTime) / float64(m.blocksMined)
}

// RecordHashRate records the current hash rate.
func (m *Metrics) RecordHashRate(rate uint64) {
	atomic.StoreUint64(&m.hashRate, rate)
}

// RecordMiningAttempt records a mining attempt.
func (m *Metrics) RecordMiningAttempt(success bool) {
	atomic.AddUint64(&m.miningAttempts, 1)
	if !success {
		atomic.AddUint64(&m.miningFailures, 1)
	}
}

// RecordResourceUsage records current resource utilization.
func (m *Metrics) RecordResourceUsage(cpu, memory float64, workers, optimalWorkers int) {
	m.cpuUtilization = cpu
	m.memoryUtilization = memory
	atomic.StoreInt32(&m.workerCount, int32(workers))
	atomic.StoreInt32(&m.optimalWorkerCount, int32(optimalWorkers))
}

// RecordCircuitBreakerState records a change in circuit breaker state.
func (m *Metrics) RecordCircuitBreakerState(state CircuitBreakerState) {
	if state == CircuitBreakerOpen && m.circuitBreakerState != CircuitBreakerOpen {
		atomic.AddUint64(&m.circuitBreakerTrips, 1)
		m.lastCircuitBreakerTrip = time.Now()
	} else if state == CircuitBreakerClosed && m.circuitBreakerState != CircuitBreakerClosed {
		atomic.AddUint64(&m.circuitBreakerResets, 1)
	}
	m.circuitBreakerState = state
}

// RecordDifficultyChange records a change in mining difficulty.
func (m *Metrics) RecordDifficultyChange(newDifficulty int) {
	m.currentDifficulty = newDifficulty
	atomic.AddUint64(&m.difficultyAdjustments, 1)
	m.lastDifficultyChange = time.Now()
}

// RecordHashOperation records a hash operation.
func (m *Metrics) RecordHashOperation() {
	atomic.AddUint64(&m.hashOperations, 1)
}

// RecordNonceIteration records a nonce iteration.
func (m *Metrics) RecordNonceIteration() {
	atomic.AddUint64(&m.nonceIterations, 1)
}

// RecordValidation records a block validation attempt.
func (m *Metrics) RecordValidation(success bool) {
	atomic.AddUint64(&m.validationChecks, 1)
	if success {
		atomic.AddUint64(&m.validationSuccess, 1)
	}
}

// GetMetrics returns a snapshot of all current metrics.
func (m *Metrics) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"blocks_mined":           atomic.LoadUint64(&m.blocksMined),
		"hash_rate":              atomic.LoadUint64(&m.hashRate),
		"avg_mining_time":        m.avgMiningTime,
		"last_mining_duration":   m.lastMiningDuration,
		"mining_attempts":        atomic.LoadUint64(&m.miningAttempts),
		"mining_failures":        atomic.LoadUint64(&m.miningFailures),
		"cpu_utilization":        m.cpuUtilization,
		"memory_utilization":     m.memoryUtilization,
		"worker_count":           atomic.LoadInt32(&m.workerCount),
		"optimal_worker_count":   atomic.LoadInt32(&m.optimalWorkerCount),
		"circuit_breaker_trips":  atomic.LoadUint64(&m.circuitBreakerTrips),
		"circuit_breaker_resets": atomic.LoadUint64(&m.circuitBreakerResets),
		"circuit_breaker_state":  m.circuitBreakerState,
		"current_difficulty":     m.currentDifficulty,
		"difficulty_adjustments": atomic.LoadUint64(&m.difficultyAdjustments),
		"hash_operations":        atomic.LoadUint64(&m.hashOperations),
		"nonce_iterations":       atomic.LoadUint64(&m.nonceIterations),
		"validation_checks":      atomic.LoadUint64(&m.validationChecks),
		"validation_success":     atomic.LoadUint64(&m.validationSuccess),
	}
}

// Reset resets all metrics to their initial values.
func (m *Metrics) Reset() {
	atomic.StoreUint64(&m.blocksMined, 0)
	atomic.StoreUint64(&m.hashRate, 0)
	m.avgMiningTime = 0
	m.totalMiningTime = 0
	atomic.StoreUint64(&m.miningAttempts, 0)
	atomic.StoreUint64(&m.miningFailures, 0)
	m.lastMiningDuration = 0
	m.cpuUtilization = 0
	m.memoryUtilization = 0
	atomic.StoreInt32(&m.workerCount, 0)
	atomic.StoreInt32(&m.optimalWorkerCount, 0)
	atomic.StoreUint64(&m.circuitBreakerTrips, 0)
	atomic.StoreUint64(&m.circuitBreakerResets, 0)
	m.circuitBreakerState = CircuitBreakerClosed
	m.currentDifficulty = TargetBits
	atomic.StoreUint64(&m.difficultyAdjustments, 0)
	atomic.StoreUint64(&m.hashOperations, 0)
	atomic.StoreUint64(&m.nonceIterations, 0)
	atomic.StoreUint64(&m.validationChecks, 0)
	atomic.StoreUint64(&m.validationSuccess, 0)
}
