package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
)

func TestPlatformConstants(t *testing.T) {
	// Ensure platform constants are defined correctly
	if handlers.PlatformDiscord != "discord" {
		t.Errorf("PlatformDiscord = %q, want %q", handlers.PlatformDiscord, "discord")
	}
	if handlers.PlatformLINE != "line" {
		t.Errorf("PlatformLINE = %q, want %q", handlers.PlatformLINE, "line")
	}
}

func TestNewMessage(t *testing.T) {
	called := false
	replyFunc := func(response string) error {
		called = true
		return nil
	}

	msg := handlers.NewMessage("msg123", "user456", handlers.PlatformDiscord, "hello world", replyFunc)

	if msg.ID != "msg123" {
		t.Errorf("ID = %q, want %q", msg.ID, "msg123")
	}
	if msg.UserID != "user456" {
		t.Errorf("UserID = %q, want %q", msg.UserID, "user456")
	}
	if msg.Platform != handlers.PlatformDiscord {
		t.Errorf("Platform = %q, want %q", msg.Platform, handlers.PlatformDiscord)
	}
	if msg.Content != "hello world" {
		t.Errorf("Content = %q, want %q", msg.Content, "hello world")
	}
	if msg.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
	if msg.Metadata == nil {
		t.Error("Metadata should be initialized")
	}

	// Test reply function
	if err := msg.ReplyFunc("response"); err != nil {
		t.Errorf("ReplyFunc returned error: %v", err)
	}
	if !called {
		t.Error("ReplyFunc was not called")
	}
}

func TestNewMessage_FromLINE(t *testing.T) {
	msg := handlers.NewMessage("line-msg-id", "U12345", handlers.PlatformLINE, "LINE message", nil)

	if msg.Platform != handlers.PlatformLINE {
		t.Errorf("Platform = %q, want %q", msg.Platform, handlers.PlatformLINE)
	}
	if msg.ID != "line-msg-id" {
		t.Errorf("ID = %q, want %q", msg.ID, "line-msg-id")
	}
}

func TestNewMessage_FromDiscord(t *testing.T) {
	msg := handlers.NewMessage("discord-msg-id", "123456789", handlers.PlatformDiscord, "Discord message", nil)

	if msg.Platform != handlers.PlatformDiscord {
		t.Errorf("Platform = %q, want %q", msg.Platform, handlers.PlatformDiscord)
	}
	if msg.ID != "discord-msg-id" {
		t.Errorf("ID = %q, want %q", msg.ID, "discord-msg-id")
	}
}

func TestNewResponse(t *testing.T) {
	resp := handlers.NewResponse("Hello!")

	if resp.Text != "Hello!" {
		t.Errorf("Text = %q, want %q", resp.Text, "Hello!")
	}
	if resp.Data == nil {
		t.Error("Data should be initialized")
	}
	if resp.Error != nil {
		t.Error("Error should be nil")
	}
}

func TestNewErrorResponse(t *testing.T) {
	err := errors.New("something went wrong")
	resp := handlers.NewErrorResponse(err)

	if !errors.Is(resp.Error, err) {
		t.Errorf("Error = %v, want %v", resp.Error, err)
	}
	if resp.Text != "" {
		t.Errorf("Text = %q, want empty string", resp.Text)
	}
}

func TestNewStatusMessage(t *testing.T) {
	status := handlers.NewStatusMessage("start", "youtube_download", "user123", handlers.PlatformDiscord)

	if status.Type != "start" {
		t.Errorf("Type = %q, want %q", status.Type, "start")
	}
	if status.ToolName != "youtube_download" {
		t.Errorf("ToolName = %q, want %q", status.ToolName, "youtube_download")
	}
	if status.UserID != "user123" {
		t.Errorf("UserID = %q, want %q", status.UserID, "user123")
	}
	if status.Platform != handlers.PlatformDiscord {
		t.Errorf("Platform = %q, want %q", status.Platform, handlers.PlatformDiscord)
	}
	if status.Result == nil {
		t.Error("Result should be initialized")
	}
}

func TestStatusMessage_WithDuration(t *testing.T) {
	status := handlers.NewStatusMessage("complete", "youtube_download", "user123", handlers.PlatformLINE)
	status.Duration = 32500 * time.Millisecond

	if status.Duration != 32500*time.Millisecond {
		t.Errorf("Duration = %v, want %v", status.Duration, 32500*time.Millisecond)
	}
}

func TestStatusMessage_WithError(t *testing.T) {
	status := handlers.NewStatusMessage("error", "gdrive_upload", "user456", handlers.PlatformDiscord)
	status.Error = errors.New("upload failed")

	if status.Error == nil {
		t.Error("Error should not be nil")
	}
	if status.Error.Error() != "upload failed" {
		t.Errorf("Error message = %q, want %q", status.Error.Error(), "upload failed")
	}
}

func TestDefaultErrorFormatter_FormatError(t *testing.T) {
	formatter := &handlers.DefaultErrorFormatter{}

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
		{
			name: "simple error",
			err:  errors.New("something failed"),
			want: "An error occurred: something failed",
		},
		{
			name: "wrapped error",
			err:  errors.New("outer: inner"),
			want: "An error occurred: outer: inner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatter.FormatError(tt.err)
			if got != tt.want {
				t.Errorf("FormatError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMessage_MetadataUsage(t *testing.T) {
	msg := handlers.NewMessage("id", "user", handlers.PlatformDiscord, "content", nil)

	// Add metadata
	msg.Metadata["channel_id"] = "123456"
	msg.Metadata["guild_id"] = "789012"

	if msg.Metadata["channel_id"] != "123456" {
		t.Errorf("Metadata[channel_id] = %v, want %v", msg.Metadata["channel_id"], "123456")
	}
	if msg.Metadata["guild_id"] != "789012" {
		t.Errorf("Metadata[guild_id] = %v, want %v", msg.Metadata["guild_id"], "789012")
	}
}

func TestResponse_DataUsage(t *testing.T) {
	resp := handlers.NewResponse("Success")

	// Add data
	resp.Data["file_path"] = "/path/to/file.mp4"
	resp.Data["file_size"] = int64(12345678)

	if resp.Data["file_path"] != "/path/to/file.mp4" {
		t.Errorf("Data[file_path] = %v, want %v", resp.Data["file_path"], "/path/to/file.mp4")
	}
	if resp.Data["file_size"] != int64(12345678) {
		t.Errorf("Data[file_size] = %v, want %v", resp.Data["file_size"], int64(12345678))
	}
}

func TestStatusMessage_AllTypes(t *testing.T) {
	types := []string{"start", "progress", "complete", "error"}

	for _, msgType := range types {
		status := handlers.NewStatusMessage(msgType, "test_tool", "user", handlers.PlatformDiscord)
		if status.Type != msgType {
			t.Errorf("Type = %q, want %q", status.Type, msgType)
		}
	}
}

func TestFormatUserFriendlyError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{
			name:    "nil error",
			err:     nil,
			wantMsg: "",
		},
		{
			name:    "context deadline exceeded",
			err:     context.DeadlineExceeded,
			wantMsg: "⏱️ Request timed out. Please try again.",
		},
		{
			name:    "context canceled",
			err:     context.Canceled,
			wantMsg: "🚫 Request was cancelled.",
		},
		{
			name:    "generic error",
			err:     errors.New("something went wrong"),
			wantMsg: "❌ An error occurred while processing your request. Please try again later.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := handlers.FormatUserFriendlyError(tt.err)
			if got != tt.wantMsg {
				t.Errorf("FormatUserFriendlyError() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}
