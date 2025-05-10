// Package pow implements a robust Proof of Work consensus mechanism with adaptive mining capabilities.
package pow

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Mining metrics
	blocksMined = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pow_blocks_mined_total",
		Help: "Total number of blocks successfully mined",
	})

	hashRate = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pow_hash_rate",
		Help: "Current hash rate in hashes per second",
	})

	miningDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "pow_mining_duration_seconds",
		Help:    "Time taken to mine blocks in seconds",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10),
	})

	miningAttempts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pow_mining_attempts_total",
		Help: "Total number of mining attempts",
	})

	miningFailures = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pow_mining_failures_total",
		Help: "Total number of failed mining attempts",
	})

	// Resource metrics
	cpuUtilization = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pow_cpu_utilization_percent",
		Help: "Current CPU utilization percentage",
	})

	memoryUtilization = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pow_memory_utilization_percent",
		Help: "Current memory utilization percentage",
	})

	workerCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pow_worker_count",
		Help: "Current number of mining workers",
	})

	optimalWorkerCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pow_optimal_worker_count",
		Help: "Optimal number of mining workers based on system resources",
	})

	// Circuit breaker metrics
	circuitBreakerTrips = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pow_circuit_breaker_trips_total",
		Help: "Total number of circuit breaker trips",
	})

	circuitBreakerResets = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pow_circuit_breaker_resets_total",
		Help: "Total number of circuit breaker resets",
	})

	circuitBreakerState = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pow_circuit_breaker_state",
		Help: "Current state of the circuit breaker (0=Closed, 1=Open, 2=Half-Open)",
	})

	// Difficulty metrics
	currentDifficulty = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pow_current_difficulty",
		Help: "Current mining difficulty",
	})

	difficultyAdjustments = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pow_difficulty_adjustments_total",
		Help: "Total number of difficulty adjustments",
	})

	// Performance metrics
	hashOperations = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pow_hash_operations_total",
		Help: "Total number of hash operations performed",
	})

	nonceIterations = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pow_nonce_iterations_total",
		Help: "Total number of nonce iterations",
	})

	validationChecks = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pow_validation_checks_total",
		Help: "Total number of block validation checks",
	})

	validationSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pow_validation_success_total",
		Help: "Total number of successful block validations",
	})

	// Additional performance metrics
	blockTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "pow_block_time_seconds",
		Help:    "Time between blocks in seconds",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10),
	})

	workerEfficiency = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "pow_worker_efficiency",
		Help: "Efficiency of each worker (hashes per second per CPU)",
	}, []string{"worker_id"})

	resourceContention = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pow_resource_contention",
		Help: "Resource contention level (0-1)",
	})
)

// updatePrometheusMetrics updates all Prometheus metrics with current values
func (m *Metrics) updatePrometheusMetrics() {
	// Mining metrics
	blocksMined.Add(float64(atomic.LoadUint64(&m.blocksMined)))
	hashRate.Set(float64(atomic.LoadUint64(&m.hashRate)))
	miningDuration.Observe(m.avgMiningTime)
	miningAttempts.Add(float64(atomic.LoadUint64(&m.miningAttempts)))
	miningFailures.Add(float64(atomic.LoadUint64(&m.miningFailures)))

	// Resource metrics
	cpuUtilization.Set(m.cpuUtilization)
	memoryUtilization.Set(m.memoryUtilization)
	workerCount.Set(float64(atomic.LoadInt32(&m.workerCount)))
	optimalWorkerCount.Set(float64(atomic.LoadInt32(&m.optimalWorkerCount)))

	// Circuit breaker metrics
	circuitBreakerTrips.Add(float64(atomic.LoadUint64(&m.circuitBreakerTrips)))
	circuitBreakerResets.Add(float64(atomic.LoadUint64(&m.circuitBreakerResets)))
	circuitBreakerState.Set(float64(m.circuitBreakerState))

	// Difficulty metrics
	currentDifficulty.Set(float64(m.currentDifficulty))
	difficultyAdjustments.Add(float64(atomic.LoadUint64(&m.difficultyAdjustments)))

	// Performance metrics
	hashOperations.Add(float64(atomic.LoadUint64(&m.hashOperations)))
	nonceIterations.Add(float64(atomic.LoadUint64(&m.nonceIterations)))
	validationChecks.Add(float64(atomic.LoadUint64(&m.validationChecks)))
	validationSuccess.Add(float64(atomic.LoadUint64(&m.validationSuccess)))
}

// RecordBlockTime records the time between blocks
func (m *Metrics) RecordBlockTime(duration time.Duration) {
	blockTime.Observe(duration.Seconds())
}

// RecordWorkerEfficiency records the efficiency of a specific worker
func (m *Metrics) RecordWorkerEfficiency(workerID string, efficiency float64) {
	workerEfficiency.WithLabelValues(workerID).Set(efficiency)
}

// RecordResourceContention records the current resource contention level
func (m *Metrics) RecordResourceContention(level float64) {
	resourceContention.Set(level)
}
