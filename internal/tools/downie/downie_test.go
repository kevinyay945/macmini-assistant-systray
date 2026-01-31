package downie_test

import (
	"context"
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

func TestTool_Execute_NotEnabled(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: false})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{"url": "https://example.com"})
	if err == nil {
		t.Error("Execute() should return error when tool is not enabled")
	}
}

func TestTool_Execute_MissingURL(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{})
	if err == nil {
		t.Error("Execute() should return error when URL is missing")
	}
}

func TestTool_Execute_ContextCanceled(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := tool.Execute(ctx, map[string]interface{}{"url": "https://example.com"})
	if err == nil {
		t.Error("Execute() should return error when context is canceled")
	}
}

func TestTool_Execute_ContextTimeout(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true})
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(1 * time.Millisecond)

	_, err := tool.Execute(ctx, map[string]interface{}{"url": "https://example.com"})
	if err == nil {
		t.Error("Execute() should return error when context times out")
	}
}

func TestTool_Execute_ValidRequest(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true})
	ctx := context.Background()

	result, err := tool.Execute(ctx, map[string]interface{}{"url": "https://example.com/video"})
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
