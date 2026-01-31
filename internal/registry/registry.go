// Package registry provides tool registration and lookup functionality.
package registry

// Registry manages registered tools and their configurations.
type Registry struct {
	tools map[string]Tool
}

// Tool represents a registered tool that can be executed.
type Tool interface {
	Name() string
	Description() string
	Execute(params map[string]interface{}) (interface{}, error)
}

// New creates a new tool registry.
func New() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tool names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}
