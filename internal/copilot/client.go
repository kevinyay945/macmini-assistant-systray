// Package copilot provides integration with GitHub Copilot SDK.
package copilot

import (
	"context"
	"errors"
)

// Sentinel errors for the Copilot client.
var (
	ErrAPIKeyNotConfigured = errors.New("copilot API key not configured")
)

// Client handles communication with the Copilot SDK.
type Client struct {
	apiKey string
}

// Config holds Copilot client configuration.
type Config struct {
	APIKey string `yaml:"api_key" json:"api_key"`
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
	// Context check should be first to fail fast
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if c.apiKey == "" {
		return "", ErrAPIKeyNotConfigured
	}

	// TODO: Implement Copilot SDK integration
	// 1. Create request with message
	// 2. Send to Copilot API
	// 3. Parse and return response
	return "", nil
}
