// Package updater provides self-update functionality.
package updater

import (
	"context"
	"strings"
)

// Updater handles application self-updates.
type Updater struct {
	currentVersion string
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
	return &Updater{
		currentVersion: cfg.CurrentVersion,
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
	return u.currentVersion
}

// IsNewerVersion compares two semantic versions and returns true if newVersion is newer.
// Versions are expected in the format "v1.2.3" or "1.2.3".
func (u *Updater) IsNewerVersion(newVersion string) bool {
	current := normalizeVersion(u.currentVersion)
	new := normalizeVersion(newVersion)

	// Simple semver comparison (major.minor.patch)
	currentParts := parseVersion(current)
	newParts := parseVersion(new)

	for i := 0; i < 3; i++ {
		if newParts[i] > currentParts[i] {
			return true
		}
		if newParts[i] < currentParts[i] {
			return false
		}
	}
	return false
}

// normalizeVersion ensures the version has a "v" prefix stripped for comparison.
func normalizeVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}

// parseVersion splits a version string into [major, minor, patch] integers.
// Returns [0, 0, 0] for invalid versions.
func parseVersion(v string) [3]int {
	parts := strings.Split(v, ".")
	var result [3]int
	for i := 0; i < 3 && i < len(parts); i++ {
		// Parse each part, ignoring any suffix (e.g., "1-beta" -> 1)
		numStr := parts[i]
		if idx := strings.IndexAny(numStr, "-+"); idx != -1 {
			numStr = numStr[:idx]
		}
		var num int
		for _, c := range numStr {
			if c >= '0' && c <= '9' {
				num = num*10 + int(c-'0')
			} else {
				break
			}
		}
		result[i] = num
	}
	return result
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
