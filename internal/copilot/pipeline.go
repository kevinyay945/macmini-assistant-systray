// Package copilot provides integration with GitHub Copilot SDK.
package copilot

import (
	"context"
	"errors"
	"fmt"

	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
	"github.com/kevinyay945/macmini-assistant-systray/internal/observability"
)

// contextKey is a type for context keys used in this package.
type contextKey string

const (
	conversationIDKey contextKey = "conversation_id"
	userIDKey         contextKey = "user_id"
	platformKey       contextKey = "platform"
)

// Pipeline handles the transformation of messages between platform format and Copilot format.
type Pipeline struct {
	client *Client
	logger *observability.Logger
}

// NewPipeline creates a new request/response pipeline.
func NewPipeline(client *Client) *Pipeline {
	logger := observability.New(observability.WithLevel(observability.LevelInfo))
	if client != nil && client.logger != nil {
		logger = client.logger
	}

	return &Pipeline{
		client: client,
		logger: logger.With("component", "pipeline"),
	}
}

// ChatRequest represents a request to the Copilot SDK.
type ChatRequest struct {
	// Message is the user's message content.
	Message string
	// UserID is the platform-specific user identifier.
	UserID string
	// Platform identifies the message source ("line" or "discord").
	Platform string
	// Metadata contains additional context data.
	Metadata map[string]string
}

// ChatResponse represents a response from the Copilot SDK.
type ChatResponse struct {
	// Message is the text response from Copilot.
	Message string
	// ToolResults contains any tool execution results.
	ToolResults map[string]interface{}
	// ToolName is the name of the tool that was executed (if any).
	ToolName string
}

// PrepareRequest transforms an incoming platform message to Copilot format.
func (p *Pipeline) PrepareRequest(msg *handlers.Message) *ChatRequest {
	metadata := make(map[string]string)
	metadata["message_id"] = msg.ID

	// Copy string metadata from the message
	for k, v := range msg.Metadata {
		if strVal, ok := v.(string); ok {
			metadata[k] = strVal
		}
	}

	return &ChatRequest{
		Message:  msg.Content,
		UserID:   msg.UserID,
		Platform: msg.Platform,
		Metadata: metadata,
	}
}

// FormatResponse transforms a Copilot response to platform-specific format.
func (p *Pipeline) FormatResponse(resp *ChatResponse, platform string) *handlers.Response {
	if resp == nil {
		return handlers.NewResponse("")
	}

	switch platform {
	case handlers.PlatformLINE:
		return p.formatForLINE(resp)
	case handlers.PlatformDiscord:
		return p.formatForDiscord(resp)
	default:
		return p.formatGeneric(resp)
	}
}

// formatForLINE formats the response for LINE platform.
// LINE has specific message format requirements and character limits.
func (p *Pipeline) formatForLINE(resp *ChatResponse) *handlers.Response {
	response := handlers.NewResponse(resp.Message)
	response.Data = resp.ToolResults

	// LINE has a 5000 character limit per message
	const lineCharLimit = 5000
	if len(response.Text) > lineCharLimit {
		response.Text = response.Text[:lineCharLimit-3] + "..."
	}

	return response
}

// formatForDiscord formats the response for Discord platform.
// Discord supports rich embeds and has different formatting options.
func (p *Pipeline) formatForDiscord(resp *ChatResponse) *handlers.Response {
	response := handlers.NewResponse(resp.Message)
	response.Data = resp.ToolResults

	// Discord has a 2000 character limit for regular messages
	const discordCharLimit = 2000
	if len(response.Text) > discordCharLimit {
		response.Text = response.Text[:discordCharLimit-3] + "..."
	}

	// Add tool execution info to Data for Discord's rich formatting
	if resp.ToolName != "" {
		if response.Data == nil {
			response.Data = make(map[string]interface{})
		}
		response.Data["executed_tool"] = resp.ToolName
	}

	return response
}

// formatGeneric formats the response for unknown platforms.
func (p *Pipeline) formatGeneric(resp *ChatResponse) *handlers.Response {
	response := handlers.NewResponse(resp.Message)
	response.Data = resp.ToolResults
	return response
}

// WithContext adds conversation context to the context.
// This allows tracking of conversation state across requests.
func (p *Pipeline) WithContext(ctx context.Context, conversationID string) context.Context {
	return context.WithValue(ctx, conversationIDKey, conversationID)
}

