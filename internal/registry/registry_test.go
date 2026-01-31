package registry_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kevinyay945/macmini-assistant-systray/internal/config"
	"github.com/kevinyay945/macmini-assistant-systray/internal/registry"
)

// mockTool implements registry.Tool for testing.
type mockTool struct {
	name        string
	description string
	schema      registry.ToolSchema
	executeFunc func(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error)
}

func (m *mockTool) Name() string {
	return m.name
}

func (m *mockTool) Description() string {
	return m.description
}

func (m *mockTool) Schema() registry.ToolSchema {
	if m.schema.Inputs == nil && m.schema.Outputs == nil {
		return registry.ToolSchema{
			Inputs: []registry.Parameter{
				{Name: "test_param", Type: "string", Required: true, Description: "A test parameter"},
			},
		}
	}
	return m.schema
}

func (m *mockTool) Execute(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, params)
	}
	return map[string]interface{}{"result": "executed"}, nil
}

func TestRegistry_New(t *testing.T) {
	r := registry.New()
	if r == nil {
		t.Error("New() returned nil")
	}
}

func TestRegistry_NewWithTimeout(t *testing.T) {
	r := registry.New(registry.WithTimeout(5 * time.Second))
	if r == nil {
		t.Error("New() returned nil")
	}
	if r.Timeout() != 5*time.Second {
		t.Errorf("Timeout() = %v, want 5s", r.Timeout())
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := registry.New()
	tool := &mockTool{name: "test_tool", description: "A test tool"}
	err := r.Register(tool)
	if err != nil {
		t.Errorf("Register() returned error: %v", err)
	}

	got, found := r.Get("test_tool")
	if !found {
		t.Error("Get() did not find registered tool")
	}
	if got.Name() != "test_tool" {
		t.Errorf("Get() returned tool with name %q, want %q", got.Name(), "test_tool")
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	r := registry.New()
	tool1 := &mockTool{name: "test_tool", description: "First tool"}
	tool2 := &mockTool{name: "test_tool", description: "Second tool"}

	if err := r.Register(tool1); err != nil {
		t.Fatalf("First Register() returned error: %v", err)
	}

	err := r.Register(tool2)
	if err == nil {
		t.Error("Second Register() should return error for duplicate name")
	}
	if !errors.Is(err, registry.ErrDuplicateTool) {
		t.Errorf("Expected ErrDuplicateTool, got: %v", err)
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	r := registry.New()
	_, found := r.Get("nonexistent")
	if found {
		t.Error("Get() should return false for nonexistent tool")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	r := registry.New()
	r.MustRegister(&mockTool{name: "test_tool", description: "A test tool"})

	// Unregister existing tool
	removed := r.Unregister("test_tool")
	if !removed {
		t.Error("Unregister() should return true for existing tool")
	}

	// Verify it's gone
	_, found := r.Get("test_tool")
	if found {
		t.Error("Get() should return false after Unregister()")
	}
}

func TestRegistry_Unregister_NotFound(t *testing.T) {
	r := registry.New()

	removed := r.Unregister("nonexistent")
	if removed {
		t.Error("Unregister() should return false for nonexistent tool")
	}
}

func TestRegistry_List(t *testing.T) {
	r := registry.New()
	r.MustRegister(&mockTool{name: "tool1", description: "Tool 1"})
	r.MustRegister(&mockTool{name: "tool2", description: "Tool 2"})

	names := r.List()
	if len(names) != 2 {
		t.Errorf("List() returned %d names, want 2", len(names))
	}
}

func TestRegistry_List_IsSorted(t *testing.T) {
	r := registry.New()
	// Register in reverse order
	r.MustRegister(&mockTool{name: "zebra", description: "Z tool"})
	r.MustRegister(&mockTool{name: "alpha", description: "A tool"})
	r.MustRegister(&mockTool{name: "beta", description: "B tool"})

	names := r.List()
	expected := []string{"alpha", "beta", "zebra"}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("List()[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

func TestRegistry_ListTools(t *testing.T) {
	r := registry.New()
	r.MustRegister(&mockTool{name: "tool1", description: "Tool 1"})
	r.MustRegister(&mockTool{name: "tool2", description: "Tool 2"})

	tools := r.ListTools()
	if len(tools) != 2 {
		t.Errorf("ListTools() returned %d tools, want 2", len(tools))
	}
}

func TestRegistry_Execute(t *testing.T) {
	r := registry.New()
	r.MustRegister(&mockTool{
		name: "test_tool",
		executeFunc: func(_ context.Context, params map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{"result": params["test_param"]}, nil
		},
	})

	result, err := r.Execute(context.Background(), "test_tool", map[string]interface{}{
		"test_param": "hello",
	})
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	if result["result"] != "hello" {
		t.Errorf("Execute() result = %v, want 'hello'", result["result"])
	}
}

func TestRegistry_Execute_NotFound(t *testing.T) {
	r := registry.New()

	_, err := r.Execute(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Error("Execute() should return error for nonexistent tool")
	}
	if !errors.Is(err, registry.ErrToolNotFound) {
		t.Errorf("Expected ErrToolNotFound, got: %v", err)
	}
}

func TestRegistry_Execute_MissingRequiredParam(t *testing.T) {
	r := registry.New()
	r.MustRegister(&mockTool{
		name: "test_tool",
		schema: registry.ToolSchema{
			Inputs: []registry.Parameter{
				{Name: "required_param", Type: "string", Required: true},
			},
		},
	})

	_, err := r.Execute(context.Background(), "test_tool", map[string]interface{}{})
	if err == nil {
		t.Error("Execute() should return error for missing required parameter")
	}
}

func TestRegistry_Execute_WithTimeout(t *testing.T) {
	r := registry.New(registry.WithTimeout(50 * time.Millisecond))
	r.MustRegister(&mockTool{
		name: "slow_tool",
		schema: registry.ToolSchema{
			Inputs: []registry.Parameter{}, // no required params
		},
		executeFunc: func(ctx context.Context, _ map[string]interface{}) (map[string]interface{}, error) {
			select {
			case <-time.After(200 * time.Millisecond):
				return map[string]interface{}{"result": "done"}, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	})

	_, err := r.Execute(context.Background(), "slow_tool", map[string]interface{}{})
	if err == nil {
		t.Error("Execute() should return error for timeout")
	}
	if !errors.Is(err, registry.ErrToolTimeout) {
		t.Errorf("Expected ErrToolTimeout, got: %v", err)
	}
}

func TestRegistry_LoadFromConfig(t *testing.T) {
	r := registry.New()

	// Register a factory
	r.RegisterFactory("test_type", func(cfg config.ToolConfig) (registry.Tool, error) {
		return &mockTool{
			name:        cfg.Name,
			description: "Created from config",
		}, nil
	})

	tools := []config.ToolConfig{
		{Name: "tool1", Type: "test_type", Enabled: true},
		{Name: "tool2", Type: "test_type", Enabled: true},
		{Name: "disabled_tool", Type: "test_type", Enabled: false},
	}

	err := r.LoadFromConfig(tools)
	if err != nil {
		t.Fatalf("LoadFromConfig() returned error: %v", err)
	}

	names := r.List()
	if len(names) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(names))
	}
}

func TestRegistry_LoadFromConfig_UnknownType(t *testing.T) {
	r := registry.New()

	tools := []config.ToolConfig{
		{Name: "tool1", Type: "unknown_type", Enabled: true},
	}

	err := r.LoadFromConfig(tools)
	if err == nil {
		t.Error("LoadFromConfig() should return error for unknown type")
	}
}

func TestRegistry_LoadFromConfig_FactoryError(t *testing.T) {
	r := registry.New()

	r.RegisterFactory("failing_type", func(_ config.ToolConfig) (registry.Tool, error) {
		return nil, errors.New("factory error")
	})

	tools := []config.ToolConfig{
		{Name: "tool1", Type: "failing_type", Enabled: true},
	}

	err := r.LoadFromConfig(tools)
	if err == nil {
		t.Error("LoadFromConfig() should return error when factory fails")
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := registry.New()
	var wg sync.WaitGroup
	const numGoroutines = 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = r.Register(&mockTool{
				name:        fmt.Sprintf("tool_%d", n),
				description: "A concurrent tool",
			})
		}(i)
	}

	// Concurrent reads while writing
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			r.Get(fmt.Sprintf("tool_%d", n))
			r.List()
		}(i)
	}

	wg.Wait()

	names := r.List()
	if len(names) != numGoroutines {
		t.Errorf("Expected %d tools, got %d", numGoroutines, len(names))
	}
}

func TestRegistry_SetTimeout(t *testing.T) {
	r := registry.New()

	r.SetTimeout(30 * time.Second)
	if r.Timeout() != 30*time.Second {
		t.Errorf("Timeout() = %v, want 30s", r.Timeout())
	}
}

func TestToolSchema_Validation(t *testing.T) {
	schema := registry.ToolSchema{
		Inputs: []registry.Parameter{
			{Name: "url", Type: "string", Required: true, Description: "The URL to download"},
			{Name: "format", Type: "string", Required: false, Default: "mp4", Allowed: []string{"mp4", "mkv", "avi"}},
		},
		Outputs: []registry.Parameter{
			{Name: "file_path", Type: "string", Required: true, Description: "Path to downloaded file"},
		},
	}

	if len(schema.Inputs) != 2 {
		t.Errorf("Expected 2 inputs, got %d", len(schema.Inputs))
	}
	if len(schema.Outputs) != 1 {
		t.Errorf("Expected 1 output, got %d", len(schema.Outputs))
	}
}
