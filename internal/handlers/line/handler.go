// Package line provides LINE bot webhook handling.
package line

import (
	"net/http"
)

// Handler processes LINE bot webhook events.
type Handler struct {
	channelSecret string
	channelToken  string
}

// Config holds LINE handler configuration.
type Config struct {
	ChannelSecret string
	ChannelToken  string
}

// New creates a new LINE webhook handler.
func New(cfg Config) *Handler {
	return &Handler{
		channelSecret: cfg.ChannelSecret,
		channelToken:  cfg.ChannelToken,
	}
}

// Start begins the LINE webhook handler.
// Note: LINE uses webhooks, so this is a no-op. The actual HTTP server
// should be started separately and route requests to HandleWebhook.
func (h *Handler) Start() error {
	// LINE uses webhooks, so no persistent connection needed
	return nil
}

// Stop gracefully shuts down the LINE handler.
func (h *Handler) Stop() error {
	// No cleanup needed for webhook-based handler
	return nil
}

// HandleWebhook processes incoming LINE webhook requests.
// This is designed to be used with net/http or any HTTP framework.
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Always close the request body to prevent connection leaks
	if r.Body != nil {
		defer r.Body.Close()
	}

	// LINE webhooks only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement LINE webhook handling
	// 1. Validate X-Line-Signature header
	// 2. Parse webhook events from request body
	// 3. Process each event (message, follow, unfollow, etc.)
	// 4. Return 200 OK to LINE
	w.WriteHeader(http.StatusOK)
}
