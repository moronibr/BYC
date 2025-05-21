package security

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	Timestamp   time.Time   `json:"timestamp"`
	EventType   string      `json:"event_type"`
	IP          string      `json:"ip"`
	UserID      string      `json:"user_id,omitempty"`
	Description string      `json:"description"`
	Severity    string      `json:"severity"`
	Details     interface{} `json:"details,omitempty"`
}

// SecurityAuditor handles security event logging and monitoring
type SecurityAuditor struct {
	mu            sync.Mutex
	logFile       *os.File
	events        []SecurityEvent
	maxEvents     int
	alertCallback func(SecurityEvent)
}

// NewSecurityAuditor creates a new security auditor
func NewSecurityAuditor(logPath string, maxEvents int) (*SecurityAuditor, error) {
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	return &SecurityAuditor{
		logFile:   logFile,
		events:    make([]SecurityEvent, 0, maxEvents),
		maxEvents: maxEvents,
	}, nil
}

// LogEvent logs a security event
func (sa *SecurityAuditor) LogEvent(event SecurityEvent) error {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	// Add timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Add to memory buffer
	sa.events = append(sa.events, event)
	if len(sa.events) > sa.maxEvents {
		sa.events = sa.events[1:]
	}

	// Write to log file
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	if _, err := sa.logFile.Write(append(eventJSON, '\n')); err != nil {
		return fmt.Errorf("failed to write event: %v", err)
	}

	// Trigger alert for high severity events
	if event.Severity == "high" && sa.alertCallback != nil {
		go sa.alertCallback(event)
	}

	return nil
}

// GetEvents returns recent security events
func (sa *SecurityAuditor) GetEvents() []SecurityEvent {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	events := make([]SecurityEvent, len(sa.events))
	copy(events, sa.events)
	return events
}

// SetAlertCallback sets a callback function for high severity events
func (sa *SecurityAuditor) SetAlertCallback(callback func(SecurityEvent)) {
	sa.mu.Lock()
	defer sa.mu.Unlock()
	sa.alertCallback = callback
}

// Close closes the security auditor
func (sa *SecurityAuditor) Close() error {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	if sa.logFile != nil {
		return sa.logFile.Close()
	}
	return nil
}

// Common security event types
const (
	EventTypeLogin        = "login"
	EventTypeLogout       = "logout"
	EventTypeAuthFailure  = "auth_failure"
	EventTypeRateLimit    = "rate_limit"
	EventTypeKeyRotation  = "key_rotation"
	EventTypeBlockMined   = "block_mined"
	EventTypeTransaction  = "transaction"
	EventTypeSystemError  = "system_error"
	EventTypeConfigChange = "config_change"
)

// Common severity levels
const (
	SeverityLow    = "low"
	SeverityMedium = "medium"
	SeverityHigh   = "high"
)

// Helper functions for common security events

// LogLogin logs a login event
func (sa *SecurityAuditor) LogLogin(userID, ip string, success bool) error {
	eventType := EventTypeLogin
	severity := SeverityLow
	description := "Successful login"

	if !success {
		eventType = EventTypeAuthFailure
		severity = SeverityMedium
		description = "Failed login attempt"
	}

	return sa.LogEvent(SecurityEvent{
		EventType:   eventType,
		IP:          ip,
		UserID:      userID,
		Description: description,
		Severity:    severity,
	})
}

// LogRateLimit logs a rate limit event
func (sa *SecurityAuditor) LogRateLimit(ip string, limit int) error {
	return sa.LogEvent(SecurityEvent{
		EventType:   EventTypeRateLimit,
		IP:          ip,
		Description: fmt.Sprintf("Rate limit exceeded: %d requests", limit),
		Severity:    SeverityMedium,
		Details: map[string]interface{}{
			"limit": limit,
		},
	})
}

// LogKeyRotation logs a key rotation event
func (sa *SecurityAuditor) LogKeyRotation(userID string) error {
	return sa.LogEvent(SecurityEvent{
		EventType:   EventTypeKeyRotation,
		UserID:      userID,
		Description: "Encryption key rotated",
		Severity:    SeverityLow,
	})
}

// LogBlockMined logs a block mining event
func (sa *SecurityAuditor) LogBlockMined(userID, blockHash string, blockType string) error {
	return sa.LogEvent(SecurityEvent{
		EventType:   EventTypeBlockMined,
		UserID:      userID,
		Description: fmt.Sprintf("Block mined: %s", blockHash),
		Severity:    SeverityLow,
		Details: map[string]interface{}{
			"block_hash": blockHash,
			"block_type": blockType,
		},
	})
}

// LogTransaction logs a transaction event
func (sa *SecurityAuditor) LogTransaction(userID, txID string, amount float64, coinType string) error {
	return sa.LogEvent(SecurityEvent{
		EventType:   EventTypeTransaction,
		UserID:      userID,
		Description: fmt.Sprintf("Transaction created: %s", txID),
		Severity:    SeverityMedium,
		Details: map[string]interface{}{
			"tx_id":     txID,
			"amount":    amount,
			"coin_type": coinType,
		},
	})
}

// LogSystemError logs a system error
func (sa *SecurityAuditor) LogSystemError(err error, component string) error {
	return sa.LogEvent(SecurityEvent{
		EventType:   EventTypeSystemError,
		Description: fmt.Sprintf("System error in %s: %v", component, err),
		Severity:    SeverityHigh,
		Details: map[string]interface{}{
			"error":     err.Error(),
			"component": component,
		},
	})
}
