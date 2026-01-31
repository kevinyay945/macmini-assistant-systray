package downie_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kevinyay945/macmini-assistant-systray/internal/tools/downie"
)

func TestTool_New(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true})
	if tool == nil {
		t.Error("New() returned nil")
	}
}

func TestTool_Name(t *testing.T) {
	tool := downie.New(downie.Config{})
	if got := tool.Name(); got != "downie" {
		t.Errorf("Name() = %q, want %q", got, "downie")
	}
}

func TestTool_Description(t *testing.T) {
	tool := downie.New(downie.Config{})
	if got := tool.Description(); got == "" {
		t.Error("Description() returned empty string")
	}
}

func TestTool_Schema(t *testing.T) {
	tool := downie.New(downie.Config{})
	schema := tool.Schema()

	if len(schema.Inputs) != 3 {
		t.Errorf("Schema().Inputs returned %d params, want 3", len(schema.Inputs))
	}

	// Check required URL parameter
	urlParam := schema.Inputs[0]
	if urlParam.Name != "url" {
		t.Errorf("First param name = %q, want 'url'", urlParam.Name)
	}
	if !urlParam.Required {
		t.Error("URL parameter should be required")
	}

	// Check optional parameters have defaults
	formatParam := schema.Inputs[1]
	if formatParam.Required {
		t.Error("format parameter should not be required")
	}
	if formatParam.Default != "mp4" {
		t.Errorf("format default = %v, want 'mp4'", formatParam.Default)
	}

	// Check format parameter has Allowed values
	if len(formatParam.Allowed) == 0 {
		t.Error("format parameter should have Allowed values")
	}
	expectedFormats := []string{"mp4", "mkv", "webm", "m4v"}
	for _, fmt := range expectedFormats {
		found := false
		for _, allowed := range formatParam.Allowed {
			if allowed == fmt {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("format Allowed should include %q", fmt)
		}
	}

	// Check resolution parameter has Allowed values
	resParam := schema.Inputs[2]
	if len(resParam.Allowed) == 0 {
		t.Error("resolution parameter should have Allowed values")
	}
	expectedResolutions := []string{"4320p", "2160p", "1440p", "1080p", "720p", "480p", "360p"}
	for _, res := range expectedResolutions {
		found := false
		for _, allowed := range resParam.Allowed {
			if allowed == res {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("resolution Allowed should include %q", res)
		}
	}

	// Check outputs include file_path and file_size
	if len(schema.Outputs) < 2 {
		t.Errorf("Schema().Outputs returned %d params, want at least 2", len(schema.Outputs))
	}
}

func TestTool_Execute_NotEnabled(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: false})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{"url": "https://example.com"})
	if !errors.Is(err, downie.ErrNotEnabled) {
		t.Errorf("Execute() error = %v, want ErrNotEnabled", err)
	}
}

func TestTool_Execute_MissingURL(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true, DownloadFolder: "/tmp"})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{})
	if !errors.Is(err, downie.ErrMissingURL) {
		t.Errorf("Execute() error = %v, want ErrMissingURL", err)
	}
}

func TestTool_Execute_NoDownloadFolder(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true, DownloadFolder: ""})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{"url": "https://youtube.com/watch?v=test"})
	if !errors.Is(err, downie.ErrNoDownloadFolder) {
		t.Errorf("Execute() error = %v, want ErrNoDownloadFolder", err)
	}
}

func TestTool_Execute_InvalidURL(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true, DownloadFolder: "/tmp"})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{"url": "not-a-valid-url"})
	if !errors.Is(err, downie.ErrInvalidURL) {
		t.Errorf("Execute() error = %v, want ErrInvalidURL", err)
	}
}

func TestTool_Execute_UnsupportedFormat(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true, DownloadFolder: "/tmp"})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"url":    "https://youtube.com/watch?v=test",
		"format": "avi",
	})
	if !errors.Is(err, downie.ErrUnsupportedFormat) {
		t.Errorf("Execute() error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestTool_Execute_UnsupportedResolution(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true, DownloadFolder: "/tmp"})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"url":        "https://youtube.com/watch?v=test",
		"resolution": "8000p",
	})
	if !errors.Is(err, downie.ErrUnsupportedResolution) {
		t.Errorf("Execute() error = %v, want ErrUnsupportedResolution", err)
	}
}

