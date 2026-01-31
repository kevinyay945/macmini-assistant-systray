package copilot_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/copilot"
	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
	"github.com/kevinyay945/macmini-assistant-systray/internal/observability"
)

func TestNewPipeline(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)
	if pipeline == nil {
		t.Error("NewPipeline() returned nil")
	}
}

func TestPipeline_PrepareRequest(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	msg := handlers.NewMessage("msg-123", "user-456", handlers.PlatformDiscord, "hello world", nil)
	msg.Metadata["channel_id"] = "chan-789"

	req := pipeline.PrepareRequest(msg)

	if req.Message != "hello world" {
		t.Errorf("Message = %q, want %q", req.Message, "hello world")
	}
	if req.UserID != "user-456" {
		t.Errorf("UserID = %q, want %q", req.UserID, "user-456")
	}
	if req.Platform != handlers.PlatformDiscord {
		t.Errorf("Platform = %q, want %q", req.Platform, handlers.PlatformDiscord)
	}
	if req.Metadata["message_id"] != "msg-123" {
		t.Errorf("Metadata[message_id] = %q, want %q", req.Metadata["message_id"], "msg-123")
	}
	if req.Metadata["channel_id"] != "chan-789" {
		t.Errorf("Metadata[channel_id] = %q, want %q", req.Metadata["channel_id"], "chan-789")
	}
}

func TestPipeline_FormatResponse_LINE(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	chatResp := &copilot.ChatResponse{
		Message:     "Hello from Copilot",
		ToolResults: map[string]interface{}{"result": "success"},
		ToolName:    "test_tool",
	}

	resp := pipeline.FormatResponse(chatResp, handlers.PlatformLINE)

	if resp.Text != "Hello from Copilot" {
		t.Errorf("Text = %q, want %q", resp.Text, "Hello from Copilot")
	}
	if resp.Data["result"] != "success" {
		t.Errorf("Data[result] = %v, want %v", resp.Data["result"], "success")
	}
}

func TestPipeline_FormatResponse_Discord(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	chatResp := &copilot.ChatResponse{
		Message:     "Hello from Copilot",
		ToolResults: map[string]interface{}{"result": "success"},
		ToolName:    "test_tool",
	}

	resp := pipeline.FormatResponse(chatResp, handlers.PlatformDiscord)

	if resp.Text != "Hello from Copilot" {
		t.Errorf("Text = %q, want %q", resp.Text, "Hello from Copilot")
	}
	if resp.Data["executed_tool"] != "test_tool" {
		t.Errorf("Data[executed_tool] = %v, want %v", resp.Data["executed_tool"], "test_tool")
	}
}

func TestPipeline_FormatResponse_Nil(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	resp := pipeline.FormatResponse(nil, handlers.PlatformDiscord)

	if resp == nil {
		t.Error("FormatResponse(nil) returned nil, want empty Response")
	}
	if resp.Text != "" {
		t.Errorf("Text = %q, want empty string", resp.Text)
	}
}

func TestPipeline_FormatResponse_LINECharLimit(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	// Create a message longer than LINE's 5000 character limit
	longMessage := make([]byte, 6000)
	for i := range longMessage {
		longMessage[i] = 'a'
	}

	chatResp := &copilot.ChatResponse{
		Message: string(longMessage),
	}

	resp := pipeline.FormatResponse(chatResp, handlers.PlatformLINE)

	// Should be truncated to 5000 chars with "..."
	if len(resp.Text) != 5000 {
		t.Errorf("len(Text) = %d, want 5000", len(resp.Text))
	}
	if resp.Text[4997:] != "..." {
		t.Errorf("Text should end with '...', got %q", resp.Text[4997:])
	}
}

func TestPipeline_FormatResponse_DiscordCharLimit(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	// Create a message longer than Discord's 2000 character limit
	longMessage := make([]byte, 3000)
	for i := range longMessage {
		longMessage[i] = 'a'
	}

	chatResp := &copilot.ChatResponse{
		Message: string(longMessage),
	}

	resp := pipeline.FormatResponse(chatResp, handlers.PlatformDiscord)

	// Should be truncated to 2000 chars with "..."
	if len(resp.Text) != 2000 {
		t.Errorf("len(Text) = %d, want 2000", len(resp.Text))
	}
	if resp.Text[1997:] != "..." {
		t.Errorf("Text should end with '...', got %q", resp.Text[1997:])
	}
}

