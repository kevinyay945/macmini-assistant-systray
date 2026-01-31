// Package downie provides video download functionality via Downie deep links.
package downie

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/kevinyay945/macmini-assistant-systray/internal/registry"
	"github.com/kevinyay945/macmini-assistant-systray/internal/tools"
)

// Compile-time interface check
var _ registry.Tool = (*Tool)(nil)

// Default timeouts and intervals
const (
	DefaultDownloadTimeout = 5 * time.Minute
	PollInterval           = 5 * time.Second
)

// Sentinel errors for the Downie tool.
var (
	ErrNotEnabled          = errors.New("downie tool is not enabled")
	ErrMissingURL          = errors.New("url parameter is required")
	ErrInvalidURL          = errors.New("invalid video URL")
	ErrDownloadInProgress  = errors.New("download already in progress")
	ErrDownloadTimeout     = errors.New("download timed out")
	ErrDownloadCancelled   = errors.New("download was cancelled")
	ErrNoDownloadFolder    = errors.New("download folder not configured")
	ErrUnsupportedFormat   = errors.New("unsupported format")
	ErrUnsupportedResolution = errors.New("unsupported resolution")
)

// Supported formats and resolutions
var (
	SupportedFormats     = []string{"mp4", "mkv", "webm", "m4v"}
	SupportedResolutions = []string{"4320p", "2160p", "1440p", "1080p", "720p", "480p", "360p"}
)

// youtubeURLPattern matches YouTube URLs in various formats.
var youtubeURLPattern = regexp.MustCompile(
	`^(https?://)?(www\.)?(youtube\.com/watch\?v=|youtu\.be/|youtube\.com/shorts/)[\w-]+`,
)

// genericURLPattern matches any valid URL
var genericURLPattern = regexp.MustCompile(
	`^https?://[^\s/$.?#].[^\s]*$`,
)

// CommandExecutor is an interface for executing system commands (for testing).
type CommandExecutor interface {
	Execute(ctx context.Context, name string, args ...string) error
}

// RealCommandExecutor executes real system commands.
type RealCommandExecutor struct{}

// Execute runs the command with the given context.
func (e *RealCommandExecutor) Execute(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Run()
}

// Tool implements the Downie video download tool.
type Tool struct {
	enabled        bool
	downloadFolder string
	timeout        time.Duration
	executor       CommandExecutor

	mu            sync.Mutex
	isDownloading bool
	cancelFunc    context.CancelFunc
}

// Config holds Downie tool configuration.
type Config struct {
	Enabled        bool
	DownloadFolder string
	Timeout        time.Duration
	Executor       CommandExecutor // optional, for testing
}

// New creates a new Downie tool instance.
func New(cfg Config) *Tool {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultDownloadTimeout
	}

	executor := cfg.Executor
	if executor == nil {
		executor = &RealCommandExecutor{}
	}

	return &Tool{
		enabled:        cfg.Enabled,
		downloadFolder: cfg.DownloadFolder,
		timeout:        timeout,
		executor:       executor,
		isDownloading:  false,
	}
}

// Name returns the tool name.
func (t *Tool) Name() string {
	return "downie"
}

// Description returns the tool description.
func (t *Tool) Description() string {
	return "Download videos using Downie application"
}

// Schema returns the tool schema for LLM integration.
func (t *Tool) Schema() registry.ToolSchema {
	return registry.ToolSchema{
		Inputs: []registry.Parameter{
			{
				Name:        "url",
				Type:        "string",
				Required:    true,
				Description: "The video URL to download",
			},
			{
				Name:        "format",
				Type:        "string",
				Required:    false,
				Description: "Output format",
				Default:     "mp4",
				Allowed:     SupportedFormats,
			},
			{
				Name:        "resolution",
				Type:        "string",
				Required:    false,
				Description: "Video resolution",
				Default:     "1080p",
				Allowed:     SupportedResolutions,
			},
		},
		Outputs: []registry.Parameter{
			{
				Name:        "status",
				Type:        "string",
				Required:    true,
				Description: "Download status",
			},
			{
				Name:        "message",
				Type:        "string",
				Required:    true,
				Description: "Status message",
			},
			{
				Name:        "file_path",
				Type:        "string",
				Required:    false,
				Description: "Absolute path to downloaded file",
			},
			{
				Name:        "file_size",
				Type:        "integer",
				Required:    false,
				Description: "File size in bytes",
			},
		},
	}
}

// IsDownloading returns whether a download is currently in progress.
func (t *Tool) IsDownloading() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.isDownloading
}

// StopDownload stops the current download process.
func (t *Tool) StopDownload() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isDownloading {
		return errors.New("no download in progress")
	}

	if t.cancelFunc != nil {
		t.cancelFunc()
	}

	return nil
}

