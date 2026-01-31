// Package copilot provides integration with GitHub Copilot SDK.
package copilot

import (
	"context"
	"errors"
)

// Client handles communication with the Copilot SDK.
type Client struct {
	apiKey string
}

// Config holds Copilot client configuration.
type Config struct {
	APIKey string
}

// New creates a new Copilot client.
func New(cfg Config) *Client {
	return &Client{
		apiKey: cfg.APIKey,
	}
}

// ProcessMessage sends a message to Copilot and returns the response.
// The context is used to enforce timeouts (10-minute hard limit per PRD).
func (c *Client) ProcessMessage(ctx context.Context, message string) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("copilot API key not configured")
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// TODO: Implement Copilot SDK integration
	// 1. Create request with message
	// 2. Send to Copilot API
	// 3. Parse and return response
	return "", nil
}