func TestTool_Execute_ContextCanceled(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true, DownloadFolder: "/tmp"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := tool.Execute(ctx, map[string]interface{}{"url": "https://example.com"})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want context.Canceled", err)
	}
}

func TestTool_Execute_ContextDeadlineExceeded(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true, DownloadFolder: "/tmp"})
	// Use an already-expired deadline to avoid flaky race conditions
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := tool.Execute(ctx, map[string]interface{}{"url": "https://example.com"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Execute() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"empty", "", false},
		{"youtube watch", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"youtube short link", "https://youtu.be/dQw4w9WgXcQ", true},
		{"youtube shorts", "https://youtube.com/shorts/abc123", true},
		{"youtube no www", "https://youtube.com/watch?v=abc123", true},
		{"vimeo", "https://vimeo.com/123456789", true},
		{"twitter", "https://twitter.com/user/status/123", true},
		{"generic https", "https://example.com/video.mp4", true},
		{"generic http", "http://example.com/video.mp4", true},
		{"no protocol", "example.com/video", false},
		{"ftp protocol", "ftp://example.com/video", false},
		{"invalid", "not-a-url", false},
		{"javascript", "javascript:alert(1)", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := downie.IsValidURL(tt.url)
			if got != tt.expected {
				t.Errorf("IsValidURL(%q) = %v, want %v", tt.url, got, tt.expected)
			}
		})
	}
}

func TestIsValidYouTubeURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"youtube watch", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"youtube short link", "https://youtu.be/dQw4w9WgXcQ", true},
		{"youtube shorts", "https://youtube.com/shorts/abc123", true},
		{"youtube no www", "https://youtube.com/watch?v=abc123", true},
		{"vimeo", "https://vimeo.com/123456789", false},
		{"not youtube", "https://example.com/youtube", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := downie.IsValidYouTubeURL(tt.url)
			if got != tt.expected {
				t.Errorf("IsValidYouTubeURL(%q) = %v, want %v", tt.url, got, tt.expected)
			}
		})
	}
}

func TestBuildDeepLink(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		destination string
		format      string
		resolution  string
		wantContain []string
	}{
		{
			name:        "basic URL",
			url:         "https://youtube.com/watch?v=abc123",
			destination: "/tmp/downloads/123",
			format:      "mp4",
			resolution:  "1080p",
			wantContain: []string{
				"downie://XUOpenURL?url=",
				"%3Fv=abc123",    // ? encoded
				"destination=",
			},
		},
		{
			name:        "URL with ampersand",
			url:         "https://youtube.com/watch?v=abc123&t=10",
			destination: "/tmp/test",
			format:      "mp4",
			resolution:  "1080p",
			wantContain: []string{
				"%26t=10", // & encoded
			},
		},
		{
			name:        "non-mp4 format",
			url:         "https://youtube.com/watch?v=abc123",
			destination: "/tmp/test",
			format:      "mkv",
			resolution:  "1080p",
			wantContain: []string{
				"postprocessing=mkv",
			},
		},
		{
			name:        "mp4 format should not add postprocessing",
			url:         "https://youtube.com/watch?v=abc123",
			destination: "/tmp/test",
			format:      "mp4",
			resolution:  "1080p",
			wantContain: []string{
				"downie://XUOpenURL",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := downie.BuildDeepLink(tt.url, tt.destination, tt.format, tt.resolution)
			for _, want := range tt.wantContain {
				if !contains(got, want) {
					t.Errorf("BuildDeepLink() = %q, want to contain %q", got, want)
				}
			}
		})
	}
}

func TestTool_IsDownloading(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true, DownloadFolder: "/tmp"})

	// Initially should not be downloading
	if tool.IsDownloading() {
		t.Error("IsDownloading() should be false initially")
	}
}

func TestTool_StopDownload_NoDownload(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true, DownloadFolder: "/tmp"})

	err := tool.StopDownload()
	if err == nil {
		t.Error("StopDownload() should return error when no download in progress")
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
