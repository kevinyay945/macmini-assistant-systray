package gdrive_test

import (
	"context"
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

func TestTool_Execute_NotEnabled(t *testing.T) {
	tool := gdrive.New(gdrive.Config{Enabled: false})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{"file_path": "/path/to/file"})
	if err == nil {
		t.Error("Execute() should return error when tool is not enabled")
	}
}

func TestTool_Execute_MissingFilePath(t *testing.T) {
	tool := gdrive.New(gdrive.Config{Enabled: true})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{})
	if err == nil {
		t.Error("Execute() should return error when file_path is missing")
	}
}

func TestTool_Execute_ContextCanceled(t *testing.T) {
	tool := gdrive.New(gdrive.Config{Enabled: true})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := tool.Execute(ctx, map[string]interface{}{"file_path": "/path/to/file"})
	if err == nil {
		t.Error("Execute() should return error when context is canceled")
	}
}

func TestTool_Execute_ContextTimeout(t *testing.T) {
	tool := gdrive.New(gdrive.Config{Enabled: true})
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond)

	_, err := tool.Execute(ctx, map[string]interface{}{"file_path": "/path/to/file"})
	if err == nil {
		t.Error("Execute() should return error when context times out")
	}
}

func TestTool_Execute_ValidRequest(t *testing.T) {
	tool := gdrive.New(gdrive.Config{Enabled: true})
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{"file_path": "/path/to/file.mp4"})
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Execute() result is not a map")
	}

	if resultMap["status"] != "pending" {
		t.Errorf("Execute() status = %v, want 'pending'", resultMap["status"])
	}
}
