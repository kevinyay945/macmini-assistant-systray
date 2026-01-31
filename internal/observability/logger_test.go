package observability_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/observability"
)

func TestLogger_New(t *testing.T) {
	l := observability.New()
	if l == nil {
		t.Error("New() returned nil")
	}
}

func TestLogger_NewWithOptions(t *testing.T) {
	l := observability.New(
		observability.WithLevel(observability.LevelDebug),
		observability.WithJSON(),
		observability.WithSource(),
	)
	if l == nil {
		t.Error("New() with options returned nil")
	}
}

func TestLogger_Methods(t *testing.T) {
	var buf bytes.Buffer
	l := observability.New(
		observability.WithOutput(&buf),
		observability.WithLevel(observability.LevelDebug),
	)
	ctx := context.Background()

	l.Info(ctx, "info message", "key", "value")
	l.Debug(ctx, "debug message", "key", "value")
	l.Warn(ctx, "warn message", "key", "value")
	l.Error(ctx, "error message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Error("Info() should log message")
	}
	if !strings.Contains(output, "debug message") {
		t.Error("Debug() should log message")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn() should log message")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error() should log message")
	}
}

func TestLogger_With(t *testing.T) {
	l := observability.New()
	child := l.With("component", "test")

	if child == nil {
		t.Error("With() returned nil")
	}
}

func TestLogger_WithGroup(t *testing.T) {
	l := observability.New()
	grouped := l.WithGroup("request")

	if grouped == nil {
		t.Error("WithGroup() returned nil")
	}
}

func TestLogger_WithRequestID(t *testing.T) {
	var buf bytes.Buffer
	l := observability.New(
		observability.WithOutput(&buf),
		observability.WithJSON(),
	)

	child := l.WithRequestID("req-123")
	child.Info(context.Background(), "test message")

	output := buf.String()
	if !strings.Contains(output, "req-123") {
		t.Errorf("WithRequestID() should include request_id in output, got: %s", output)
	}
}

func TestLogger_WithTool(t *testing.T) {
	var buf bytes.Buffer
	l := observability.New(
		observability.WithOutput(&buf),
		observability.WithJSON(),
	)

	child := l.WithTool("youtube_download")
	child.Info(context.Background(), "test message")

	output := buf.String()
	if !strings.Contains(output, "youtube_download") {
		t.Errorf("WithTool() should include tool in output, got: %s", output)
	}
}

func TestLogger_WithPlatform(t *testing.T) {
	var buf bytes.Buffer
	l := observability.New(
		observability.WithOutput(&buf),
		observability.WithJSON(),
	)

	child := l.WithPlatform("discord")
	child.Info(context.Background(), "test message")

	output := buf.String()
	if !strings.Contains(output, "discord") {
		t.Errorf("WithPlatform() should include platform in output, got: %s", output)
	}
}

func TestLogger_SensitiveDataFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := observability.New(
		observability.WithOutput(&buf),
		observability.WithJSON(),
	)
	ctx := context.Background()

	l.Info(ctx, "test", "api_key", "super-secret-key")
	l.Info(ctx, "test", "password", "my-password")
	l.Info(ctx, "test", "token", "my-token")
	l.Info(ctx, "test", "channel_secret", "secret-value")

	output := buf.String()
	if strings.Contains(output, "super-secret-key") {
		t.Error("Logger should redact api_key values")
	}
	if strings.Contains(output, "my-password") {
		t.Error("Logger should redact password values")
	}
	if strings.Contains(output, "my-token") {
		t.Error("Logger should redact token values")
	}
	if strings.Contains(output, "secret-value") {
		t.Error("Logger should redact secret values")
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Error("Logger should show [REDACTED] for sensitive fields")
	}
}

func TestLogger_SensitiveValueFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := observability.New(
		observability.WithOutput(&buf),
		observability.WithJSON(),
	)
	ctx := context.Background()

	// Test that sensitive patterns in VALUES are filtered when they appear as whole words
	// Note: With word boundary matching, "my_api_key" won't match but "api_key" will
	l.Info(ctx, "test", "data", "api_key=xyz123")
	l.Info(ctx, "test", "message", "token: abc456")
	l.Info(ctx, "test", "config", "password=hunter2")

	output := buf.String()
	if strings.Contains(output, "xyz123") {
		t.Error("Logger should redact values containing api_key")
	}
	if strings.Contains(output, "abc456") {
		t.Error("Logger should redact values containing token")
	}
	if strings.Contains(output, "hunter2") {
		t.Error("Logger should redact values containing password")
	}
}

