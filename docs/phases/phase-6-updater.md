# Phase 6: Auto-updater

**Duration**: Week 10
**Status**: ⚪ Not Started
**Goal**: Implement self-updating from GitHub releases

---

## Overview

This phase implements automatic updates from GitHub releases. The updater will periodically check for new versions and safely update the binary.

---

## 6.1 Update Checker

**Duration**: 2 days
**Status**: ⚪ Not Started

### Tasks

- [ ] GitHub Releases API polling
- [ ] Version comparison (SemVer)
- [ ] Periodic check (every 6 hours)
- [ ] Manual trigger from tray menu

### Implementation Details

```go
// internal/updater/checker.go
package updater

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "golang.org/x/mod/semver"
)

type Release struct {
    TagName     string  `json:"tag_name"`
    Name        string  `json:"name"`
    Body        string  `json:"body"`
    PublishedAt string  `json:"published_at"`
    Assets      []Asset `json:"assets"`
}

type Asset struct {
    Name        string `json:"name"`
    DownloadURL string `json:"browser_download_url"`
    Size        int64  `json:"size"`
}

type Checker struct {
    repo           string // "username/repo"
    currentVersion string
    checkInterval  time.Duration
    httpClient     *http.Client
    logger         *observability.Logger
}

func NewChecker(repo, currentVersion string, checkInterval time.Duration) *Checker {
    return &Checker{
        repo:           repo,
        currentVersion: currentVersion,
        checkInterval:  checkInterval,
        httpClient:     &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *Checker) GetLatestRelease(ctx context.Context) (*Release, error) {
    url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", c.repo)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Accept", "application/vnd.github.v3+json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch release: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    var release Release
    if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
        return nil, fmt.Errorf("failed to decode release: %w", err)
    }

    return &release, nil
}

func (c *Checker) IsNewerVersion(remoteVersion string) bool {
    // Ensure versions have 'v' prefix for semver comparison
    current := c.currentVersion
    remote := remoteVersion

    if !strings.HasPrefix(current, "v") {
        current = "v" + current
    }
    if !strings.HasPrefix(remote, "v") {
        remote = "v" + remote
    }

    return semver.Compare(remote, current) > 0
}

func (c *Checker) StartPeriodicCheck(ctx context.Context, onUpdate func(*Release)) {
    ticker := time.NewTicker(c.checkInterval)
    defer ticker.Stop()

    // Check immediately on start
    c.check(ctx, onUpdate)

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            c.check(ctx, onUpdate)
        }
    }
}

func (c *Checker) check(ctx context.Context, onUpdate func(*Release)) {
    release, err := c.GetLatestRelease(ctx)
    if err != nil {
        c.logger.Warn("Failed to check for updates", "error", err)
        return
    }

    if c.IsNewerVersion(release.TagName) {
        c.logger.Info("New version available", "current", c.currentVersion, "latest", release.TagName)
        onUpdate(release)
    }
}
```

### Test Cases

```go
// internal/updater/checker_test.go
func TestUpdateChecker_CheckLatestRelease(t *testing.T)
func TestUpdateChecker_VersionComparison(t *testing.T)
func TestUpdateChecker_PeriodicPolling(t *testing.T)
func TestUpdateChecker_ManualTrigger(t *testing.T)
func TestUpdateChecker_IsNewerVersion(t *testing.T)
func TestUpdateChecker_SemVerComparison(t *testing.T)
func TestUpdateChecker_NetworkError(t *testing.T)
func TestUpdateChecker_RateLimiting(t *testing.T)
```

### Acceptance Criteria

- [ ] Polls releases every 6 hours
- [ ] Detects newer versions correctly
- [ ] Manual check available
- [ ] No API rate limit issues
- [ ] Handles network errors gracefully

### Notes

<!-- Add your notes here -->

---

## 6.2 Binary Updater

**Duration**: 3 days
**Status**: ⚪ Not Started

### Tasks

- [ ] Download release binary
- [ ] Checksum verification
- [ ] Atomic binary replacement
- [ ] Graceful restart
- [ ] Rollback on failure

### Implementation Details

