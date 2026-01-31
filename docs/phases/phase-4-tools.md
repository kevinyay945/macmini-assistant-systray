# Phase 4: Tool Implementation

**Duration**: Weeks 7-8
**Status**: ⚪ Not Started
**Goal**: Implement YouTube download and Google Drive upload tools

---

## Overview

This phase implements the two core tools: YouTube download via Downie and Google Drive upload. These are the primary value-add features of the bot.

---

## 4.1 Downie Tool (YouTube Download)

**Duration**: 3 days
**Status**: ⚪ Not Started

### Tasks

- [ ] Deep link URL construction
- [ ] Format/resolution validation
- [ ] File path resolution
- [ ] Download completion detection
- [ ] Error handling (invalid URL, unsupported format)

### Implementation Details

```go
// internal/tools/downie/downie.go
package downie

import (
    "context"
    "fmt"
    "os/exec"
    "path/filepath"
    "time"
)

type DownieTool struct {
    downloadFolder string
    logger         *observability.Logger
}

func NewDownieTool(downloadFolder string) *DownieTool

func (d *DownieTool) Name() string { return "youtube_download" }

func (d *DownieTool) Description() string {
    return "Download YouTube video using Downie"
}

func (d *DownieTool) Schema() registry.ToolSchema {
    return registry.ToolSchema{
        Inputs: []registry.Parameter{
            {Name: "youtube_url", Type: "string", Required: true, Description: "YouTube video URL"},
            {Name: "file_format", Type: "string", Required: false, Default: "mp4", Allowed: []string{"mp4", "mkv", "webm"}, Description: "Output file format"},
            {Name: "resolution", Type: "string", Required: false, Default: "1080p", Allowed: []string{"4320p", "2160p", "1440p", "1080p", "720p", "480p", "360p"}, Description: "Video resolution"},
        },
        Outputs: []registry.Parameter{
            {Name: "file_path", Type: "string", Description: "Absolute path to downloaded file"},
            {Name: "file_size", Type: "integer", Description: "File size in bytes"},
        },
    }
}

func (d *DownieTool) Execute(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
    url := params["youtube_url"].(string)
    format := getStringParam(params, "file_format", "mp4")
    resolution := getStringParam(params, "resolution", "1080p")

    // Validate URL
    if !isValidYouTubeURL(url) {
        return nil, fmt.Errorf("invalid YouTube URL: %s", url)
    }

    // Construct deep link
    deepLink := d.buildDeepLink(url, format, resolution)

    // Open Downie with deep link
    if err := exec.CommandContext(ctx, "open", deepLink).Run(); err != nil {
        return nil, fmt.Errorf("failed to open Downie: %w", err)
    }

    // Wait for download completion
    filePath, err := d.waitForDownload(ctx, url)
    if err != nil {
        return nil, err
    }

    // Get file info
    fileInfo, err := os.Stat(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to get file info: %w", err)
    }

    return map[string]interface{}{
        "file_path": filePath,
        "file_size": fileInfo.Size(),
    }, nil
}

func (d *DownieTool) buildDeepLink(url, format, resolution string) string {
    // downie://[url]?format=[format]&quality=[resolution]
    return fmt.Sprintf("downie://%s?format=%s&quality=%s",
        url.QueryEscape(url), format, resolution)
}

func (d *DownieTool) waitForDownload(ctx context.Context, url string) (string, error) {
    // Poll download folder for new file
    // Implementation depends on Downie's behavior
}
```

### Test Cases

```go
// internal/tools/downie/downie_test.go
//go:build local

func TestDownie_DeepLinkConstruction(t *testing.T)
func TestDownie_DownloadMP4_1080p(t *testing.T)
func TestDownie_InvalidURL(t *testing.T)
func TestDownie_UnsupportedFormat(t *testing.T)
func TestDownie_FilePathResolution(t *testing.T)
func TestDownie_FileSizeCalculation(t *testing.T)
func TestDownie_ValidateYouTubeURL(t *testing.T)
func TestDownie_ValidateResolution(t *testing.T)
func TestDownie_ContextCancellation(t *testing.T)
```

**Build Tag**: `local` (requires Downie installed)

### Acceptance Criteria

- [ ] Downloads YouTube videos via Downie
- [ ] Supports all specified formats (mp4, mkv, webm)
- [ ] Supports all specified resolutions
- [ ] Returns absolute file path and size
- [ ] Handles errors gracefully
- [ ] Files saved to configured download folder

### Notes

<!-- Add your notes here -->

---

## 4.2 Google Drive Upload Tool

**Duration**: 4 days
**Status**: ⚪ Not Started

### Tasks

- [ ] OAuth2 authentication flow (service account)
- [ ] File upload with resumable upload
- [ ] Progress tracking (optional for v1)
- [ ] Share link generation
- [ ] Timeout handling

### Implementation Details

