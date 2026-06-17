//go:build darwin

// Package systray provides macOS system tray functionality.
package systray

import (
	_ "embed"
)

// Icon assets for system tray.
// These are embedded at compile time using go:embed directives.

//go:embed assets/icon_idle.png
var iconIdle []byte

//go:embed assets/icon_active.png
var iconActive []byte

//go:embed assets/icon_error.png
var iconError []byte
