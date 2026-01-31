// Package handlers provides common interfaces for message platform handlers.
package handlers

// Handler defines the interface for message platform handlers.
// Both LINE and Discord handlers implement this interface for unified lifecycle management.
type Handler interface {
	// Start begins listening for events from the messaging platform.
	Start() error
	// Stop gracefully shuts down the handler.
	Stop() error
}
