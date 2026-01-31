// Package handlers provides common interfaces for message platform handlers.
package handlers

import (
	"context"
	"time"
)

// Message represents a platform-agnostic incoming message.
// This allows the orchestrator to process messages uniformly regardless of source.
type Message struct {
	// ID is the unique identifier for this message from the source platform.
	ID string
	// UserID is the platform-specific user identifier.
	UserID string
	// Platform identifies the message source ("line" or "discord").
	Platform string
	// Content is the text content of the message.
	Content string
	// Timestamp is when the message was received.
	Timestamp time.Time
	// ReplyFunc is the callback to send a response back to the user.
	// This abstracts platform-specific reply mechanisms.
	ReplyFunc func(response string) error
	// Metadata contains platform-specific additional data.
	Metadata map[string]interface{}
}

// MessageRouter routes messages to the orchestrator for processing.
// Implementations handle the integration with the Copilot SDK and tool execution.
type MessageRouter interface {
	// Route processes an incoming message and returns a response.
	// The context should contain timeout and cancellation signals.
	Route(ctx context.Context, msg *Message) (*Response, error)
}

// Response represents the result of message processing.
type Response struct {
	// Text is the primary text response to send back to the user.
	Text string
	// Data contains structured data from tool execution results.
	Data map[string]interface{}
	// Error holds any error that occurred during processing.
	// This is separate from the function return error to allow partial responses.
	Error error
}

// StatusMessage represents a status update to post to the status channel.
// Used primarily by Discord to post execution status to a dedicated channel.
type StatusMessage struct {
	// Type indicates the status type: "start", "progress", "complete", or "error".
	Type string
	// ToolName is the name of the tool being executed.
	ToolName string
	// UserID is the user who initiated the action.
	UserID string
	// Platform is the source platform of the request.
	Platform string
	// Duration is the execution time (only set for "complete" type).
	Duration time.Duration
	// Result contains tool execution result data.
	Result map[string]interface{}
	// Error contains error details (only set for "error" type).
	Error error
	// Message is an optional human-readable status message.
	Message string
}

// StatusReporter defines the interface for posting status updates.
// Typically implemented by Discord handler to post to a status channel.
type StatusReporter interface {
	// PostStatus sends a status message to the configured status channel.
	PostStatus(ctx context.Context, msg StatusMessage) error
}

// ErrorFormatter provides platform-specific error message formatting.
type ErrorFormatter interface {
	// FormatError converts an error into a user-friendly message.
	// Different platforms may have different formatting requirements.
	FormatError(err error) string
}

// DefaultErrorFormatter provides a basic error formatting implementation.
type DefaultErrorFormatter struct{}

// FormatError returns a user-friendly error message.
func (f *DefaultErrorFormatter) FormatError(err error) string {
	if err == nil {
		return ""
	}
	return "An error occurred: " + err.Error()
}

// NewMessage creates a new Message with the given parameters.
func NewMessage(id, userID, platform, content string, replyFunc func(string) error) *Message {
	return &Message{
		ID:        id,
		UserID:    userID,
		Platform:  platform,
		Content:   content,
		Timestamp: time.Now(),
		ReplyFunc: replyFunc,
		Metadata:  make(map[string]interface{}),
	}
}

// NewResponse creates a new Response with the given text.
func NewResponse(text string) *Response {
	return &Response{
		Text: text,
		Data: make(map[string]interface{}),
	}
}

// NewErrorResponse creates a new Response with an error.
func NewErrorResponse(err error) *Response {
	return &Response{
		Error: err,
	}
}

// NewStatusMessage creates a new StatusMessage for the given type and tool.
func NewStatusMessage(msgType, toolName, userID, platform string) StatusMessage {
	return StatusMessage{
		Type:     msgType,
		ToolName: toolName,
		UserID:   userID,
		Platform: platform,
		Result:   make(map[string]interface{}),
	}
}
