// Package updater provides self-update functionality.
package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Release represents a GitHub release.
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	PublishedAt string    `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents a release asset.
type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
}

// DefaultCheckInterval is the default interval for periodic update checks (6 hours).
const DefaultCheckInterval = 6 * time.Hour

// DefaultHTTPTimeout is the default timeout for HTTP requests.
const DefaultHTTPTimeout = 30 * time.Second

// Checker handles periodic update checks.
type Checker struct {
	repo           string // "owner/repo"
	currentVersion string
	checkInterval  time.Duration
	httpClient     *http.Client
}

// CheckerOption configures the Checker.
type CheckerOption func(*Checker)

// WithCheckInterval sets the check interval.
func WithCheckInterval(interval time.Duration) CheckerOption {
	return func(c *Checker) {
		c.checkInterval = interval
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) CheckerOption {
	return func(c *Checker) {
		c.httpClient = client
	}
}

// NewChecker creates a new update checker.
func NewChecker(repo, currentVersion string, opts ...CheckerOption) *Checker {
	c := &Checker{
		repo:           repo,
		currentVersion: currentVersion,
		checkInterval:  DefaultCheckInterval,
		httpClient:     &http.Client{Timeout: DefaultHTTPTimeout},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// GetLatestRelease fetches the latest release from GitHub.
func (c *Checker) GetLatestRelease(ctx context.Context) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", c.repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "macmini-assistant-systray-updater")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found for repository %s", c.repo)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}

	return &release, nil
}

// CheckInterval returns the configured check interval.
func (c *Checker) CheckInterval() time.Duration {
	return c.checkInterval
}

// CurrentVersion returns the current version string.
func (c *Checker) CurrentVersion() string {
	return c.currentVersion
}

// Repo returns the repository identifier.
func (c *Checker) Repo() string {
	return c.repo
}

// StartPeriodicCheck starts a goroutine that periodically checks for updates.
// It calls onUpdate when a newer version is available.
// The function blocks until the context is canceled.
func (c *Checker) StartPeriodicCheck(ctx context.Context, updater *Updater, onUpdate func(*Release)) {
	ticker := time.NewTicker(c.checkInterval)
	defer ticker.Stop()

	// Check immediately on start
	c.check(ctx, updater, onUpdate)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.check(ctx, updater, onUpdate)
		}
	}
}

// check performs a single update check.
func (c *Checker) check(ctx context.Context, updater *Updater, onUpdate func(*Release)) {
	release, err := c.GetLatestRelease(ctx)
	if err != nil {
		// Log error but continue - network issues are expected
		return
	}

	if updater.IsNewerVersion(release.TagName) {
		onUpdate(release)
	}
}

// ManualCheck performs an immediate update check.
// Returns the release if an update is available, nil otherwise.
func (c *Checker) ManualCheck(ctx context.Context, updater *Updater) (*Release, error) {
	release, err := c.GetLatestRelease(ctx)
	if err != nil {
		return nil, err
	}

	if updater.IsNewerVersion(release.TagName) {
		return release, nil
	}

	return nil, nil
}