func TestPipeline_WithContext(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)
	ctx := context.Background()

	ctx = pipeline.WithContext(ctx, "conv-123")

	if copilot.ConversationIDFromContext(ctx) != "conv-123" {
		t.Errorf("ConversationIDFromContext() = %q, want %q",
			copilot.ConversationIDFromContext(ctx), "conv-123")
	}
}

func TestPipeline_WithUserID(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)
	ctx := context.Background()

	ctx = pipeline.WithUserID(ctx, "user-456")

	if copilot.UserIDFromContext(ctx) != "user-456" {
		t.Errorf("UserIDFromContext() = %q, want %q",
			copilot.UserIDFromContext(ctx), "user-456")
	}
}

func TestPipeline_WithPlatform(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)
	ctx := context.Background()

	ctx = pipeline.WithPlatform(ctx, handlers.PlatformDiscord)

	if copilot.PlatformFromContext(ctx) != handlers.PlatformDiscord {
		t.Errorf("PlatformFromContext() = %q, want %q",
			copilot.PlatformFromContext(ctx), handlers.PlatformDiscord)
	}
}

func TestConversationIDFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	if copilot.ConversationIDFromContext(ctx) != "" {
		t.Error("ConversationIDFromContext() should return empty string for empty context")
	}
}

func TestUserIDFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	if copilot.UserIDFromContext(ctx) != "" {
		t.Error("UserIDFromContext() should return empty string for empty context")
	}
}

func TestPlatformFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	if copilot.PlatformFromContext(ctx) != "" {
		t.Error("PlatformFromContext() should return empty string for empty context")
	}
}

func TestPipeline_HandleError_Nil(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	resp := pipeline.HandleError(nil, handlers.PlatformDiscord)

	if resp != nil {
		t.Error("HandleError(nil) should return nil")
	}
}

func TestPipeline_HandleError_ContextDeadlineExceeded(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	resp := pipeline.HandleError(context.DeadlineExceeded, handlers.PlatformDiscord)

	if resp == nil {
		t.Fatal("HandleError() returned nil")
	}
	if resp.Text == "" {
		t.Error("Response text should not be empty")
	}
	// Check that it contains both an emoji and relevant text
	if !strings.Contains(resp.Text, "⏱️") {
		t.Errorf("Text should contain timeout emoji ⏱️, got %q", resp.Text)
	}
	if !strings.Contains(resp.Text, "timed out") {
		t.Errorf("Text should contain 'timed out', got %q", resp.Text)
	}
}

func TestPipeline_HandleError_ContextCanceled(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	resp := pipeline.HandleError(context.Canceled, handlers.PlatformDiscord)

	if resp == nil {
		t.Fatal("HandleError() returned nil")
	}
	if resp.Text == "" {
		t.Error("Response text should not be empty")
	}
	// Check that it contains both an emoji and relevant text
	if !strings.Contains(resp.Text, "🚫") {
		t.Errorf("Text should contain cancelled emoji 🚫, got %q", resp.Text)
	}
	if !strings.Contains(resp.Text, "cancelled") {
		t.Errorf("Text should contain 'cancelled', got %q", resp.Text)
	}
}

func TestPipeline_HandleError_ErrAPIKeyNotConfigured(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	resp := pipeline.HandleError(copilot.ErrAPIKeyNotConfigured, handlers.PlatformDiscord)

	if resp == nil {
		t.Fatal("HandleError() returned nil")
	}
	if resp.Text == "" {
		t.Error("Response text should not be empty")
	}
	// Check that it contains both an emoji and relevant text
	if !strings.Contains(resp.Text, "⚙️") {
		t.Errorf("Text should contain config emoji ⚙️, got %q", resp.Text)
	}
	if !strings.Contains(resp.Text, "configured") {
		t.Errorf("Text should contain 'configured', got %q", resp.Text)
	}
}

