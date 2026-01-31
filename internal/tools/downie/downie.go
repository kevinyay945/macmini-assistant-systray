// Package downie provides video download functionality via Downie deep links.
package downie

import (
	"context"
	"errors"
	"fmt"

	"github.com/kevinyay945/macmini-assistant-systray/internal/registry"
	"github.com/kevinyay945/macmini-assistant-systray/internal/tools"
)

// Compile-time interface check
var _ registry.Tool = (*Tool)(nil)

// Sentinel errors for the Downie tool.
var (
	ErrNotEnabled = errors.New("downie tool is not enabled")
	ErrMissingURL = errors.New("url parameter is required")
)

// Tool implements the Downie video download tool.
type Tool struct {
	enabled bool
}

// Config holds Downie tool configuration.
type Config struct {
	Enabled bool
}

// New creates a new Downie tool instance.
func New(cfg Config) *Tool {
	return &Tool{
		enabled: cfg.Enabled,
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

// Parameters returns the tool parameter definitions for LLM integration.
func (t *Tool) Parameters() []registry.ParameterDef {
	return []registry.ParameterDef{
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
			Description: "Output format (e.g., mp4, mkv)",
			Default:     "mp4",
		},
		{
			Name:        "resolution",
			Type:        "string",
			Required:    false,
			Description: "Video resolution (e.g., 1080p, 720p)",
			Default:     "1080p",
		},
	}
}

// Execute runs the Downie download with the given parameters.
// Parameters:
//   - url: The video URL to download (required)
//   - format: Output format (optional, default: mp4)
//   - resolution: Video resolution (optional, default: 1080p)
func (t *Tool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Context check should be first to fail fast
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if !t.enabled {
		return nil, ErrNotEnabled
	}

	url, err := tools.GetRequiredString(params, "url")
	if err != nil {
		return nil, ErrMissingURL
	}

	format := tools.GetOptionalString(params, "format", "mp4")
	resolution := tools.GetOptionalString(params, "resolution", "1080p")

	// TODO: Implement Downie deep link execution
	// Format: downie://XcallbackURL/open?url=<encoded_url>
	return map[string]interface{}{
		"status":     "pending",
		"message":    fmt.Sprintf("Download request queued for: %s", url),
		"format":     format,
		"resolution": resolution,
	}, nil
}
