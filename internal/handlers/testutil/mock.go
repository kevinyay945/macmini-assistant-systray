// Package testutil provides shared test utilities for handler tests.
package testutil

import (
	"context"
	"sync"

	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
)

// MockRouter implements handlers.MessageRouter for testing.
// This is the canonical mock implementation - use this instead of
// creating duplicate mocks in individual test files.
// It is safe for concurrent use.
type MockRouter struct {
	Response *handlers.Response
	Err      error

	mu      sync.RWMutex
	called  bool
	lastMsg *handlers.Message
}

// Route processes an incoming message and returns a configured response.
// It is safe for concurrent use.
func (m *MockRouter) Route(_ context.Context, msg *handlers.Message) (*handlers.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called = true
	m.lastMsg = msg
	return m.Response, m.Err
}

// Called returns whether Route was called.
// It is safe for concurrent use.
func (m *MockRouter) Called() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.called
}

// LastMsg returns the last message passed to Route.
// It is safe for concurrent use.
func (m *MockRouter) LastMsg() *handlers.Message {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastMsg
}

// Reset clears the mock state for reuse between tests.
// It is safe for concurrent use.
func (m *MockRouter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called = false
	m.lastMsg = nil
}
