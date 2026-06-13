# Phase 3: GitHub Copilot SDK Integration

**Duration**: Week 6
**Status**: ✅ Complete
**Goal**: Connect orchestrator to Copilot SDK for AI-powered tool routing

---

## Overview

This phase integrates the GitHub Copilot SDK to provide AI-powered intent recognition and tool routing. The SDK will analyze user messages and determine which tools to invoke.

---

## 3.1 Copilot SDK Client

**Duration**: 4 days
**Status**: ✅ Complete

### Tasks

- [x] Initialize Copilot SDK client
- [x] Register tools with SDK on startup
- [x] Stream tool execution requests
- [x] Handle tool responses and errors
- [x] Implement timeout enforcement (10 min)

### Implementation Details

```go
// internal/copilot/client.go
package copilot

import (
    "context"
    "time"

    copilot "github.com/github/copilot-sdk-go"
)

type Client struct {
    sdk      *copilot.Client
    registry *registry.Registry
    timeout  time.Duration
    logger   *observability.Logger
}

func NewClient(apiKey string, registry *registry.Registry, timeout time.Duration) (*Client, error)
func (c *Client) RegisterTools() error
func (c *Client) ProcessMessage(ctx context.Context, message string, userID string) (*Response, error)
func (c *Client) handleToolRequest(ctx context.Context, req *copilot.ToolRequest) (*copilot.ToolResponse, error)
```

**Tool Registration**:
```go
// Convert internal tool schema to Copilot SDK format
func (c *Client) RegisterTools() error {
    for _, tool := range c.registry.List() {
        schema := tool.Schema()
        copilotTool := &copilot.Tool{
            Name:        tool.Name(),
            Description: tool.Description(),
            Parameters:  convertSchema(schema),
            Handler:     c.createHandler(tool),
        }
        if err := c.sdk.RegisterTool(copilotTool); err != nil {
            return fmt.Errorf("failed to register tool %s: %w", tool.Name(), err)
        }
    }
    return nil
}
```

**Timeout Enforcement**:
```go
func (c *Client) ProcessMessage(ctx context.Context, message string, userID string) (*Response, error) {
    // Create context with 10-minute timeout
    ctx, cancel := context.WithTimeout(ctx, c.timeout)
    defer cancel()

    // Process with Copilot SDK
    result, err := c.sdk.Chat(ctx, &copilot.ChatRequest{
        Message: message,
        UserID:  userID,
    })

    if errors.Is(err, context.DeadlineExceeded) {
        return nil, observability.ErrToolTimeout.WithCause(err)
    }

    return &Response{
        Text: result.Message,
        Data: result.ToolResults,
    }, err
}
```

### Test Cases

```go
// internal/copilot/client_test.go
//go:build integration

func TestCopilotClient_Initialize(t *testing.T)
func TestCopilotClient_RegisterTools(t *testing.T)
func TestCopilotClient_StreamRequests(t *testing.T)
func TestCopilotClient_ExecuteToolWithTimeout(t *testing.T)
func TestCopilotClient_HandleLLMResponse(t *testing.T)
func TestCopilotClient_ErrorHandling(t *testing.T)
func TestCopilotClient_ConcurrentRequests(t *testing.T)
func TestCopilotClient_ContextCancellation(t *testing.T)
```

**Build Tag**: `integration` (requires Copilot API credentials)

### Acceptance Criteria

- [x] All registered tools available in Copilot
- [x] Tool requests routed to correct handlers
- [x] Hard timeout at 10 minutes
- [x] Concurrent requests handled safely
- [x] LLM responses returned to messaging platforms

### Notes

<!-- Add your notes here -->

---

## 3.2 Request/Response Pipeline

**Duration**: 2 days
**Status**: ✅ Complete

### Tasks

- [x] Message → Copilot request transformation
- [x] Copilot response → Platform message transformation
- [x] Context preservation across requests
- [x] Error propagation

### Implementation Details

```go
// internal/copilot/pipeline.go
package copilot

import "context"

type Pipeline struct {
    client *Client
    logger *observability.Logger
}

func NewPipeline(client *Client) *Pipeline

// Transform incoming message to Copilot format
func (p *Pipeline) PrepareRequest(msg *handlers.Message) *copilot.ChatRequest {
    return &copilot.ChatRequest{
        Message:  msg.Content,
        UserID:   msg.UserID,
        Platform: msg.Platform,
        Metadata: map[string]string{
            "message_id": msg.ID,
        },
    }
}

// Transform Copilot response to platform-specific format
func (p *Pipeline) FormatResponse(resp *copilot.ChatResponse, platform string) *handlers.Response {
    // Format based on platform capabilities
    switch platform {
    case "line":
        return p.formatForLINE(resp)
    case "discord":
        return p.formatForDiscord(resp)
    default:
        return p.formatGeneric(resp)
    }
}

// Preserve conversation context
func (p *Pipeline) WithContext(ctx context.Context, conversationID string) context.Context {
    return context.WithValue(ctx, "conversation_id", conversationID)
}
```

**Error Propagation**:
```go
func (p *Pipeline) HandleError(err error, platform string) *handlers.Response {
    var appErr *observability.AppError
    if errors.As(err, &appErr) {
        switch appErr.Code {
        case "TOOL_TIMEOUT":
            return &handlers.Response{
                Text: "⏱️ The operation timed out. Please try again with a simpler request.",
            }
        case "TOOL_NOT_FOUND":
            return &handlers.Response{
                Text: "🔍 I don't have a tool to handle that request.",
            }
        default:
            return &handlers.Response{
                Text: fmt.Sprintf("❌ Error: %s", appErr.Message),
            }
        }
    }
    return &handlers.Response{
        Text: "❌ An unexpected error occurred. Please try again.",
    }
}
```

### Test Cases

```go
// internal/copilot/pipeline_test.go
func TestPipeline_MessageToCopilot(t *testing.T)
func TestPipeline_CopilotToMessage(t *testing.T)
func TestPipeline_ContextPreservation(t *testing.T)
func TestPipeline_ErrorPropagation(t *testing.T)
func TestPipeline_FormatForLINE(t *testing.T)
func TestPipeline_FormatForDiscord(t *testing.T)
```

### Acceptance Criteria

- [x] Messages transformed correctly for Copilot
- [x] Responses formatted appropriately per platform
- [x] Conversation context preserved
- [x] Errors mapped to user-friendly messages

### Notes

<!-- Add your notes here -->

---

## Deliverables

By the end of Phase 3:

- [x] Copilot SDK client initialized and connected
- [x] All tools registered with SDK
- [x] Request/response pipeline functional
- [x] Timeout enforcement working

---

## Dependencies

```go
// go.mod additions
require (
    github.com/github/copilot-sdk/go v0.1.20
)
```

---

## Environment Variables

```bash
GITHUB_COPILOT_API_KEY=your-api-key
```

---

## Time Tracking

| Task | Estimated | Actual | Notes |
|------|-----------|--------|-------|
| 3.1 Copilot Client | 4 days | | |
| 3.2 Pipeline | 2 days | | |
| **Total** | **6 days** | | |

---

**Previous**: [Phase 2: Messaging Platform Integration](./phase-2-messaging.md)
**Next**: [Phase 4: Tool Implementation](./phase-4-tools.md)
