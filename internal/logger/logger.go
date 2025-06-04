package logger

import (
	"fmt"

	"go.uber.org/zap"
)

var log *zap.Logger

// Init initializes the logger
func Init() error {
	var err error
	log, err = zap.NewProduction()
	if err != nil {
		return err
	}
	return nil
}

// checkLogger ensures the logger is initialized
func checkLogger() {
	if log == nil {
		panic(fmt.Errorf("logger not initialized, call logger.Init() first"))
	}
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	checkLogger()
	log.Info(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	checkLogger()
	log.Error(msg, fields...)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	checkLogger()
	log.Debug(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	checkLogger()
	log.Warn(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	checkLogger()
	log.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	checkLogger()
	return log.Sync()
}
