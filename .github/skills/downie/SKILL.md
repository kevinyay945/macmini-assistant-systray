---
name: downie
description: Expert guidance for using Downie deep link to download videos from various platforms with Go, including API reference, best practices, and common usage patterns.
---

# Downie Video Downloader - Skill Guide

This skill provides comprehensive guidance for using Downie (macOS video downloader app) via deep links with Go programming language.

## Reference Documentation Locations

When working with the Downie downloader package, always refer to these documentation sources:

### Core Implementation
- **Downloader Package**: [./reference/downloader.go](./reference/downloader.go) - Core downloader implementation with all methods and types
- **Downloader Tests**: [./reference/downloader_test.go](./reference/downloader_test.go) - Test examples showing URL building and validation
- **HTTP Handlers**: [../../selfhost-mac-helper/internal/handlers/handlers.go](../../selfhost-mac-helper/internal/handlers/handlers.go) - HTTP API implementation using the downloader
- **Models**: [../../selfhost-mac-helper/pkg/models/models.go](../../selfhost-mac-helper/pkg/models/models.go) - Request/response models for API

## Overview

The Downie downloader package provides a Go interface to control the Downie macOS application via custom URL schemes and command execution. It supports:

- Video downloads from YouTube, Vimeo, and 1000+ other sites
- Post-processing options (MP4 conversion, audio extraction, permutation)
- User-Guided Extraction (UGE) for complex sites
- Download progress tracking and cancellation
- Concurrent download prevention

## Installation Requirements

1. **Downie Application**: Install Downie 4 (standard) or Downie (Setapp version)
   - Standard: `brew install downie --cask`
   - Purchase from: https://software.charliemonroe.net/downie/

2. **Go Package**: Import the downloader package
   ```go
   import "github.com/kevin/slefhost-mac-helper/internal/downloader"
   ```

## Core Concepts

### 1. Downloader Creation

```go
import "github.com/kevin/slefhost-mac-helper/internal/downloader"

// Create a new downloader with base path for downloads
dl := downloader.New("/path/to/downloads")
```

**Parameters:**
- `basePath`: Base directory where download folders will be created

### 2. Download Options

```go
type DownloadOptions struct {
    URL            string  // Required: Video URL to download
    PostProcessing string  // Optional: "mp4", "audio", "permute"
    UseUGE         bool    // Optional: Enable User-Guided Extraction
    Destination    string  // Optional: Custom destination (set automatically)
}
```

**PostProcessing Options:**
- `"mp4"`: Convert to MP4 format
- `"audio"`: Extract audio only
- `"permute"`: Try multiple extraction methods
- `""`: No post-processing (default)

**UGE (User-Guided Extraction):**
- Opens Downie's interactive browser for complex sites
- Useful when automatic extraction fails
- Requires user interaction

### 3. Download Result

```go
type DownloadResult struct {
    FileName string  // Name of downloaded file
    FilePath string  // Relative path: "{timestamp}/{filename}"
}
```

## Basic Usage

### Pattern 1: Simple Download

```go
package main

import (
    "fmt"
    "log"

    "github.com/kevin/slefhost-mac-helper/internal/downloader"
)

func main() {
    // Create downloader
    dl := downloader.New("/Users/username/Downloads")

    // Download a video
    result, err := dl.Download(downloader.DownloadOptions{
        URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    })
    if err != nil {
        log.Fatalf("Download failed: %v", err)
    }

    fmt.Printf("Downloaded: %s\n", result.FileName)
    fmt.Printf("Path: %s\n", result.FilePath)
}
```

### Pattern 2: Download with Post-Processing

```go
// Download and convert to MP4
result, err := dl.Download(downloader.DownloadOptions{
    URL:            "https://vimeo.com/123456789",
    PostProcessing: "mp4",
})
if err != nil {
    log.Fatalf("Download failed: %v", err)
}

// Extract audio only
audioResult, err := dl.Download(downloader.DownloadOptions{
    URL:            "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    PostProcessing: "audio",
})
```

### Pattern 3: Check Download Status

```go
// Check if download is in progress
if dl.IsDownloading() {
    fmt.Println("Download already in progress")
    return
}

// Start download
go func() {
    result, err := dl.Download(downloader.DownloadOptions{
        URL: videoURL,
    })
    // Handle result...
}()

// Later: stop if needed
if err := dl.StopDownload(); err != nil {
    fmt.Printf("Failed to stop: %v\n", err)
}
```

