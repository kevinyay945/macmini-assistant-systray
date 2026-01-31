// Package updater provides self-update functionality.
package updater

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/inconshreveable/go-update"
)

// Updater handles application self-updates.
type Updater struct {
	currentVersion *semver.Version
	rawVersion     string
	repoOwner      string
	repoName       string
	httpClient     *http.Client
}

// GitHubRelease represents a GitHub release response.
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	HTMLURL    string `json:"html_url"`
	Assets     []GitHubAsset `json:"assets"`
	Prerelease bool   `json:"prerelease"`
}

// GitHubAsset represents a GitHub release asset.
type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Config holds updater configuration.
type Config struct {
	CurrentVersion string
	RepoOwner      string
	RepoName       string
	HTTPClient     *http.Client // Optional, defaults to http.DefaultClient
}

// New creates a new updater instance.
func New(cfg Config) *Updater {
	rawVersion := cfg.CurrentVersion
	// Normalize version for semver parsing
	normalized := normalizeVersion(rawVersion)
	version, _ := semver.NewVersion(normalized)

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &Updater{
		currentVersion: version,
		rawVersion:     rawVersion,
		repoOwner:      cfg.RepoOwner,
		repoName:       cfg.RepoName,
		httpClient:     httpClient,
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

	// Fetch latest release from GitHub API
	release, err := u.fetchLatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	// Skip pre-releases
	if release.Prerelease {
		return &UpdateInfo{Available: false}, nil
	}

	// Compare with current version
	if !u.IsNewerVersion(release.TagName) {
		return &UpdateInfo{Available: false}, nil
	}

	// Find the appropriate asset for this platform
	assetURL, err := u.findPlatformAsset(release)
	if err != nil {
		return nil, fmt.Errorf("failed to find platform asset: %w", err)
	}

	return &UpdateInfo{
		Available:   true,
		Version:     release.TagName,
		ReleaseURL:  release.HTMLURL,
		DownloadURL: assetURL,
		Changelog:   release.Body,
	}, nil
}

// fetchLatestRelease fetches the latest release from GitHub API.
func (u *Updater) fetchLatestRelease(ctx context.Context) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.repoOwner, u.repoName)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// findPlatformAsset finds the download URL for the current platform.
func (u *Updater) findPlatformAsset(release *GitHubRelease) (string, error) {
	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	expectedSuffix := fmt.Sprintf("%s.tar.gz", platform)

	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, expectedSuffix) {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("no asset found for platform %s", platform)
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

	// Download the archive
	archivePath, err := u.downloadArchive(ctx, info.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}
	defer os.Remove(archivePath)

	// Download and verify checksums
	checksumURL := u.getChecksumURL(info.DownloadURL)
	if err := u.verifyChecksum(ctx, archivePath, checksumURL); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	// Extract binary from archive
	binaryPath, err := u.extractBinary(archivePath)
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}
	defer os.Remove(binaryPath)

	// Apply the update
	if err := u.applyUpdate(binaryPath); err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	return nil
}

// downloadArchive downloads the release archive to a temporary file.
func (u *Updater) downloadArchive(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "update-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// getChecksumURL derives the checksum file URL from the archive URL.
func (u *Updater) getChecksumURL(archiveURL string) string {
	// Assumes checksums.txt is in the same location as the archive
	baseURL := archiveURL[:strings.LastIndex(archiveURL, "/")+1]
	return baseURL + "checksums.txt"
}

// verifyChecksum verifies the SHA256 checksum of the downloaded file.
func (u *Updater) verifyChecksum(ctx context.Context, filePath, checksumURL string) error {
	// Download checksums file
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checksumURL, nil)
	if err != nil {
		return err
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("checksums file not found: %d", resp.StatusCode)
	}

	checksumData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Calculate actual checksum
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	actualChecksum := hex.EncodeToString(hash.Sum(nil))

	// Find expected checksum
	fileName := filepath.Base(filePath)
	expectedChecksum, err := u.parseChecksum(checksumData, fileName)
	if err != nil {
		return err
	}

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// parseChecksum parses the checksums file and returns the checksum for the given file.
func (u *Updater) parseChecksum(checksumData []byte, fileName string) (string, error) {
	lines := strings.Split(string(checksumData), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 && strings.Contains(line, fileName) {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("checksum not found for %s", fileName)
}

// extractBinary extracts the binary from the tar.gz archive.
func (u *Updater) extractBinary(archivePath string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Find the binary file (usually named "orchestrator" or similar)
	expectedBinaryName := "orchestrator"
	if runtime.GOOS == "windows" {
		expectedBinaryName += ".exe"
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for the binary file
		if filepath.Base(header.Name) == expectedBinaryName {
			tmpFile, err := os.CreateTemp("", "binary-*")
			if err != nil {
				return "", err
			}
			defer tmpFile.Close()

			if _, err := io.Copy(tmpFile, tr); err != nil {
				os.Remove(tmpFile.Name())
				return "", err
			}

			// Make it executable
			if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
				os.Remove(tmpFile.Name())
				return "", err
			}

			return tmpFile.Name(), nil
		}
	}

	return "", fmt.Errorf("binary %s not found in archive", expectedBinaryName)
}

// applyUpdate applies the update using go-update library.
func (u *Updater) applyUpdate(newBinaryPath string) error {
	newBinary, err := os.Open(newBinaryPath)
	if err != nil {
		return err
	}
	defer newBinary.Close()

	err = update.Apply(newBinary, update.Options{})
	if err != nil {
		// Rollback is handled automatically by go-update
		return err
	}

	return nil
}
