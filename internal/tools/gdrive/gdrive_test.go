package gdrive_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/api/drive/v3"

	"github.com/kevinyay945/macmini-assistant-systray/internal/tools/gdrive"
)

// MockDriveService is a mock implementation of DriveService for testing.
type MockDriveService struct {
	UploadFunc     func(ctx context.Context, filePath, name, folderID string) (*drive.File, error)
	PermissionFunc func(ctx context.Context, fileID string) error
}

func (m *MockDriveService) UploadFile(ctx context.Context, filePath, name, folderID string) (*drive.File, error) {
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, filePath, name, folderID)
	}
	return &drive.File{Id: "test-file-id", Name: name}, nil
}

func (m *MockDriveService) SetPublicPermission(ctx context.Context, fileID string) error {
	if m.PermissionFunc != nil {
		return m.PermissionFunc(ctx, fileID)
	}
	return nil
}

func TestTool_New(t *testing.T) {
	tool := gdrive.New(gdrive.Config{Enabled: true})
	if tool == nil {
		t.Error("New() returned nil")
	}
}

func TestTool_Name(t *testing.T) {
	tool := gdrive.New(gdrive.Config{})
	if got := tool.Name(); got != "google_drive" {
		t.Errorf("Name() = %q, want %q", got, "google_drive")
	}
}

func TestTool_Description(t *testing.T) {
	tool := gdrive.New(gdrive.Config{})
	if got := tool.Description(); got == "" {
		t.Error("Description() returned empty string")
	}
}

func TestTool_Schema(t *testing.T) {
	tool := gdrive.New(gdrive.Config{})
	schema := tool.Schema()

	if len(schema.Inputs) != 4 {
		t.Errorf("Schema().Inputs returned %d params, want 4", len(schema.Inputs))
	}

	// Check required file_path parameter
	filePathParam := schema.Inputs[0]
	if filePathParam.Name != "file_path" {
		t.Errorf("First param name = %q, want 'file_path'", filePathParam.Name)
	}
	if !filePathParam.Required {
		t.Error("file_path parameter should be required")
	}

	// Check optional parameters
	folderIDParam := schema.Inputs[1]
	if folderIDParam.Required {
		t.Error("folder_id parameter should not be required")
	}

	// Check timeout parameter has default
	timeoutParam := schema.Inputs[3]
	if timeoutParam.Name != "timeout" {
		t.Errorf("Fourth param name = %q, want 'timeout'", timeoutParam.Name)
	}
	if timeoutParam.Default != 300 {
		t.Errorf("timeout default = %v, want 300", timeoutParam.Default)
	}

	// Check outputs include share_link
	if len(schema.Outputs) < 3 {
		t.Errorf("Schema().Outputs returned %d params, want at least 3", len(schema.Outputs))
	}
	foundShareLink := false
	for _, out := range schema.Outputs {
		if out.Name == "share_link" {
			foundShareLink = true
			break
		}
	}
	if !foundShareLink {
		t.Error("Schema().Outputs should include share_link")
	}
}

func TestTool_Execute_NotEnabled(t *testing.T) {
	tool := gdrive.New(gdrive.Config{Enabled: false})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{"file_path": "/path/to/file"})
	if !errors.Is(err, gdrive.ErrNotEnabled) {
		t.Errorf("Execute() error = %v, want ErrNotEnabled", err)
	}
}

func TestTool_Execute_MissingFilePath(t *testing.T) {
	tool := gdrive.New(gdrive.Config{
		Enabled:      true,
		DriveService: &MockDriveService{},
	})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{})
	if !errors.Is(err, gdrive.ErrMissingFilePath) {
		t.Errorf("Execute() error = %v, want ErrMissingFilePath", err)
	}
}

func TestTool_Execute_FileNotFound(t *testing.T) {
	tool := gdrive.New(gdrive.Config{
		Enabled:      true,
		DriveService: &MockDriveService{},
	})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{"file_path": "/nonexistent/file.txt"})
	if !errors.Is(err, gdrive.ErrFileNotFound) {
		t.Errorf("Execute() error = %v, want ErrFileNotFound", err)
	}
}

func TestTool_Execute_ContextCanceled(t *testing.T) {
	tool := gdrive.New(gdrive.Config{Enabled: true})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := tool.Execute(ctx, map[string]interface{}{"file_path": "/path/to/file"})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want context.Canceled", err)
	}
}

