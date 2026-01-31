---
name: go-update
description: Expert guidance for implementing secure self-updating Go programs using the go-update library, including basic updates, binary patching, checksum verification, and cryptographic signature validation.
---

# go-update - Self-Updating Go Programs

This skill provides comprehensive guidance for implementing secure, self-updating Go programs using the `inconshreveable/go-update` library.

## Installation

```bash
go get github.com/inconshreveable/go-update
```

## Core Concepts

The go-update library enables Go programs to update themselves by replacing their executable file with a new version. It supports multiple update methods with varying security levels.

### Key Features
- Cross-platform support (including Windows)
- Simple HTTP-based updates
- Binary patch application (bsdiff)
- SHA256 checksum verification
- Cryptographic signature verification (ECDSA)
- Automatic rollback on failure
- Custom file target updates

## Basic Update Pattern

### Simple HTTP Update

The most basic pattern downloads a complete binary from a URL:

```go
import (
    "fmt"
    "net/http"
    
    "github.com/inconshreveable/go-update"
)

func doUpdate(url string) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    err = update.Apply(resp.Body, update.Options{})
    if err != nil {
        if rerr := update.RollbackError(err); rerr != nil {
            fmt.Printf("Failed to rollback from bad update: %v\n", rerr)
        }
    }
    return err
}
```

**Key points:**
- Use `update.Apply()` with an io.Reader (typically HTTP response body)
- Always handle rollback errors with `update.RollbackError(err)`
- Empty `update.Options{}` uses defaults (SHA256, no verification)

## Binary Patch Updates

For large binaries, use binary patches to reduce download size:

```go
import (
    "io"
    "github.com/inconshreveable/go-update"
)

func updateWithPatch(patch io.Reader) error {
    err := update.Apply(patch, update.Options{
        Patcher: update.NewBSDiffPatcher(),
    })
    if err != nil {
        // Handle error
    }
    return err
}
```

**Benefits:**
- Significantly smaller download size
- Same security features available
- Custom patcher interface support

## Security: Checksum Verification

Always verify checksums to ensure update integrity:

```go
import (
    "crypto"
    _ "crypto/sha256"
    "encoding/hex"
    "io"
    
    "github.com/inconshreveable/go-update"
)

func updateWithChecksum(binary io.Reader, hexChecksum string) error {
    checksum, err := hex.DecodeString(hexChecksum)
    if err != nil {
        return err
    }
    
    err = update.Apply(binary, update.Options{
        Hash:     crypto.SHA256,  // Default, can omit
        Checksum: checksum,
    })
    if err != nil {
        // Handle error
    }
    return err
}
```

**Important:**
- SHA256 is the default hash algorithm
- Checksum must be retrieved via a secure channel
- Without signature verification, checksums alone don't prevent MITM attacks

## Security: Cryptographic Signature Verification

For production systems, always use signature verification:

```go
import (
    "crypto"
    _ "crypto/sha256"
    "encoding/hex"
    "io"
    
    "github.com/inconshreveable/go-update"
)

var publicKey = []byte(`
-----BEGIN PUBLIC KEY-----
MFYwEAYHKoZIzj0CAQYFK4EEAAoDQgAEtrVmBxQvheRArXjg2vG1xIprWGuCyESx
MMY8pjmjepSy2kuz+nl9aFLqmr+rDNdYvEBqQaZrYMc6k29gjvoQnQ==
-----END PUBLIC KEY-----
`)

func verifiedUpdate(binary io.Reader, hexChecksum, hexSignature string) error {
    checksum, err := hex.DecodeString(hexChecksum)
    if err != nil {
        return err
    }
    signature, err := hex.DecodeString(hexSignature)
    if err != nil {
        return err
    }
    
    opts := update.Options{
        Checksum:  checksum,
        Signature: signature,
        Hash:      crypto.SHA256,            // Default
        Verifier:  update.NewECDSAVerifier(), // Default
    }
    
    err = opts.SetPublicKeyPEM(publicKey)
    if err != nil {
        return err
    }
    
    err = update.Apply(binary, opts)
    if err != nil {
        // Handle error
    }
    return err
}
```

