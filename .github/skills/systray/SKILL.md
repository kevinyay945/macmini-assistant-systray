```skill
---
name: systray
description: Expert guidance for using the github.com/getlantern/systray package in Go to create cross-platform system tray applications with icons, menus, and notifications. Use when working with: (1) Creating system tray applications, (2) Adding menu items to system tray, (3) Handling tray icon clicks, (4) Managing menu item state (enabled/disabled, checked/unchecked), (5) Setting tray icons and tooltips, or any other system tray functionality on macOS, Windows, or Linux.
---

# Systray - Cross-Platform System Tray for Go

This skill provides comprehensive guidance for using the `github.com/getlantern/systray` package to create system tray applications in Go.

## Reference Documentation

**API Reference**: [./references/api-reference.txt](./references/api-reference.txt) - Complete package documentation with all functions, types, and methods

## Installation

```bash
go get github.com/getlantern/systray
```

## Core Concepts

### 1. Application Lifecycle

The systray package provides two main functions for initializing the application:

**Run (blocking):**
```go
import "github.com/getlantern/systray"

func main() {
    systray.Run(onReady, onExit)
}

func onReady() {
    // Set up tray icon and menu
    systray.SetIcon(icon.Data)
    systray.SetTitle("My App")
    systray.SetTooltip("My Application Tooltip")
}

func onExit() {
    // Cleanup before exit
}
```

**Register (non-blocking):**
```go
func main() {
    systray.Register(onReady, onExit)
    // Show other UI (e.g., webview)
    // Run event loop elsewhere
}
```

**Key differences:**
- `Run()`: Blocks until `systray.Quit()` is called - use for tray-only apps
- `Register()`: Returns immediately - use when showing other UI elements alongside the tray
- On macOS pre-Catalina, `Register()` behaves identically to `Run()`

### 2. Setting Tray Icon

**Standard icon:**
```go
systray.SetIcon(iconBytes)
```

**Template icon (macOS-specific with fallback):**
```go
// On macOS: uses template icon (adapts to menu bar theme)
// On Windows/Linux: falls back to regular icon
systray.SetTemplateIcon(templateIconBytes, regularIconBytes)
```

**Icon format requirements:**
- Windows: `.ico`
- macOS/Linux: `.ico`, `.jpg`, or `.png`

### 3. Creating Menu Items

**Basic menu item:**
```go
mItem := systray.AddMenuItem("Menu Item", "Tooltip text")
```

**Menu item with checkbox (Linux-specific):**
```go
// Linux: creates checkable menu item
// Windows/macOS: same as AddMenuItem (items are checkable by default)
mItemCheckbox := systray.AddMenuItemCheckbox("Checkable Item", "Tooltip", true)
```

**Sub-menu items:**
```go
parentItem := systray.AddMenuItem("Parent", "Parent tooltip")
subItem := parentItem.AddSubMenuItem("Child", "Child tooltip")

// With checkbox on Linux
subItemCheck := parentItem.AddSubMenuItemCheckbox("Checkable Child", "Tooltip", false)
```

**Separator:**
```go
systray.AddSeparator()
```

### 4. Handling Menu Item Clicks

Menu items have a `ClickedCh` channel that receives notifications when clicked:

```go
mItem := systray.AddMenuItem("Click Me", "Tooltip")

// Handle clicks in a goroutine
go func() {
    for range mItem.ClickedCh {
        fmt.Println("Menu item clicked!")
        // Perform action
    }
}()
```

**Pattern for quit button:**
```go
mQuit := systray.AddMenuItem("Quit", "Exit the application")

go func() {
    <-mQuit.ClickedCh
    systray.Quit()
}()
```

### 5. Menu Item State Management

**Enable/Disable:**
```go
mItem.Disable()  // Grays out the item
mItem.Enable()   // Makes it clickable again
mItem.Disabled() // Returns bool status
```

**Check/Uncheck:**
```go
mItem.Check()    // Shows checkmark
mItem.Uncheck()  // Removes checkmark
mItem.Checked()  // Returns bool status
```

**Show/Hide:**
```go
mItem.Hide()  // Hides the menu item
mItem.Show()  // Shows the menu item again
```

**Set title and tooltip:**
```go
mItem.SetTitle("New Title")
mItem.SetTooltip("New Tooltip")
```

**Set icon (macOS/Windows only):**
```go
mItem.SetIcon(iconBytes)
mItem.SetTemplateIcon(templateIconBytes, regularIconBytes)
```

### 6. Setting Tray Title and Tooltip

**Title (macOS/Linux only):**
```go
systray.SetTitle("App Name")
```

**Tooltip (macOS/Windows only):**
```go
systray.SetTooltip("Hover text for the tray icon")
```

### 7. Quitting the Application

```go
systray.Quit()  // Stops the event loop and triggers onExit callback
```

## Common Patterns

### Pattern 1: Simple Tray Application

```go
package main

import (
    "github.com/getlantern/systray"
    "github.com/getlantern/systray/example/icon"
)

