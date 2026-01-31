package registry_test

import (
	"context"
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/registry"
)

// mockTool implements registry.Tool for testing.
type mockTool struct {
	name        string
	description string
}

func (m *mockTool) Name() string {
	return m.name
}

func (m *mockTool) Description() string {
	return m.description
}

func (m *mockTool) Execute(_ context.Context, _ map[string]interface{}) (interface{}, error) {
	return "executed", nil
}

func TestRegistry_New(t *testing.T) {
	r := registry.New()
	if r == nil {
		t.Error("New() returned nil")
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := registry.New()
	tool := &mockTool{name: "test_tool", description: "A test tool"}
	r.Register(tool)

	got, found := r.Get("test_tool")
	if !found {
		t.Error("Get() did not find registered tool")
	}
	if got.Name() != "test_tool" {
		t.Errorf("Get() returned tool with name %q, want %q", got.Name(), "test_tool")
	}
}

func TestRegistry_GetNotFound(t *testing.T) {
	r := registry.New()
	_, found := r.Get("nonexistent")
	if found {
		t.Error("Get() should return false for nonexistent tool")
	}
}

func TestRegistry_List(t *testing.T) {
	r := registry.New()
	r.Register(&mockTool{name: "tool1", description: "Tool 1"})
	r.Register(&mockTool{name: "tool2", description: "Tool 2"})

	names := r.List()
	if len(names) != 2 {
		t.Errorf("List() returned %d names, want 2", len(names))
	}
}
