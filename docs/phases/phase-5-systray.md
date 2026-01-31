# Phase 5: System Tray & Auto-start

**Duration**: Week 9
**Status**: ⚪ Not Started
**Goal**: Create macOS system tray interface and startup integration

---

## Overview

This phase implements the macOS system tray (menu bar) interface and auto-start functionality. The tray provides user control over the application.

---

## 5.1 System Tray Application

**Duration**: 3 days
**Status**: ⚪ Not Started

### Tasks

- [ ] System tray icon and menu
- [ ] Menu items: Start/Stop, Settings, Check Updates, Quit
- [ ] Icon state indicators (active, error, idle)
- [ ] Graceful shutdown handling

### Implementation Details

```go
// internal/systray/tray.go
package systray

import (
    "github.com/getlantern/systray"
)

type TrayApp struct {
    orchestrator *Orchestrator
    logger       *observability.Logger
    state        AppState

    // Menu items
    mStatus      *systray.MenuItem
    mToggle      *systray.MenuItem
    mSettings    *systray.MenuItem
    mCheckUpdate *systray.MenuItem
    mQuit        *systray.MenuItem
}

type AppState int

const (
    StateIdle AppState = iota
    StateActive
    StateError
)

func NewTrayApp(orchestrator *Orchestrator) *TrayApp

func (t *TrayApp) Run() {
    systray.Run(t.onReady, t.onExit)
}

func (t *TrayApp) onReady() {
    // Set icon
    systray.SetIcon(iconIdle)
    systray.SetTitle("")
    systray.SetTooltip("MacMini Assistant")

    // Create menu items
    t.mStatus = systray.AddMenuItem("Status: Idle", "Current status")
    t.mStatus.Disable()

    systray.AddSeparator()

    t.mToggle = systray.AddMenuItem("Start", "Start/Stop the bot")
    t.mSettings = systray.AddMenuItem("Settings...", "Open settings")
    t.mCheckUpdate = systray.AddMenuItem("Check for Updates", "Check for updates")

    systray.AddSeparator()

    t.mQuit = systray.AddMenuItem("Quit", "Quit the application")

    // Handle menu clicks
    go t.handleMenuClicks()
}

func (t *TrayApp) handleMenuClicks() {
    for {
        select {
        case <-t.mToggle.ClickedCh:
            t.toggleBot()
        case <-t.mSettings.ClickedCh:
            t.openSettings()
        case <-t.mCheckUpdate.ClickedCh:
            t.checkUpdates()
        case <-t.mQuit.ClickedCh:
            t.quit()
            return
        }
    }
}

func (t *TrayApp) SetState(state AppState) {
    t.state = state
    switch state {
    case StateIdle:
        systray.SetIcon(iconIdle)
        t.mStatus.SetTitle("Status: Idle")
        t.mToggle.SetTitle("Start")
    case StateActive:
        systray.SetIcon(iconActive)
        t.mStatus.SetTitle("Status: Running")
        t.mToggle.SetTitle("Stop")
    case StateError:
        systray.SetIcon(iconError)
        t.mStatus.SetTitle("Status: Error")
    }
}

func (t *TrayApp) quit() {
    t.logger.Info("Shutting down...")
    t.orchestrator.Stop()
    systray.Quit()
}
```

**Icon Assets**:
```go
// internal/systray/icons.go
package systray

// Embed icons as byte arrays
// Use png2ico or similar to convert PNG to ICO/ICNS

var (
    iconIdle   []byte // Gray icon
    iconActive []byte // Green icon
    iconError  []byte // Red icon
)

//go:embed assets/icon_idle.png
var iconIdle []byte

//go:embed assets/icon_active.png
var iconActive []byte

//go:embed assets/icon_error.png
var iconError []byte
```

### Test Cases

```go
// internal/systray/tray_test.go
//go:build local

func TestSysTray_Initialize(t *testing.T)
func TestSysTray_MenuItems(t *testing.T)
func TestSysTray_IconStates(t *testing.T)
func TestSysTray_GracefulShutdown(t *testing.T)
func TestSysTray_ToggleBot(t *testing.T)
func TestSysTray_StateTransitions(t *testing.T)
```

**Build Tag**: `local` (requires macOS GUI)

### Acceptance Criteria

- [ ] Icon appears in macOS menu bar
- [ ] Menu items functional
- [ ] Icon reflects application state (colors)
- [ ] App shuts down gracefully on Quit
- [ ] Start/Stop toggles correctly

### Notes

<!-- Add your notes here -->

---

## 5.2 Auto-start on Login

**Duration**: 2 days
**Status**: ⚪ Not Started

