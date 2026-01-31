// Package registry provides tool registration and lookup functionality.
package registry

import (
	"context"
	"slices"
	"sync"
)

// Registry manages registered tools and their configurations.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// Tool represents a registered tool that can be executed.
type Tool interface {
	Name() string
	Description() string
	Parameters() []ParameterDef
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
}

// ParameterDef describes a tool parameter for LLM integration.
type ParameterDef struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "int", "bool"
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     any    `json:"default,omitempty"`
}

// New creates a new tool registry.
func New() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (r *Registry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

// Unregister removes a tool from the registry.
// Returns true if the tool was found and removed, false otherwise.
func (r *Registry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.tools[name]
	if exists {
		delete(r.tools, name)
	}
	return exists
}

// List returns all registered tool names in sorted order for deterministic output.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
