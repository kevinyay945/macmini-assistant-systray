// Package updater provides self-update functionality.
package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/inconshreveable/go-update"
)

// BinaryDownloadTimeout is the timeout for downloading binaries (5 minutes).
const BinaryDownloadTimeout = 5 * time.Minute

// Updater handles application self-updates.
type Updater struct {
	currentVersion *semver.Version
	rawVersion     string
	repoOwner      string
	repoName       string
	httpClient     *http.Client
}

// Config holds updater configuration.
type Config struct {
	CurrentVersion string
	RepoOwner      string
	RepoName       string
}

// New creates a new updater instance.
func New(cfg Config) *Updater {
	rawVersion := cfg.CurrentVersion
	// Normalize version for semver parsing
	normalized := normalizeVersion(rawVersion)
	version, _ := semver.NewVersion(normalized)

	return &Updater{
		currentVersion: version,
		rawVersion:     rawVersion,
		repoOwner:      cfg.RepoOwner,
		repoName:       cfg.RepoName,
		httpClient:     &http.Client{Timeout: BinaryDownloadTimeout},
	}
}

// UpdateInfo contains information about an available update.
type UpdateInfo struct {
	Available   bool
	Version     string
	ReleaseURL  string
	DownloadURL string
	Changelog   string
}

// CurrentVersion returns the currently running version.
func (u *Updater) CurrentVersion() string {
	return u.rawVersion
}

// IsNewerVersion compares two semantic versions and returns true if newVersion is newer.
// Versions are expected in the format "v1.2.3" or "1.2.3".
func (u *Updater) IsNewerVersion(newVersion string) bool {
	if u.currentVersion == nil {
		// If current version is invalid (e.g., "dev"), treat any valid version as newer
		normalized := normalizeVersion(newVersion)
		_, err := semver.NewVersion(normalized)
		return err == nil
	}

	normalized := normalizeVersion(newVersion)
	newVer, err := semver.NewVersion(normalized)
	if err != nil {
		return false
	}

	return newVer.GreaterThan(u.currentVersion)
}

// normalizeVersion ensures the version is suitable for semver parsing.
func normalizeVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	// Handle empty or "dev" versions
	if v == "" || v == "dev" || v == "none" {
		return "0.0.0"
	}
	return v
}

// CheckForUpdate checks if a new version is available.
func (u *Updater) CheckForUpdate(ctx context.Context) (*UpdateInfo, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Create a checker for the repository
	repo := fmt.Sprintf("%s/%s", u.repoOwner, u.repoName)
	if repo == "/" {
		// No repo configured, return no update available
		return &UpdateInfo{Available: false}, nil
	}

	checker := NewChecker(repo, u.rawVersion)

	release, err := checker.GetLatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("check latest release: %w", err)
	}

	available := u.IsNewerVersion(release.TagName)
	info := &UpdateInfo{
		Available:  available,
		Version:    release.TagName,
		ReleaseURL: release.HTMLURL,
		Changelog:  release.Body,
	}

	// Find download URL for current platform
	if available {
		assetName := u.getAssetName(release.TagName)
		for _, asset := range release.Assets {
			if asset.Name == assetName {
				info.DownloadURL = asset.DownloadURL
				break
			}
		}
	}

	return info, nil
}

// getAssetName returns the expected asset name for the current platform.
func (u *Updater) getAssetName(version string) string {
	// Remove 'v' prefix if present for asset naming consistency
	ver := strings.TrimPrefix(version, "v")
	return fmt.Sprintf("orchestrator_%s_%s_%s.tar.gz", ver, runtime.GOOS, runtime.GOARCH)
}