func TestTool_Execute_ContextDeadlineExceeded(t *testing.T) {
	tool := gdrive.New(gdrive.Config{Enabled: true})
	// Use an already-expired deadline to avoid flaky race conditions
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := tool.Execute(ctx, map[string]interface{}{"file_path": "/path/to/file"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Execute() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestTool_Execute_ValidRequest(t *testing.T) {
	// Create a temp file to upload
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	mockService := &MockDriveService{
		UploadFunc: func(ctx context.Context, filePath, name, folderID string) (*drive.File, error) {
			return &drive.File{Id: "test-id-123", Name: name}, nil
		},
		PermissionFunc: func(ctx context.Context, fileID string) error {
			return nil
		},
	}

	tool := gdrive.New(gdrive.Config{
		Enabled:      true,
		DriveService: mockService,
	})
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{"file_path": tmpFile})
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	if result["status"] != "completed" {
		t.Errorf("Execute() status = %v, want 'completed'", result["status"])
	}

	if result["file_id"] != "test-id-123" {
		t.Errorf("Execute() file_id = %v, want 'test-id-123'", result["file_id"])
	}

	shareLink, ok := result["share_link"].(string)
	if !ok || shareLink == "" {
		t.Error("Execute() should return a share_link")
	}
}

func TestTool_Execute_WithCustomName(t *testing.T) {
	// Create a temp file to upload
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	var uploadedName string
	mockService := &MockDriveService{
		UploadFunc: func(ctx context.Context, filePath, name, folderID string) (*drive.File, error) {
			uploadedName = name
			return &drive.File{Id: "test-id", Name: name}, nil
		},
	}

	tool := gdrive.New(gdrive.Config{
		Enabled:      true,
		DriveService: mockService,
	})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"file_path": tmpFile,
		"name":      "custom-name.txt",
	})
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	if uploadedName != "custom-name.txt" {
		t.Errorf("Upload name = %q, want 'custom-name.txt'", uploadedName)
	}
}

func TestTool_Execute_WithFolderID(t *testing.T) {
	// Create a temp file to upload
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	var uploadedFolderID string
	mockService := &MockDriveService{
		UploadFunc: func(ctx context.Context, filePath, name, folderID string) (*drive.File, error) {
			uploadedFolderID = folderID
			return &drive.File{Id: "test-id", Name: name}, nil
		},
	}

	tool := gdrive.New(gdrive.Config{
		Enabled:      true,
		DriveService: mockService,
	})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{
		"file_path": tmpFile,
		"folder_id": "folder-123",
	})
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	if uploadedFolderID != "folder-123" {
		t.Errorf("Folder ID = %q, want 'folder-123'", uploadedFolderID)
	}
}

func TestTool_Execute_UploadFailure(t *testing.T) {
	// Create a temp file to upload
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	mockService := &MockDriveService{
		UploadFunc: func(ctx context.Context, filePath, name, folderID string) (*drive.File, error) {
			return nil, gdrive.ErrUploadFailed
		},
	}

	tool := gdrive.New(gdrive.Config{
		Enabled:      true,
		DriveService: mockService,
	})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{"file_path": tmpFile})
	if !errors.Is(err, gdrive.ErrUploadFailed) {
		t.Errorf("Execute() error = %v, want ErrUploadFailed", err)
	}
}

func TestTool_Execute_PermissionFailure(t *testing.T) {
	// Create a temp file to upload
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test content"), 0o644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	mockService := &MockDriveService{
		UploadFunc: func(ctx context.Context, filePath, name, folderID string) (*drive.File, error) {
			return &drive.File{Id: "test-id", Name: name}, nil
		},
		PermissionFunc: func(ctx context.Context, fileID string) error {
			return gdrive.ErrPermissionFailed
		},
	}

	tool := gdrive.New(gdrive.Config{
		Enabled:      true,
		DriveService: mockService,
	})
	ctx := context.Background()

	// Should still succeed but without share link
	result, err := tool.Execute(ctx, map[string]interface{}{"file_path": tmpFile})
	if err != nil {
		t.Errorf("Execute() should not error on permission failure, got: %v", err)
	}

	if result["status"] != "completed" {
		t.Errorf("Execute() status = %v, want 'completed'", result["status"])
	}

	// Should have file_id but no share_link
	if result["file_id"] != "test-id" {
		t.Errorf("Execute() file_id = %v, want 'test-id'", result["file_id"])
	}
	if _, ok := result["share_link"]; ok {
		t.Error("Execute() should not have share_link on permission failure")
	}
}

func TestTool_InitService_NoCredentials(t *testing.T) {
	tool := gdrive.New(gdrive.Config{
		Enabled:            true,
		CredentialsPath:    "",
		ServiceAccountPath: "",
	})

	err := tool.InitService(context.Background())
	if !errors.Is(err, gdrive.ErrInvalidCredentials) {
		t.Errorf("InitService() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestTool_InitService_FileNotFound(t *testing.T) {
	tool := gdrive.New(gdrive.Config{
		Enabled:            true,
		ServiceAccountPath: "/nonexistent/credentials.json",
	})

	err := tool.InitService(context.Background())
	if !errors.Is(err, gdrive.ErrInvalidCredentials) {
		t.Errorf("InitService() error = %v, want ErrInvalidCredentials", err)
	}
}
