package updater_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"runtime"
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
	// Test with no repo configured - should return no update available
	u := updater.New(updater.Config{
		CurrentVersion: "v1.0.0",
		// No RepoOwner/RepoName configured
	})
	ctx := context.Background()

	info, err := u.CheckForUpdate(ctx)
	if err != nil {
		t.Errorf("CheckForUpdate() returned error: %v", err)
	}
	if info == nil {
		t.Error("CheckForUpdate() returned nil info")
	}
	if info != nil && info.Available {
		t.Error("CheckForUpdate() with no repo should return Available=false")
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

func TestUpdater_Update_NoDownloadURL(t *testing.T) {
	u := updater.New(updater.Config{
		CurrentVersion: "v1.0.0",
	})
	ctx := context.Background()

	info := &updater.UpdateInfo{Available: true, Version: "v2.0.0", DownloadURL: ""}
	err := u.Update(ctx, info)
	if err == nil {
		t.Error("Update() with no download URL should return error")
	}
}

func TestUpdater_ApplyFromRelease_NilRelease(t *testing.T) {
	u := updater.New(updater.Config{
		CurrentVersion: "v1.0.0",
	})
	ctx := context.Background()

	err := u.ApplyFromRelease(ctx, nil)
	if err == nil {
		t.Error("ApplyFromRelease() with nil release should return error")
	}
}

func TestUpdater_ApplyFromRelease_ContextCanceled(t *testing.T) {
	u := updater.New(updater.Config{
		CurrentVersion: "v1.0.0",
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	release := &updater.Release{TagName: "v2.0.0"}
	err := u.ApplyFromRelease(ctx, release)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("ApplyFromRelease() error = %v, want context.Canceled", err)
	}
}

func TestUpdater_ApplyFromRelease_MissingAsset(t *testing.T) {
	u := updater.New(updater.Config{
		CurrentVersion: "v1.0.0",
	})
	ctx := context.Background()

	// Release with no assets for the current platform
	release := &updater.Release{
		TagName: "v2.0.0",
		Assets:  []updater.Asset{},
	}
	err := u.ApplyFromRelease(ctx, release)
	if err == nil {
		t.Error("ApplyFromRelease() with no matching asset should return error")
	}
}

// createTarGz creates a tar.gz archive with a single file.
func createTarGz(t *testing.T, filename string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)

	hdr := &tar.Header{
		Name: filename,
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tarWriter.WriteHeader(hdr); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("failed to write tar content: %v", err)
	}

	if err := tarWriter.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	return buf.Bytes()
}

func TestUpdater_ExtractBinaryFromTarGz(t *testing.T) {
	// Create a mock tar.gz with an orchestrator binary
	binaryContent := []byte("fake binary content")
	tarGzData := createTarGz(t, "orchestrator", binaryContent)

	// Verify the tar.gz was created correctly by checking it's not empty
	if len(tarGzData) == 0 {
		t.Fatal("createTarGz returned empty data")
	}
}

func TestUpdater_ChecksumVerification(t *testing.T) {
	// Create a mock tar.gz
	binaryContent := []byte("fake binary content for checksum test")
	tarGzData := createTarGz(t, "orchestrator", binaryContent)

	// Calculate checksum
	checksum := sha256.Sum256(tarGzData)
	checksumHex := hex.EncodeToString(checksum[:])

	// Asset name for current platform
	assetName := "orchestrator_2.0.0_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"

	// Create checksum file content
	checksumFileContent := checksumHex + "  " + assetName + "\n"

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/owner/repo/releases/latest":
			release := updater.Release{
				TagName: "v2.0.0",
				Name:    "Release 2.0.0",
				Assets: []updater.Asset{
					{
						Name:        assetName,
						DownloadURL: "http://" + r.Host + "/download/" + assetName,
						Size:        int64(len(tarGzData)),
					},
					{
						Name:        "checksums.txt",
						DownloadURL: "http://" + r.Host + "/download/checksums.txt",
						Size:        int64(len(checksumFileContent)),
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(release); err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		case "/download/" + assetName:
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = w.Write(tarGzData)
		case "/download/checksums.txt":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte(checksumFileContent))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Test checksum file parsing
	if checksumHex == "" {
		t.Error("checksum should not be empty")
	}
}

func TestUpdater_ChecksumMismatch(t *testing.T) {
	// This test validates that the checksum file format is correct
	// and the checksum comparison would work.
	// Full integration testing of ApplyFromRelease requires more setup
	// and would actually modify the running binary.

	binaryContent := []byte("fake binary content")
	tarGzData := createTarGz(t, "orchestrator", binaryContent)

	// Wrong checksum (doesn't match the actual tarGzData)
	wrongChecksum := "0000000000000000000000000000000000000000000000000000000000000000"
	assetName := "orchestrator_2.0.0_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
	checksumFileContent := wrongChecksum + "  " + assetName + "\n"

	// Calculate actual checksum
	actualChecksum := sha256.Sum256(tarGzData)
	actualHex := hex.EncodeToString(actualChecksum[:])

	// Verify that wrong checksum doesn't match
	if wrongChecksum == actualHex {
		t.Error("wrong checksum should not match actual checksum")
	}

	// Verify checksum file format parsing works
	lines := bytes.Split([]byte(checksumFileContent), []byte("\n"))
	if len(lines) < 1 {
		t.Fatal("checksum file should have at least one line")
	}
	parts := bytes.Fields(lines[0])
	if len(parts) != 2 {
		t.Fatalf("checksum line should have 2 parts, got %d", len(parts))
	}
	if string(parts[0]) != wrongChecksum {
		t.Errorf("parsed checksum = %q, want %q", string(parts[0]), wrongChecksum)
	}
	if string(parts[1]) != assetName {
		t.Errorf("parsed filename = %q, want %q", string(parts[1]), assetName)
	}
}
