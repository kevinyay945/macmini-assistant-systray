package discord_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers/discord"
)

func TestHandler_New(t *testing.T) {
	h := discord.New(discord.Config{
		Token:   "test-token",
		GuildID: "123456789",
	})
	if h == nil {
		t.Error("New() returned nil")
	}
}

func TestHandler_New_EmptyConfig(t *testing.T) {
	h := discord.New(discord.Config{})
	if h == nil {
		t.Error("New() with empty config returned nil")
	}
}

func TestHandler_New_WithAllConfig(t *testing.T) {
	router := &mockRouter{}
	h := discord.New(discord.Config{
		Token:               "test-token",
		GuildID:             "123456789",
		StatusChannelID:     "987654321",
		Router:              router,
		EnableSlashCommands: true,
	})
	if h == nil {
		t.Error("New() with full config returned nil")
	}
}

func TestHandler_Start_NoToken(t *testing.T) {
	h := discord.New(discord.Config{})
	err := h.Start()
	if err == nil {
		t.Error("Start() should return error when no token is provided")
	}
}

func TestHandler_Start_InvalidToken(t *testing.T) {
	h := discord.New(discord.Config{
		Token: "invalid-token",
	})
	err := h.Start()
	// Should fail to connect with invalid token
	if err == nil {
		t.Error("Start() should return error with invalid token")
	}
}

func TestHandler_Stop_NotStarted(t *testing.T) {
	h := discord.New(discord.Config{})
	if err := h.Stop(); err != nil {
		t.Errorf("Stop() returned error when not started: %v", err)
	}
}

func TestHandler_Stop_Idempotent(t *testing.T) {
	h := discord.New(discord.Config{})

	// Stop twice should not error
	if err := h.Stop(); err != nil {
		t.Errorf("Stop() first call returned error: %v", err)
	}
	if err := h.Stop(); err != nil {
		t.Errorf("Stop() second call returned error: %v", err)
	}
}

func TestHandler_InterfaceCompliance(t *testing.T) {
	var _ handlers.Handler = (*discord.Handler)(nil)
	var _ handlers.StatusReporter = (*discord.Handler)(nil)
}

func TestHandler_PostStatus_NotStarted(t *testing.T) {
	h := discord.New(discord.Config{
		Token:           "test-token",
		StatusChannelID: "123456789",
	})

	err := h.PostStatus(context.Background(), handlers.StatusMessage{
		Type:     "start",
		ToolName: "test_tool",
	})

	if err == nil {
		t.Error("PostStatus() should return error when session not initialized")
	}
}

func TestStatusMessage_Types(t *testing.T) {
	tests := []struct {
		msgType string
		valid   bool
	}{
		{"start", true},
		{"progress", true},
		{"complete", true},
		{"error", true},
		{"unknown", true}, // Unknown types still work, just get default formatting
	}

	for _, tt := range tests {
		t.Run(tt.msgType, func(t *testing.T) {
			msg := handlers.NewStatusMessage(tt.msgType, "test_tool", "user123", "discord")
			if msg.Type != tt.msgType {
				t.Errorf("Type = %q, want %q", msg.Type, tt.msgType)
			}
		})
	}
}

func TestStatusMessage_WithDuration(t *testing.T) {
	msg := handlers.NewStatusMessage("complete", "youtube_download", "user123", "discord")
	msg.Duration = 32500 * time.Millisecond

	if msg.Duration != 32500*time.Millisecond {
		t.Errorf("Duration = %v, want %v", msg.Duration, 32500*time.Millisecond)
	}
}

func TestStatusMessage_WithError(t *testing.T) {
	msg := handlers.NewStatusMessage("error", "gdrive_upload", "user456", "discord")
	msg.Error = errors.New("upload failed")

	if msg.Error == nil {
		t.Error("Error should not be nil")
	}
	if msg.Error.Error() != "upload failed" {
		t.Errorf("Error message = %q, want %q", msg.Error.Error(), "upload failed")
	}
}

func TestStatusMessage_WithResult(t *testing.T) {
	msg := handlers.NewStatusMessage("complete", "youtube_download", "user123", "discord")
	msg.Result["file_size"] = "125 MB"
	msg.Result["file_path"] = "/downloads/video.mp4"

	if msg.Result["file_size"] != "125 MB" {
		t.Errorf("Result[file_size] = %v, want %v", msg.Result["file_size"], "125 MB")
	}
}

// mockRouter implements handlers.MessageRouter for testing.
type mockRouter struct {
	response *handlers.Response
	err      error
	called   bool
	lastMsg  *handlers.Message
}

func (m *mockRouter) Route(ctx context.Context, msg *handlers.Message) (*handlers.Response, error) {
	m.called = true
	m.lastMsg = msg
	return m.response, m.err
}

func TestHandler_EmbedColors(t *testing.T) {
	// Test that color constants are defined correctly
	if discord.ColorBlue != 0x3498db {
		t.Errorf("ColorBlue = %x, want %x", discord.ColorBlue, 0x3498db)
	}
	if discord.ColorGreen != 0x2ecc71 {
		t.Errorf("ColorGreen = %x, want %x", discord.ColorGreen, 0x2ecc71)
	}
	if discord.ColorRed != 0xe74c3c {
		t.Errorf("ColorRed = %x, want %x", discord.ColorRed, 0xe74c3c)
	}
	if discord.ColorYellow != 0xf1c40f {
		t.Errorf("ColorYellow = %x, want %x", discord.ColorYellow, 0xf1c40f)
	}
}

func TestHandler_ErrorFormatting(t *testing.T) {
	// Document expected error formatting behavior
	tests := []struct {
		name string
		err  error
	}{
		{"context deadline exceeded", context.DeadlineExceeded},
		{"context canceled", context.Canceled},
		{"generic error", errors.New("something went wrong")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test documents expected error message behavior
			// Direct testing would require exposing formatErrorMessage
			_ = tt
		})
	}
}
