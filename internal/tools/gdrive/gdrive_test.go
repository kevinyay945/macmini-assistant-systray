package gdrive_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kevinyay945/macmini-assistant-systray/internal/tools/gdrive"
)

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

	if len(schema.Inputs) != 3 {
		t.Errorf("Schema().Inputs returned %d params, want 3", len(schema.Inputs))
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
	tool := gdrive.New(gdrive.Config{Enabled: true})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{})
	if !errors.Is(err, gdrive.ErrMissingFilePath) {
		t.Errorf("Execute() error = %v, want ErrMissingFilePath", err)
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
	tool := gdrive.New(gdrive.Config{Enabled: true})
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{"file_path": "/path/to/file.mp4"})
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	if result["status"] != "pending" {
		t.Errorf("Execute() status = %v, want 'pending'", result["status"])
	}
}
