package line_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers/line"
	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers/testutil"
)

func TestHandler_New(t *testing.T) {
	h := line.New(line.Config{
		ChannelSecret: "secret",
		ChannelToken:  "token",
	})
	if h == nil {
		t.Error("New() returned nil")
	}
}

func TestHandler_New_EmptyConfig(t *testing.T) {
	h := line.New(line.Config{})
	if h == nil {
		t.Error("New() with empty config returned nil")
	}
}

func TestHandler_New_WithRouter(t *testing.T) {
	router := testutil.NewMockRouter()
	h := line.New(line.Config{
		ChannelSecret: "secret",
		ChannelToken:  "token",
		Router:        router,
	})
	if h == nil {
		t.Error("New() with router returned nil")
	}
}

func TestHandler_HandleWebhook_MethodNotAllowed(t *testing.T) {
	h := line.New(line.Config{})

	testCases := []struct {
		name   string
		method string
	}{
		{"GET", http.MethodGet},
		{"PUT", http.MethodPut},
		{"DELETE", http.MethodDelete},
		{"PATCH", http.MethodPatch},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/webhook", nil)
			w := httptest.NewRecorder()

			h.HandleWebhook(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("HandleWebhook() status = %d, want %d for %s", w.Code, http.StatusMethodNotAllowed, tc.method)
			}
		})
	}
}

func TestHandler_HandleWebhook_InvalidSignature(t *testing.T) {
	h := line.New(line.Config{
		ChannelSecret: "test-secret",
		ChannelToken:  "test-token",
	})

	body := strings.NewReader(`{"events":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Line-Signature", "invalid-signature")
	w := httptest.NewRecorder()

	h.HandleWebhook(w, req)

	// Should return 400 for invalid signature
	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleWebhook() status = %d, want %d for invalid signature", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_HandleWebhook_EmptyBody(t *testing.T) {
	h := line.New(line.Config{
		ChannelSecret: "secret",
		ChannelToken:  "token",
	})

	body := strings.NewReader(``)
	req := httptest.NewRequest(http.MethodPost, "/webhook", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWebhook(w, req)

	// Should return 400 for empty body (parsing will fail)
	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleWebhook() status = %d, want %d for empty body", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_Start(t *testing.T) {
	h := line.New(line.Config{})
	if err := h.Start(); err != nil {
		t.Errorf("Start() returned error: %v", err)
	}
}

func TestHandler_Start_Idempotent(t *testing.T) {
	h := line.New(line.Config{})

	// Start twice should not error
	if err := h.Start(); err != nil {
		t.Errorf("Start() first call returned error: %v", err)
	}
	if err := h.Start(); err != nil {
		t.Errorf("Start() second call returned error: %v", err)
	}
}

func TestHandler_Stop(t *testing.T) {
	h := line.New(line.Config{})
	if err := h.Stop(); err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}
}

func TestHandler_Stop_Idempotent(t *testing.T) {
	h := line.New(line.Config{})

	_ = h.Start()

	// Stop twice should not error
	if err := h.Stop(); err != nil {
		t.Errorf("Stop() first call returned error: %v", err)
	}
	if err := h.Stop(); err != nil {
		t.Errorf("Stop() second call returned error: %v", err)
	}
}

func TestHandler_StartStop_Lifecycle(t *testing.T) {
	h := line.New(line.Config{
		ChannelSecret: "secret",
		ChannelToken:  "token",
	})

	// Start
	if err := h.Start(); err != nil {
		t.Errorf("Start() returned error: %v", err)
	}

	// Stop
	if err := h.Stop(); err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	// Start again
	if err := h.Start(); err != nil {
		t.Errorf("Start() after Stop() returned error: %v", err)
	}
}

func TestHandler_InterfaceCompliance(t *testing.T) {
	var _ handlers.Handler = (*line.Handler)(nil)
}

func TestHandler_UsesCanonicalErrorFormatter(t *testing.T) {
	// Full error formatting is tested in handlers/interface_test.go
	// This test just verifies the function exists and is accessible
	err := errors.New("test error")
	result := handlers.FormatUserFriendlyError(err)
	if result == "" {
		t.Error("FormatUserFriendlyError should return non-empty string for non-nil error")
	}
}

func TestHandler_ParseMessage_Errors(t *testing.T) {
	h := line.New(line.Config{
		ChannelSecret: "test-secret",
		ChannelToken:  "test-token",
	})

	// Test with nil Message - this would cause unsupported message type error
	// We can't directly test webhook.MessageEvent creation without importing internal types,
	// but we can verify the exported constants and error types

	if line.ErrEmptyMessage == nil {
		t.Error("ErrEmptyMessage should be defined")
	}

	if line.ErrEmptyMessage.Error() != "line: empty message content" {
		t.Errorf("ErrEmptyMessage = %q, want %q", line.ErrEmptyMessage.Error(), "line: empty message content")
	}

	// Verify MaxMessageLength constant
	if line.MaxMessageLength != 5000 {
		t.Errorf("MaxMessageLength = %d, want %d", line.MaxMessageLength, 5000)
	}

	// Verify DefaultReplyTimeout constant
	if line.DefaultReplyTimeout != 30*time.Second {
		t.Errorf("DefaultReplyTimeout = %v, want %v", line.DefaultReplyTimeout, 30*time.Second)
	}

	// Handler should not be nil
	if h == nil {
		t.Error("New() returned nil handler")
	}
}
