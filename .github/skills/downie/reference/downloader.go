package downloader

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// DownloadOptions holds the options for downloading a video
type DownloadOptions struct {
	URL            string
	PostProcessing string // Options: mp4, audio, permute
	UseUGE         bool   // User-Guided Extraction
	Destination    string
}

// DownloadResult holds the result of a download operation
type DownloadResult struct {
	FileName string
	FilePath string
}

// Downloader handles downloading videos using Downie
type Downloader struct {
	basePath        string
	mu              sync.Mutex
	isDownloading   bool
	cancelFunc      context.CancelFunc
	downloadContext context.Context
}

// New creates a new Downloader instance
func New(basePath string) *Downloader {
	return &Downloader{
		basePath:      basePath,
		isDownloading: false,
	}
}

// IsDownloading returns whether a download is currently in progress
func (d *Downloader) IsDownloading() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.isDownloading
}

// StopDownload stops the current download process
func (d *Downloader) StopDownload() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isDownloading {
		return fmt.Errorf("no download in progress")
	}

	if d.cancelFunc != nil {
		d.cancelFunc()
	}

	return nil
}

// Download downloads a video using Downie and waits for completion
func (d *Downloader) Download(opts DownloadOptions) (*DownloadResult, error) {
	// Acquire lock
	d.mu.Lock()
	if d.isDownloading {
		d.mu.Unlock()
		return nil, fmt.Errorf("download already in progress")
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	d.downloadContext = ctx
	d.cancelFunc = cancel
	d.isDownloading = true
	d.mu.Unlock()

	// Ensure we clean up when done
	defer func() {
		d.mu.Lock()
		d.isDownloading = false
		d.cancelFunc = nil
		d.downloadContext = nil
		d.mu.Unlock()
	}()

	// Create destination folder
	randomFolder := time.Now().Unix()
	destination := fmt.Sprintf("%s/%d", d.basePath, randomFolder)

	if err := os.MkdirAll(destination, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination folder: %w", err)
	}

	// Execute Downie command
	if err := d.executeDownie(opts.URL, destination, opts.PostProcessing, opts.UseUGE); err != nil {
		return nil, err
	}

	// Wait for download to complete
	result, err := d.waitForDownload(ctx, destination, randomFolder)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// executeDownie executes the Downie command
func (d *Downloader) executeDownie(url, destination, postProcessing string, useUGE bool) error {
	var cmd *exec.Cmd

	// Check if we should use the custom URL scheme for advanced options
	if postProcessing != "" || destination != "" || useUGE {
		downieURL := d.buildDownieURL(url, destination, postProcessing, useUGE)
		cmd = exec.Command("open", downieURL)
	} else {
		// Try with "Downie 4" first (standard version)
		cmd = exec.Command("open", "-a", "Downie 4", url)
	}

	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		errorMsg := string(output)
		// Check if it might be a Setapp version issue
		if strings.Contains(errorMsg, "Unable to find application") && postProcessing == "" && !useUGE {
			// Try with just "Downie" (Setapp version)
			cmd = exec.Command("open", "-a", "Downie", url)
			output, err = cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to execute Downie command: %s", string(output))
			}
		} else {
			return fmt.Errorf("failed to execute Downie command: %s", string(output))
		}
	}

	return nil
}

// buildDownieURL builds the custom Downie URL scheme
func (d *Downloader) buildDownieURL(url, destination, postProcessing string, useUGE bool) string {
	// Replace ? with %3F and & with %26 in the URL to properly encode it
	encodedURL := strings.ReplaceAll(url, "?", "%3F")
	encodedURL = strings.ReplaceAll(encodedURL, "&", "%26")

	// Start building the custom URL
	downieURL := fmt.Sprintf("downie://XUOpenURL?url=%s", encodedURL)

	// Add optional parameters if provided
	if postProcessing != "" {
		downieURL += fmt.Sprintf("&postprocessing=%s", postProcessing)
	}

	if destination != "" {
		downieURL += fmt.Sprintf("&destination=%s", destination)
	}

	if useUGE {
		downieURL += "&action=open_in_uge"
	}

	return downieURL
}

// waitForDownload waits for the download to complete and returns the result
func (d *Downloader) waitForDownload(ctx context.Context, destination string, folderID int64) (*DownloadResult, error) {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("download was cancelled")
		case <-timeout:
			return nil, fmt.Errorf("download timed out after 5 minutes")
		case <-ticker.C:
			// Check if the destination folder exists and is not empty
			files, err := os.ReadDir(destination)
			if err == nil && len(files) > 0 {
				// Check if any file has an extension containing "downiepart"
				containsDowniePart := false
				var fileName string
				for _, file := range files {
					if strings.Contains(file.Name(), "downiepart") {
						containsDowniePart = true
						break
					}
					fileName = file.Name()
				}

				// If no file contains "downiepart", return the file name
				if !containsDowniePart && fileName != "" {
					return &DownloadResult{
						FileName: fileName,
						FilePath: fmt.Sprintf("%d/%s", folderID, fileName),
					}, nil
				}
			}
		}
	}
}
