// Package copilot provides integration with GitHub Copilot SDK.
package copilot

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	copilot "github.com/github/copilot-sdk/go"

	"github.com/kevinyay945/macmini-assistant-systray/internal/observability"
	"github.com/kevinyay945/macmini-assistant-systray/internal/registry"
)

// Default timeout for Copilot SDK operations (10 minutes per PRD).
const DefaultTimeout = 10 * time.Minute

// Sentinel errors for the Copilot client.
var (
	ErrAPIKeyNotConfigured = errors.New("copilot API key not configured")
	ErrClientNotStarted    = errors.New("copilot client not started")
	ErrSessionNotCreated   = errors.New("copilot session not created")
)

// Response represents the response from Copilot after processing a message.
type Response struct {
	// Text is the main text response from the LLM.
	Text string
	// Data contains any tool execution results.
	Data map[string]interface{}
	// ToolName is the name of the tool that was executed (if any).
	ToolName string
}

// Client handles communication with the Copilot SDK.
type Client struct {
	apiKey   string
	registry *registry.Registry
	timeout  time.Duration
	logger   *observability.Logger

	sdk     *copilot.Client
	mu      sync.RWMutex
	started bool
}

// Config holds Copilot client configuration.
type Config struct {
	APIKey   string               `yaml:"api_key" json:"api_key"`
	Registry *registry.Registry   `yaml:"-" json:"-"`
	Timeout  time.Duration        `yaml:"timeout" json:"timeout"`
	Logger   *observability.Logger `yaml:"-" json:"-"`
}

// NewClient creates a new Copilot client with the given configuration.
// If timeout is 0, it defaults to 10 minutes.
func NewClient(cfg Config) (*Client, error) {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	logger := cfg.Logger
	if logger == nil {
		logger = observability.New(observability.WithLevel(observability.LevelInfo))
	}

	return &Client{
		apiKey:   cfg.APIKey,
		registry: cfg.Registry,
		timeout:  timeout,
		logger:   logger.With("component", "copilot"),
	}, nil
}

// New creates a new Copilot client (backwards-compatible constructor).
func New(cfg Config) *Client {
	client, _ := NewClient(cfg)
	return client
}

// Start initializes the Copilot SDK client.
func (c *Client) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return nil
	}

	// Create the Copilot SDK client
	opts := &copilot.ClientOptions{
		UseStdio:  true,
		AutoStart: copilot.Bool(true),
		LogLevel:  "error",
	}

	// Set GitHub token if API key is provided
	if c.apiKey != "" {
		opts.GithubToken = c.apiKey
	}

	c.sdk = copilot.NewClient(opts)

	if err := c.sdk.Start(); err != nil {
		return fmt.Errorf("failed to start Copilot SDK: %w", err)
	}

	c.started = true
	c.logger.Info(context.Background(), "Copilot client started")

	return nil
}

// Stop gracefully shuts down the Copilot SDK client.
func (c *Client) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		return nil
	}

	if c.sdk != nil {
		c.sdk.Stop()
		c.sdk = nil
	}

	c.started = false
	c.logger.Info(context.Background(), "Copilot client stopped")

	return nil
}

// RegisterTools registers all tools from the registry with the Copilot SDK.
// This converts internal tool schemas to Copilot SDK format.
func (c *Client) RegisterTools() ([]copilot.Tool, error) {
	if c.registry == nil {
		return nil, nil
	}

	tools := c.registry.ListTools()
	copilotTools := make([]copilot.Tool, 0, len(tools))

	for _, tool := range tools {
		copilotTool := c.createCopilotTool(tool)
		copilotTools = append(copilotTools, copilotTool)
		c.logger.Info(context.Background(), "registered tool with Copilot",
			"tool", tool.Name(),
		)
	}

	return copilotTools, nil
}

// createCopilotTool converts an internal Tool to a Copilot SDK Tool.
func (c *Client) createCopilotTool(tool registry.Tool) copilot.Tool {
	schema := tool.Schema()

	// Convert internal schema to Copilot SDK parameters format
	parameters := convertSchemaToParameters(schema)

	// Create the Copilot tool with a handler that delegates to our internal tool
	return copilot.Tool{
		Name:        tool.Name(),
		Description: tool.Description(),
		Parameters:  parameters,
		Handler:     c.createToolHandler(tool),
	}
}

// convertSchemaToParameters converts internal ToolSchema to Copilot SDK parameter format.
func convertSchemaToParameters(schema registry.ToolSchema) map[string]interface{} {
	properties := make(map[string]interface{})
	required := make([]string, 0)

	for _, input := range schema.Inputs {
		prop := map[string]interface{}{
			"type":        input.Type,
			"description": input.Description,
		}
		if len(input.Allowed) > 0 {
			prop["enum"] = input.Allowed
		}
		properties[input.Name] = prop

		if input.Required {
			required = append(required, input.Name)
		}
	}

	params := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		params["required"] = required
	}

	return params
}

