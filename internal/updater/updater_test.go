package updater_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kevinyay945/macmini-assistant-systray/internal/updater"
)

func TestUpdater_New(t *testing.T) {
	u := updater.New(updater.Config{
		CurrentVersion: "v1.0.0",
		RepoOwner:      "kevinyay945",
		RepoName:       "macmini-assistant-systray",
	})
	if u == nil {
		t.Error("New() returned nil")
	}
}

func TestUpdater_CurrentVersion(t *testing.T) {
	u := updater.New(updater.Config{
		CurrentVersion: "v1.2.3",
	})

	if got := u.CurrentVersion(); got != "v1.2.3" {
		t.Errorf("CurrentVersion() = %q, want %q", got, "v1.2.3")
	}
}

func TestUpdater_IsNewerVersion(t *testing.T) {
	testCases := []struct {
		name           string
		currentVersion string
		newVersion     string
		wantNewer      bool
	}{
		{"newer major", "v1.0.0", "v2.0.0", true},
		{"newer minor", "v1.0.0", "v1.1.0", true},
		{"newer patch", "v1.0.0", "v1.0.1", true},
		{"same version", "v1.0.0", "v1.0.0", false},
		{"older major", "v2.0.0", "v1.0.0", false},
		{"older minor", "v1.1.0", "v1.0.0", false},
		{"older patch", "v1.0.1", "v1.0.0", false},
		{"no v prefix current", "1.0.0", "v1.0.1", true},
		{"no v prefix new", "v1.0.0", "1.0.1", true},
		{"no v prefix both", "1.0.0", "1.0.1", true},
		{"complex newer", "v1.9.0", "v1.10.0", true},
		{"dev version", "dev", "v1.0.0", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := updater.New(updater.Config{
				CurrentVersion: tc.currentVersion,
			})

			got := u.IsNewerVersion(tc.newVersion)
			if got != tc.wantNewer {
				t.Errorf("IsNewerVersion(%q) = %v, want %v", tc.newVersion, got, tc.wantNewer)
			}
		})
	}
}

func TestUpdater_CheckForUpdate(t *testing.T) {
	u := updater.New(updater.Config{
		CurrentVersion: "v1.0.0",
		RepoOwner:      "kevinyay945",
		RepoName:       "macmini-assistant-systray",
	})
	ctx := context.Background()

	info, err := u.CheckForUpdate(ctx)
	if err != nil {
		t.Errorf("CheckForUpdate() returned error: %v", err)
	}
	if info == nil {
		t.Error("CheckForUpdate() returned nil info")
	}
}

func TestUpdater_CheckForUpdate_ContextCanceled(t *testing.T) {
	u := updater.New(updater.Config{
		CurrentVersion: "v1.0.0",
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := u.CheckForUpdate(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("CheckForUpdate() error = %v, want context.Canceled", err)
	}
}

func TestUpdater_CheckForUpdate_ContextDeadlineExceeded(t *testing.T) {
	u := updater.New(updater.Config{
		CurrentVersion: "v1.0.0",
	})
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := u.CheckForUpdate(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("CheckForUpdate() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestUpdater_Update_NilInfo(t *testing.T) {
	u := updater.New(updater.Config{
		CurrentVersion: "v1.0.0",
	})
	ctx := context.Background()

	err := u.Update(ctx, nil)
	if err != nil {
		t.Errorf("Update() with nil info returned error: %v", err)
	}
}

func TestUpdater_Update_NotAvailable(t *testing.T) {
	u := updater.New(updater.Config{
		CurrentVersion: "v1.0.0",
	})
	ctx := context.Background()

	info := &updater.UpdateInfo{Available: false}
	err := u.Update(ctx, info)
	if err != nil {
		t.Errorf("Update() with unavailable update returned error: %v", err)
	}
}

func TestUpdater_Update_ContextCanceled(t *testing.T) {
	u := updater.New(updater.Config{
		CurrentVersion: "v1.0.0",
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	info := &updater.UpdateInfo{Available: true, Version: "v2.0.0"}
	err := u.Update(ctx, info)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Update() error = %v, want context.Canceled", err)
	}
}
