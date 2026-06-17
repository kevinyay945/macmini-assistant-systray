//go:build integration

package updater_test

import (
	"context"
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/updater"
)

func TestUpdater_CheckForUpdate_Integration(t *testing.T) {
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
