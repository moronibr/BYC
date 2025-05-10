// Package pow implements a robust Proof of Work consensus mechanism with adaptive mining capabilities.
package pow

import (
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	metrics := NewMetrics()

	// Test mining metrics
	metrics.RecordBlockMined(time.Second * 2)
	metrics.RecordHashRate(1000)
	metrics.RecordMiningAttempt(true)
	metrics.RecordMiningAttempt(false)

	// Test resource metrics
	metrics.RecordResourceUsage(75.5, 60.0, 4, 4)

	// Test circuit breaker metrics
	metrics.RecordCircuitBreakerState(CircuitBreakerOpen)
	metrics.RecordCircuitBreakerState(CircuitBreakerClosed)

	// Test difficulty metrics
	metrics.RecordDifficultyChange(25)

	// Test performance metrics
	metrics.RecordHashOperation()
	metrics.RecordNonceIteration()
	metrics.RecordValidation(true)
	metrics.RecordValidation(false)

	// Get all metrics
	allMetrics := metrics.GetMetrics()

	// Verify mining metrics
	if allMetrics["blocks_mined"].(uint64) != 1 {
		t.Errorf("Expected 1 block mined, got %d", allMetrics["blocks_mined"])
	}
	if allMetrics["hash_rate"].(uint64) != 1000 {
		t.Errorf("Expected hash rate 1000, got %d", allMetrics["hash_rate"])
	}
	if allMetrics["mining_attempts"].(uint64) != 2 {
		t.Errorf("Expected 2 mining attempts, got %d", allMetrics["mining_attempts"])
	}
	if allMetrics["mining_failures"].(uint64) != 1 {
		t.Errorf("Expected 1 mining failure, got %d", allMetrics["mining_failures"])
	}

	// Verify resource metrics
	if allMetrics["cpu_utilization"].(float64) != 75.5 {
		t.Errorf("Expected CPU utilization 75.5, got %f", allMetrics["cpu_utilization"])
	}
	if allMetrics["memory_utilization"].(float64) != 60.0 {
		t.Errorf("Expected memory utilization 60.0, got %f", allMetrics["memory_utilization"])
	}
	if allMetrics["worker_count"].(int32) != 4 {
		t.Errorf("Expected 4 workers, got %d", allMetrics["worker_count"])
	}

	// Verify circuit breaker metrics
	if allMetrics["circuit_breaker_trips"].(uint64) != 1 {
		t.Errorf("Expected 1 circuit breaker trip, got %d", allMetrics["circuit_breaker_trips"])
	}
	if allMetrics["circuit_breaker_resets"].(uint64) != 1 {
		t.Errorf("Expected 1 circuit breaker reset, got %d", allMetrics["circuit_breaker_resets"])
	}

	// Verify difficulty metrics
	if allMetrics["current_difficulty"].(int) != 25 {
		t.Errorf("Expected difficulty 25, got %d", allMetrics["current_difficulty"])
	}
	if allMetrics["difficulty_adjustments"].(uint64) != 1 {
		t.Errorf("Expected 1 difficulty adjustment, got %d", allMetrics["difficulty_adjustments"])
	}

	// Verify performance metrics
	if allMetrics["hash_operations"].(uint64) != 1 {
		t.Errorf("Expected 1 hash operation, got %d", allMetrics["hash_operations"])
	}
	if allMetrics["nonce_iterations"].(uint64) != 1 {
		t.Errorf("Expected 1 nonce iteration, got %d", allMetrics["nonce_iterations"])
	}
	if allMetrics["validation_checks"].(uint64) != 2 {
		t.Errorf("Expected 2 validation checks, got %d", allMetrics["validation_checks"])
	}
	if allMetrics["validation_success"].(uint64) != 1 {
		t.Errorf("Expected 1 validation success, got %d", allMetrics["validation_success"])
	}

	// Test reset
	metrics.Reset()
	allMetrics = metrics.GetMetrics()

	// Verify all metrics are reset
	if allMetrics["blocks_mined"].(uint64) != 0 {
		t.Errorf("Expected 0 blocks mined after reset, got %d", allMetrics["blocks_mined"])
	}
	if allMetrics["hash_rate"].(uint64) != 0 {
		t.Errorf("Expected hash rate 0 after reset, got %d", allMetrics["hash_rate"])
	}
	if allMetrics["mining_attempts"].(uint64) != 0 {
		t.Errorf("Expected 0 mining attempts after reset, got %d", allMetrics["mining_attempts"])
	}
}

func TestMetricsConcurrent(t *testing.T) {
	metrics := NewMetrics()
	done := make(chan bool)

	// Start multiple goroutines to update metrics concurrently
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				metrics.RecordHashOperation()
				metrics.RecordNonceIteration()
				metrics.RecordValidation(true)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final counts
	allMetrics := metrics.GetMetrics()
	expectedHashOps := uint64(1000) // 10 goroutines * 100 operations
	expectedNonceIterations := uint64(1000)
	expectedValidations := uint64(1000)

	if allMetrics["hash_operations"].(uint64) != expectedHashOps {
		t.Errorf("Expected %d hash operations, got %d", expectedHashOps, allMetrics["hash_operations"])
	}
	if allMetrics["nonce_iterations"].(uint64) != expectedNonceIterations {
		t.Errorf("Expected %d nonce iterations, got %d", expectedNonceIterations, allMetrics["nonce_iterations"])
	}
	if allMetrics["validation_checks"].(uint64) != expectedValidations {
		t.Errorf("Expected %d validation checks, got %d", expectedValidations, allMetrics["validation_checks"])
	}
}
