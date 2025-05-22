package wallet

import (
	"fmt"
	"sync"
	"time"
)

// Error types for different failure scenarios
type (
	// InsufficientFundsError occurs when there are not enough funds for a transaction
	InsufficientFundsError struct {
		Required  float64
		Available float64
		CoinType  string
	}

	// InvalidAddressError occurs when an address is invalid
	InvalidAddressError struct {
		Address string
		Reason  string
	}

	// InvalidAmountError occurs when a transaction amount is invalid
	InvalidAmountError struct {
		Amount float64
		Reason string
	}

	// EncryptionError occurs during wallet encryption/decryption
	EncryptionError struct {
		Operation string
		Reason    string
	}

	// TransactionError occurs during transaction creation or processing
	TransactionError struct {
		Operation string
		Reason    string
		TxID      string
	}

	// BackupError occurs during wallet backup/restore
	BackupError struct {
		Path    string
		Reason  string
		Details map[string]interface{}
	}

	// RateLimitError occurs when operation rate limit is exceeded
	RateLimitError struct {
		Operation string
		Limit     int
		Window    time.Duration
	}

	// NetworkError occurs during network operations
	NetworkError struct {
		Operation string
		Reason    string
	}

	// SecurityError occurs during security-sensitive operations
	SecurityError struct {
		Operation string
		Reason    string
	}

	// ValidationError occurs during data validation
	ValidationError struct {
		Field  string
		Value  interface{}
		Reason string
	}

	// StateError occurs when wallet is in an invalid state
	StateError struct {
		CurrentState  string
		RequiredState string
		Operation     string
	}

	// RecoveryError occurs during wallet recovery
	RecoveryError struct {
		Method  string
		Reason  string
		Details map[string]interface{}
	}

	// MonitoringError occurs during monitoring operations
	MonitoringError struct {
		Component string
		Metric    string
		Threshold float64
		Value     float64
	}

	// AnalyticsError occurs during analytics operations
	AnalyticsError struct {
		Operation string
		Data      interface{}
		Reason    string
	}

	// WalletError represents a wallet-related error
	WalletError struct {
		Operation string
		Reason    string
		Details   map[string]interface{}
	}

	// RestoreError represents a wallet restore error
	RestoreError struct {
		Path    string
		Reason  string
		Details map[string]interface{}
	}
)

// Error messages and recovery suggestions
const (
	ErrInsufficientFundsMsg      = "insufficient funds: required %f %s, available %f %s"
	ErrInsufficientFundsRecovery = "Please ensure you have enough funds before attempting the transaction"

	ErrInvalidAddressMsg      = "invalid address '%s': %s"
	ErrInvalidAddressRecovery = "Please check the address format and ensure it's a valid wallet address"

	ErrInvalidAmountMsg      = "invalid amount %f: %s"
	ErrInvalidAmountRecovery = "Please ensure the amount is greater than 0 and within valid limits"

	ErrEncryptionMsg      = "encryption error during %s: %s"
	ErrEncryptionRecovery = "Please ensure you're using a strong password and try again"

	ErrTransactionMsg      = "transaction error during %s: %s"
	ErrTransactionRecovery = "Please check your transaction details and try again"

	ErrBackupMsg      = "backup error during %s at path '%s': %s"
	ErrBackupRecovery = "Please ensure you have write permissions and sufficient disk space"

	ErrRateLimitMsg      = "rate limit exceeded for %s: %d operations per %v"
	ErrRateLimitRecovery = "Please wait before attempting the operation again"

	ErrNetworkMsg      = "network error during %s: %s"
	ErrNetworkRecovery = "Please check your network connection and try again"

	ErrSecurityMsg      = "security error during %s: %s"
	ErrSecurityRecovery = "Please ensure you're following security best practices"

	ErrValidationMsg      = "validation error for field '%s' with value '%v': %s"
	ErrValidationRecovery = "Please check the input values and ensure they meet the requirements"

	ErrStateMsg      = "invalid wallet state for operation '%s': current state '%s', required state '%s'"
	ErrStateRecovery = "Please ensure the wallet is in the correct state before performing this operation"

	ErrRecoveryMsg      = "recovery error using method '%s': %s"
	ErrRecoveryRecovery = "Please check the recovery data and try again with valid information"

	ErrMonitoringMsg      = "monitoring error in component '%s' for metric '%s': value %.2f exceeds threshold %.2f"
	ErrMonitoringRecovery = "Please check the system resources and configuration"

	ErrAnalyticsMsg      = "analytics error during '%s': %s"
	ErrAnalyticsRecovery = "Please check the analytics configuration and data format"
)

// Error methods
func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf(ErrInsufficientFundsMsg, e.Required, e.CoinType, e.Available, e.CoinType)
}

func (e *InsufficientFundsError) Recovery() string {
	return ErrInsufficientFundsRecovery
}

func (e *InvalidAddressError) Error() string {
	return fmt.Sprintf(ErrInvalidAddressMsg, e.Address, e.Reason)
}

func (e *InvalidAddressError) Recovery() string {
	return ErrInvalidAddressRecovery
}

func (e *InvalidAmountError) Error() string {
	return fmt.Sprintf(ErrInvalidAmountMsg, e.Amount, e.Reason)
}

func (e *InvalidAmountError) Recovery() string {
	return ErrInvalidAmountRecovery
}

func (e *EncryptionError) Error() string {
	return fmt.Sprintf(ErrEncryptionMsg, e.Operation, e.Reason)
}

func (e *EncryptionError) Recovery() string {
	return ErrEncryptionRecovery
}

func (e *TransactionError) Error() string {
	return fmt.Sprintf(ErrTransactionMsg, e.Operation, e.Reason)
}