### Pattern 4: Concurrent Download Prevention

```go
// Attempt download
result, err := dl.Download(downloader.DownloadOptions{
    URL: url,
})

if err != nil {
    if err.Error() == "download already in progress" {
        fmt.Println("Please wait for current download to complete")
        return
    }
    log.Fatalf("Download failed: %v", err)
}
```

## HTTP API Integration

### Setup Handler

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/kevin/slefhost-mac-helper/internal/downloader"
    "github.com/kevin/slefhost-mac-helper/internal/handlers"
)

func main() {
    // Create downloader
    dl := downloader.New("/path/to/downloads")

    // Create handler
    handler := handlers.New(dl)

    // Setup routes
    r := gin.Default()
    r.GET("/api/health", handler.HealthCheck)
    r.GET("/api/download/youtube", handler.DownloadYoutube)
    r.POST("/api/download/stop", handler.StopDownload)

    r.Run(":8080")
}
```

### API Endpoints

#### Download Video
```bash
# Basic download
GET /api/download/youtube?ytPath=https://youtube.com/watch?v=VIDEO_ID

# With post-processing
GET /api/download/youtube?ytPath=URL&postProcessing=mp4

# With UGE
GET /api/download/youtube?ytPath=URL&useUGE=true

# Combined options
GET /api/download/youtube?ytPath=URL&postProcessing=audio&useUGE=false
```

**Response (Success - 200 OK):**
```json
{
  "message": "Download completed successfully",
  "url": "https://youtube.com/watch?v=VIDEO_ID",
  "fileName": "video_title.mp4",
  "files": "1234567890/video_title.mp4"
}
```

**Response (Conflict - 409):**
```json
{
  "error": "waiting for the process",
  "details": "A download is already in progress. Please wait for it to complete or stop it first."
}
```

**Response (Timeout - 408):**
```json
{
  "error": "download timed out after 5 minutes"
}
```

#### Stop Download
```bash
POST /api/download/stop
```

**Response (Success - 200 OK):**
```json
{
  "message": "Download stopped successfully"
}
```

## Advanced Features

### 1. Download Lifecycle Management

The downloader uses mutex-based locking to prevent concurrent downloads:

```go
type Downloader struct {
    basePath        string
    mu              sync.Mutex
    isDownloading   bool
    cancelFunc      context.CancelFunc
    downloadContext context.Context
}
```

**Thread-safe operations:**
- Only one download at a time
- Graceful cancellation via context
- Automatic cleanup on completion

### 2. Downie URL Scheme

The package supports two methods to launch Downie:

**Method 1: Simple Open (no options)**
```go
// Uses: open -a "Downie 4" <url>
// Falls back to: open -a "Downie" <url> (Setapp)
```

**Method 2: Custom URL Scheme (with options)**
```go
// Format: downie://XUOpenURL?url=<encoded_url>&param=value
// Example: downie://XUOpenURL?url=https%3A%2F%2Fyoutube.com&postprocessing=mp4
```

### 3. URL Encoding

Special characters in video URLs are properly encoded:

```go
func buildDownieURL(url, destination, postProcessing string, useUGE bool) string {
    // URL encoding
    encodedURL := strings.ReplaceAll(url, "?", "%3F")
    encodedURL = strings.ReplaceAll(encodedURL, "&", "%26")

    // Build custom scheme URL
    downieURL := fmt.Sprintf("downie://XUOpenURL?url=%s", encodedURL)

    // Add parameters...
    return downieURL
}
```

### 4. Download Detection

The package waits for download completion by monitoring the destination folder:

```go
// Polls every 5 seconds for up to 5 minutes
// Checks for:
// 1. Files exist in destination folder
// 2. No files contain "downiepart" extension (incomplete)
// 3. Returns when complete file is found
```

**Behavior:**
- Timeout: 5 minutes
- Poll interval: 5 seconds
- Detects incomplete downloads by `.downiepart` extension
- Returns first complete file found

## Error Handling

### Common Errors

```go
// 1. Download already in progress
if err.Error() == "download already in progress" {
    // Wait or stop current download
}

// 2. Timeout
if err.Error() == "download timed out after 5 minutes" {
    // Retry or check Downie app
}

// 3. Cancelled
if err.Error() == "download was cancelled" {
    // User or system cancelled
}

// 4. No download to stop
if err.Error() == "no download in progress" {
    // Can't stop non-existent download
}

