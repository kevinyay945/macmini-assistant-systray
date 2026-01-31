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

// Test timing constants - using multiples to ensure reliable timeout behavior.
const (
	testShortTimeout  = 50 * time.Millisecond  // Short timeout for testing
	testLongOperation = 200 * time.Millisecond // 4x timeout to ensure timeout triggers reliably
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
	r := registry.New(registry.WithTimeout(testShortTimeout))
	r.MustRegister(&mockTool{
		name: "slow_tool",
		schema: registry.ToolSchema{
			Inputs: []registry.Parameter{}, // no required params
		},
		executeFunc: func(ctx context.Context, _ map[string]interface{}) (map[string]interface{}, error) {
			select {
			case <-time.After(testLongOperation):
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
	err := r.RegisterFactory("test_type", func(cfg config.ToolConfig) (registry.Tool, error) {
		return &mockTool{
			name:        cfg.Name,
			description: "Created from config",
		}, nil
	})
	if err != nil {
		t.Fatalf("RegisterFactory() returned error: %v", err)
	}

	tools := []config.ToolConfig{
		{Name: "tool1", Type: "test_type", Enabled: true},
		{Name: "tool2", Type: "test_type", Enabled: true},
		{Name: "disabled_tool", Type: "test_type", Enabled: false},
	}

	err = r.LoadFromConfig(tools)
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

	err := r.RegisterFactory("failing_type", func(_ config.ToolConfig) (registry.Tool, error) {
		return nil, errors.New("factory error")
	})
	if err != nil {
		t.Fatalf("RegisterFactory() returned error: %v", err)
	}

	tools := []config.ToolConfig{
		{Name: "tool1", Type: "failing_type", Enabled: true},
	}

	err = r.LoadFromConfig(tools)
	if err == nil {
		t.Error("LoadFromConfig() should return error when factory fails")
	}
}

func TestRegistry_RegisterFactory_Duplicate(t *testing.T) {
	r := registry.New()

	factory := func(_ config.ToolConfig) (registry.Tool, error) {
		return &mockTool{name: "test"}, nil
	}

	if err := r.RegisterFactory("test_type", factory); err != nil {
		t.Fatalf("First RegisterFactory() returned error: %v", err)
	}

	err := r.RegisterFactory("test_type", factory)
	if err == nil {
		t.Error("Second RegisterFactory() should return error for duplicate type")
	}
	if !errors.Is(err, registry.ErrDuplicateFactory) {
		t.Errorf("Expected ErrDuplicateFactory, got: %v", err)
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

func TestRegistry_Execute_InvalidParamType(t *testing.T) {
	r := registry.New()
	r.MustRegister(&mockTool{
		name: "test_tool",
		schema: registry.ToolSchema{
			Inputs: []registry.Parameter{
				{Name: "count", Type: "integer", Required: true},
			},
		},
	})

	// Pass a string instead of integer
	_, err := r.Execute(context.Background(), "test_tool", map[string]interface{}{
		"count": "not-an-integer",
	})
	if err == nil {
		t.Error("Execute() should return error for invalid parameter type")
	}
	if !errors.Is(err, registry.ErrInvalidParamType) {
		t.Errorf("Expected ErrInvalidParamType, got: %v", err)
	}
}

func TestRegistry_Execute_ValidParamTypes(t *testing.T) {
	tests := []struct {
		name      string
		paramType string
		value     interface{}
	}{
		{"string", "string", "hello"},
		{"integer_int", "integer", 42},
		{"integer_int64", "integer", int64(42)},
		{"integer_float64", "integer", float64(42)}, // JSON numbers
		{"boolean", "boolean", true},
		{"array_interface", "array", []interface{}{"a", "b"}},
		{"array_string", "array", []string{"a", "b"}},
		{"array_bool", "array", []bool{true, false}},
		{"array_int", "array", []int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := registry.New()
			r.MustRegister(&mockTool{
				name: "test_tool",
				schema: registry.ToolSchema{
					Inputs: []registry.Parameter{
						{Name: "param", Type: tt.paramType, Required: true},
					},
				},
			})

			_, err := r.Execute(context.Background(), "test_tool", map[string]interface{}{
				"param": tt.value,
			})
			if err != nil {
				t.Errorf("Execute() returned error for valid %s: %v", tt.paramType, err)
			}
		})
	}
}

func TestRegistry_Execute_DefaultValues(t *testing.T) {
	r := registry.New()

	var receivedParams map[string]interface{}
	r.MustRegister(&mockTool{
		name: "test_tool",
		schema: registry.ToolSchema{
			Inputs: []registry.Parameter{
				{Name: "required_param", Type: "string", Required: true},
				{Name: "optional_with_default", Type: "string", Required: false, Default: "default_value"},
				{Name: "optional_no_default", Type: "string", Required: false},
			},
		},
		executeFunc: func(_ context.Context, params map[string]interface{}) (map[string]interface{}, error) {
			receivedParams = params
			return map[string]interface{}{"status": "ok"}, nil
		},
	})

	_, err := r.Execute(context.Background(), "test_tool", map[string]interface{}{
		"required_param": "value",
	})
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	// Check that default value was applied
	if receivedParams["optional_with_default"] != "default_value" {
		t.Errorf("Default value not applied, got: %v", receivedParams["optional_with_default"])
	}

	// Check that optional without default is not in params
	if _, exists := receivedParams["optional_no_default"]; exists {
		t.Error("Optional param without default should not be in params")
	}
}

func TestRegistry_Execute_DoesNotMutateOriginalParams(t *testing.T) {
	r := registry.New()
	r.MustRegister(&mockTool{
		name: "test_tool",
		schema: registry.ToolSchema{
			Inputs: []registry.Parameter{
				{Name: "required_param", Type: "string", Required: true},
				{Name: "optional_with_default", Type: "string", Required: false, Default: "default_value"},
			},
		},
	})

	originalParams := map[string]interface{}{
		"required_param": "value",
	}

	_, err := r.Execute(context.Background(), "test_tool", originalParams)
	if err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	// Check that original params were not modified
	if _, exists := originalParams["optional_with_default"]; exists {
		t.Error("Execute() should not mutate original params map")
	}
}

func TestRegistry_Execute_NumberType(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"float64", float64(3.14)},
		{"float32", float32(3.14)},
		{"int", 42},
		{"int64", int64(42)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := registry.New()
			r.MustRegister(&mockTool{
				name: "test_tool",
				schema: registry.ToolSchema{
					Inputs: []registry.Parameter{
						{Name: "param", Type: "number", Required: true},
					},
				},
			})

			_, err := r.Execute(context.Background(), "test_tool", map[string]interface{}{
				"param": tt.value,
			})
			if err != nil {
				t.Errorf("Execute() returned error for valid number %v: %v", tt.value, err)
			}
		})
	}
}

func TestRegistry_Execute_NumberType_Invalid(t *testing.T) {
	r := registry.New()
	r.MustRegister(&mockTool{
		name: "test_tool",
		schema: registry.ToolSchema{
			Inputs: []registry.Parameter{
				{Name: "param", Type: "number", Required: true},
			},
		},
	})

	_, err := r.Execute(context.Background(), "test_tool", map[string]interface{}{
		"param": "not-a-number",
	})
	if err == nil {
		t.Error("Execute() should return error for invalid number type")
	}
	if !errors.Is(err, registry.ErrInvalidParamType) {
		t.Errorf("Expected ErrInvalidParamType, got: %v", err)
	}
}

func TestRegistry_Execute_AllowedValues(t *testing.T) {
	r := registry.New()
	r.MustRegister(&mockTool{
		name: "test_tool",
		schema: registry.ToolSchema{
			Inputs: []registry.Parameter{
				{Name: "format", Type: "string", Required: true, Allowed: []string{"mp4", "mkv", "avi"}},
			},
		},
	})

	// Valid value
	_, err := r.Execute(context.Background(), "test_tool", map[string]interface{}{
		"format": "mp4",
	})
	if err != nil {
		t.Errorf("Execute() returned error for valid allowed value: %v", err)
	}

	// Invalid value
	_, err = r.Execute(context.Background(), "test_tool", map[string]interface{}{
		"format": "invalid_format",
	})
	if err == nil {
		t.Error("Execute() should return error for value not in allowed list")
	}
	if !errors.Is(err, registry.ErrInvalidParamType) {
		t.Errorf("Expected ErrInvalidParamType, got: %v", err)
	}
}
