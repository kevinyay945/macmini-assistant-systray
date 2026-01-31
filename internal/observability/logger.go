// Package observability provides logging, error handling, and metrics functionality.
package observability

import (
	"context"
	"io"
	"log/slog"
	"os"
	"regexp"
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

// contextKey is a type for context keys used in this package.
type contextKey string

const (
	requestIDKey contextKey = "request_id"
)

// sensitivePatterns are pre-compiled regex patterns for filtering sensitive data.
// These are defined at package level to avoid recompilation on every log call.
var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password|credential|auth)`),
}

// ParseLevel converts a string level name to a slog.Level.
func ParseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

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
	output    io.Writer
}

// WithLevel sets the minimum logging level.
func WithLevel(level slog.Level) Option {
	return func(o *loggerOptions) {
		o.level = level
	}
}

// WithLevelString sets the minimum logging level from a string.
func WithLevelString(level string) Option {
	return func(o *loggerOptions) {
		o.level = ParseLevel(level)
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

// WithOutput sets the output writer for logs.
func WithOutput(w io.Writer) Option {
	return func(o *loggerOptions) {
		o.output = w
	}
}

// sensitiveFieldFilter wraps a handler to filter sensitive data.
type sensitiveFieldFilter struct {
	slog.Handler
	patterns []*regexp.Regexp
}

func newSensitiveFieldFilter(handler slog.Handler) *sensitiveFieldFilter {
	return &sensitiveFieldFilter{
		Handler:  handler,
		patterns: sensitivePatterns,
	}
}

func (f *sensitiveFieldFilter) Handle(ctx context.Context, r slog.Record) error {
	filteredRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)

	r.Attrs(func(a slog.Attr) bool {
		// Check if key contains sensitive patterns
		for _, pattern := range f.patterns {
			if pattern.MatchString(a.Key) {
				filteredRecord.AddAttrs(slog.String(a.Key, "[REDACTED]"))
				return true
			}
		}
		// Also check if string value contains sensitive patterns
		if strVal, ok := a.Value.Any().(string); ok {
			for _, pattern := range f.patterns {
				if pattern.MatchString(strVal) {
					filteredRecord.AddAttrs(slog.String(a.Key, "[REDACTED]"))
					return true
				}
			}
		}
		filteredRecord.AddAttrs(a)
		return true
	})

	return f.Handler.Handle(ctx, filteredRecord)
}

func (f *sensitiveFieldFilter) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &sensitiveFieldFilter{
		Handler:  f.Handler.WithAttrs(attrs),
		patterns: f.patterns,
	}
}

func (f *sensitiveFieldFilter) WithGroup(name string) slog.Handler {
	return &sensitiveFieldFilter{
		Handler:  f.Handler.WithGroup(name),
		patterns: f.patterns,
	}
}

// New creates a new logger instance with the given options.
func New(opts ...Option) *Logger {
	options := &loggerOptions{
		level:     slog.LevelInfo,
		jsonMode:  false,
		addSource: false,
		output:    os.Stdout,
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
		handler = slog.NewJSONHandler(options.output, handlerOpts)
	} else {
		handler = slog.NewTextHandler(options.output, handlerOpts)
	}

	// Wrap with sensitive data filter
	handler = newSensitiveFieldFilter(handler)

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

// WithRequestID returns a new logger with the request ID attached.
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.With("request_id", requestID)
}

// WithTool returns a new logger with the tool name attached.
func (l *Logger) WithTool(toolName string) *Logger {
	return l.With("tool", toolName)
}

// WithPlatform returns a new logger with the platform name attached.
func (l *Logger) WithPlatform(platform string) *Logger {
	return l.With("platform", platform)
}

// ContextWithRequestID adds a request ID to the context.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext retrieves the request ID from context.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