### Tasks

- [ ] LaunchAgent plist generation
- [ ] Installation to `~/Library/LaunchAgents/`
- [ ] Enable/disable via config
- [ ] Uninstall functionality

### Implementation Details

```go
// internal/systray/autostart.go
package systray

import (
    "fmt"
    "os"
    "path/filepath"
    "text/template"
)

const launchAgentName = "com.macmini-assistant.orchestrator"

type AutoStart struct {
    enabled bool
    plistPath string
    execPath  string
}

func NewAutoStart() *AutoStart {
    homeDir, _ := os.UserHomeDir()
    return &AutoStart{
        plistPath: filepath.Join(homeDir, "Library", "LaunchAgents", launchAgentName+".plist"),
    }
}

func (a *AutoStart) Enable() error {
    // Get current executable path
    execPath, err := os.Executable()
    if err != nil {
        return fmt.Errorf("failed to get executable path: %w", err)
    }
    a.execPath = execPath

    // Generate plist
    plist, err := a.generatePlist()
    if err != nil {
        return fmt.Errorf("failed to generate plist: %w", err)
    }

    // Write plist file
    if err := os.WriteFile(a.plistPath, []byte(plist), 0644); err != nil {
        return fmt.Errorf("failed to write plist: %w", err)
    }

    // Load launch agent
    if err := exec.Command("launchctl", "load", a.plistPath).Run(); err != nil {
        return fmt.Errorf("failed to load launch agent: %w", err)
    }

    a.enabled = true
    return nil
}

func (a *AutoStart) Disable() error {
    // Unload launch agent
    exec.Command("launchctl", "unload", a.plistPath).Run()

    // Remove plist file
    if err := os.Remove(a.plistPath); err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("failed to remove plist: %w", err)
    }

    a.enabled = false
    return nil
}

func (a *AutoStart) IsEnabled() bool {
    _, err := os.Stat(a.plistPath)
    return err == nil
}

func (a *AutoStart) generatePlist() (string, error) {
    const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.ExecPath}}</string>
        <string>start</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
    <key>StandardOutPath</key>
    <string>{{.LogPath}}/stdout.log</string>
    <key>StandardErrorPath</key>
    <string>{{.LogPath}}/stderr.log</string>
</dict>
</plist>`

    homeDir, _ := os.UserHomeDir()
    logPath := filepath.Join(homeDir, ".macmini-assistant", "logs")

    // Ensure log directory exists
    os.MkdirAll(logPath, 0755)

    tmpl, err := template.New("plist").Parse(plistTemplate)
    if err != nil {
        return "", err
    }

    var buf bytes.Buffer
    err = tmpl.Execute(&buf, map[string]string{
        "Label":    launchAgentName,
        "ExecPath": a.execPath,
        "LogPath":  logPath,
    })
    if err != nil {
        return "", err
    }

    return buf.String(), nil
}
```

### Test Cases

```go
// internal/systray/autostart_test.go
//go:build local

func TestAutoStart_PlistGeneration(t *testing.T)
func TestAutoStart_Install(t *testing.T)
func TestAutoStart_Uninstall(t *testing.T)
func TestAutoStart_ConfigToggle(t *testing.T)
func TestAutoStart_IsEnabled(t *testing.T)
func TestAutoStart_InvalidPath(t *testing.T)
```

**Build Tag**: `local` (requires macOS)

### Acceptance Criteria

- [ ] App starts on login when enabled
- [ ] LaunchAgent correctly configured
- [ ] Can be disabled via config
- [ ] Uninstall removes LaunchAgent
- [ ] Logs written to appropriate location

### Notes

<!-- Add your notes here -->

---

## Deliverables

By the end of Phase 5:

- [ ] System tray icon and menu working
- [ ] Start/Stop functionality
- [ ] Auto-start on login (optional)
- [ ] Graceful shutdown

---

## Dependencies

```go
// go.mod additions
require (
    github.com/getlantern/systray v1.2.x
)
```

---

## Icon Requirements

Create three PNG icons (16x16 or 22x22 for menu bar):
- `icon_idle.png` - Gray/neutral color
- `icon_active.png` - Green color
- `icon_error.png` - Red color

---

## Time Tracking

| Task | Estimated | Actual | Notes |
|------|-----------|--------|-------|
| 5.1 System Tray | 3 days | | |
| 5.2 Auto-start | 2 days | | |
| **Total** | **5 days** | | |

---

**Previous**: [Phase 4: Tool Implementation](./phase-4-tools.md)
**Next**: [Phase 6: Auto-updater](./phase-6-updater.md)
