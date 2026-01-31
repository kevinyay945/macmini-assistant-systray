// Package registry provides tool registration and lookup functionality.
package registry

import (
	"context"
	"errors"
	"fmt"
	"reflect"
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

// ErrInvalidParamType is returned when a parameter has an invalid type.
var ErrInvalidParamType = errors.New("invalid parameter type")

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
	Type        string   `json:"type"` // string, integer, number, boolean, array, object
	Required    bool     `json:"required"`
	Default     any      `json:"default,omitempty"`
	Description string   `json:"description"`
	Allowed     []string `json:"allowed,omitempty"` // Only applicable for Type="string" - enum validation
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

// ErrDuplicateFactory is returned when attempting to register a factory with a type that already exists.
var ErrDuplicateFactory = errors.New("factory already registered")

// RegisterFactory registers a factory function for a tool type.
// The factory will be used by LoadFromConfig to create tools.
// Returns ErrDuplicateFactory if a factory with the same type is already registered.
func (r *Registry) RegisterFactory(toolType string, factory ToolFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.factories[toolType]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateFactory, toolType)
	}
	r.factories[toolType] = factory
	return nil
}

// MustRegisterFactory registers a factory function, panicking on error.
func (r *Registry) MustRegisterFactory(toolType string, factory ToolFactory) {
	if err := r.RegisterFactory(toolType, factory); err != nil {
		panic(err)
	}
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
// IMPORTANT: Tool implementations MUST check ctx.Done() to properly support cancellation.
// Tools that block indefinitely without checking context will cause goroutine leaks.
func (r *Registry) Execute(ctx context.Context, name string, params map[string]interface{}) (map[string]interface{}, error) {
	tool, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, name)
	}

	// Make a copy of params to avoid mutating the original
	execParams := make(map[string]interface{}, len(params))
	for k, v := range params {
		execParams[k] = v
	}

	// Validate and apply defaults for parameters
	schema := tool.Schema()
	for _, param := range schema.Inputs {
		val, exists := execParams[param.Name]
		if !exists {
			if param.Required {
				return nil, fmt.Errorf("missing required parameter: %s", param.Name)
			}
			// Apply default value if available
			if param.Default != nil {
				execParams[param.Name] = param.Default
			}
			continue
		}
		// Validate parameter type
		if err := validateParamType(val, param.Type, param.Allowed); err != nil {
			return nil, fmt.Errorf("%w for parameter %s: %w", ErrInvalidParamType, param.Name, err)
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
		output, err := tool.Execute(timeoutCtx, execParams)
		select {
		case resultCh <- result{output, err}:
		case <-timeoutCtx.Done():
			// Context cancelled, discard result and exit
		}
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

// validateParamType validates that a value matches the expected parameter type.
// If allowed is non-empty and the value is a string, it also validates against allowed values.
func validateParamType(val interface{}, expectedType string, allowed []string) error {
	if val == nil {
		return nil // nil is acceptable for optional params
	}

	switch expectedType {
	case "string":
		strVal, ok := val.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", val)
		}
		// Validate against allowed values if specified
		if len(allowed) > 0 {
			for _, a := range allowed {
				if strVal == a {
					return nil
				}
			}
			return fmt.Errorf("value %q not in allowed values: %v", strVal, allowed)
		}
	case "integer":
		// Handle all integer types including unsigned, plus float64 (JSON unmarshal produces float64 for numbers)
		switch v := val.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			// ok
		case float64:
			// Check if it's actually an integer value
			if v != float64(int64(v)) {
				return fmt.Errorf("expected integer, got float %v", v)
			}
		default:
			return fmt.Errorf("expected integer, got %T", val)
		}
	case "number":
		// Handle any numeric type (int, uint, or float)
		switch val.(type) {
		case float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			// ok
		default:
			return fmt.Errorf("expected number, got %T", val)
		}
	case "boolean":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", val)
		}
	case "array":
		rv := reflect.ValueOf(val)
		if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
			return fmt.Errorf("expected array, got %T", val)
		}
	case "object":
		if _, ok := val.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", val)
		}
	default:
		// Unknown type, skip validation
	}
	return nil
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
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.timeout
}

// SetTimeout updates the timeout for tool execution.
func (r *Registry) SetTimeout(timeout time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.timeout = timeout
}