```go
// internal/updater/updater.go
package updater

import (
    "context"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "net/http"
    "os"
    "runtime"

    "github.com/inconshreveable/go-update"
)

type Updater struct {
    checker    *Checker
    httpClient *http.Client
    logger     *observability.Logger
}

func NewUpdater(checker *Checker) *Updater {
    return &Updater{
        checker:    checker,
        httpClient: &http.Client{Timeout: 5 * time.Minute},
    }
}

func (u *Updater) Update(ctx context.Context, release *Release) error {
    // Find the correct asset for this platform
    assetName := fmt.Sprintf("orchestrator_%s_%s_%s.tar.gz",
        release.TagName, runtime.GOOS, runtime.GOARCH)

    var targetAsset *Asset
    var checksumAsset *Asset

    for _, asset := range release.Assets {
        if asset.Name == assetName {
            targetAsset = &asset
        }
        if asset.Name == "checksums.txt" {
            checksumAsset = &asset
        }
    }

    if targetAsset == nil {
        return fmt.Errorf("no asset found for platform %s/%s", runtime.GOOS, runtime.GOARCH)
    }

    // Download checksums
    expectedChecksum, err := u.fetchChecksum(ctx, checksumAsset, assetName)
    if err != nil {
        return fmt.Errorf("failed to fetch checksum: %w", err)
    }

    // Download binary
    binary, err := u.downloadAsset(ctx, targetAsset)
    if err != nil {
        return fmt.Errorf("failed to download binary: %w", err)
    }

    // Verify checksum
    actualChecksum := sha256.Sum256(binary)
    if hex.EncodeToString(actualChecksum[:]) != expectedChecksum {
        return fmt.Errorf("checksum mismatch")
    }

    // Extract binary from tarball
    binaryReader, err := extractBinary(binary)
    if err != nil {
        return fmt.Errorf("failed to extract binary: %w", err)
    }

    // Apply update
    err = update.Apply(binaryReader, update.Options{})
    if err != nil {
        // Attempt rollback
        if rollbackErr := update.RollbackError(err); rollbackErr != nil {
            u.logger.Error("Rollback failed", "error", rollbackErr)
        }
        return fmt.Errorf("failed to apply update: %w", err)
    }

    u.logger.Info("Update applied successfully", "version", release.TagName)
    return nil
}

func (u *Updater) downloadAsset(ctx context.Context, asset *Asset) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", asset.DownloadURL, nil)
    if err != nil {
        return nil, err
    }

    resp, err := u.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    return io.ReadAll(resp.Body)
}

func (u *Updater) fetchChecksum(ctx context.Context, asset *Asset, filename string) (string, error) {
    data, err := u.downloadAsset(ctx, asset)
    if err != nil {
        return "", err
    }

    // Parse checksums.txt format:
    // sha256sum  filename
    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        parts := strings.Fields(line)
        if len(parts) == 2 && parts[1] == filename {
            return parts[0], nil
        }
    }

    return "", fmt.Errorf("checksum not found for %s", filename)
}

func (u *Updater) Restart() error {
    // Get current executable
    execPath, err := os.Executable()
    if err != nil {
        return err
    }

    // Start new process
    cmd := exec.Command(execPath, os.Args[1:]...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Start(); err != nil {
        return err
    }

    // Exit current process
    os.Exit(0)
    return nil
}
```

### Test Cases

```go
// internal/updater/updater_test.go
func TestUpdater_DownloadBinary(t *testing.T)
func TestUpdater_ChecksumVerification(t *testing.T)
func TestUpdater_BinaryReplacement(t *testing.T)
func TestUpdater_GracefulRestart(t *testing.T)
func TestUpdater_RollbackOnFailure(t *testing.T)
func TestUpdater_ExtractTarGz(t *testing.T)
func TestUpdater_ChecksumMismatch(t *testing.T)
func TestUpdater_MissingAsset(t *testing.T)
```

### Acceptance Criteria

- [ ] Binary downloads and verifies checksums
- [ ] Replacement is atomic (no partial updates)
- [ ] App restarts automatically after update
- [ ] Rolls back if new version crashes on startup
- [ ] User notified of update status

### Notes

<!-- Add your notes here -->

---

## Deliverables

By the end of Phase 6:

- [ ] Automatic update checking every 6 hours
- [ ] Manual update check from tray menu
- [ ] Safe binary replacement with rollback
- [ ] Graceful restart after update

---

## Dependencies

```go
// go.mod additions
require (
    github.com/inconshreveable/go-update v0.0.0-20160112193335-8152e7eb6ccf
    golang.org/x/mod v0.x.x
)
```

---

## Security Considerations

1. **Checksum Verification**: Always verify downloaded binary against checksum
2. **HTTPS Only**: All downloads over HTTPS
3. **Signature Verification** (optional): Consider GPG signing releases
4. **Atomic Replacement**: Use go-update for safe binary replacement

---

## Time Tracking

| Task | Estimated | Actual | Notes |
|------|-----------|--------|-------|
| 6.1 Update Checker | 2 days | | |
| 6.2 Binary Updater | 3 days | | |
| **Total** | **5 days** | | |

---

**Previous**: [Phase 5: System Tray & Auto-start](./phase-5-systray.md)
**Next**: [Phase 7: Integration & Testing](./phase-7-integration.md)