// Update downloads and applies the latest update.
func (u *Updater) Update(ctx context.Context, info *UpdateInfo) error {
	if info == nil || !info.Available {
		return nil
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if info.DownloadURL == "" {
		return fmt.Errorf("no download URL for platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return u.applyUpdate(ctx, info.DownloadURL)
}

// ApplyFromRelease downloads and applies an update from a Release.
func (u *Updater) ApplyFromRelease(ctx context.Context, release *Release) error {
	if release == nil {
		return fmt.Errorf("release is nil")
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Find the correct asset for this platform
	assetName := u.getAssetName(release.TagName)
	checksumAssetName := "checksums.txt"

	var targetAsset *Asset
	var checksumAsset *Asset

	for i := range release.Assets {
		asset := &release.Assets[i]
		if asset.Name == assetName {
			targetAsset = asset
		}
		if asset.Name == checksumAssetName {
			checksumAsset = asset
		}
	}

	if targetAsset == nil {
		return fmt.Errorf("no asset found for platform %s/%s (looking for %s)", runtime.GOOS, runtime.GOARCH, assetName)
	}

	// Download and verify checksum if available
	if checksumAsset != nil {
		return u.applyUpdateWithChecksum(ctx, targetAsset, checksumAsset, assetName)
	}

	// Apply without checksum verification
	return u.applyUpdate(ctx, targetAsset.DownloadURL)
}

// applyUpdateWithChecksum downloads and applies an update with checksum verification.
func (u *Updater) applyUpdateWithChecksum(ctx context.Context, targetAsset, checksumAsset *Asset, assetName string) error {
	// Download checksums
	expectedChecksum, err := u.fetchChecksum(ctx, checksumAsset.DownloadURL, assetName)
	if err != nil {
		return fmt.Errorf("fetch checksum: %w", err)
	}

	// Download binary
	binaryData, err := u.downloadAsset(ctx, targetAsset.DownloadURL)
	if err != nil {
		return fmt.Errorf("download binary: %w", err)
	}

	// Verify checksum
	actualChecksum := sha256.Sum256(binaryData)
	actualHex := hex.EncodeToString(actualChecksum[:])
	if actualHex != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualHex)
	}

	// Extract binary from tarball
	binaryReader, err := extractBinaryFromTarGz(binaryData)
	if err != nil {
		return fmt.Errorf("extract binary: %w", err)
	}

	// Apply update
	err = update.Apply(binaryReader, update.Options{})
	if err != nil {
		// Attempt rollback
		if rerr := update.RollbackError(err); rerr != nil {
			return fmt.Errorf("apply update failed: %w, rollback also failed: %v", err, rerr)
		}
		return fmt.Errorf("apply update failed (rolled back): %w", err)
	}

	return nil
}

// applyUpdate downloads and applies an update without checksum verification.
func (u *Updater) applyUpdate(ctx context.Context, downloadURL string) error {
	// Download binary
	binaryData, err := u.downloadAsset(ctx, downloadURL)
	if err != nil {
		return fmt.Errorf("download binary: %w", err)
	}

	// Extract binary from tarball
	binaryReader, err := extractBinaryFromTarGz(binaryData)
	if err != nil {
		return fmt.Errorf("extract binary: %w", err)
	}

	// Apply update
	err = update.Apply(binaryReader, update.Options{})
	if err != nil {
		// Attempt rollback
		if rerr := update.RollbackError(err); rerr != nil {
			return fmt.Errorf("apply update failed: %w, rollback also failed: %v", err, rerr)
		}
		return fmt.Errorf("apply update failed (rolled back): %w", err)
	}

	return nil
}

// downloadAsset downloads an asset and returns its content.
func (u *Updater) downloadAsset(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// fetchChecksum fetches and parses the checksum file.
func (u *Updater) fetchChecksum(ctx context.Context, checksumURL, filename string) (string, error) {
	data, err := u.downloadAsset(ctx, checksumURL)
	if err != nil {
		return "", err
	}

	// Parse checksums.txt format:
	// sha256sum  filename
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == filename {
			return parts[0], nil
		}
	}

	return "", fmt.Errorf("checksum not found for %s", filename)
}

// extractBinaryFromTarGz extracts the binary from a tar.gz archive.
func extractBinaryFromTarGz(data []byte) (io.Reader, error) {
	// Create a reader from the data
	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	// Look for the executable file (typically named "orchestrator" or similar)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar: %w", err)
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Look for executable files (non-zero mode with execute bit)
		// or files that look like our binary
		if header.Typeflag == tar.TypeReg {
			name := header.Name
			// Check for common binary names
			if strings.Contains(name, "orchestrator") ||
				(header.Mode&0o111 != 0 && !strings.HasSuffix(name, ".txt") && !strings.HasSuffix(name, ".md")) {
				// Read the entire file content
				content, err := io.ReadAll(tarReader)
				if err != nil {
					return nil, fmt.Errorf("read binary: %w", err)
				}
				return bytes.NewReader(content), nil
			}
		}
	}

	return nil, fmt.Errorf("no executable found in archive")
}

// Restart spawns a new process and exits the current one.
// This should be called after a successful update.
func (u *Updater) Restart() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	// Start new process with the same arguments
	cmd := exec.Command(execPath, os.Args[1:]...) // #nosec G204 - Using os.Executable() which returns the current process path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start new process: %w", err)
	}

	// Exit current process - the new process will continue running
	os.Exit(0)
	return nil // unreachable, but required for compilation
}