**Key steps:**
1. Embed public key in application at build time
2. Keep private key secure for signing releases
3. Distribute signature with each release
4. Verify signature before applying update

## Complete Self-Update Implementation

A production-ready pattern with version checking:

```go
import (
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"
    "time"
    
    "github.com/Masterminds/semver/v3"
    "github.com/inconshreveable/go-update"
    log "github.com/sirupsen/logrus"
)

var (
    Version        string
    semVersion     *semver.Version
    releaseChannel = "stable"
)

func selfUpdate() {
    var err error
    
    semVersion, err = semver.NewVersion(Version)
    if err != nil {
        log.Errorf("Invalid version: %v", err)
        return
    }
    
    log.Infof("Current version: %s", Version)
    
    // Check immediately
    checkForUpdate()
    
    // Check periodically
    go func() {
        for range time.Tick(5 * time.Minute) {
            checkForUpdate()
        }
    }()
}

func checkForUpdate() {
    latestVersion, err := getLatestVersion()
    if err != nil {
        log.Errorf("Failed to get latest version: %v", err)
        return
    }
    
    semLatestVersion, err := semver.NewVersion(latestVersion)
    if err != nil {
        log.Errorf("Invalid latest version: %v", err)
        return
    }
    
    if semVersion.Compare(semLatestVersion) >= 0 {
        return // Already up to date
    }
    
    log.Infof("Updating from %s to %s", Version, latestVersion)
    
    if err := downloadAndApply(latestVersion); err != nil {
        log.Errorf("Update failed: %v", err)
        return
    }
    
    log.Infof("Updated successfully, restarting...")
    os.Exit(0) // Process manager should restart
}

func getLatestVersion() (string, error) {
    url := fmt.Sprintf("https://releases.example.com/%s/VERSION", releaseChannel)
    
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    
    if resp.StatusCode != 200 {
        return "", fmt.Errorf("status code %d", resp.StatusCode)
    }
    
    return strings.TrimRight(string(body), "\n"), nil
}

func downloadAndApply(version string) error {
    url := fmt.Sprintf("https://releases.example.com/%s/myapp.exe", version)
    
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("status code %d: %s", resp.StatusCode, string(body))
    }
    
    // Apply update with automatic rollback on error
    if err := update.Apply(resp.Body, update.Options{}); err != nil {
        return fmt.Errorf("apply failed: %v", err)
    }
    
    return nil
}
```

**Pattern highlights:**
- Semantic versioning for comparison
- Periodic update checks in background goroutine
- Release channel support (stable/beta/etc)
- Graceful error handling
- Restart on successful update (requires process manager)

## Update Options Reference

```go
type Options struct {
    // Target file path (default: current executable)
    TargetPath string
    
    // Target file mode (default: current mode)
    TargetMode os.FileMode
    
    // Checksum to verify (optional)
    Checksum []byte
    
    // Hash algorithm for checksum (default: SHA256)
    Hash crypto.Hash
    
    // Signature to verify (optional)
    Signature []byte
    
    // Signature verifier (default: ECDSA)
    Verifier Verifier
    
    // Public key for signature verification
    PublicKey crypto.PublicKey
    
    // Binary patcher (optional, for patch-based updates)
    Patcher Patcher
    
    // Old save path (for custom rollback location)
    OldSavePath string
}
```

## Best Practices