// WithUserID adds user ID to the context.
func (p *Pipeline) WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// WithPlatform adds platform info to the context.
func (p *Pipeline) WithPlatform(ctx context.Context, platform string) context.Context {
	return context.WithValue(ctx, platformKey, platform)
}

// ConversationIDFromContext retrieves the conversation ID from context.
func ConversationIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(conversationIDKey).(string); ok {
		return id
	}
	return ""
}

// UserIDFromContext retrieves the user ID from context.
func UserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}

// PlatformFromContext retrieves the platform from context.
func PlatformFromContext(ctx context.Context) string {
	if platform, ok := ctx.Value(platformKey).(string); ok {
		return platform
	}
	return ""
}

// HandleError converts an error to a user-friendly response.
// This maps internal errors to appropriate user-facing messages.
func (p *Pipeline) HandleError(err error, platform string) *handlers.Response {
	if err == nil {
		return nil
	}

	var appErr *observability.AppError
	if errors.As(err, &appErr) {
		return p.handleAppError(appErr, platform)
	}

	// Handle standard context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return handlers.NewResponse("⏱️ The operation timed out. Please try again with a simpler request.")
	}
	if errors.Is(err, context.Canceled) {
		return handlers.NewResponse("🚫 The operation was cancelled.")
	}

	// Handle Copilot-specific errors
	if errors.Is(err, ErrAPIKeyNotConfigured) {
		return handlers.NewResponse("⚙️ The AI service is not configured. Please contact the administrator.")
	}
	if errors.Is(err, ErrClientNotStarted) {
		return handlers.NewResponse("🔌 The AI service is starting up. Please try again in a moment.")
	}
	if errors.Is(err, ErrSessionNotCreated) {
		return handlers.NewResponse("🔄 Unable to create a session. Please try again.")
	}

	// Generic error response
	p.logger.Error(context.Background(), "unhandled error in pipeline",
		"error", err,
		"platform", platform,
	)

	return handlers.NewResponse("❌ An unexpected error occurred. Please try again.")
}

// handleAppError converts an AppError to a user-friendly response.
func (p *Pipeline) handleAppError(appErr *observability.AppError, _ string) *handlers.Response {
	switch appErr.Code {
	case observability.CodeToolTimeout:
		return handlers.NewResponse("⏱️ The operation timed out. Please try again with a simpler request.")
	case observability.CodeToolNotFound:
		return handlers.NewResponse("🔍 I don't have a tool to handle that request.")
	case observability.CodeInvalidParams:
		return handlers.NewResponse(fmt.Sprintf("⚠️ Invalid input: %s", appErr.Message))
	case observability.CodeCopilotConnection:
		return handlers.NewResponse("🔌 Unable to connect to the AI service. Please try again later.")
	case observability.CodeAuthFailed:
		return handlers.NewResponse("🔐 Authentication failed. Please check your credentials.")
	case observability.CodeMessageFailed:
		return handlers.NewResponse("📨 Failed to send the message. Please try again.")
	case observability.CodeConfigNotFound:
		return handlers.NewResponse("⚙️ Configuration error. Please contact the administrator.")
	default:
		return handlers.NewResponse(fmt.Sprintf("❌ Error: %s", appErr.UserMessage()))
	}
}

// Process handles the full request/response cycle for a message.
// This is a convenience method that combines PrepareRequest, Copilot processing, and FormatResponse.
func (p *Pipeline) Process(ctx context.Context, msg *handlers.Message) (*handlers.Response, error) {
	if p.client == nil {
		return nil, fmt.Errorf("copilot client is not configured")
	}

	// Prepare the request
	req := p.PrepareRequest(msg)

	// Add context information
	ctx = p.WithContext(ctx, msg.ID)
	ctx = p.WithUserID(ctx, msg.UserID)
	ctx = p.WithPlatform(ctx, msg.Platform)

	// Process with Copilot
	resp, err := p.client.ProcessMessageWithUserID(ctx, req.Message, req.UserID)
	if err != nil {
		p.logger.Error(ctx, "failed to process message with Copilot",
			"error", err,
			"user_id", req.UserID,
			"platform", req.Platform,
		)
		return p.HandleError(err, req.Platform), err
	}

	// Convert to ChatResponse
	chatResp := &ChatResponse{
		Message:     resp.Text,
		ToolResults: resp.Data,
		ToolName:    resp.ToolName,
	}

	// Format response for the platform
	return p.FormatResponse(chatResp, req.Platform), nil
}