func (e *TransactionError) Recovery() string {
	return ErrTransactionRecovery
}

func (e *BackupError) Error() string {
	return e.Reason
}

func (e *BackupError) Recovery() string {
	return ErrBackupRecovery
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf(ErrRateLimitMsg, e.Operation, e.Limit, e.Window)
}

func (e *RateLimitError) Recovery() string {
	return ErrRateLimitRecovery
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf(ErrNetworkMsg, e.Operation, e.Reason)
}

func (e *NetworkError) Recovery() string {
	return ErrNetworkRecovery
}

func (e *SecurityError) Error() string {
	return fmt.Sprintf(ErrSecurityMsg, e.Operation, e.Reason)
}

func (e *SecurityError) Recovery() string {
	return ErrSecurityRecovery
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf(ErrValidationMsg, e.Field, e.Value, e.Reason)
}

func (e *ValidationError) Recovery() string {
	return ErrValidationRecovery
}

func (e *StateError) Error() string {
	return fmt.Sprintf(ErrStateMsg, e.Operation, e.CurrentState, e.RequiredState)
}

func (e *StateError) Recovery() string {
	return ErrStateRecovery
}

func (e *RecoveryError) Error() string {
	return fmt.Sprintf(ErrRecoveryMsg, e.Method, e.Reason)
}

func (e *RecoveryError) Recovery() string {
	return ErrRecoveryRecovery
}

func (e *MonitoringError) Error() string {
	return fmt.Sprintf(ErrMonitoringMsg, e.Component, e.Metric, e.Value, e.Threshold)
}

func (e *MonitoringError) Recovery() string {
	return ErrMonitoringRecovery
}

func (e *AnalyticsError) Error() string {
	return fmt.Sprintf(ErrAnalyticsMsg, e.Operation, e.Reason)
}

func (e *AnalyticsError) Recovery() string {
	return ErrAnalyticsRecovery
}

// RateLimiter implements rate limiting for sensitive operations
type RateLimiter struct {
	operations map[string][]time.Time
	limits     map[string]struct {
		count  int
		window time.Duration
	}
	mu sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		operations: make(map[string][]time.Time),
		limits: map[string]struct {
			count  int
			window time.Duration
		}{
			"create_transaction": {10, time.Minute},
			"encrypt_wallet":     {5, time.Hour},
			"decrypt_wallet":     {5, time.Hour},
			"backup_wallet":      {3, time.Hour},
			"restore_wallet":     {3, time.Hour},
		},
	}
}

// CheckRateLimit checks if an operation is within rate limits
func (rl *RateLimiter) CheckRateLimit(operation string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, exists := rl.limits[operation]
	if !exists {
		return nil
	}

	now := time.Now()
	windowStart := now.Add(-limit.window)

	// Clean up old operations
	var validOps []time.Time
	for _, t := range rl.operations[operation] {
		if t.After(windowStart) {
			validOps = append(validOps, t)
		}
	}
	rl.operations[operation] = validOps

	// Check if limit is exceeded
	if len(validOps) >= limit.count {
		return &RateLimitError{
			Operation: operation,
			Limit:     limit.count,
			Window:    limit.window,
		}
	}

	// Add new operation
	rl.operations[operation] = append(rl.operations[operation], now)
	return nil
}

// ResetRateLimit resets the rate limit for an operation
func (rl *RateLimiter) ResetRateLimit(operation string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.operations, operation)
}

// ErrorMonitor handles error monitoring and alerting
type ErrorMonitor struct {
	errors     map[string][]error
	thresholds map[string]int
	alerts     chan error
	mu         sync.RWMutex
}

// NewErrorMonitor creates a new error monitor
func NewErrorMonitor() *ErrorMonitor {
	return &ErrorMonitor{
		errors: make(map[string][]error),
		thresholds: map[string]int{
			"validation":  10,
			"security":    5,
			"transaction": 20,
			"network":     15,
			"recovery":    3,
			"monitoring":  5,
			"analytics":   10,
		},
		alerts: make(chan error, 100),
	}
}

// RecordError records an error and checks if it exceeds the threshold
func (em *ErrorMonitor) RecordError(err error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Get error type
	var errorType string
	switch err.(type) {
	case *ValidationError:
		errorType = "validation"
	case *SecurityError:
		errorType = "security"
	case *TransactionError:
		errorType = "transaction"
	case *NetworkError:
		errorType = "network"
	case *RecoveryError:
		errorType = "recovery"
	case *MonitoringError:
		errorType = "monitoring"
	case *AnalyticsError:
		errorType = "analytics"
	default:
		errorType = "unknown"
	}

	// Record error
	em.errors[errorType] = append(em.errors[errorType], err)

	// Check threshold
	if threshold, exists := em.thresholds[errorType]; exists {
		if len(em.errors[errorType]) >= threshold {
			em.alerts <- err
		}
	}
}

// GetErrorStats returns error statistics
func (em *ErrorMonitor) GetErrorStats() map[string]int {
	em.mu.RLock()
	defer em.mu.RUnlock()

	stats := make(map[string]int)
	for errorType, errors := range em.errors {
		stats[errorType] = len(errors)
	}
	return stats
}

// ResetErrors resets error counts for a specific type
func (em *ErrorMonitor) ResetErrors(errorType string) {
	em.mu.Lock()
	defer em.mu.Unlock()
	delete(em.errors, errorType)
}

// GetAlerts returns the alerts channel
func (em *ErrorMonitor) GetAlerts() <-chan error {
	return em.alerts
}

func (e *WalletError) Error() string {
	return e.Reason
}

func (e *RestoreError) Error() string {
	return e.Reason
}
