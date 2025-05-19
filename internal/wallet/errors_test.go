package wallet

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWalletErrors tests the error types and their messages
func TestWalletErrors(t *testing.T) {
	// Test InsufficientFundsError
	err := &InsufficientFundsError{
		Required:  100.0,
		Available: 50.0,
		CoinType:  "Leah",
	}
	assert.Equal(t, "insufficient funds: required 100.000000 Leah, available 50.000000 Leah", err.Error())
	assert.Equal(t, "Please ensure you have enough funds before attempting the transaction", err.Recovery())

	// Test InvalidAddressError
	err2 := &InvalidAddressError{
		Address: "invalid-address",
		Reason:  "invalid format",
	}
	assert.Equal(t, "invalid address 'invalid-address': invalid format", err2.Error())
	assert.Equal(t, "Please check the address format and ensure it's a valid wallet address", err2.Recovery())

	// Test InvalidAmountError
	err3 := &InvalidAmountError{
		Amount: -1.0,
		Reason: "amount must be positive",
	}
	assert.Equal(t, "invalid amount -1.000000: amount must be positive", err3.Error())
	assert.Equal(t, "Please ensure the amount is greater than 0 and within valid limits", err3.Recovery())

	// Test EncryptionError
	err4 := &EncryptionError{
		Operation: "encrypt",
		Reason:    "invalid password",
	}
	assert.Equal(t, "encryption error during encrypt: invalid password", err4.Error())
	assert.Equal(t, "Please ensure you're using a strong password and try again", err4.Recovery())

	// Test TransactionError
	err5 := &TransactionError{
		Operation: "create",
		Reason:    "invalid input",
		TxID:      "tx123",
	}
	assert.Equal(t, "transaction error during create: invalid input", err5.Error())
	assert.Equal(t, "Please check your transaction details and try again", err5.Recovery())

	// Test BackupError
	err6 := &BackupError{
		Operation: "write",
		Path:      "/path/to/backup",
		Reason:    "permission denied",
	}
	assert.Equal(t, "backup error during write at path '/path/to/backup': permission denied", err6.Error())
	assert.Equal(t, "Please ensure you have write permissions and sufficient disk space", err6.Recovery())

	// Test RateLimitError
	err7 := &RateLimitError{
		Operation: "create_transaction",
		Limit:     10,
		Window:    time.Minute,
	}
	assert.Equal(t, "rate limit exceeded for create_transaction: 10 operations per 1m0s", err7.Error())
	assert.Equal(t, "Please wait before attempting the operation again", err7.Recovery())

	// Test NetworkError
	err8 := &NetworkError{
		Operation: "broadcast",
		Reason:    "connection refused",
	}
	assert.Equal(t, "network error during broadcast: connection refused", err8.Error())
	assert.Equal(t, "Please check your network connection and try again", err8.Recovery())

	// Test SecurityError
	err9 := &SecurityError{
		Operation: "sign",
		Reason:    "invalid key",
	}
	assert.Equal(t, "security error during sign: invalid key", err9.Error())
	assert.Equal(t, "Please ensure you're following security best practices", err9.Recovery())
}

// TestRateLimiter tests the rate limiter functionality
func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter()

	// Test within limits
	for i := 0; i < 10; i++ {
		err := limiter.CheckRateLimit("create_transaction")
		require.NoError(t, err)
	}

	// Test exceeding limit
	err := limiter.CheckRateLimit("create_transaction")
	require.Error(t, err)
	assert.IsType(t, &RateLimitError{}, err)

	// Test different operations
	err = limiter.CheckRateLimit("encrypt_wallet")
	require.NoError(t, err)

	// Test reset
	limiter.ResetRateLimit("create_transaction")
	err = limiter.CheckRateLimit("create_transaction")
	require.NoError(t, err)

	// Test window expiration
	time.Sleep(time.Minute)
	for i := 0; i < 10; i++ {
		err := limiter.CheckRateLimit("create_transaction")
		require.NoError(t, err)
	}
}

