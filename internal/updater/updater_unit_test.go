package updater_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/updater"
)

const mockReleasesPath = "/repos/test/repo/releases/latest"

func TestUpdater_CheckForUpdate_NoUpdateAvailable(t *testing.T) {
	// Mock server returning older version
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mockReleasesPath {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"tag_name": "v0.9.0",
				"name": "Release 0.9.0",
				"body": "Old release",
				"html_url": "https://github.com/test/repo/releases/tag/v0.9.0",
				"prerelease": false,
				"assets": []
			}`)
		}
	}))
	defer server.Close()

	// For now, this test documents the expected behavior
	// Unit tests with HTTP mocking require URL override capability
	t.Skip("Unit test requires URL override capability - covered by integration tests")
}

func TestUpdater_CheckForUpdate_UpdateAvailable(t *testing.T) {
	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	assetName := fmt.Sprintf("app_%s.tar.gz", platform)

	// Mock server returning newer version
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mockReleasesPath {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"tag_name": "v2.0.0",
				"name": "Release 2.0.0",
				"body": "New features",
				"html_url": "https://github.com/test/repo/releases/tag/v2.0.0",
				"prerelease": false,
				"assets": [
					{
						"name": "%s",
						"browser_download_url": "https://github.com/test/repo/releases/download/v2.0.0/%s"
					}
				]
			}`, assetName, assetName)
		}
	}))
	defer server.Close()

	t.Skip("Unit test requires URL override capability - covered by integration tests")
}

func TestUpdater_CheckForUpdate_PreRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == mockReleasesPath {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"tag_name": "v2.0.0-beta.1",
				"name": "Release 2.0.0 Beta",
				"body": "Beta release",
				"html_url": "https://github.com/test/repo/releases/tag/v2.0.0-beta.1",
				"prerelease": true,
				"assets": []
			}`)
		}
	}))
	defer server.Close()

	t.Skip("Unit test requires URL override capability - covered by integration tests")
}

func TestUpdater_IsNewerVersion_EdgeCases(t *testing.T) {
	testCases := []struct {
		name           string
		currentVersion string
		newVersion     string
		wantNewer      bool
	}{
		// Note: semver considers v1.0.1-beta.1 > v1.0.0 (higher patch version, even with prerelease)
		{"prerelease with higher patch", "v1.0.0", "v1.0.1-beta.1", true},
		{"prerelease vs stable same base", "v1.0.0", "v1.0.0-beta.1", false},
		{"stable vs prerelease", "v1.0.0-beta.1", "v1.0.0", true},
		{"empty new version", "v1.0.0", "", false},
		{"invalid new version", "v1.0.0", "invalid", false},
		{"build metadata", "v1.0.0", "v1.0.0+build.1", false},
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

func TestUpdater_Update_NoUpdate(t *testing.T) {
	u := updater.New(updater.Config{
		CurrentVersion: "v1.0.0",
	})
	ctx := context.Background()

	// Test with nil info
	if err := u.Update(ctx, nil); err != nil {
		t.Errorf("Update(nil) should not return error, got: %v", err)
	}

	// Test with unavailable update
	info := &updater.UpdateInfo{Available: false}
	if err := u.Update(ctx, info); err != nil {
		t.Errorf("Update(unavailable) should not return error, got: %v", err)
	}
}
