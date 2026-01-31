package registry_test

import (
	"context"
	"fmt"
	"sync"
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

func (m *mockTool) Parameters() []registry.ParameterDef {
	return []registry.ParameterDef{
		{Name: "test_param", Type: "string", Required: true, Description: "A test parameter"},
	}
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

func TestRegistry_Unregister(t *testing.T) {
	r := registry.New()
	r.Register(&mockTool{name: "test_tool", description: "A test tool"})

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
	r.Register(&mockTool{name: "tool1", description: "Tool 1"})
	r.Register(&mockTool{name: "tool2", description: "Tool 2"})

	names := r.List()
	if len(names) != 2 {
		t.Errorf("List() returned %d names, want 2", len(names))
	}
}

func TestRegistry_List_IsSorted(t *testing.T) {
	r := registry.New()
	// Register in reverse order
	r.Register(&mockTool{name: "zebra", description: "Z tool"})
	r.Register(&mockTool{name: "alpha", description: "A tool"})
	r.Register(&mockTool{name: "beta", description: "B tool"})

	names := r.List()
	expected := []string{"alpha", "beta", "zebra"}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("List()[%d] = %q, want %q", i, name, expected[i])
		}
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
			r.Register(&mockTool{
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
