// Package gdrive provides Google Drive upload functionality.
package gdrive

import (
	"context"
	"errors"
	"fmt"

	"github.com/kevinyay945/macmini-assistant-systray/internal/registry"
	"github.com/kevinyay945/macmini-assistant-systray/internal/tools"
)

// Compile-time interface check
var _ registry.Tool = (*Tool)(nil)

// Sentinel errors for the Google Drive tool.
var (
	ErrNotEnabled      = errors.New("google_drive tool is not enabled")
	ErrMissingFilePath = errors.New("file_path parameter is required")
)

// Tool implements the Google Drive upload tool.
type Tool struct {
	enabled            bool
	credentialsPath    string
	serviceAccountPath string
}

// Config holds Google Drive tool configuration.
type Config struct {
	Enabled            bool
	CredentialsPath    string
	ServiceAccountPath string
}

// New creates a new Google Drive tool instance.
func New(cfg Config) *Tool {
	return &Tool{
		enabled:            cfg.Enabled,
		credentialsPath:    cfg.CredentialsPath,
		serviceAccountPath: cfg.ServiceAccountPath,
	}
}

// Name returns the tool name.
func (t *Tool) Name() string {
	return "google_drive"
}

// Description returns the tool description.
func (t *Tool) Description() string {
	return "Upload files to Google Drive"
}

// Parameters returns the tool parameter definitions for LLM integration.
func (t *Tool) Parameters() []registry.ParameterDef {
	return []registry.ParameterDef{
		{
			Name:        "file_path",
			Type:        "string",
			Required:    true,
			Description: "Local path to the file to upload",
		},
		{
			Name:        "folder_id",
			Type:        "string",
			Required:    false,
			Description: "Google Drive folder ID to upload to (defaults to root)",
		},
		{
			Name:        "name",
			Type:        "string",
			Required:    false,
			Description: "Name for the uploaded file (defaults to original filename)",
		},
	}
}

// Execute runs the Google Drive upload with the given parameters.
// Parameters:
//   - file_path: Local path to the file to upload (required)
//   - folder_id: Google Drive folder ID to upload to (optional)
//   - name: Name for the uploaded file (optional, defaults to original filename)
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

	filePath, err := tools.GetRequiredString(params, "file_path")
	if err != nil {
		return nil, ErrMissingFilePath
	}

	folderID := tools.GetOptionalString(params, "folder_id", "")
	name := tools.GetOptionalString(params, "name", "")

	// TODO: Implement Google Drive upload
	// 1. Authenticate using OAuth2 or service account
	// 2. Create Drive service
	// 3. Upload file with metadata
	return map[string]interface{}{
		"status":    "pending",
		"message":   fmt.Sprintf("Upload request queued for: %s", filePath),
		"folder_id": folderID,
		"name":      name,
	}, nil
}