func TestPipeline_HandleError_ErrClientNotStarted(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	resp := pipeline.HandleError(copilot.ErrClientNotStarted, handlers.PlatformDiscord)

	if resp == nil {
		t.Fatal("HandleError() returned nil")
	}
	if resp.Text == "" {
		t.Error("Response text should not be empty")
	}
	// Check that it contains both an emoji and relevant text
	if !strings.Contains(resp.Text, "🔌") {
		t.Errorf("Text should contain plug emoji 🔌, got %q", resp.Text)
	}
	if !strings.Contains(resp.Text, "starting") {
		t.Errorf("Text should contain 'starting', got %q", resp.Text)
	}
}

func TestPipeline_HandleError_AppError_ToolTimeout(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	err := observability.ErrToolTimeout.WithMessage("test timeout")

	resp := pipeline.HandleError(err, handlers.PlatformDiscord)

	if resp == nil {
		t.Fatal("HandleError() returned nil")
	}
	if resp.Text == "" {
		t.Error("Response text should not be empty")
	}
	// Check that it contains both an emoji and relevant text
	if !strings.Contains(resp.Text, "⏱️") {
		t.Errorf("Text should contain timeout emoji ⏱️, got %q", resp.Text)
	}
	if !strings.Contains(resp.Text, "timed out") {
		t.Errorf("Text should contain 'timed out', got %q", resp.Text)
	}
}

func TestPipeline_HandleError_AppError_ToolNotFound(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	err := observability.ErrToolNotFound.WithMessage("unknown tool")

	resp := pipeline.HandleError(err, handlers.PlatformDiscord)

	if resp == nil {
		t.Fatal("HandleError() returned nil")
	}
	if resp.Text == "" {
		t.Error("Response text should not be empty")
	}
	// Check that it contains both an emoji and relevant text
	if !strings.Contains(resp.Text, "🔍") {
		t.Errorf("Text should contain search emoji 🔍, got %q", resp.Text)
	}
	if !strings.Contains(resp.Text, "tool") {
		t.Errorf("Text should contain 'tool', got %q", resp.Text)
	}
}

func TestPipeline_HandleError_GenericError(t *testing.T) {
	pipeline := copilot.NewPipeline(nil)

	err := errors.New("some random error")

	resp := pipeline.HandleError(err, handlers.PlatformDiscord)

	if resp == nil {
		t.Fatal("HandleError() returned nil")
	}
	if resp.Text == "" {
		t.Error("Response text should not be empty")
	}
	// Check that it contains both an emoji and relevant text
	if !strings.Contains(resp.Text, "❌") {
		t.Errorf("Text should contain error emoji ❌, got %q", resp.Text)
	}
	if !strings.Contains(resp.Text, "error") {
		t.Errorf("Text should contain 'error', got %q", resp.Text)
	}
}

func TestChatRequest_Fields(t *testing.T) {
	req := &copilot.ChatRequest{
		Message:  "test message",
		UserID:   "user-123",
		Platform: handlers.PlatformLINE,
		Metadata: map[string]string{"key": "value"},
	}

	if req.Message != "test message" {
		t.Errorf("Message = %q, want %q", req.Message, "test message")
	}
	if req.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", req.UserID, "user-123")
	}
	if req.Platform != handlers.PlatformLINE {
		t.Errorf("Platform = %q, want %q", req.Platform, handlers.PlatformLINE)
	}
	if req.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %q, want %q", req.Metadata["key"], "value")
	}
}

func TestChatResponse_Fields(t *testing.T) {
	resp := &copilot.ChatResponse{
		Message:     "response text",
		ToolResults: map[string]interface{}{"result": "data"},
		ToolName:    "my_tool",
	}

	if resp.Message != "response text" {
		t.Errorf("Message = %q, want %q", resp.Message, "response text")
	}
	if resp.ToolResults["result"] != "data" {
		t.Errorf("ToolResults[result] = %v, want %v", resp.ToolResults["result"], "data")
	}
	if resp.ToolName != "my_tool" {
		t.Errorf("ToolName = %q, want %q", resp.ToolName, "my_tool")
	}
}
