package systray

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const launchAgentName = "com.macmini-assistant.orchestrator"

// plistTemplate is the LaunchAgent plist template.
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

// AutoStart manages macOS LaunchAgent for auto-start on login.
type AutoStart struct {
	enabled   bool
	plistPath string
	execPath  string
	logPath   string
}

// AutoStartOption configures AutoStart.
type AutoStartOption func(*AutoStart)

// WithExecPath sets the executable path.
func WithExecPath(path string) AutoStartOption {
	return func(a *AutoStart) {
		a.execPath = path
	}
}

// WithLogPath sets the log directory path.
func WithLogPath(path string) AutoStartOption {
	return func(a *AutoStart) {
		a.logPath = path
	}
}

// WithPlistPath sets the plist file path.
func WithPlistPath(path string) AutoStartOption {
	return func(a *AutoStart) {
		a.plistPath = path
	}
}

// NewAutoStart creates a new AutoStart manager.
func NewAutoStart(opts ...AutoStartOption) (*AutoStart, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	a := &AutoStart{
		plistPath: filepath.Join(homeDir, "Library", "LaunchAgents", launchAgentName+".plist"),
		logPath:   filepath.Join(homeDir, ".macmini-assistant", "logs"),
	}

	for _, opt := range opts {
		opt(a)
	}

	// Update enabled state based on whether plist exists
	a.enabled = a.IsEnabled()

	return a, nil
}

// Enable installs and loads the LaunchAgent.
func (a *AutoStart) Enable() error {
	// Get current executable path if not set
	if err := a.ensureExecPath(); err != nil {
		return err
	}

	// Generate plist content
	plist, err := a.generatePlist()
	if err != nil {
		return fmt.Errorf("failed to generate plist: %w", err)
	}

	// Ensure LaunchAgents directory exists
	launchAgentsDir := filepath.Dir(a.plistPath)
	if err := os.MkdirAll(launchAgentsDir, 0o750); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	// Ensure log directory exists
	if err := os.MkdirAll(a.logPath, 0o750); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Write plist file
	if err := os.WriteFile(a.plistPath, []byte(plist), 0o644); err != nil {
		return fmt.Errorf("failed to write plist: %w", err)
	}

	// Load launch agent using launchctl
	// #nosec G204 - plistPath is derived from user's home directory, not user input
	cmd := exec.Command("launchctl", "load", a.plistPath)
	if err := cmd.Run(); err != nil {
		// Clean up plist file on failure
		_ = os.Remove(a.plistPath)
		return fmt.Errorf("failed to load launch agent: %w", err)
	}

	a.enabled = true
	return nil
}

// Disable unloads and removes the LaunchAgent.
func (a *AutoStart) Disable() error {
	// Unload launch agent (ignore error if not loaded)
	// #nosec G204 - plistPath is derived from user's home directory, not user input
	cmd := exec.Command("launchctl", "unload", a.plistPath)
	_ = cmd.Run()

	// Remove plist file
	if err := os.Remove(a.plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist: %w", err)
	}

	a.enabled = false
	return nil
}

// IsEnabled checks if the LaunchAgent is installed.
func (a *AutoStart) IsEnabled() bool {
	_, err := os.Stat(a.plistPath)
	return err == nil
}

// GetPlistPath returns the path to the plist file.
func (a *AutoStart) GetPlistPath() string {
	return a.plistPath
}

// GetLogPath returns the path to the log directory.
func (a *AutoStart) GetLogPath() string {
	return a.logPath
}

// ensureExecPath sets the executable path if not already set.
func (a *AutoStart) ensureExecPath() error {
	if a.execPath == "" {
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}
		a.execPath = execPath
	}
	return nil
}

// generatePlist generates the LaunchAgent plist content.
func (a *AutoStart) generatePlist() (string, error) {
	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse plist template: %w", err)
	}

	data := map[string]string{
		"Label":    launchAgentName,
		"ExecPath": a.execPath,
		"LogPath":  a.logPath,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute plist template: %w", err)
	}

	return buf.String(), nil
}

// GeneratePlistContent returns the plist content that would be generated.
// This is useful for testing or preview purposes.
func (a *AutoStart) GeneratePlistContent() (string, error) {
	// Set execPath if not set
	if err := a.ensureExecPath(); err != nil {
		return "", err
	}
	return a.generatePlist()
}
