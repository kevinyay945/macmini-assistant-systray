// Package observability provides logging and metrics functionality.
package observability

import (
	"log"
	"os"
)

// Logger provides structured logging capabilities.
type Logger struct {
	logger *log.Logger
}

// New creates a new logger instance.
func New() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// Info logs an informational message.
func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Printf("[INFO] "+msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+msg, args...)
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logger.Printf("[DEBUG] "+msg, args...)
}