// 5. Failed to create folder
if strings.Contains(err.Error(), "failed to create destination folder") {
    // Check permissions
}

// 6. Downie command failed
if strings.Contains(err.Error(), "failed to execute Downie command") {
    // Check if Downie is installed
}
```

### Best Practices for Error Handling

```go
func handleDownload(dl *downloader.Downloader, url string) error {
    result, err := dl.Download(downloader.DownloadOptions{
        URL: url,
    })

    if err != nil {
        switch {
        case err.Error() == "download already in progress":
            return fmt.Errorf("concurrent download prevented: %w", err)

        case err.Error() == "download timed out after 5 minutes":
            // Might need manual intervention
            return fmt.Errorf("download took too long, check Downie app: %w", err)

        case err.Error() == "download was cancelled":
            return fmt.Errorf("user cancelled: %w", err)

        case strings.Contains(err.Error(), "Unable to find application"):
            return fmt.Errorf("Downie not installed: %w", err)

        default:
            return fmt.Errorf("unexpected error: %w", err)
        }
    }

    fmt.Printf("Success: %s at %s\n", result.FileName, result.FilePath)
    return nil
}
```

## Configuration

### Directory Structure

Downloads are organized by timestamp:

```
basePath/
├── 1706745600/          # Unix timestamp
│   └── video_title.mp4
├── 1706745800/
│   └── another_video.mp4
└── 1706746000/
    └── audio_only.m4a
```

**Folder naming:**
- Format: Unix timestamp (`time.Now().Unix()`)
- Ensures unique folders for each download
- Easy to sort chronologically

### Timeouts and Intervals

```go
const (
    MaxDownloadTime = 5 * time.Minute  // Overall timeout
    PollInterval    = 5 * time.Second  // Check interval
)
```

**Customization** (if needed):
- Modify constants in `waitForDownload()` method
- Consider video size and network speed
- Balance between responsiveness and resource usage

## Testing

### Unit Tests

```go
func TestNew(t *testing.T) {
    basePath := "/tmp/downloads"
    dl := downloader.New(basePath)

    if dl == nil {
        t.Fatal("New() returned nil")
    }

    if dl.basePath != basePath {
        t.Errorf("basePath = %v, want %v", dl.basePath, basePath)
    }
}
```

### URL Building Tests

```go
func TestBuildDownieURL(t *testing.T) {
    dl := downloader.New("/tmp")

    tests := []struct {
        name           string
        url            string
        destination    string
        postProcessing string
        useUGE         bool
        wantContains   []string
    }{
        {
            name:         "basic URL",
            url:          "https://youtube.com/watch?v=123",
            destination:  "/tmp/test",
            wantContains: []string{"downie://XUOpenURL?url=", "destination=/tmp/test"},
        },
        {
            name:           "with post processing",
            url:            "https://youtube.com/watch?v=123",
            postProcessing: "mp4",
            wantContains:   []string{"postprocessing=mp4"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := dl.buildDownieURL(tt.url, tt.destination, tt.postProcessing, tt.useUGE)
            // Assert contains...
        })
    }
}
```

## Common Use Cases

### Use Case 1: Batch Download Service

```go
type BatchDownloader struct {
    dl    *downloader.Downloader
    queue chan string
}

func (b *BatchDownloader) Start() {
    for url := range b.queue {
        result, err := b.dl.Download(downloader.DownloadOptions{
            URL: url,
        })
        if err != nil {
            log.Printf("Failed to download %s: %v", url, err)
            continue
        }
        log.Printf("Downloaded: %s", result.FileName)
    }
}

func (b *BatchDownloader) Add(url string) {
    b.queue <- url
}
```

### Use Case 2: API with Progress Tracking

```go
type DownloadStatus struct {
    URL        string
    Status     string // "pending", "downloading", "completed", "failed"
    Result     *downloader.DownloadResult
    Error      error
    StartedAt  time.Time
    FinishedAt time.Time
}

