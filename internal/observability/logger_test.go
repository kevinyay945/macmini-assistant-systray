package observability_test

import (
	"context"
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
	l := observability.New()
	ctx := context.Background()

	l.Info(ctx, "info message", "key", "value")
	l.Debug(ctx, "debug message", "key", "value")
	l.Warn(ctx, "warn message", "key", "value")
	l.Error(ctx, "error message", "key", "value")
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
