package systray_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/systray"
)

func TestNewAutoStart(t *testing.T) {
	a, err := systray.NewAutoStart()
	if err != nil {
		t.Fatalf("NewAutoStart() error = %v", err)
	}
	if a == nil {
		t.Error("NewAutoStart() returned nil")
	}
}

func TestNewAutoStart_WithOptions(t *testing.T) {
	tmpDir := t.TempDir()
	execPath := "/test/binary"
	logPath := filepath.Join(tmpDir, "logs")
	plistPath := filepath.Join(tmpDir, "test.plist")

	a, err := systray.NewAutoStart(
		systray.WithExecPath(execPath),
		systray.WithLogPath(logPath),
		systray.WithPlistPath(plistPath),
	)
	if err != nil {
		t.Fatalf("NewAutoStart() with options error = %v", err)
	}
	if a == nil {
		t.Error("NewAutoStart() with options returned nil")
	}
	if a.GetPlistPath() != plistPath {
		t.Errorf("GetPlistPath() = %v, want %v", a.GetPlistPath(), plistPath)
	}
	if a.GetLogPath() != logPath {
		t.Errorf("GetLogPath() = %v, want %v", a.GetLogPath(), logPath)
	}
}

func TestAutoStart_IsEnabled_NotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := filepath.Join(tmpDir, "nonexistent.plist")

	a, err := systray.NewAutoStart(
		systray.WithPlistPath(plistPath),
	)
	if err != nil {
		t.Fatalf("NewAutoStart() error = %v", err)
	}

	if a.IsEnabled() {
		t.Error("IsEnabled() should return false for non-existent plist")
	}
}

func TestAutoStart_GeneratePlistContent(t *testing.T) {
	tmpDir := t.TempDir()
	execPath := "/Applications/MacMiniAssistant.app/Contents/MacOS/orchestrator"
	logPath := filepath.Join(tmpDir, "logs")
	plistPath := filepath.Join(tmpDir, "test.plist")

	a, err := systray.NewAutoStart(
		systray.WithExecPath(execPath),
		systray.WithLogPath(logPath),
		systray.WithPlistPath(plistPath),
	)
	if err != nil {
		t.Fatalf("NewAutoStart() error = %v", err)
	}

	content, err := a.GeneratePlistContent()
	if err != nil {
		t.Fatalf("GeneratePlistContent() error = %v", err)
	}

	// Verify plist content contains expected elements
	expectedElements := []string{
		`<string>com.macmini-assistant.orchestrator</string>`,
		`<string>` + execPath + `</string>`,
		`<string>start</string>`,
		`<true/>`,
		`<string>` + logPath + `/stdout.log</string>`,
		`<string>` + logPath + `/stderr.log</string>`,
	}

	for _, expected := range expectedElements {
		if !strings.Contains(content, expected) {
			t.Errorf("GeneratePlistContent() missing expected element: %s", expected)
		}
	}
}

func TestAutoStart_GetPaths(t *testing.T) {
	a, err := systray.NewAutoStart()
	if err != nil {
		t.Fatalf("NewAutoStart() error = %v", err)
	}

	homeDir, _ := os.UserHomeDir()

	// Check plist path
	expectedPlistPath := filepath.Join(homeDir, "Library", "LaunchAgents", "com.macmini-assistant.orchestrator.plist")
	if a.GetPlistPath() != expectedPlistPath {
		t.Errorf("GetPlistPath() = %v, want %v", a.GetPlistPath(), expectedPlistPath)
	}

	// Check log path
	expectedLogPath := filepath.Join(homeDir, ".macmini-assistant", "logs")
	if a.GetLogPath() != expectedLogPath {
		t.Errorf("GetLogPath() = %v, want %v", a.GetLogPath(), expectedLogPath)
	}
}

func TestAutoStart_Disable_NotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := filepath.Join(tmpDir, "nonexistent.plist")

	a, err := systray.NewAutoStart(
		systray.WithPlistPath(plistPath),
	)
	if err != nil {
		t.Fatalf("NewAutoStart() error = %v", err)
	}

	// Disable should succeed even if not installed
	if err := a.Disable(); err != nil {
		t.Errorf("Disable() error = %v, want nil", err)
	}
}

func TestAutoStart_PlistFormat(t *testing.T) {
	tmpDir := t.TempDir()
	execPath := "/usr/local/bin/orchestrator"
	logPath := filepath.Join(tmpDir, "logs")

	a, err := systray.NewAutoStart(
		systray.WithExecPath(execPath),
		systray.WithLogPath(logPath),
		systray.WithPlistPath(filepath.Join(tmpDir, "test.plist")),
	)
	if err != nil {
		t.Fatalf("NewAutoStart() error = %v", err)
	}

	content, err := a.GeneratePlistContent()
	if err != nil {
		t.Fatalf("GeneratePlistContent() error = %v", err)
	}

	// Verify it's a valid plist structure
	if !strings.HasPrefix(content, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Error("Plist should start with XML declaration")
	}
	if !strings.Contains(content, `<!DOCTYPE plist`) {
		t.Error("Plist should contain DOCTYPE")
	}
	if !strings.Contains(content, `<plist version="1.0">`) {
		t.Error("Plist should contain plist version")
	}
	if !strings.HasSuffix(strings.TrimSpace(content), `</plist>`) {
		t.Error("Plist should end with </plist>")
	}
}