func main() {
    systray.Run(onReady, onExit)
}

func onReady() {
    systray.SetIcon(icon.Data)
    systray.SetTitle("My App")
    systray.SetTooltip("My Application")

    mQuit := systray.AddMenuItem("Quit", "Quit the application")

    go func() {
        <-mQuit.ClickedCh
        systray.Quit()
    }()
}

func onExit() {
    // Cleanup
}
```

### Pattern 2: Menu with Multiple Actions

```go
func onReady() {
    systray.SetIcon(icon.Data)
    systray.SetTooltip("Multi-action App")

    mShow := systray.AddMenuItem("Show Window", "Display main window")
    mSettings := systray.AddMenuItem("Settings", "Open settings")
    systray.AddSeparator()
    mQuit := systray.AddMenuItem("Quit", "Exit")

    go func() {
        for {
            select {
            case <-mShow.ClickedCh:
                showMainWindow()
            case <-mSettings.ClickedCh:
                openSettings()
            case <-mQuit.ClickedCh:
                systray.Quit()
                return
            }
        }
    }()
}
```

### Pattern 3: Checkable Menu Items

```go
func onReady() {
    systray.SetIcon(icon.Data)

    mEnable := systray.AddMenuItemCheckbox("Enable Feature", "Toggle feature", false)

    go func() {
        for range mEnable.ClickedCh {
            if mEnable.Checked() {
                mEnable.Uncheck()
                disableFeature()
            } else {
                mEnable.Check()
                enableFeature()
            }
        }
    }()
}
```

### Pattern 4: Dynamic Menu Updates

```go
func onReady() {
    systray.SetIcon(icon.Data)

    mStatus := systray.AddMenuItem("Status: Idle", "Current status")
    mStatus.Disable() // Status display only

    mStart := systray.AddMenuItem("Start", "Start process")
    mStop := systray.AddMenuItem("Stop", "Stop process")
    mStop.Disable() // Initially disabled

    go func() {
        for range mStart.ClickedCh {
            mStatus.SetTitle("Status: Running")
            mStart.Disable()
            mStop.Enable()
            startProcess()
        }
    }()

    go func() {
        for range mStop.ClickedCh {
            mStatus.SetTitle("Status: Idle")
            mStart.Enable()
            mStop.Disable()
            stopProcess()
        }
    }()
}
```

### Pattern 5: Sub-menus for Organization

```go
func onReady() {
    systray.SetIcon(icon.Data)

    // Main menu
    mFile := systray.AddMenuItem("File", "File operations")
    mEdit := systray.AddMenuItem("Edit", "Edit operations")
    systray.AddSeparator()
    mQuit := systray.AddMenuItem("Quit", "Exit")

    // File sub-menu
    mNew := mFile.AddSubMenuItem("New", "Create new file")
    mOpen := mFile.AddSubMenuItem("Open", "Open file")
    mSave := mFile.AddSubMenuItem("Save", "Save file")

    // Edit sub-menu
    mCopy := mEdit.AddSubMenuItem("Copy", "Copy")
    mPaste := mEdit.AddSubMenuItem("Paste", "Paste")

    // Handle clicks
    go func() {
        for {
            select {
            case <-mNew.ClickedCh:
                createNewFile()
            case <-mOpen.ClickedCh:
                openFile()
            case <-mSave.ClickedCh:
                saveFile()
            case <-mCopy.ClickedCh:
                copyToClipboard()
            case <-mPaste.ClickedCh:
                pasteFromClipboard()
            case <-mQuit.ClickedCh:
                systray.Quit()
                return
            }
        }
    }()
}
```

### Pattern 6: Icon in Embedded Assets

```go
import _ "embed"

//go:embed icon.png
var iconData []byte

func onReady() {
    systray.SetIcon(iconData)
    // ...
}
```

## Platform-Specific Behavior

### macOS
- Template icons (via `SetTemplateIcon`) adapt to light/dark menu bar themes
- Title visible in menu bar (via `SetTitle`)
- Tooltip shows on hover (via `SetTooltip`)
- Menu items checkable by default
- `Register()` behaves like `Run()` on pre-Catalina versions

### Windows
- Only `.ico` format for icons
- Tooltip shows on hover (via `SetTooltip`)
- Menu items checkable by default
- Title not supported

### Linux
- Icons support `.ico`, `.jpg`, `.png`
- Title visible (via `SetTitle`)
- Tooltip not supported
- Use `AddMenuItemCheckbox` for checkable items (not automatic)

## Best Practices

### 1. Always Use Goroutines for Click Handlers

Menu item click handlers should run in separate goroutines to avoid blocking the UI:

```go
mItem := systray.AddMenuItem("Action", "Perform action")

