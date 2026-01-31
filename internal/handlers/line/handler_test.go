package line_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers/line"
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
	router := &mockRouter{}
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

// mockRouter implements handlers.MessageRouter for testing.
type mockRouter struct {
	response *handlers.Response
	err      error
	called   bool
	lastMsg  *handlers.Message
}

func (m *mockRouter) Route(ctx context.Context, msg *handlers.Message) (*handlers.Response, error) {
	m.called = true
	m.lastMsg = msg
	return m.response, m.err
}

func TestHandler_InterfaceCompliance(t *testing.T) {
	var _ handlers.Handler = (*line.Handler)(nil)
}

func TestHandler_ErrorFormatting(t *testing.T) {
	// Error formatting is tested in handlers/interface_test.go
	// This test verifies the canonical FormatUserFriendlyError is used consistently
	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{
			name:    "nil error",
			err:     nil,
			wantMsg: "",
		},
		{
			name:    "context deadline exceeded",
			err:     context.DeadlineExceeded,
			wantMsg: "⏱️ Request timed out. Please try again.",
		},
		{
			name:    "context canceled",
			err:     context.Canceled,
			wantMsg: "🚫 Request was cancelled.",
		},
		{
			name:    "generic error",
			err:     errors.New("something went wrong"),
			wantMsg: "❌ An error occurred while processing your request. Please try again later.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.FormatUserFriendlyError(tt.err)
			if got != tt.wantMsg {
				t.Errorf("FormatUserFriendlyError() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}