// TestRateLimiterConcurrent tests concurrent rate limiter operations
func TestRateLimiterConcurrent(t *testing.T) {
	limiter := NewRateLimiter()
	done := make(chan bool)
	errors := make(chan error, 20)

	// Start multiple goroutines
	for i := 0; i < 20; i++ {
		go func() {
			err := limiter.CheckRateLimit("create_transaction")
			errors <- err
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	// Count errors
	errorCount := 0
	for i := 0; i < 20; i++ {
		if err := <-errors; err != nil {
			errorCount++
		}
	}

	// Should have exactly 10 errors (limit is 10 per minute)
	assert.Equal(t, 10, errorCount)
}

// TestRateLimiterDifferentOperations tests rate limiting for different operations
func TestRateLimiterDifferentOperations(t *testing.T) {
	limiter := NewRateLimiter()

	// Test create_transaction limit
	for i := 0; i < 10; i++ {
		err := limiter.CheckRateLimit("create_transaction")
		require.NoError(t, err)
	}
	err := limiter.CheckRateLimit("create_transaction")
	require.Error(t, err)

	// Test encrypt_wallet limit
	for i := 0; i < 5; i++ {
		err := limiter.CheckRateLimit("encrypt_wallet")
		require.NoError(t, err)
	}
	err = limiter.CheckRateLimit("encrypt_wallet")
	require.Error(t, err)

	// Test backup_wallet limit
	for i := 0; i < 3; i++ {
		err := limiter.CheckRateLimit("backup_wallet")
		require.NoError(t, err)
	}
	err = limiter.CheckRateLimit("backup_wallet")
	require.Error(t, err)

	// Test unknown operation (should not be limited)
	for i := 0; i < 100; i++ {
		err := limiter.CheckRateLimit("unknown_operation")
		require.NoError(t, err)
	}
}

// TestNewErrorTypes tests the new error types and their messages
func TestNewErrorTypes(t *testing.T) {
	// Test ValidationError
	err := &ValidationError{
		Field:  "amount",
		Value:  -100.0,
		Reason: "must be positive",
	}
	assert.Equal(t, "validation error for field 'amount' with value '-100': must be positive", err.Error())
	assert.Equal(t, "Please check the input values and ensure they meet the requirements", err.Recovery())

	// Test StateError
	err2 := &StateError{
		CurrentState:  "encrypted",
		RequiredState: "decrypted",
		Operation:     "sign_transaction",
	}
	assert.Equal(t, "invalid wallet state for operation 'sign_transaction': current state 'encrypted', required state 'decrypted'", err2.Error())
	assert.Equal(t, "Please ensure the wallet is in the correct state before performing this operation", err2.Recovery())

	// Test RecoveryError
	err3 := &RecoveryError{
		Method:  "mnemonic",
		Reason:  "invalid phrase",
		Details: map[string]interface{}{"words": 11},
	}
	assert.Equal(t, "recovery error using method 'mnemonic': invalid phrase", err3.Error())
	assert.Equal(t, "Please check the recovery data and try again with valid information", err3.Recovery())

	// Test MonitoringError
	err4 := &MonitoringError{
		Component: "memory",
		Metric:    "usage",
		Threshold: 80.0,
		Value:     90.0,
	}
	assert.Equal(t, "monitoring error in component 'memory' for metric 'usage': value 90.00 exceeds threshold 80.00", err4.Error())
	assert.Equal(t, "Please check the system resources and configuration", err4.Recovery())

	// Test AnalyticsError
	err5 := &AnalyticsError{
		Operation: "process_transaction",
		Data:      map[string]interface{}{"tx_id": "123"},
		Reason:    "invalid format",
	}
	assert.Equal(t, "analytics error during 'process_transaction': invalid format", err5.Error())
	assert.Equal(t, "Please check the analytics configuration and data format", err5.Recovery())
}

// TestErrorMonitor tests the error monitoring functionality
func TestErrorMonitor(t *testing.T) {
	monitor := NewErrorMonitor()

	// Test recording errors
	for i := 0; i < 5; i++ {
		monitor.RecordError(&SecurityError{
			Operation: "sign",
			Reason:    "invalid key",
		})
	}

	// Test threshold exceeded
	monitor.RecordError(&SecurityError{
		Operation: "sign",
		Reason:    "invalid key",
	})

	// Check if alert was sent
	select {
	case err := <-monitor.GetAlerts():
		assert.IsType(t, &SecurityError{}, err)
	default:
		t.Error("Expected alert for security error threshold exceeded")
	}

	// Test error stats
	stats := monitor.GetErrorStats()
	assert.Equal(t, 6, stats["security"])

	// Test reset
	monitor.ResetErrors("security")
	stats = monitor.GetErrorStats()
	assert.Equal(t, 0, stats["security"])
}

// TestErrorMonitorConcurrent tests concurrent error monitoring
func TestErrorMonitorConcurrent(t *testing.T) {
	monitor := NewErrorMonitor()
	done := make(chan bool)

	// Start multiple goroutines
	for i := 0; i < 50; i++ {
		go func() {
			monitor.RecordError(&TransactionError{
				Operation: "create",
				Reason:    "invalid input",
			})
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 50; i++ {
		<-done
	}

	// Check error stats
	stats := monitor.GetErrorStats()
	assert.Equal(t, 50, stats["transaction"])

	// Check alerts
	alertCount := 0
	for {
		select {
		case <-monitor.GetAlerts():
			alertCount++
		default:
			goto done
		}
	}
done:
	assert.Equal(t, 3, alertCount) // Should have 3 alerts (threshold is 20)
}

// TestErrorMonitorDifferentTypes tests monitoring different error types
func TestErrorMonitorDifferentTypes(t *testing.T) {
	monitor := NewErrorMonitor()

	// Record different types of errors
	monitor.RecordError(&ValidationError{
		Field:  "amount",
		Value:  -100.0,
		Reason: "must be positive",
	})

	monitor.RecordError(&NetworkError{
		Operation: "broadcast",
		Reason:    "connection refused",
	})

	monitor.RecordError(&MonitoringError{
		Component: "memory",
		Metric:    "usage",
		Threshold: 80.0,
		Value:     90.0,
	})

	// Check stats
	stats := monitor.GetErrorStats()
	assert.Equal(t, 1, stats["validation"])
	assert.Equal(t, 1, stats["network"])
	assert.Equal(t, 1, stats["monitoring"])

	// Test unknown error type
	monitor.RecordError(fmt.Errorf("unknown error"))
	stats = monitor.GetErrorStats()
	assert.Equal(t, 1, stats["unknown"])
}
