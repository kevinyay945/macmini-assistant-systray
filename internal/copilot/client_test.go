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

func TestClient_NewClient(t *testing.T) {
	client, err := copilot.NewClient(copilot.Config{APIKey: "test-key"})
	if err != nil {
		t.Errorf("NewClient() returned error: %v", err)
	}
	if client == nil {
		t.Error("NewClient() returned nil")
	}
}

func TestClient_New_EmptyConfig(t *testing.T) {
	client := copilot.New(copilot.Config{})
	if client == nil {
		t.Error("New() with empty config returned nil")
	}
}

func TestClient_NewClient_WithTimeout(t *testing.T) {
	timeout := 5 * time.Minute
	client, err := copilot.NewClient(copilot.Config{
		APIKey:  "test-key",
		Timeout: timeout,
	})
	if err != nil {
		t.Errorf("NewClient() returned error: %v", err)
	}
	if client == nil {
		t.Error("NewClient() returned nil")
	}
	if client.Timeout() != timeout {
		t.Errorf("Timeout() = %v, want %v", client.Timeout(), timeout)
	}
}

func TestClient_NewClient_DefaultTimeout(t *testing.T) {
	client, err := copilot.NewClient(copilot.Config{APIKey: "test-key"})
	if err != nil {
		t.Errorf("NewClient() returned error: %v", err)
	}
	if client.Timeout() != copilot.DefaultTimeout {
		t.Errorf("Timeout() = %v, want %v", client.Timeout(), copilot.DefaultTimeout)
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

func TestClient_ProcessMessage_ClientNotStarted(t *testing.T) {
	client := copilot.New(copilot.Config{APIKey: "test-key"})
	ctx := context.Background()

	// Client not started should return ErrClientNotStarted
	_, err := client.ProcessMessage(ctx, "hello")
	if err == nil {
		t.Error("ProcessMessage() should return error when client is not started")
	}
	if !errors.Is(err, copilot.ErrClientNotStarted) {
		t.Errorf("ProcessMessage() error = %v, want ErrClientNotStarted", err)
	}
}

func TestClient_IsStarted_Default(t *testing.T) {
	client := copilot.New(copilot.Config{APIKey: "test-key"})
	if client.IsStarted() {
		t.Error("IsStarted() should return false for new client")
	}
}

func TestResponse_Fields(t *testing.T) {
	resp := &copilot.Response{
		Text:     "test response",
		Data:     map[string]interface{}{"key": "value"},
		ToolName: "test_tool",
	}

	if resp.Text != "test response" {
		t.Errorf("Text = %q, want %q", resp.Text, "test response")
	}
	if resp.Data["key"] != "value" {
		t.Errorf("Data[key] = %v, want %v", resp.Data["key"], "value")
	}
	if resp.ToolName != "test_tool" {
		t.Errorf("ToolName = %q, want %q", resp.ToolName, "test_tool")
	}
}
