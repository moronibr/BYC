package security

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/youngchain/internal/core/block"
	"github.com/youngchain/internal/core/monitoring"
	"github.com/youngchain/internal/core/transaction"
)

// ErrorType represents the category of error
type ErrorType string

const (
	// Device Errors
	ErrorTypeDeviceConnection ErrorType = "device_connection"
	ErrorTypeDeviceTimeout    ErrorType = "device_timeout"
	ErrorTypeDeviceLocked     ErrorType = "device_locked"
	ErrorTypeDeviceWiped      ErrorType = "device_wiped"

	// Security Errors
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypeAuthorization  ErrorType = "authorization"
	ErrorTypeInvalidPin     ErrorType = "invalid_pin"
	ErrorTypeInvalidKey     ErrorType = "invalid_key"

	// Transaction Errors
	ErrorTypeTransactionFailed ErrorType = "transaction_failed"
	ErrorTypeInvalidAmount     ErrorType = "invalid_amount"
	ErrorTypeInsufficientFunds ErrorType = "insufficient_funds"

	// System Errors
	ErrorTypeSystemFailure ErrorType = "system_failure"
	ErrorTypeConfiguration ErrorType = "configuration"
	ErrorTypeNetwork       ErrorType = "network"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	ErrorSeverityLow      ErrorSeverity = "low"
	ErrorSeverityMedium   ErrorSeverity = "medium"
	ErrorSeverityHigh     ErrorSeverity = "high"
	ErrorSeverityCritical ErrorSeverity = "critical"
)

// Error represents a detailed error with metadata
type Error struct {
	Type      ErrorType
	Severity  ErrorSeverity
	Message   string
	Code      int
	Timestamp time.Time
	Context   map[string]interface{}
	Original  error
}

// ErrorHandler manages error handling and recovery
type ErrorHandler struct {
	mu              sync.RWMutex
	errorCounts     map[ErrorType]int
	errorThreshold  map[ErrorType]int
	recoveryFuncs   map[ErrorType]func(context.Context) error
	notifiers       []ErrorNotifier
	circuitBreakers map[ErrorType]*CircuitBreaker
	logger          *log.Logger
	file            *os.File
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu           sync.RWMutex
	failures     int
	threshold    int
	lastFailure  time.Time
	state        CircuitState
	resetTimeout time.Duration
}

type CircuitState string

const (
	CircuitStateClosed   CircuitState = "closed"
	CircuitStateOpen     CircuitState = "open"
	CircuitStateHalfOpen CircuitState = "half-open"
)

// ErrorNotifier defines the interface for error notification systems
type ErrorNotifier interface {
	NotifyError(error *Error) error
}

// NewErrorHandler creates a new error handler with default configurations
func NewErrorHandler(logPath string) (*ErrorHandler, error) {
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	logger := log.New(file, "", log.LstdFlags)
	handler := &ErrorHandler{
		errorCounts:     make(map[ErrorType]int),
		errorThreshold:  make(map[ErrorType]int),
		recoveryFuncs:   make(map[ErrorType]func(context.Context) error),
		notifiers:       make([]ErrorNotifier, 0),
		circuitBreakers: make(map[ErrorType]*CircuitBreaker),
		logger:          logger,
		file:            file,
	}

	// Set default thresholds
	handler.SetErrorThreshold(ErrorTypeDeviceConnection, 5)
	handler.SetErrorThreshold(ErrorTypeAuthentication, 3)
	handler.SetErrorThreshold(ErrorTypeTransactionFailed, 10)

	// Initialize circuit breakers
	handler.InitializeCircuitBreakers()

	return handler, nil
}

// InitializeCircuitBreakers sets up circuit breakers for critical operations
func (h *ErrorHandler) InitializeCircuitBreakers() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Device operations circuit breaker
	h.circuitBreakers[ErrorTypeDeviceConnection] = &CircuitBreaker{
		threshold:    5,
		resetTimeout: 30 * time.Second,
		state:        CircuitStateClosed,
	}

	// Authentication circuit breaker
	h.circuitBreakers[ErrorTypeAuthentication] = &CircuitBreaker{
		threshold:    3,
		resetTimeout: 60 * time.Second,
		state:        CircuitStateClosed,
	}

	// Transaction circuit breaker
	h.circuitBreakers[ErrorTypeTransactionFailed] = &CircuitBreaker{
		threshold:    10,
		resetTimeout: 5 * time.Minute,
		state:        CircuitStateClosed,
	}
}

// HandleError processes an error and implements recovery mechanisms
func (h *ErrorHandler) HandleError(ctx context.Context, err *Error) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Update error counts
	h.errorCounts[err.Type]++

	// Check circuit breaker
	if breaker, exists := h.circuitBreakers[err.Type]; exists {
		if !breaker.AllowRequest() {
			return fmt.Errorf("circuit breaker open for error type %s", err.Type)
		}
	}

	// Attempt recovery if a recovery function exists
	if recoveryFunc, exists := h.recoveryFuncs[err.Type]; exists {
		if recoveryErr := recoveryFunc(ctx); recoveryErr != nil {
			// If recovery fails, notify all notifiers
			h.notifyError(err)
			return recoveryErr
		}
	}

	// Check if error threshold is exceeded
	if h.errorCounts[err.Type] >= h.errorThreshold[err.Type] {
		h.notifyError(err)
		return fmt.Errorf("error threshold exceeded for type %s", err.Type)
	}

	return nil
}