### 1. Security First
- **Always use signature verification** in production
- Embed public key at build time
- Distribute checksums and signatures via secure channels
- Consider using [equinox.io](https://equinox.io) for managed update infrastructure

### 2. Error Handling
```go
if err := update.Apply(binary, opts); err != nil {
    // Check for rollback errors
    if rerr := update.RollbackError(err); rerr != nil {
        log.Fatalf("Rollback failed: %v", rerr)
    }
    return fmt.Errorf("update failed: %w", err)
}
```

### 3. Version Management
- Use semantic versioning (semver)
- Compare versions before downloading
- Support release channels (stable/beta/nightly)
- Store version in build-time variable: `go build -ldflags "-X main.Version=1.2.3"`

### 4. Single Binary Requirement
- go-update only works for single-file executables
- Embed static assets with tools like [go-bindata](https://github.com/jteeuwen/go-bindata)
- Cannot update multi-file applications

### 5. Process Management
- Call `os.Exit(0)` after successful update
- Require process manager (systemd, supervisor, etc.) to restart
- Verify new version on restart
- Consider graceful shutdown before exit

### 6. Testing Updates
```go
func testUpdate() error {
    // Use TargetPath to test without replacing current binary
    opts := update.Options{
        TargetPath: "/tmp/myapp-test",
    }
    return update.Apply(binary, opts)
}
```

### 7. Network Considerations
- Use HTTPS for update downloads
- Implement timeout and retry logic
- Support resumable downloads for large binaries
- Consider binary patches for bandwidth efficiency

## Common Patterns

### Pattern 1: Update on Startup
```go
func main() {
    if shouldCheckUpdate() {
        if err := checkAndApplyUpdate(); err != nil {
            log.Warnf("Update check failed: %v", err)
        }
    }
    
    // Continue normal startup
    run()
}
```

### Pattern 2: Background Update Checks
```go
func startUpdateChecker() {
    go func() {
        ticker := time.NewTicker(1 * time.Hour)
        defer ticker.Stop()
        
        for range ticker.C {
            if err := checkAndApplyUpdate(); err != nil {
                log.Warnf("Update failed: %v", err)
            }
        }
    }()
}
```

### Pattern 3: User-Triggered Updates
```go
func handleUpdateCommand() error {
    fmt.Println("Checking for updates...")
    
    hasUpdate, version, err := checkUpdateAvailable()
    if err != nil {
        return err
    }
    
    if !hasUpdate {
        fmt.Println("Already up to date")
        return nil
    }
    
    fmt.Printf("Update available: %s. Updating...\n", version)
    return downloadAndApply(version)
}
```

### Pattern 4: Conditional Updates (Release Channels)
```go
func getUpdateChannel() string {
    if os.Getenv("BETA_UPDATES") == "true" {
        return "beta"
    }
    return "stable"
}

func getLatestVersion() (string, error) {
    channel := getUpdateChannel()
    url := fmt.Sprintf("https://releases.example.com/%s/VERSION", channel)
    // ... fetch version
}
```

## Troubleshooting

### Permission Errors
- Update requires write permission to executable
- On Unix: ensure executable not in use by other processes
- On Windows: may require elevation

### Rollback Failures
- Old binary saved to temporary location before update
- If rollback fails, manual recovery may be needed
- Always log rollback errors for debugging

### Signature Verification Issues
- Ensure public key format is correct (PEM)
- Verify signature was created with matching private key
- Check hash algorithm matches between signing and verification

## Additional Resources

For complete reference documentation and implementation details, see:
- [./references/README.md](./references/README.md) - Complete go-update library documentation
- [./references/doc.go](./references/doc.go) - Package documentation with detailed examples

## Quick Reference

```go
// Basic update
update.Apply(httpResponse.Body, update.Options{})

// With checksum
update.Apply(binary, update.Options{
    Checksum: checksumBytes,
})

// With signature (recommended)
opts := update.Options{
    Checksum:  checksumBytes,
    Signature: signatureBytes,
}
opts.SetPublicKeyPEM(publicKeyPEM)
update.Apply(binary, opts)

// With binary patch
update.Apply(patchFile, update.Options{
    Patcher: update.NewBSDiffPatcher(),
})

// Error handling
if err := update.Apply(binary, opts); err != nil {
    if rerr := update.RollbackError(err); rerr != nil {
        log.Fatal("Rollback failed:", rerr)
    }
    return err
}
```
