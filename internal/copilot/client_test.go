package copilot_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kevinyay945/macmini-assistant-systray/internal/copilot"
)

func TestClient_New(t *testing.T) {
	client := copilot.New(copilot.Config{APIKey: "test-key"})
	if client == nil {
		t.Error("New() returned nil")
	}
}

func TestClient_New_EmptyConfig(t *testing.T) {
	client := copilot.New(copilot.Config{})
	if client == nil {
		t.Error("New() with empty config returned nil")
	}
}

func TestClient_ProcessMessage_NoAPIKey(t *testing.T) {
	client := copilot.New(copilot.Config{})
	ctx := context.Background()

	_, err := client.ProcessMessage(ctx, "hello")
	if err == nil {
		t.Error("ProcessMessage() should return error when API key is not configured")
	}
	if !errors.Is(err, copilot.ErrAPIKeyNotConfigured) {
		t.Errorf("ProcessMessage() error = %v, want ErrAPIKeyNotConfigured", err)
	}
}

func TestClient_ProcessMessage_ContextCanceled(t *testing.T) {
	client := copilot.New(copilot.Config{APIKey: "test-key"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.ProcessMessage(ctx, "hello")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("ProcessMessage() error = %v, want context.Canceled", err)
	}
}

func TestClient_ProcessMessage_ContextDeadlineExceeded(t *testing.T) {
	client := copilot.New(copilot.Config{APIKey: "test-key"})
	// Use an already-expired deadline to avoid flaky race conditions
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := client.ProcessMessage(ctx, "hello")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("ProcessMessage() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestClient_ProcessMessage_ValidRequest(t *testing.T) {
	client := copilot.New(copilot.Config{APIKey: "test-key"})
	ctx := context.Background()

	// Currently returns empty string as TODO implementation
	result, err := client.ProcessMessage(ctx, "hello")
	if err != nil {
		t.Errorf("ProcessMessage() returned error: %v", err)
	}
	// Empty result is expected for stub implementation
	if result != "" {
		t.Errorf("ProcessMessage() = %q, want empty string (stub)", result)
	}
}