// Execute runs the Downie download with the given parameters.
// Parameters:
//   - url: The video URL to download (required)
//   - format: Output format (optional, default: mp4)
//   - resolution: Video resolution (optional, default: 1080p)
func (t *Tool) Execute(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	// Context check should be first to fail fast
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if !t.enabled {
		return nil, ErrNotEnabled
	}

	// Check if download folder is configured
	if t.downloadFolder == "" {
		return nil, ErrNoDownloadFolder
	}

	// Parse parameters
	videoURL, err := tools.GetRequiredString(params, "url")
	if err != nil {
		return nil, ErrMissingURL
	}

	// Validate URL
	if !IsValidURL(videoURL) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidURL, videoURL)
	}

	format := tools.GetOptionalString(params, "format", "mp4")
	resolution := tools.GetOptionalString(params, "resolution", "1080p")

	// Validate format
	if !isValidFormat(format) {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, format)
	}

	// Validate resolution
	if !isValidResolution(resolution) {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedResolution, resolution)
	}

	// Check if a download is already in progress
	t.mu.Lock()
	if t.isDownloading {
		t.mu.Unlock()
		return nil, ErrDownloadInProgress
	}

	// Create a cancellable context for this download
	downloadCtx, cancel := context.WithCancel(ctx)
	t.cancelFunc = cancel
	t.isDownloading = true
	t.mu.Unlock()

	// Ensure cleanup when done
	defer func() {
		t.mu.Lock()
		t.isDownloading = false
		t.cancelFunc = nil
		t.mu.Unlock()
	}()

	// Create destination folder with timestamp
	timestamp := time.Now().Unix()
	destination := filepath.Join(t.downloadFolder, fmt.Sprintf("%d", timestamp))
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create destination folder: %w", err)
	}

	// Build and execute Downie deep link
	deepLink := BuildDeepLink(videoURL, destination, format, resolution)
	if err := t.executor.Execute(downloadCtx, "open", deepLink); err != nil {
		// Clean up folder on failure
		_ = os.Remove(destination)
		return nil, fmt.Errorf("failed to execute Downie command: %w", err)
	}

	// Wait for download completion
	filePath, err := t.waitForDownload(downloadCtx, destination, timestamp)
	if err != nil {
		return nil, err
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return map[string]interface{}{
		"status":    "completed",
		"message":   fmt.Sprintf("Downloaded: %s", filepath.Base(filePath)),
		"file_path": filePath,
		"file_size": fileInfo.Size(),
	}, nil
}

// waitForDownload polls the destination folder until a complete file is found.
func (t *Tool) waitForDownload(ctx context.Context, destination string, folderID int64) (string, error) {
	timeout := time.After(t.timeout)
	ticker := time.NewTicker(PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ErrDownloadCancelled
		case <-timeout:
			return "", ErrDownloadTimeout
		case <-ticker.C:
			// Check if the destination folder has files
			files, err := os.ReadDir(destination)
			if err != nil || len(files) == 0 {
				continue
			}

			// Look for complete files (no .downiepart extension)
			var completeFile string
			hasIncomplete := false
			for _, file := range files {
				if strings.Contains(file.Name(), "downiepart") {
					hasIncomplete = true
					break
				}
				if !file.IsDir() {
					completeFile = file.Name()
				}
			}

			// Return if we have a complete file and no incomplete files
			if !hasIncomplete && completeFile != "" {
				return filepath.Join(destination, completeFile), nil
			}
		}
	}
}

// BuildDeepLink constructs the Downie deep link URL.
func BuildDeepLink(videoURL, destination, format, resolution string) string {
	// URL-encode the video URL (encode ? and & characters)
	encodedURL := strings.ReplaceAll(videoURL, "?", "%3F")
	encodedURL = strings.ReplaceAll(encodedURL, "&", "%26")

	// Start building the custom URL scheme
	deepLink := fmt.Sprintf("downie://XUOpenURL?url=%s", encodedURL)

	// Add destination if provided
	if destination != "" {
		deepLink += fmt.Sprintf("&destination=%s", url.QueryEscape(destination))
	}

	// Add post-processing format if not mp4 (mp4 is default)
	if format != "" && format != "mp4" {
		deepLink += fmt.Sprintf("&postprocessing=%s", format)
	}

	// Note: Downie doesn't have a direct quality parameter in the URL scheme
	// Quality is typically controlled through Downie's preferences

	return deepLink
}

// IsValidURL checks if the URL is a valid video URL.
func IsValidURL(videoURL string) bool {
	if videoURL == "" {
		return false
	}

	// Check for YouTube URLs (most common)
	if youtubeURLPattern.MatchString(videoURL) {
		return true
	}

	// Check for any valid HTTP/HTTPS URL
	return genericURLPattern.MatchString(videoURL)
}

// IsValidYouTubeURL checks if the URL is a valid YouTube URL.
func IsValidYouTubeURL(videoURL string) bool {
	return youtubeURLPattern.MatchString(videoURL)
}

// isValidFormat checks if the format is supported.
func isValidFormat(format string) bool {
	for _, f := range SupportedFormats {
		if f == format {
			return true
		}
	}
	return false
}

// isValidResolution checks if the resolution is supported.
func isValidResolution(resolution string) bool {
	for _, r := range SupportedResolutions {
		if r == resolution {
			return true
		}
	}
	return false
}
