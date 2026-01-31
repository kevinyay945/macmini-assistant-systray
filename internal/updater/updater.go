// Package updater provides self-update functionality.
package updater

import (
	"context"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Updater handles application self-updates.
type Updater struct {
	currentVersion *semver.Version
	rawVersion     string
	repoOwner      string
	repoName       string
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

	// TODO: Implement update check via GitHub releases
	// 1. Fetch latest release from GitHub API
	// 2. Compare with current version
	// 3. Return update info if newer version available
	return &UpdateInfo{Available: false}, nil
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

	// TODO: Implement self-update using github.com/inconshreveable/go-update
	// 1. Download new binary
	// 2. Verify checksum
	// 3. Apply update
	return nil
}
