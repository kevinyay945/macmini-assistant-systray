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

func TestApp_Run(t *testing.T) {
	app := systray.New()
	if err := app.Run(); err != nil {
		t.Errorf("Run() returned error: %v", err)
	}
}

func TestApp_Quit(t *testing.T) {
	app := systray.New()
	app.Quit()
}
