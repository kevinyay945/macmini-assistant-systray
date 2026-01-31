// Package systray provides macOS system tray functionality.
package systray

// App manages the system tray icon and menu.
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

// Run starts the system tray application.
// This blocks until Quit is called or the application terminates.
func (a *App) Run() error {
	// TODO: Implement systray using github.com/getlantern/systray
	// 1. Set icon and tooltip
	// 2. Add menu items (Status, Settings, Quit)
	// 3. Handle menu item clicks
	// systray.Run(a.onReady, a.onExit)
	return nil
}

// Quit gracefully shuts down the system tray application.
func (a *App) Quit() {
	// TODO: Implement graceful shutdown
	// systray.Quit()
}
