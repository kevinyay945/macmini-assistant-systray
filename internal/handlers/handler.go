// Package handlers provides common interfaces for message platform handlers.
package handlers

import "context"

// Handler defines the interface for message platform handlers.
// Both LINE and Discord handlers implement this interface for unified lifecycle management.
type Handler interface {
	// Start begins listening for events from the messaging platform.
	Start() error
	// Stop gracefully shuts down the handler.
	Stop() error
}

// HealthChecker defines the interface for health check operations.
// Implementations should return the current health status of the handler.
type HealthChecker interface {
	// HealthCheck returns the current health status of the handler.
	HealthCheck(ctx context.Context) HealthStatus
}

// HealthStatus represents the health status of a handler.
type HealthStatus struct {
	// Healthy indicates whether the handler is functioning properly.
	Healthy bool
	// Message provides a human-readable status message.
	Message string
	// Details contains additional health check information.
	Details map[string]interface{}
}

// NewHealthStatus creates a new HealthStatus with initialized Details map.
func NewHealthStatus(healthy bool, message string) HealthStatus {
	return HealthStatus{
		Healthy: healthy,
		Message: message,
		Details: make(map[string]interface{}),
	}
}