```go
// internal/tools/gdrive/gdrive.go
package gdrive

import (
    "context"
    "fmt"
    "io"
    "os"

    "google.golang.org/api/drive/v3"
    "google.golang.org/api/option"
)

type GDriveTool struct {
    service        *drive.Service
    credentialsPath string
    logger         *observability.Logger
}

func NewGDriveTool(credentialsPath string) (*GDriveTool, error) {
    ctx := context.Background()

    // Read service account credentials
    b, err := os.ReadFile(credentialsPath)
    if err != nil {
        return nil, fmt.Errorf("unable to read credentials file: %w", err)
    }

    // Create Drive service
    service, err := drive.NewService(ctx, option.WithCredentialsJSON(b))
    if err != nil {
        return nil, fmt.Errorf("unable to create Drive service: %w", err)
    }

    return &GDriveTool{
        service:        service,
        credentialsPath: credentialsPath,
    }, nil
}

func (g *GDriveTool) Name() string { return "gdrive_upload" }

func (g *GDriveTool) Description() string {
    return "Upload file to Google Drive and generate share link"
}

func (g *GDriveTool) Schema() registry.ToolSchema {
    return registry.ToolSchema{
        Inputs: []registry.Parameter{
            {Name: "file_path", Type: "string", Required: true, Description: "Absolute path to file"},
            {Name: "upload_name", Type: "string", Required: false, Description: "Name for uploaded file"},
            {Name: "timeout", Type: "integer", Required: false, Default: 300, Description: "Upload timeout in seconds"},
        },
        Outputs: []registry.Parameter{
            {Name: "share_link", Type: "string", Description: "Public share link to uploaded file"},
            {Name: "file_id", Type: "string", Description: "Google Drive file ID"},
        },
    }
}

func (g *GDriveTool) Execute(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
    filePath := params["file_path"].(string)
    uploadName := getStringParam(params, "upload_name", filepath.Base(filePath))
    timeout := getIntParam(params, "timeout", 300)

    // Create context with timeout
    ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
    defer cancel()

    // Open file
    f, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    defer f.Close()

    // Create file metadata
    fileMetadata := &drive.File{
        Name: uploadName,
    }

    // Upload file
    file, err := g.service.Files.Create(fileMetadata).
        Media(f).
        Context(ctx).
        Do()
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, fmt.Errorf("upload timeout exceeded")
        }
        return nil, fmt.Errorf("failed to upload file: %w", err)
    }

    // Set file to public (anyone with link can view)
    permission := &drive.Permission{
        Type: "anyone",
        Role: "reader",
    }
    _, err = g.service.Permissions.Create(file.Id, permission).Context(ctx).Do()
    if err != nil {
        return nil, fmt.Errorf("failed to set permissions: %w", err)
    }

    // Generate share link
    shareLink := fmt.Sprintf("https://drive.google.com/file/d/%s/view?usp=sharing", file.Id)

    return map[string]interface{}{
        "share_link": shareLink,
        "file_id":    file.Id,
    }, nil
}
```

### Test Cases

```go
// internal/tools/gdrive/gdrive_test.go
//go:build integration

func TestGDrive_OAuth2Flow(t *testing.T)
func TestGDrive_UploadFile(t *testing.T)
func TestGDrive_CustomFileName(t *testing.T)
func TestGDrive_ShareLinkGeneration(t *testing.T)
func TestGDrive_UploadTimeout(t *testing.T)
func TestGDrive_LargeFileUpload(t *testing.T)
func TestGDrive_FileNotFound(t *testing.T)
func TestGDrive_InvalidCredentials(t *testing.T)
```

**Build Tag**: `integration` (requires Google Cloud credentials)

### Acceptance Criteria

- [ ] Service account authentication working
- [ ] Files uploaded successfully
- [ ] Share links generated and public-accessible
- [ ] Custom file names supported
- [ ] Timeout respected
- [ ] Handles network interruptions

### Notes

<!-- Add your notes here -->

---

## Deliverables

By the end of Phase 4:

- [ ] YouTube download via Downie working
- [ ] Google Drive upload working
- [ ] Both tools registered and callable via Copilot

---

## Dependencies

```go
// go.mod additions
require (
    google.golang.org/api v0.x.x
    golang.org/x/oauth2 v0.x.x
)
```

---

## Setup Requirements

### Downie Setup
1. Install Downie from App Store or website
2. Configure download folder to match config
3. Enable deep link support (should be enabled by default)

### Google Drive Setup
1. Create Google Cloud Project
2. Enable Drive API
3. Create Service Account
4. Download JSON credentials
5. Place at `~/.macmini-assistant/gdrive-creds.json`
6. Share target folder with service account email (optional)

---

## Time Tracking

| Task | Estimated | Actual | Notes |
|------|-----------|--------|-------|
| 4.1 Downie Tool | 3 days | | |
| 4.2 GDrive Tool | 4 days | | |
| **Total** | **7 days** | | |

---

**Previous**: [Phase 3: GitHub Copilot SDK Integration](./phase-3-copilot.md)
**Next**: [Phase 5: System Tray & Auto-start](./phase-5-systray.md)
