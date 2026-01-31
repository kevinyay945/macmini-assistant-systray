package systray_test

import (
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/systray"
)

func TestApp_New(t *testing.T) {
	app := systray.New()
	if app == nil {
		t.Error("New() returned nil")
	}
}

func TestApp_NewWithOnReady(t *testing.T) {
	app := systray.New(
		systray.WithOnReady(func() {}),
	)
	if app == nil {
		t.Error("New() with OnReady option returned nil")
	}
}

func TestApp_NewWithOnExit(t *testing.T) {
	app := systray.New(
		systray.WithOnExit(func() {}),
	)
	if app == nil {
		t.Error("New() with OnExit option returned nil")
	}
}

func TestApp_NewWithAllOptions(t *testing.T) {
	app := systray.New(
		systray.WithOnReady(func() {}),
		systray.WithOnExit(func() {}),
	)
	if app == nil {
		t.Error("New() with all options returned nil")
	}
}

// TestAppState_String tests the string representation of AppState.
func TestAppState_String(t *testing.T) {
	tests := []struct {
		state    systray.AppState
		expected string
	}{
		{systray.StateIdle, "Idle"},
		{systray.StateActive, "Running"},
		{systray.StateError, "Error"},
		{systray.AppState(99), "Unknown"}, // Unknown state
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("AppState.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestNewTrayApp tests creating a new TrayApp.
func TestNewTrayApp(t *testing.T) {
	app := systray.NewTrayApp()
	if app == nil {
		t.Error("NewTrayApp() returned nil")
	}
}

// TestNewTrayApp_WithOptions tests creating TrayApp with options.
func TestNewTrayApp_WithOptions(t *testing.T) {
	onReadyCalled := false
	onExitCalled := false

	app := systray.NewTrayApp(
		systray.WithOnReadyCallback(func() { onReadyCalled = true }),
		systray.WithOnExitCallback(func() { onExitCalled = true }),
	)
	if app == nil {
		t.Error("NewTrayApp() with options returned nil")
	}
	// Note: callbacks would be called during Run(), which requires GUI
	_ = onReadyCalled
	_ = onExitCalled
}

// TestTrayApp_GetState tests getting the current state.
func TestTrayApp_GetState(t *testing.T) {
	app := systray.NewTrayApp()
	if app.GetState() != systray.StateIdle {
		t.Errorf("Initial state should be StateIdle, got %v", app.GetState())
	}
}

// mockOrchestrator is a mock implementation of the Orchestrator interface.
type mockOrchestrator struct {
	running   bool
	startErr  error
	stopErr   error
	startCall int
	stopCall  int
}

func (m *mockOrchestrator) Start() error {
	m.startCall++
	if m.startErr != nil {
		return m.startErr
	}
	m.running = true
	return nil
}

func (m *mockOrchestrator) Stop() error {
	m.stopCall++
	if m.stopErr != nil {
		return m.stopErr
	}
	m.running = false
	return nil
}

func (m *mockOrchestrator) IsRunning() bool {
	return m.running
}

// TestNewTrayApp_WithOrchestrator tests creating TrayApp with orchestrator.
func TestNewTrayApp_WithOrchestrator(t *testing.T) {
	mock := &mockOrchestrator{}
	app := systray.NewTrayApp(
		systray.WithOrchestrator(mock),
	)
	if app == nil {
		t.Error("NewTrayApp() with orchestrator returned nil")
	}
}

// mockSettingsOpener is a mock implementation of SettingsOpener.
type mockSettingsOpener struct {
	openCalled bool
}

func (m *mockSettingsOpener) OpenSettings() error {
	m.openCalled = true
	return nil
}

// TestNewTrayApp_WithSettingsOpener tests creating TrayApp with settings opener.
func TestNewTrayApp_WithSettingsOpener(t *testing.T) {
	mock := &mockSettingsOpener{}
	app := systray.NewTrayApp(
		systray.WithSettingsOpener(mock),
	)
	if app == nil {
		t.Error("NewTrayApp() with settings opener returned nil")
	}
}

// mockUpdateChecker is a mock implementation of UpdateChecker.
type mockUpdateChecker struct {
	checkCalled bool
}

func (m *mockUpdateChecker) CheckForUpdates() error {
	m.checkCalled = true
	return nil
}

// TestNewTrayApp_WithUpdateChecker tests creating TrayApp with update checker.
func TestNewTrayApp_WithUpdateChecker(t *testing.T) {
	mock := &mockUpdateChecker{}
	app := systray.NewTrayApp(
		systray.WithUpdateChecker(mock),
	)
	if app == nil {
		t.Error("NewTrayApp() with update checker returned nil")
	}
}

// TestTrayApp_AllOptions tests creating TrayApp with all options.
func TestTrayApp_AllOptions(t *testing.T) {
	mockOrch := &mockOrchestrator{}
	mockSettings := &mockSettingsOpener{}
	mockUpdater := &mockUpdateChecker{}

	app := systray.NewTrayApp(
		systray.WithOrchestrator(mockOrch),
		systray.WithSettingsOpener(mockSettings),
		systray.WithUpdateChecker(mockUpdater),
		systray.WithOnReadyCallback(func() {}),
		systray.WithOnExitCallback(func() {}),
	)
	if app == nil {
		t.Error("NewTrayApp() with all options returned nil")
	}
}
