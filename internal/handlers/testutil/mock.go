// Package testutil provides shared test utilities for handler tests.
package testutil

import (
	"context"

	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
)

// MockRouter implements handlers.MessageRouter for testing.
// This is the canonical mock implementation - use this instead of
// creating duplicate mocks in individual test files.
type MockRouter struct {
	Response *handlers.Response
	Err      error
	Called   bool
	LastMsg  *handlers.Message
}

// Route processes an incoming message and returns a configured response.
func (m *MockRouter) Route(_ context.Context, msg *handlers.Message) (*handlers.Response, error) {
	m.Called = true
	m.LastMsg = msg
	return m.Response, m.Err
}

// Reset clears the mock state for reuse between tests.
func (m *MockRouter) Reset() {
	m.Called = false
	m.LastMsg = nil
}