// SetErrorThreshold sets the threshold for a specific error type
func (h *ErrorHandler) SetErrorThreshold(errorType ErrorType, threshold int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.errorThreshold[errorType] = threshold
}

// SetRecoveryFunction sets a recovery function for a specific error type
func (h *ErrorHandler) SetRecoveryFunction(errorType ErrorType, recoveryFunc func(context.Context) error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.recoveryFuncs[errorType] = recoveryFunc
}

// AddNotifier adds an error notifier
func (h *ErrorHandler) AddNotifier(notifier ErrorNotifier) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.notifiers = append(h.notifiers, notifier)
}

// notifyError sends error notifications to all registered notifiers
func (h *ErrorHandler) notifyError(err *Error) {
	for _, notifier := range h.notifiers {
		go notifier.NotifyError(err)
	}
}

// AllowRequest checks if a request should be allowed based on circuit breaker state
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitStateClosed:
		return true
	case CircuitStateOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = CircuitStateHalfOpen
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case CircuitStateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordFailure records a failure and updates circuit breaker state
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.failures >= cb.threshold {
		cb.state = CircuitStateOpen
	}
}

// RecordSuccess records a success and resets circuit breaker if in half-open state
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == CircuitStateHalfOpen {
		cb.failures = 0
		cb.state = CircuitStateClosed
	}
}

// GetErrorCounts returns the current error counts
func (h *ErrorHandler) GetErrorCounts() map[ErrorType]int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	counts := make(map[ErrorType]int)
	for k, v := range h.errorCounts {
		counts[k] = v
	}
	return counts
}

// ResetErrorCounts resets all error counts
func (h *ErrorHandler) ResetErrorCounts() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for k := range h.errorCounts {
		h.errorCounts[k] = 0
	}
}

// GetCircuitBreakerState returns the current state of a circuit breaker
func (h *ErrorHandler) GetCircuitBreakerState(errorType ErrorType) CircuitState {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if breaker, exists := h.circuitBreakers[errorType]; exists {
		return breaker.state
	}
	return CircuitStateClosed
}

// Close closes the error handler
func (eh *ErrorHandler) Close() error {
	return eh.file.Close()
}

// LogError logs an error with the specified severity
func (h *ErrorHandler) LogError(err error, severity ErrorSeverity) {
	message := fmt.Sprintf("[%s] %s", h.getSeverityString(severity), err.Error())
	h.logger.Println(message)

	// Log to console for high severity errors
	if severity == ErrorSeverityHigh || severity == ErrorSeverityCritical {
		fmt.Fprintf(os.Stderr, "%s\n", message)
	}

	// For critical errors, panic
	if severity == ErrorSeverityCritical {
		panic(err)
	}
}

// LogSecurityEvent logs a security event
func (eh *ErrorHandler) LogSecurityEvent(event string, details map[string]interface{}) {
	timestamp := time.Now().Format(time.RFC3339)
	message := fmt.Sprintf("[SECURITY] %s - %s", timestamp, event)

	if len(details) > 0 {
		var detailStrs []string
		for k, v := range details {
			detailStrs = append(detailStrs, fmt.Sprintf("%s=%v", k, v))
		}
		message += " - " + strings.Join(detailStrs, ", ")
	}

	eh.logger.Println(message)
}

// getSeverityString returns the string representation of a severity level
func (eh *ErrorHandler) getSeverityString(severity ErrorSeverity) string {
	switch severity {
	case ErrorSeverityLow:
		return "LOW"
	case ErrorSeverityMedium:
		return "MEDIUM"
	case ErrorSeverityHigh:
		return "HIGH"
	case ErrorSeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// WrapError wraps an error with additional context
func (eh *ErrorHandler) WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// IsSecurityError checks if an error is security-related
func (eh *ErrorHandler) IsSecurityError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "security") ||
		strings.Contains(msg, "unauthorized") ||
		strings.Contains(msg, "forbidden") ||
		strings.Contains(msg, "invalid") ||
		strings.Contains(msg, "malformed")
}

type CustomCollector struct{}

func (c *CustomCollector) Collect() (monitoring.Metric, error) {
	// Your custom logic
	return monitoring.Metric{
		Type:      "custom_metric",
		Value:     42,
		Timestamp: time.Now(),
		Labels:    map[string]string{"info": "custom"},
	}, nil
}

func main() {
	// ... other setup ...
	blockchain := block.NewBlockchain() // or however you initialize
	txPool := transaction.NewTxPool()   // or however you initialize

	monitor := monitoring.NewMonitor(blockchain, txPool, 10*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	monitor.Start(ctx)
	defer monitor.Stop()

	// ... rest of your main logic ...
}
