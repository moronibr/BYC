package monitoring

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// LevelDebug represents debug level messages
	LevelDebug LogLevel = iota
	// LevelInfo represents info level messages
	LevelInfo
	// LevelWarning represents warning level messages
	LevelWarning
	// LevelError represents error level messages
	LevelError
	// LevelFatal represents fatal level messages
	LevelFatal
)

var (
	levelStrings = map[LogLevel]string{
		LevelDebug:   "DEBUG",
		LevelInfo:    "INFO",
		LevelWarning: "WARNING",
		LevelError:   "ERROR",
		LevelFatal:   "FATAL",
	}
)

// Logger manages logging with different levels and outputs
type Logger struct {
	mu      sync.Mutex
	level   LogLevel
	loggers map[LogLevel]*log.Logger
	file    *os.File
	metrics *MetricsCollector
}

// NewLogger creates a new logger
func NewLogger(logPath string, level LogLevel, metrics *MetricsCollector) (*Logger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// Create multi-writer for both file and stdout
	multiWriter := io.MultiWriter(file, os.Stdout)

	// Create loggers for each level
	loggers := make(map[LogLevel]*log.Logger)
	for l := LevelDebug; l <= LevelFatal; l++ {
		loggers[l] = log.New(multiWriter, fmt.Sprintf("[%s] ", levelStrings[l]), log.LstdFlags)
	}

	return &Logger{
		level:   level,
		loggers: loggers,
		file:    file,
		metrics: metrics,
	}, nil
}

// Close closes the logger
func (l *Logger) Close() error {
	return l.file.Close()
}

// log logs a message with the given level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Format message
	message := fmt.Sprintf(format, args...)
	location := fmt.Sprintf("%s:%d", filepath.Base(file), line)
	logMessage := fmt.Sprintf("%s [%s] %s", location, levelStrings[level], message)

	// Log message
	l.loggers[level].Println(logMessage)

	// Update metrics for errors
	if level >= LevelError {
		l.metrics.IncrementError(levelStrings[level])
	}

	// Exit for fatal errors
	if level == LevelFatal {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warning logs a warning message
func (l *Logger) Warning(format string, args ...interface{}) {
	l.log(LevelWarning, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelFatal, format, args...)
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// Rotate rotates the log file
func (l *Logger) Rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Close current file
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %v", err)
	}

	// Create new file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	newPath := fmt.Sprintf("%s.%s", l.file.Name(), timestamp)
	if err := os.Rename(l.file.Name(), newPath); err != nil {
		return fmt.Errorf("failed to rename log file: %v", err)
	}

	// Open new file
	file, err := os.OpenFile(l.file.Name(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open new log file: %v", err)
	}

	// Update loggers with new file
	multiWriter := io.MultiWriter(file, os.Stdout)
	for level, logger := range l.loggers {
		l.loggers[level] = log.New(multiWriter, logger.Prefix(), logger.Flags())
	}

	l.file = file
	return nil
}
