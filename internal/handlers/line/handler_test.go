package line_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func TestHandler_HandleWebhook_POST_OK(t *testing.T) {
	h := line.New(line.Config{
		ChannelSecret: "secret",
		ChannelToken:  "token",
	})

	body := strings.NewReader(`{"events":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleWebhook(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleWebhook() status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandler_Start(t *testing.T) {
	h := line.New(line.Config{})
	if err := h.Start(); err != nil {
		t.Errorf("Start() returned error: %v", err)
	}
}

func TestHandler_Stop(t *testing.T) {
	h := line.New(line.Config{})
	if err := h.Stop(); err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}
}
