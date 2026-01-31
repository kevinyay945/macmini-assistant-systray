// Package systray provides macOS system tray functionality.
package systray

import (
	"sync"
)

// AppState represents the current state of the application.
type AppState int

// Application states.
const (
	StateIdle AppState = iota
	StateActive
	StateError
)

// String returns the string representation of the state.
func (s AppState) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateActive:
		return "Running"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// Orchestrator defines the interface for the orchestrator that the tray app controls.
type Orchestrator interface {
	Start() error
	Stop() error
	IsRunning() bool
}

// SettingsOpener defines the interface for opening settings.
type SettingsOpener interface {
	OpenSettings() error
}

// UpdateChecker defines the interface for checking updates.
type UpdateChecker interface {
	CheckForUpdates() error
}

// TrayApp manages the system tray icon and menu.
type TrayApp struct {
	orchestrator   Orchestrator
	settingsOpener SettingsOpener
	updateChecker  UpdateChecker
	state          AppState
	stateMu        sync.RWMutex

	// Callbacks
	onReady func()
	onExit  func()

	// Stop channel for graceful shutdown
	stopCh chan struct{}

	// Platform-specific fields are in tray_darwin.go
}

// TrayOption configures the TrayApp.
type TrayOption func(*TrayApp)

// WithOrchestrator sets the orchestrator for the tray app.
func WithOrchestrator(o Orchestrator) TrayOption {
	return func(t *TrayApp) {
		t.orchestrator = o
	}
}

// WithSettingsOpener sets the settings opener for the tray app.
func WithSettingsOpener(s SettingsOpener) TrayOption {
	return func(t *TrayApp) {
		t.settingsOpener = s
	}
}

// WithUpdateChecker sets the update checker for the tray app.
func WithUpdateChecker(u UpdateChecker) TrayOption {
	return func(t *TrayApp) {
		t.updateChecker = u
	}
}

// WithOnReadyCallback sets the callback for when the systray is ready.
func WithOnReadyCallback(fn func()) TrayOption {
	return func(t *TrayApp) {
		t.onReady = fn
	}
}

// WithOnExitCallback sets the callback for when the systray is about to exit.
func WithOnExitCallback(fn func()) TrayOption {
	return func(t *TrayApp) {
		t.onExit = fn
	}
}

// NewTrayApp creates a new TrayApp with the given options.
func NewTrayApp(opts ...TrayOption) *TrayApp {
	t := &TrayApp{
		state:   StateIdle,
		onReady: func() {},
		onExit:  func() {},
		stopCh:  make(chan struct{}),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// GetState returns the current application state.
func (t *TrayApp) GetState() AppState {
	t.stateMu.RLock()
	defer t.stateMu.RUnlock()
	return t.state
}

// setStateInternal updates the state without UI updates.
// Used internally and by stub implementation.
func (t *TrayApp) setStateInternal(state AppState) {
	t.stateMu.Lock()
	t.state = state
	t.stateMu.Unlock()
}

// --- Legacy App type for backward compatibility ---

// App manages the system tray icon and menu (legacy interface).
type App struct {
	onReady func()
	onExit  func()
}

// Option configures the system tray application.
type Option func(*App)

// WithOnReady sets the callback for when the systray is ready.
func WithOnReady(fn func()) Option {
	return func(a *App) {
		a.onReady = fn
	}
}

// WithOnExit sets the callback for when the systray is about to exit.
func WithOnExit(fn func()) Option {
	return func(a *App) {
		a.onExit = fn
	}
}

// New creates a new system tray application.
func New(opts ...Option) *App {
	app := &App{
		onReady: func() {},
		onExit:  func() {},
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}
