package registry
package registry_test

import (
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/registry"
)




























































}	}		t.Errorf("List() returned %d names, want 2", len(names))	if len(names) != 2 {	names := r.List()	r.Register(&mockTool{name: "tool2", description: "Tool 2"})	r.Register(&mockTool{name: "tool1", description: "Tool 1"})	r := registry.New()func TestRegistry_List(t *testing.T) {}	}		t.Error("Get() should return false for nonexistent tool")	if found {	_, found := r.Get("nonexistent")	r := registry.New()func TestRegistry_GetNotFound(t *testing.T) {}	}		t.Errorf("Get() returned tool with name %q, want %q", got.Name(), "test_tool")	if got.Name() != "test_tool" {	}		t.Error("Get() did not find registered tool")	if !found {	got, found := r.Get("test_tool")	r.Register(tool)	tool := &mockTool{name: "test_tool", description: "A test tool"}	r := registry.New()func TestRegistry_RegisterAndGet(t *testing.T) {}	}		t.Error("New() returned nil")	if r == nil {	r := registry.New()func TestRegistry_New(t *testing.T) {}	return "executed", nilfunc (m *mockTool) Execute(params map[string]interface{}) (interface{}, error) {}	return m.descriptionfunc (m *mockTool) Description() string {}	return m.namefunc (m *mockTool) Name() string {}	description string	name        stringtype mockTool struct {// mockTool implements registry.Tool for testing.