func trackDownload(dl *downloader.Downloader, url string) *DownloadStatus {
    status := &DownloadStatus{
        URL:       url,
        Status:    "pending",
        StartedAt: time.Now(),
    }

    go func() {
        status.Status = "downloading"
        result, err := dl.Download(downloader.DownloadOptions{URL: url})

        status.FinishedAt = time.Now()
        if err != nil {
            status.Status = "failed"
            status.Error = err
        } else {
            status.Status = "completed"
            status.Result = result
        }
    }()

    return status
}
```

### Use Case 3: Webhook Notifications

```go
func downloadWithNotification(dl *downloader.Downloader, url, webhookURL string) error {
    result, err := dl.Download(downloader.DownloadOptions{URL: url})

    notification := map[string]interface{}{
        "url":    url,
        "status": "completed",
    }

    if err != nil {
        notification["status"] = "failed"
        notification["error"] = err.Error()
    } else {
        notification["fileName"] = result.FileName
        notification["filePath"] = result.FilePath
    }

    // Send webhook
    sendWebhook(webhookURL, notification)

    return err
}
```

## Supported Platforms

Downie supports 1000+ websites including:

- **Video Platforms**: YouTube, Vimeo, Dailymotion, Facebook, Instagram
- **Streaming**: Twitch, TikTok, Twitter/X
- **Educational**: Coursera, Udemy, Khan Academy
- **News**: CNN, BBC, Reuters
- **And many more...**

For full list, check: https://software.charliemonroe.net/downie/

## Troubleshooting

### Issue 1: "Unable to find application"

**Cause**: Downie not installed or different version installed

**Solution**:
```go
// The code automatically tries both:
// 1. "Downie 4" (standard purchase)
// 2. "Downie" (Setapp version)

// Verify installation:
// $ ls /Applications | grep -i downie
```

### Issue 2: Download timeout

**Cause**: Large file, slow network, or Downie stuck

**Solutions**:
1. Check Downie app manually
2. Increase timeout in `waitForDownload()`
3. Use `StopDownload()` and retry

### Issue 3: Download not starting

**Cause**: Downie URL scheme not recognized

**Solutions**:
1. Ensure Downie is installed
2. Run Downie at least once to register URL scheme
3. Check system permissions

### Issue 4: Concurrent download errors

**Cause**: Multiple downloads attempted simultaneously

**Solution**:
```go
// Always check before starting
if dl.IsDownloading() {
    return errors.New("please wait for current download")
}

// Or use a queue system
```

## Best Practices

### 1. Resource Management

```go
// Always use a single Downloader instance per basePath
var globalDownloader = downloader.New("/downloads")

// Don't:
// dl1 := downloader.New("/downloads")
// dl2 := downloader.New("/downloads") // Same path!
```

### 2. Timeout Handling

```go
// Set realistic timeouts based on expected file size
// Default: 5 minutes
// For large files (>1GB), consider increasing timeout
```

### 3. Error Logging

```go
result, err := dl.Download(opts)
if err != nil {
    log.Printf("Download failed - URL: %s, Error: %v", opts.URL, err)
    // Include context for debugging
}
```

### 4. Cleanup

```go
// Periodically clean up old download folders
func cleanupOldDownloads(basePath string, daysOld int) error {
    // Remove folders older than X days
    // Based on timestamp folder names
}
```

### 5. Validation

```go
// Validate URLs before downloading
func isValidURL(url string) bool {
    _, err := url.Parse(url)
    return err == nil && (strings.HasPrefix(url, "http://") ||
                          strings.HasPrefix(url, "https://"))
}
```

## Quick Reference

```go
// Create downloader
dl := downloader.New("/path/to/downloads")

// Simple download
result, err := dl.Download(downloader.DownloadOptions{
    URL: "https://youtube.com/watch?v=VIDEO_ID",
})

// Download with options
result, err := dl.Download(downloader.DownloadOptions{
    URL:            "https://vimeo.com/123456789",
    PostProcessing: "mp4",      // or "audio", "permute"
    UseUGE:         false,
})

// Check status
isActive := dl.IsDownloading()

// Stop download
err := dl.StopDownload()

// Handle result
if err == nil {
    fmt.Printf("File: %s\n", result.FileName)
    fmt.Printf("Path: %s\n", result.FilePath)
}
```

## Related Documentation

- [Downie Official Documentation](https://software.charliemonroe.net/downie/)
- [Downie URL Scheme Reference](https://software.charliemonroe.net/downie/url-scheme.php)
- [Project Main README](../../selfhost-mac-helper/README.md)

## Summary

This package provides a robust Go interface to Downie for automated video downloading with:

- ✅ Simple, type-safe API
- ✅ Concurrent download prevention
- ✅ Graceful cancellation
- ✅ Multiple post-processing options
- ✅ HTTP API integration ready
- ✅ Comprehensive error handling
- ✅ Production-tested patterns

Always refer to the implementation files in [./reference/](./reference/) for the most up-to-date code examples and patterns.