func TestLogger_SensitiveValueFiltering_WordBoundary(t *testing.T) {
	var buf bytes.Buffer
	l := observability.New(
		observability.WithOutput(&buf),
		observability.WithJSON(),
	)
	ctx := context.Background()

	// Test that word boundary prevents false positives
	// "my_api_key" should NOT be redacted because "api" is not at a word boundary
	// "author" should NOT be redacted because it's not "auth" as a word
	l.Info(ctx, "test", "author", "John Doe")
	l.Info(ctx, "test", "custom_field", "my_custom_api_key_value")

	output := buf.String()
	if !strings.Contains(output, "John Doe") {
		t.Error("Logger should NOT redact 'author' field (word boundary)")
	}
	if !strings.Contains(output, "my_custom_api_key_value") {
		t.Error("Logger should NOT redact values where 'api_key' is not at word boundary")
	}
}

func TestLogger_BearerTokenFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := observability.New(
		observability.WithOutput(&buf),
		observability.WithJSON(),
	)
	ctx := context.Background()

	// Test that Bearer tokens in values are filtered
	l.Info(ctx, "auth header", "value", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
	l.Info(ctx, "auth header", "value", "bearer abc123-def456")

	output := buf.String()
	if strings.Contains(output, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Error("Logger should redact Bearer token values")
	}
	if strings.Contains(output, "abc123-def456") {
		t.Error("Logger should redact bearer token values (lowercase)")
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Error("Logger should show [REDACTED] for Bearer tokens")
	}
}

func TestLogger_NonSensitiveDataNotFiltered(t *testing.T) {
	var buf bytes.Buffer
	l := observability.New(
		observability.WithOutput(&buf),
		observability.WithJSON(),
	)
	ctx := context.Background()

	l.Info(ctx, "test", "user_id", "12345", "action", "download")

	output := buf.String()
	if !strings.Contains(output, "12345") {
		t.Error("Logger should not filter non-sensitive data")
	}
	if !strings.Contains(output, "download") {
		t.Error("Logger should not filter non-sensitive data")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected observability.Level
	}{
		{"debug", observability.LevelDebug},
		{"info", observability.LevelInfo},
		{"warn", observability.LevelWarn},
		{"error", observability.LevelError},
		{"unknown", observability.LevelInfo}, // default
		{"", observability.LevelInfo},        // default
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := observability.ParseLevel(tc.input)
			if result != tc.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestLogger_WithLevelString(t *testing.T) {
	var buf bytes.Buffer
	l := observability.New(
		observability.WithOutput(&buf),
		observability.WithLevelString("warn"),
	)
	ctx := context.Background()

	l.Debug(ctx, "debug message")
	l.Info(ctx, "info message")
	l.Warn(ctx, "warn message")

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("Debug should be filtered at warn level")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info should be filtered at warn level")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn should be logged at warn level")
	}
}

func TestContextWithRequestID(t *testing.T) {
	ctx := context.Background()
	ctx = observability.ContextWithRequestID(ctx, "test-request-id")

	id := observability.RequestIDFromContext(ctx)
	if id != "test-request-id" {
		t.Errorf("RequestIDFromContext() = %q, want %q", id, "test-request-id")
	}
}

func TestRequestIDFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	id := observability.RequestIDFromContext(ctx)
	if id != "" {
		t.Errorf("RequestIDFromContext() = %q, want empty string", id)
	}
}

func TestLogger_StructuredOutput(t *testing.T) {
	var buf bytes.Buffer
	l := observability.New(
		observability.WithOutput(&buf),
		observability.WithJSON(),
	)
	ctx := context.Background()

	l.Info(ctx, "structured test", "count", 42, "enabled", true)

	output := buf.String()
	// JSON output should contain the structured fields
	if !strings.Contains(output, `"count"`) {
		t.Error("JSON output should contain structured field names")
	}
	if !strings.Contains(output, "42") {
		t.Error("JSON output should contain structured field values")
	}
}