// createToolHandler creates a Copilot SDK tool handler that delegates to our internal tool.
func (c *Client) createToolHandler(tool registry.Tool) copilot.ToolHandler {
	return func(inv copilot.ToolInvocation) (copilot.ToolResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
		defer cancel()

		c.logger.Info(ctx, "executing tool via Copilot SDK",
			"tool", tool.Name(),
			"tool_call_id", inv.ToolCallID,
		)

		// Convert arguments to map[string]interface{}
		params, ok := inv.Arguments.(map[string]interface{})
		if !ok {
			return copilot.ToolResult{
				TextResultForLLM: "Invalid arguments format",
				ResultType:       "error",
				Error:            "arguments must be an object",
			}, nil
		}

		// Execute the tool
		result, err := tool.Execute(ctx, params)
		if err != nil {
			c.logger.Error(ctx, "tool execution failed",
				"tool", tool.Name(),
				"error", err,
			)

			// Check for timeout
			if errors.Is(err, context.DeadlineExceeded) {
				return copilot.ToolResult{
					TextResultForLLM: "Tool execution timed out",
					ResultType:       "error",
					Error:            err.Error(),
				}, nil
			}

			return copilot.ToolResult{
				TextResultForLLM: fmt.Sprintf("Tool execution failed: %s", err.Error()),
				ResultType:       "error",
				Error:            err.Error(),
			}, nil
		}

		// Format result as text for LLM
		resultText := formatToolResult(result)

		c.logger.Info(ctx, "tool execution completed",
			"tool", tool.Name(),
		)

		return copilot.ToolResult{
			TextResultForLLM: resultText,
			ResultType:       "success",
		}, nil
	}
}

// formatToolResult converts a tool result map to a string for the LLM.
func formatToolResult(result map[string]interface{}) string {
	if result == nil {
		return "Operation completed successfully"
	}

	// Try to get a "message" or "result" field first
	if msg, ok := result["message"].(string); ok {
		return msg
	}
	if res, ok := result["result"].(string); ok {
		return res
	}

	// Format as key-value pairs
	var resultText string
	for k, v := range result {
		resultText += fmt.Sprintf("%s: %v\n", k, v)
	}

	if resultText == "" {
		return "Operation completed successfully"
	}

	return resultText
}

// ProcessMessage sends a message to Copilot and returns the response.
// The context is used to enforce timeouts (10-minute hard limit per PRD).
func (c *Client) ProcessMessage(ctx context.Context, message string) (string, error) {
	// Context check should be first to fail fast
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if c.apiKey == "" {
		return "", ErrAPIKeyNotConfigured
	}

	c.mu.RLock()
	started := c.started
	sdk := c.sdk
	c.mu.RUnlock()

	if !started || sdk == nil {
		return "", ErrClientNotStarted
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Create a session for this request
	tools, _ := c.RegisterTools()
	session, err := sdk.CreateSession(&copilot.SessionConfig{
		Tools:     tools,
		Streaming: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Copilot session: %w", err)
	}
	defer session.Destroy()

	// Collect response
	responseCh := make(chan string, 1)
	errCh := make(chan error, 1)

	session.On(func(event copilot.SessionEvent) {
		if event.Type == "assistant.message" && event.Data.Content != nil {
			select {
			case responseCh <- *event.Data.Content:
			default:
			}
		}
		if event.Type == "session.idle" {
			select {
			case responseCh <- "":
			default:
			}
		}
	})

	// Send the message
	_, err = session.Send(copilot.MessageOptions{
		Prompt: message,
	})
	if err != nil {
		return "", fmt.Errorf("failed to send message to Copilot: %w", err)
	}

	// Wait for response or timeout
	select {
	case <-timeoutCtx.Done():
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return "", observability.ErrToolTimeout.WithCause(timeoutCtx.Err())
		}
		return "", timeoutCtx.Err()
	case err := <-errCh:
		return "", err
	case response := <-responseCh:
		return response, nil
	}
}

// ProcessMessageWithUserID sends a message to Copilot with user context and returns the response.
// This method provides more context to Copilot about the user making the request.
func (c *Client) ProcessMessageWithUserID(ctx context.Context, message, userID string) (*Response, error) {
	// Context check should be first to fail fast
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if c.apiKey == "" {
		return nil, ErrAPIKeyNotConfigured
	}

	c.mu.RLock()
	started := c.started
	sdk := c.sdk
	c.mu.RUnlock()

	if !started || sdk == nil {
		return nil, ErrClientNotStarted
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Create a session for this request with user-specific ID
	tools, _ := c.RegisterTools()
	sessionID := fmt.Sprintf("user-%s-%d", userID, time.Now().UnixNano())
	session, err := sdk.CreateSession(&copilot.SessionConfig{
		SessionID: sessionID,
		Tools:     tools,
		Streaming: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Copilot session: %w", err)
	}
	defer session.Destroy()

	// Collect response and tool results
	var response Response
	responseCh := make(chan struct{}, 1)

	session.On(func(event copilot.SessionEvent) {
		switch event.Type {
		case "assistant.message":
			if event.Data.Content != nil {
				response.Text = *event.Data.Content
			}
		case "session.idle":
			select {
			case responseCh <- struct{}{}:
			default:
			}
		}
	})

	// Send the message
	_, err = session.Send(copilot.MessageOptions{
		Prompt: message,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send message to Copilot: %w", err)
	}

	// Wait for response or timeout
	select {
	case <-timeoutCtx.Done():
		if errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
			return nil, observability.ErrToolTimeout.WithCause(timeoutCtx.Err())
		}
		return nil, timeoutCtx.Err()
	case <-responseCh:
		return &response, nil
	}
}

// IsStarted returns whether the client has been started.
func (c *Client) IsStarted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.started
}

// Timeout returns the configured timeout.
func (c *Client) Timeout() time.Duration {
	return c.timeout
}
