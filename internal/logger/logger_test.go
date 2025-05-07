package logger

import (
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestInit(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "byc-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create log file path
	logPath := filepath.Join(tmpDir, "test.log")

	// Test logger initialization
	config := Config{
		Level:      "debug",
		Filename:   logPath,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	if err := Init(config); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Test logging functions
	Debug("debug message", String("key", "value"))
	Info("info message", Int("number", 42))
	Warn("warning message", Bool("flag", true))
	Error("error message", Error2(err))

	// Verify log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// Test child logger
	childLogger := With(String("child", "test"))
	if childLogger == nil {
		t.Error("Child logger is nil")
	}

	// Test sync
	if err := Sync(); err != nil {
		t.Errorf("Failed to sync logger: %v", err)
	}
}

func TestLogLevels(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "byc-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test different log levels
	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		logPath := filepath.Join(tmpDir, level+".log")
		config := Config{
			Level:      level,
			Filename:   logPath,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}

		if err := Init(config); err != nil {
			t.Errorf("Failed to initialize logger with level %s: %v", level, err)
			continue
		}

		// Log messages at different levels
		Debug("debug message")
		Info("info message")
		Warn("warning message")
		Error("error message")

		// Verify log file exists
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Errorf("Log file was not created for level %s", level)
		}
	}
}

func TestFieldConstructors(t *testing.T) {
	// Test all field constructors
	fields := []zap.Field{
		String("string", "value"),
		Int("int", 42),
		Int64("int64", 42),
		Float64("float64", 42.0),
		Bool("bool", true),
		Error2(nil),
	}

	for _, field := range fields {
		if field.Type == 0 {
			t.Error("Field constructor returned invalid field type")
		}
	}
}
