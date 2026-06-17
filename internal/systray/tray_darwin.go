//go:build darwin

// Package systray provides macOS system tray functionality.
package systray

import (
	"github.com/getlantern/systray"
)

// Run starts the system tray application.
// This blocks until Quit is called or the application terminates.
func (t *TrayApp) Run() {
	systray.Run(t.setupTray, t.cleanupTray)
}

// setupTray is the internal handler that sets up the tray menu.
func (t *TrayApp) setupTray() {
	// Set initial icon and tooltip
	systray.SetIcon(iconIdle)
	systray.SetTitle("")
	systray.SetTooltip("MacMini Assistant")

	// Create menu items
	mStatus := systray.AddMenuItem("Status: Idle", "Current status")
	mStatus.Disable()

	systray.AddSeparator()

	mToggle := systray.AddMenuItem("Start", "Start/Stop the bot")
	mSettings := systray.AddMenuItem("Settings...", "Open settings")
	mCheckUpdate := systray.AddMenuItem("Check for Updates", "Check for updates")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// Call user's onReady callback
	if t.onReady != nil {
		t.onReady()
	}

	// Handle menu clicks in a goroutine
	go t.handleMenuClicks(mStatus, mToggle, mSettings, mCheckUpdate, mQuit)
}

// cleanupTray is the internal handler for tray cleanup.
func (t *TrayApp) cleanupTray() {
	// Signal stop to handleMenuClicks
	close(t.stopCh)

	// Call user's onExit callback
	if t.onExit != nil {
		t.onExit()
	}
}

// handleMenuClicks handles menu item click events.
func (t *TrayApp) handleMenuClicks(mStatus, mToggle, mSettings, mCheckUpdate, mQuit *systray.MenuItem) {
	for {
		select {
		case <-t.stopCh:
			return
		case <-mToggle.ClickedCh:
			t.toggleBotDarwin(mStatus, mToggle)
		case <-mSettings.ClickedCh:
			t.openSettings()
		case <-mCheckUpdate.ClickedCh:
			t.checkUpdates()
		case <-mQuit.ClickedCh:
			t.Quit()
			return
		}
	}
}

// openSettings opens the settings interface.
func (t *TrayApp) openSettings() {
	if t.settingsOpener != nil {
		_ = t.settingsOpener.OpenSettings()
	}
}

// checkUpdates checks for application updates.
func (t *TrayApp) checkUpdates() {
	if t.updateChecker != nil {
		_ = t.updateChecker.CheckForUpdates()
	}
}

// toggleBotDarwin starts or stops the orchestrator with UI updates.
func (t *TrayApp) toggleBotDarwin(mStatus, mToggle *systray.MenuItem) {
	t.stateMu.RLock()
	currentState := t.state
	t.stateMu.RUnlock()

	if t.orchestrator == nil {
		return
	}

	if currentState == StateActive {
		if err := t.orchestrator.Stop(); err != nil {
			t.SetStateDarwin(StateError, mStatus, mToggle)
			return
		}
		t.SetStateDarwin(StateIdle, mStatus, mToggle)
	} else {
		if err := t.orchestrator.Start(); err != nil {
			t.SetStateDarwin(StateError, mStatus, mToggle)
			return
		}
		t.SetStateDarwin(StateActive, mStatus, mToggle)
	}
}

// SetState updates the application state and UI.
func (t *TrayApp) SetState(state AppState) {
	t.setStateInternal(state)

	switch state {
	case StateIdle:
		systray.SetIcon(iconIdle)
	case StateActive:
		systray.SetIcon(iconActive)
	case StateError:
		systray.SetIcon(iconError)
	}
}

// SetStateDarwin updates the application state and UI with menu items.
func (t *TrayApp) SetStateDarwin(state AppState, mStatus, mToggle *systray.MenuItem) {
	t.setStateInternal(state)

	switch state {
	case StateIdle:
		systray.SetIcon(iconIdle)
		if mStatus != nil {
			mStatus.SetTitle("Status: Idle")
		}
		if mToggle != nil {
			mToggle.SetTitle("Start")
		}
	case StateActive:
		systray.SetIcon(iconActive)
		if mStatus != nil {
			mStatus.SetTitle("Status: Running")
		}
		if mToggle != nil {
			mToggle.SetTitle("Stop")
		}
	case StateError:
		systray.SetIcon(iconError)
		if mStatus != nil {
			mStatus.SetTitle("Status: Error")
		}
	}
}

// Quit gracefully shuts down the system tray application.
func (t *TrayApp) Quit() {
	// Stop the orchestrator if running
	if t.orchestrator != nil && t.orchestrator.IsRunning() {
		_ = t.orchestrator.Stop()
	}
	systray.Quit()
}

// Run starts the system tray application (legacy interface).
// This blocks until Quit is called or the application terminates.
func (a *App) Run() error {
	systray.Run(a.onReady, a.onExit)
	return nil
}

// Quit gracefully shuts down the system tray application (legacy interface).
func (a *App) Quit() {
	systray.Quit()
}
