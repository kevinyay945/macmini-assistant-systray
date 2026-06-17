// Package gdrive provides Google Drive upload functionality.
package gdrive

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/kevinyay945/macmini-assistant-systray/internal/registry"
	"github.com/kevinyay945/macmini-assistant-systray/internal/tools"
)

// Compile-time interface check
var _ registry.Tool = (*Tool)(nil)

// Default timeout for uploads
const DefaultUploadTimeout = 5 * time.Minute

// Sentinel errors for the Google Drive tool.
var (
	ErrNotEnabled          = errors.New("google_drive tool is not enabled")
	ErrMissingFilePath     = errors.New("file_path parameter is required")
	ErrFileNotFound        = errors.New("file not found")
	ErrInvalidCredentials  = errors.New("invalid or missing credentials")
	ErrUploadFailed        = errors.New("upload failed")
	ErrPermissionFailed    = errors.New("failed to set permissions")
	ErrUploadTimeout       = errors.New("upload timeout exceeded")
	ErrServiceNotInitialized = errors.New("drive service not initialized")
)

// DriveService defines the interface for Google Drive operations (for testing).
type DriveService interface {
	UploadFile(ctx context.Context, filePath, name, folderID string) (*drive.File, error)
	SetPublicPermission(ctx context.Context, fileID string) error
}

// RealDriveService implements DriveService using the actual Google Drive API.
type RealDriveService struct {
	service *drive.Service
}

// UploadFile uploads a file to Google Drive.
func (s *RealDriveService) UploadFile(ctx context.Context, filePath, name, folderID string) (*drive.File, error) {
	// Open the file
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFileNotFound, err)
	}
	defer f.Close()

	// Create file metadata
	fileMetadata := &drive.File{
		Name: name,
	}

	// Set parent folder if specified
	if folderID != "" {
		fileMetadata.Parents = []string{folderID}
	}

	// Upload the file
	file, err := s.service.Files.Create(fileMetadata).
		Media(f).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUploadFailed, err)
	}

	return file, nil
}

// SetPublicPermission sets the file to be publicly accessible via link.
func (s *RealDriveService) SetPublicPermission(ctx context.Context, fileID string) error {
	permission := &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}

	_, err := s.service.Permissions.Create(fileID, permission).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPermissionFailed, err)
	}

	return nil
}

// Tool implements the Google Drive upload tool.
type Tool struct {
	enabled            bool
	credentialsPath    string
	serviceAccountPath string
	timeout            time.Duration
	driveService       DriveService
}

// Config holds Google Drive tool configuration.
type Config struct {
	Enabled            bool
	CredentialsPath    string
	ServiceAccountPath string
	Timeout            time.Duration
	DriveService       DriveService // optional, for testing
}

// New creates a new Google Drive tool instance.
func New(cfg Config) *Tool {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultUploadTimeout
	}

	return &Tool{
		enabled:            cfg.Enabled,
		credentialsPath:    cfg.CredentialsPath,
		serviceAccountPath: cfg.ServiceAccountPath,
		timeout:            timeout,
		driveService:       cfg.DriveService,
	}
}

// InitService initializes the Google Drive service using credentials.
// This is called lazily on first Execute if not already initialized.
func (t *Tool) InitService(ctx context.Context) error {
	if t.driveService != nil {
		return nil // Already initialized
	}

	// Determine which credentials to use
	credPath := t.serviceAccountPath
	if credPath == "" {
		credPath = t.credentialsPath
	}
	if credPath == "" {
		return ErrInvalidCredentials
	}

	// Read credentials file
	// #nosec G304 - Path comes from user config, validated at config load time
	data, err := os.ReadFile(credPath)
	if err != nil {
		return fmt.Errorf("%w: unable to read credentials file: %v", ErrInvalidCredentials, err)
	}

	// Create credentials from JSON
	// For service accounts, use the default credentials
	config, err := google.JWTConfigFromJSON(data, drive.DriveFileScope)
	if err != nil {
		// If not a service account, the error will indicate that
		return fmt.Errorf("%w: unable to parse credentials: %v", ErrInvalidCredentials, err)
	}

	// Create HTTP client with credentials
	client := config.Client(ctx)

	// Create Drive service
	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("%w: unable to create Drive service: %v", ErrInvalidCredentials, err)
	}

	t.driveService = &RealDriveService{service: service}
	return nil
}

// Name returns the tool name.
func (t *Tool) Name() string {
	return "google_drive"
}

// Description returns the tool description.
func (t *Tool) Description() string {
	return "Upload files to Google Drive and generate share links"
}

// Schema returns the tool schema for LLM integration.
func (t *Tool) Schema() registry.ToolSchema {
	return registry.ToolSchema{
		Inputs: []registry.Parameter{
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
			{
				Name:        "timeout",
				Type:        "integer",
				Required:    false,
				Default:     300,
				Description: "Upload timeout in seconds",
			},
		},
		Outputs: []registry.Parameter{
			{
				Name:        "status",
				Type:        "string",
				Required:    true,
				Description: "Upload status",
			},
			{
				Name:        "file_id",
				Type:        "string",
				Required:    false,
				Description: "Google Drive file ID",
			},
			{
				Name:        "share_link",
				Type:        "string",
				Required:    false,
				Description: "Public share link to uploaded file",
			},
		},
	}
}

// Execute runs the Google Drive upload with the given parameters.
// Parameters:
//   - file_path: Local path to the file to upload (required)
//   - folder_id: Google Drive folder ID to upload to (optional)
//   - name: Name for the uploaded file (optional, defaults to original filename)
//   - timeout: Upload timeout in seconds (optional, default: 300)
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

	// Parse parameters
	filePath, err := tools.GetRequiredString(params, "file_path")
	if err != nil {
		return nil, ErrMissingFilePath
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrFileNotFound, filePath)
	}

	folderID := tools.GetOptionalString(params, "folder_id", "")
	name := tools.GetOptionalString(params, "name", filepath.Base(filePath))
	timeoutSec := tools.GetOptionalInt(params, "timeout", 300)

	// Create context with timeout
	uploadCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	// Initialize service if needed
	if err := t.InitService(uploadCtx); err != nil {
		return nil, err
	}

	if t.driveService == nil {
		return nil, ErrServiceNotInitialized
	}

	// Upload file
	file, err := t.driveService.UploadFile(uploadCtx, filePath, name, folderID)
	if err != nil {
		if errors.Is(uploadCtx.Err(), context.DeadlineExceeded) {
			return nil, ErrUploadTimeout
		}
		return nil, err
	}

	// Set public permissions
	if err := t.driveService.SetPublicPermission(uploadCtx, file.Id); err != nil {
		// Log the error but don't fail - file was uploaded successfully
		// Return with file ID but no share link
		return map[string]interface{}{
			"status":  "completed",
			"file_id": file.Id,
			"message": fmt.Sprintf("File uploaded but sharing failed: %v", err),
		}, nil
	}

	// Generate share link
	shareLink := fmt.Sprintf("https://drive.google.com/file/d/%s/view?usp=sharing", file.Id)

	return map[string]interface{}{
		"status":     "completed",
		"file_id":    file.Id,
		"share_link": shareLink,
	}, nil
}
