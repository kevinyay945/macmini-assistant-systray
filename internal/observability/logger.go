// Package observability provides logging and metrics functionality.
package observability

import (
	"context"
	"log/slog"
	"os"
)

// Level represents the logging level.
type Level = slog.Level

// Logging levels.
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Logger provides structured logging capabilities.
type Logger struct {
	logger *slog.Logger
}

// Option configures the logger.
type Option func(*loggerOptions)

type loggerOptions struct {
	level     slog.Level
	jsonMode  bool
	addSource bool
}

// WithLevel sets the minimum logging level.
func WithLevel(level slog.Level) Option {
	return func(o *loggerOptions) {
		o.level = level
	}
}

// WithJSON enables JSON output format.
func WithJSON() Option {
	return func(o *loggerOptions) {
		o.jsonMode = true
	}
}

// WithSource adds source file information to logs.
func WithSource() Option {
	return func(o *loggerOptions) {
		o.addSource = true
	}
}

// New creates a new logger instance with the given options.
func New(opts ...Option) *Logger {
	options := &loggerOptions{
		level:     slog.LevelInfo,
		jsonMode:  false,
		addSource: false,
	}

	for _, opt := range opts {
		opt(options)
	}

	handlerOpts := &slog.HandlerOptions{
		Level:     options.level,
		AddSource: options.addSource,
	}

	var handler slog.Handler
	if options.jsonMode {
		handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	return &Logger{
		logger: slog.New(handler),
	}
}

// Info logs an informational message with structured fields.
func (l *Logger) Info(ctx context.Context, msg string, attrs ...any) {
	l.logger.InfoContext(ctx, msg, attrs...)
}

// Error logs an error message with structured fields.
func (l *Logger) Error(ctx context.Context, msg string, attrs ...any) {
	l.logger.ErrorContext(ctx, msg, attrs...)
}

// Debug logs a debug message with structured fields.
func (l *Logger) Debug(ctx context.Context, msg string, attrs ...any) {
	l.logger.DebugContext(ctx, msg, attrs...)
}

// Warn logs a warning message with structured fields.
func (l *Logger) Warn(ctx context.Context, msg string, attrs ...any) {
	l.logger.WarnContext(ctx, msg, attrs...)
}

// With returns a new logger with the given attributes added to every log.
func (l *Logger) With(attrs ...any) *Logger {
	return &Logger{logger: l.logger.With(attrs...)}
}

// WithGroup returns a new logger with the given group name.
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{logger: l.logger.WithGroup(name)}
}
