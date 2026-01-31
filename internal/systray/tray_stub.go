//go:build !darwin

// Package systray provides macOS system tray functionality.
// This file provides stub implementations for non-darwin platforms.
package systray

import (
	"errors"
)

// ErrNotSupported is returned when systray is not supported on the current platform.
var ErrNotSupported = errors.New("systray is only supported on macOS")

// Run starts the system tray application.
// On non-darwin platforms, this is a no-op.
func (t *TrayApp) Run() {
	// Call callbacks for testing purposes
	if t.onReady != nil {
		t.onReady()
	}
	// Block until stop is signaled
	<-t.stopCh
	if t.onExit != nil {
		t.onExit()
	}
}

// SetState updates the application state.
// On non-darwin platforms, this only updates the internal state.
func (t *TrayApp) SetState(state AppState) {
	t.setStateInternal(state)
}

// Quit gracefully shuts down the system tray application.
func (t *TrayApp) Quit() {
	// Stop the orchestrator if running
	if t.orchestrator != nil && t.orchestrator.IsRunning() {
		_ = t.orchestrator.Stop()
	}
	// Signal stop
	select {
	case <-t.stopCh:
		// Already closed
	default:
		close(t.stopCh)
	}
}

// Run starts the system tray application (legacy interface).
// On non-darwin platforms, this is a no-op.
func (a *App) Run() error {
	if a.onReady != nil {
		a.onReady()
	}
	if a.onExit != nil {
		a.onExit()
	}
	return nil
}

// Quit gracefully shuts down the system tray application (legacy interface).
func (a *App) Quit() {
	// No-op on non-darwin
}
