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

// HandleWebhook processes incoming LINE webhook requests.
// This is designed to be used with net/http or any HTTP framework.
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// Always close the request body to prevent connection leaks
	if r.Body != nil {
		defer r.Body.Close()
	}

	// TODO: Implement LINE webhook handling
	// 1. Validate X-Line-Signature header
	// 2. Parse webhook events from request body
	// 3. Process each event (message, follow, unfollow, etc.)
	// 4. Return 200 OK to LINE
	w.WriteHeader(http.StatusOK)
}
