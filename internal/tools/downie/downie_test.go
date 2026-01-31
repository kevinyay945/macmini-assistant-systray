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

func TestTool_Parameters(t *testing.T) {
	tool := downie.New(downie.Config{})
	params := tool.Parameters()

	if len(params) != 3 {
		t.Errorf("Parameters() returned %d params, want 3", len(params))
	}

	// Check required URL parameter
	urlParam := params[0]
	if urlParam.Name != "url" {
		t.Errorf("First param name = %q, want 'url'", urlParam.Name)
	}
	if !urlParam.Required {
		t.Error("URL parameter should be required")
	}

	// Check optional parameters have defaults
	formatParam := params[1]
	if formatParam.Required {
		t.Error("format parameter should not be required")
	}
	if formatParam.Default != "mp4" {
		t.Errorf("format default = %v, want 'mp4'", formatParam.Default)
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
	tool := downie.New(downie.Config{Enabled: true})
	ctx := context.Background()

	_, err := tool.Execute(ctx, map[string]interface{}{})
	if !errors.Is(err, downie.ErrMissingURL) {
		t.Errorf("Execute() error = %v, want ErrMissingURL", err)
	}
}

func TestTool_Execute_ContextCanceled(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := tool.Execute(ctx, map[string]interface{}{"url": "https://example.com"})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() error = %v, want context.Canceled", err)
	}
}

func TestTool_Execute_ContextDeadlineExceeded(t *testing.T) {
	tool := downie.New(downie.Config{Enabled: true})
	// Use an already-expired deadline to avoid flaky race conditions
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := tool.Execute(ctx, map[string]interface{}{"url": "https://example.com"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Execute() error = %v, want context.DeadlineExceeded", err)
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