go func() {
    for range mItem.ClickedCh {
        performAction()
    }
}()
```

### 2. Proper Cleanup

Use the `onExit` callback for cleanup:

```go
func onExit() {
    // Close connections
    // Save state
    // Release resources
}
```

### 3. Icon Formats

Provide appropriate icon formats for each platform:
- Windows: require `.ico`
- macOS/Linux: prefer `.png` for clarity

### 4. Menu Organization

- Use separators to group related items
- Place quit/exit at the bottom
- Use sub-menus for complex hierarchies
- Keep menu depth minimal (max 2-3 levels)

### 5. State Management

Keep menu items in sync with application state:

```go
func updateMenu(running bool) {
    if running {
        mStart.Disable()
        mStop.Enable()
        mStatus.SetTitle("Status: Running")
    } else {
        mStart.Enable()
        mStop.Disable()
        mStatus.SetTitle("Status: Stopped")
    }
}
```

### 6. Error Handling

The systray package doesn't return errors from most functions. Handle errors in your click handlers:

```go
go func() {
    for range mItem.ClickedCh {
        if err := performAction(); err != nil {
            log.Printf("Action failed: %v", err)
            // Update menu to reflect error state
            mItem.SetTitle("Action Failed")
        }
    }
}()
```

## Common Use Cases

### Use Case 1: Background Service Monitor

Create a tray app that monitors a background service and allows starting/stopping:

```go
func onReady() {
    systray.SetIcon(iconStopped)

    mStatus := systray.AddMenuItem("Service: Stopped", "")
    mStatus.Disable()

    mStart := systray.AddMenuItem("Start Service", "")
    mStop := systray.AddMenuItem("Stop Service", "")
    mStop.Disable()

    systray.AddSeparator()
    mQuit := systray.AddMenuItem("Quit", "")

    // Handle service state changes
    go monitorService(mStatus, mStart, mStop)
}
```

### Use Case 2: Quick Actions Menu

Provide quick access to common actions:

```go
func onReady() {
    systray.SetIcon(icon.Data)

    mAction1 := systray.AddMenuItem("Quick Action 1", "")
    mAction2 := systray.AddMenuItem("Quick Action 2", "")
    mAction3 := systray.AddMenuItem("Quick Action 3", "")

    systray.AddSeparator()
    mQuit := systray.AddMenuItem("Quit", "")

    // Handle actions
    go handleQuickActions(mAction1, mAction2, mAction3, mQuit)
}
```

### Use Case 3: Settings Toggle

Simple on/off toggle for features:

```go
func onReady() {
    systray.SetIcon(icon.Data)

    mFeature := systray.AddMenuItemCheckbox("Feature Enabled", "", loadSetting())

    go func() {
        for range mFeature.ClickedCh {
            enabled := !mFeature.Checked()
            if enabled {
                mFeature.Check()
            } else {
                mFeature.Uncheck()
            }
            saveSetting(enabled)
            applyFeature(enabled)
        }
    }()
}
```

## Troubleshooting

### Menu Items Not Responding

Ensure click handlers run in goroutines:
```go
go func() {
    for range mItem.ClickedCh {
        // Handle click
    }
}()
```

### Icon Not Showing

- Verify icon format is correct for the platform
- Check that icon data is valid
- Ensure icon is set in `onReady` callback

### Application Not Quitting

Ensure `systray.Quit()` is called:
```go
go func() {
    <-mQuit.ClickedCh
    systray.Quit()  // Don't forget this!
}()
```

### Template Icon Not Working

Template icons only work on macOS. On other platforms, provide a regular icon as fallback:
```go
systray.SetTemplateIcon(templateIconBytes, regularIconBytes)
```

## Quick Reference

```go
// Lifecycle
systray.Run(onReady, onExit)           // Blocking
systray.Register(onReady, onExit)      // Non-blocking
systray.Quit()                         // Exit application

// Tray setup
systray.SetIcon(iconBytes)             // Set icon
systray.SetTemplateIcon(tpl, reg)      // Template icon (macOS)
systray.SetTitle(title)                // Set title (macOS/Linux)
systray.SetTooltip(tooltip)            // Set tooltip (macOS/Windows)

// Menu items
systray.AddMenuItem(title, tooltip)                    // Add menu item
systray.AddMenuItemCheckbox(title, tooltip, checked)   // Checkable (Linux)
systray.AddSeparator()                                 // Add separator

item.AddSubMenuItem(title, tooltip)                    // Add sub-item
item.AddSubMenuItemCheckbox(title, tooltip, checked)   // Checkable sub-item (Linux)

// Menu item state
item.Enable() / item.Disable() / item.Disabled()       // Enable/disable
item.Check() / item.Uncheck() / item.Checked()         // Check/uncheck
item.Show() / item.Hide()                              // Show/hide
item.SetTitle(title)                                   // Update title
item.SetTooltip(tooltip)                               // Update tooltip
item.SetIcon(iconBytes)                                // Set icon (macOS/Windows)
item.SetTemplateIcon(tpl, reg)                         // Template icon (macOS/Windows)

// Click handling
<-item.ClickedCh                       // Wait for click (in goroutine)
```
```
