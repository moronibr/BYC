package security

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	// SeverityLow represents low severity errors
	SeverityLow ErrorSeverity = iota
	// SeverityMedium represents medium severity errors
	SeverityMedium
	// SeverityHigh represents high severity errors
	SeverityHigh
	// SeverityCritical represents critical severity errors
	SeverityCritical
)

// ErrorHandler manages error handling and logging
type ErrorHandler struct {
	logger *log.Logger
	file   *os.File
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logPath string) (*ErrorHandler, error) {
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	logger := log.New(file, "", log.LstdFlags)
	return &ErrorHandler{
		logger: logger,
		file:   file,
	}, nil
}

// Close closes the error handler
func (eh *ErrorHandler) Close() error {
	return eh.file.Close()
}

// LogError logs an error with the given severity
func (eh *ErrorHandler) LogError(err error, severity ErrorSeverity) {
	if err == nil {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Format error message
	severityStr := eh.getSeverityString(severity)
	message := fmt.Sprintf("[%s] %s (%s:%d)", severityStr, err.Error(), file, line)

	// Log to file
	eh.logger.Println(message)

	// Log to console for high severity errors
	if severity >= SeverityHigh {
		fmt.Fprintf(os.Stderr, "%s\n", message)
	}
}

// HandleError handles an error with the given severity
func (eh *ErrorHandler) HandleError(err error, severity ErrorSeverity) error {
	if err == nil {
		return nil
	}

	eh.LogError(err, severity)

	// For critical errors, panic
	if severity == SeverityCritical {
		panic(err)
	}

	return err
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
	case SeverityLow:
		return "LOW"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityHigh:
		return "HIGH"
	case SeverityCritical:
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
