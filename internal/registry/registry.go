// Package registry provides tool registration and lookup functionality.
package registry

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/kevinyay945/macmini-assistant-systray/internal/config"
)

// ErrToolNotFound is returned when a tool is not found in the registry.
var ErrToolNotFound = errors.New("tool not found")

// ErrToolTimeout is returned when a tool execution exceeds the timeout.
var ErrToolTimeout = errors.New("tool execution timed out")

// ErrDuplicateTool is returned when attempting to register a tool with a name that already exists.
var ErrDuplicateTool = errors.New("tool already registered")

// Tool represents a registered tool that can be executed.
type Tool interface {
	Name() string
	Description() string
	Schema() ToolSchema
	Execute(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error)
}

// ToolSchema describes the input/output schema for a tool.
type ToolSchema struct {
	Inputs  []Parameter `json:"inputs"`
	Outputs []Parameter `json:"outputs"`
}

// Parameter describes a tool parameter.
type Parameter struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"` // string, integer, boolean, array
	Required    bool     `json:"required"`
	Default     any      `json:"default,omitempty"`
	Description string   `json:"description"`
	Allowed     []string `json:"allowed,omitempty"` // for enum types
}

// ToolFactory is a function that creates a tool from configuration.
type ToolFactory func(cfg config.ToolConfig) (Tool, error)

// Registry manages registered tools and their configurations.
type Registry struct {
	mu        sync.RWMutex
	tools     map[string]Tool
	factories map[string]ToolFactory
	timeout   time.Duration
}

// Option configures the registry.
type Option func(*Registry)

// WithTimeout sets the default execution timeout for tools.
func WithTimeout(timeout time.Duration) Option {
	return func(r *Registry) {
		r.timeout = timeout
	}
}

// New creates a new tool registry.
func New(opts ...Option) *Registry {
	r := &Registry{
		tools:     make(map[string]Tool),
		factories: make(map[string]ToolFactory),
		timeout:   10 * time.Minute, // default 10 minute timeout
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// RegisterFactory registers a factory function for a tool type.
// The factory will be used by LoadFromConfig to create tools.
func (r *Registry) RegisterFactory(toolType string, factory ToolFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[toolType] = factory
}

// Register adds a tool to the registry.
// Returns ErrDuplicateTool if a tool with the same name is already registered.
func (r *Registry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[tool.Name()]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateTool, tool.Name())
	}
	r.tools[tool.Name()] = tool
	return nil
}

// MustRegister adds a tool to the registry, panicking on error.
func (r *Registry) MustRegister(tool Tool) {
	if err := r.Register(tool); err != nil {
		panic(err)
	}
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

// ListTools returns all registered tools.
func (r *Registry) ListTools() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tools := make([]Tool, 0, len(r.tools))
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		tools = append(tools, r.tools[name])
	}
	return tools
}

// Execute runs a tool with the given parameters, respecting the timeout.
func (r *Registry) Execute(ctx context.Context, name string, params map[string]interface{}) (map[string]interface{}, error) {
	tool, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, name)
	}

	// Validate required parameters
	schema := tool.Schema()
	for _, param := range schema.Inputs {
		if param.Required {
			if _, ok := params[param.Name]; !ok {
				return nil, fmt.Errorf("missing required parameter: %s", param.Name)
			}
		}
	}

	// Apply timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Create result channel
	type result struct {
		output map[string]interface{}
		err    error
	}
	resultCh := make(chan result, 1)

	go func() {
		output, err := tool.Execute(timeoutCtx, params)
		resultCh <- result{output, err}
	}()

	select {
	case <-timeoutCtx.Done():
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: %s after %v", ErrToolTimeout, name, r.timeout)
		}
		return nil, timeoutCtx.Err()
	case res := <-resultCh:
		return res.output, res.err
	}
}

// LoadFromConfig creates and registers tools from configuration.
func (r *Registry) LoadFromConfig(tools []config.ToolConfig) error {
	var errs []error

	for _, toolCfg := range tools {
		if !toolCfg.Enabled {
			continue
		}

		factory, ok := r.factories[toolCfg.Type]
		if !ok {
			errs = append(errs, fmt.Errorf("unknown tool type %q for tool %q", toolCfg.Type, toolCfg.Name))
			continue
		}

		tool, err := factory(toolCfg)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create tool %q: %w", toolCfg.Name, err))
			continue
		}

		if err := r.Register(tool); err != nil {
			errs = append(errs, fmt.Errorf("failed to register tool %q: %w", toolCfg.Name, err))
		}
	}

	return errors.Join(errs...)
}

// Timeout returns the current timeout setting.
func (r *Registry) Timeout() time.Duration {
	return r.timeout
}

// SetTimeout updates the timeout for tool execution.
func (r *Registry) SetTimeout(timeout time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.timeout = timeout
}
