package discord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
	"github.com/kevinyay945/macmini-assistant-systray/internal/registry"
)

func TestCreateStatusEmbed_Start(t *testing.T) {
	h := New(Config{})
	msg := handlers.StatusMessage{
		Type:     "start",
		ToolName: "youtube_download",
		UserID:   "123456",
		Platform: "discord",
	}
	embed := h.createStatusEmbed(msg)
	if embed.Title != "🎬 youtube_download Started" {
		t.Errorf("Title = %q, want %q", embed.Title, "🎬 youtube_download Started")
	}
	if embed.Color != ColorBlue {
		t.Errorf("Color = %x, want %x", embed.Color, ColorBlue)
	}
}

func TestCreateStatusEmbed_Progress(t *testing.T) {
	h := New(Config{})
	msg := handlers.StatusMessage{
		Type:     "progress",
		ToolName: "gdrive_upload",
		Message:  "Uploading 50%...",
	}
	embed := h.createStatusEmbed(msg)
	if embed.Title != "⏳ gdrive_upload In Progress" {
		t.Errorf("Title = %q, want %q", embed.Title, "⏳ gdrive_upload In Progress")
	}
	if embed.Color != ColorYellow {
		t.Errorf("Color = %x, want %x", embed.Color, ColorYellow)
	}
	if embed.Description != "Uploading 50%..." {
		t.Errorf("Description = %q, want %q", embed.Description, "Uploading 50%...")
	}
}

func TestCreateStatusEmbed_Complete(t *testing.T) {
	h := New(Config{})
	msg := handlers.StatusMessage{
		Type:     "complete",
		ToolName: "youtube_download",
		Duration: 30 * time.Second,
	}
	embed := h.createStatusEmbed(msg)
	if embed.Title != "✅ youtube_download Complete" {
		t.Errorf("Title = %q, want %q", embed.Title, "✅ youtube_download Complete")
	}
	if embed.Color != ColorGreen {
		t.Errorf("Color = %x, want %x", embed.Color, ColorGreen)
	}
}

func TestCreateStatusEmbed_Error(t *testing.T) {
	h := New(Config{})
	msg := handlers.StatusMessage{
		Type:     "error",
		ToolName: "gdrive_upload",
		Error:    errors.New("upload failed"),
	}
	embed := h.createStatusEmbed(msg)
	if embed.Title != "❌ gdrive_upload Failed" {
		t.Errorf("Title = %q, want %q", embed.Title, "❌ gdrive_upload Failed")
	}
	if embed.Color != ColorRed {
		t.Errorf("Color = %x, want %x", embed.Color, ColorRed)
	}
}

func TestCreateStatusEmbed_Default(t *testing.T) {
	h := New(Config{})
	msg := handlers.StatusMessage{
		Type:     "unknown",
		ToolName: "some_tool",
	}
	embed := h.createStatusEmbed(msg)
	if embed.Color != ColorBlue {
		t.Errorf("Color = %x, want default blue %x", embed.Color, ColorBlue)
	}
}

func TestHandleStatusCommand_Online(t *testing.T) {
	h := New(Config{})
	h.started = true
	resp := h.handleStatusCommand(context.Background())
	if resp.Type != discordgo.InteractionResponseChannelMessageWithSource {
		t.Errorf("Response Type = %v, want %v", resp.Type, discordgo.InteractionResponseChannelMessageWithSource)
	}
	if len(resp.Data.Embeds) == 0 {
		t.Fatal("Expected embed in response")
	}
	embed := resp.Data.Embeds[0]
	if embed.Description != "🟢 Online" {
		t.Errorf("Description = %q, want %q", embed.Description, "🟢 Online")
	}
}

func TestHandleStatusCommand_Offline(t *testing.T) {
	h := New(Config{})
	h.started = false
	resp := h.handleStatusCommand(context.Background())
	if len(resp.Data.Embeds) == 0 {
		t.Fatal("Expected embed in response")
	}
	embed := resp.Data.Embeds[0]
	if embed.Description != "🔴 Offline" {
		t.Errorf("Description = %q, want %q", embed.Description, "🔴 Offline")
	}
}

func TestHandleToolsCommand_NoRegistry(t *testing.T) {
	h := New(Config{})
	resp := h.handleToolsCommand(context.Background())
	if resp.Data.Content == "" {
		t.Error("Content should not be empty")
	}
}

func TestHandleToolsCommand_WithRegistry(t *testing.T) {
	reg := registry.New()
	h := New(Config{Registry: reg})
	resp := h.handleToolsCommand(context.Background())
	if resp.Data.Content == "" {
		t.Error("Content should not be empty")
	}
}

func TestHandleHelpCommand(t *testing.T) {
	h := New(Config{})
	resp := h.handleHelpCommand(context.Background())
	if len(resp.Data.Embeds) == 0 {
		t.Fatal("Expected embed in response")
	}
	embed := resp.Data.Embeds[0]
	if embed.Title != "MacMini Assistant Help" {
		t.Errorf("Title = %q, want %q", embed.Title, "MacMini Assistant Help")
	}
}

func TestSendMessage_NilSession(t *testing.T) {
	h := New(Config{})
	err := h.SendMessage(context.Background(), "channel123", "test")
	if err == nil {
		t.Error("SendMessage should return error when session is nil")
	}
}

func TestSendEmbed_NilSession(t *testing.T) {
	h := New(Config{})
	embed := &discordgo.MessageEmbed{Title: "Test"}
	err := h.SendEmbed(context.Background(), "channel123", embed)
	if err == nil {
		t.Error("SendEmbed should return error when session is nil")
	}
}

func TestPostStatus_NilSession(t *testing.T) {
	h := New(Config{StatusChannelID: "channel123"})
	err := h.PostStatus(context.Background(), handlers.StatusMessage{Type: "start", ToolName: "test"})
	if err == nil {
		t.Error("PostStatus should return error when session is nil")
	}
}

func TestRegisterSlashCommands_NilSession(t *testing.T) {
	h := New(Config{})
	err := h.registerSlashCommands()
	if err == nil {
		t.Error("registerSlashCommands should return error when session is nil")
	}
}

func TestUnregisterSlashCommands_NilSession(t *testing.T) {
	h := New(Config{})
	h.unregisterSlashCommands() // Should not panic
}

func TestSlashCommandsDefinition(t *testing.T) {
	if len(slashCommands) != 3 {
		t.Errorf("Expected 3 slash commands, got %d", len(slashCommands))
	}
}

func TestHandler_StartIdempotent(t *testing.T) {
	h := New(Config{})
	h.started = true
	err := h.Start()
	if err != nil {
		t.Errorf("Start() on already started handler returned error: %v", err)
	}
}

func TestHandler_Start_NoToken(t *testing.T) {
	h := New(Config{Token: ""})
	err := h.Start()
	if err == nil {
		t.Error("Start() should return error when token is empty")
	}
}

func TestHandler_StopIdempotent(t *testing.T) {
	h := New(Config{})
	h.started = false
	err := h.Stop()
	if err != nil {
		t.Errorf("Stop() on non-started handler returned error: %v", err)
	}
}

func TestHandleComponentInteraction(t *testing.T) {
	h := New(Config{})
	i := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionMessageComponent,
			Data: discordgo.MessageComponentInteractionData{CustomID: "test_button"},
		},
	}
	h.handleComponentInteraction(context.Background(), nil, i) // Should not panic
}

func TestIsBotMentioned_NoMentions(t *testing.T) {
	h := New(Config{})
	session := &discordgo.Session{
		State: &discordgo.State{Ready: discordgo.Ready{User: &discordgo.User{ID: "bot123"}}},
	}
	msg := &discordgo.MessageCreate{Message: &discordgo.Message{Mentions: []*discordgo.User{}}}
	result := h.isBotMentioned(session, msg)
	if result {
		t.Error("isBotMentioned should return false when no mentions")
	}
}

func TestIsBotMentioned_BotMentioned(t *testing.T) {
	h := New(Config{})
	session := &discordgo.Session{
		State: &discordgo.State{Ready: discordgo.Ready{User: &discordgo.User{ID: "bot123"}}},
	}
	msg := &discordgo.MessageCreate{Message: &discordgo.Message{Mentions: []*discordgo.User{{ID: "bot123"}}}}
	result := h.isBotMentioned(session, msg)
	if !result {
		t.Error("isBotMentioned should return true when bot is mentioned")
	}
}

func TestCleanMentions(t *testing.T) {
	h := New(Config{})
	session := &discordgo.Session{
		State: &discordgo.State{Ready: discordgo.Ready{User: &discordgo.User{ID: "bot123"}}},
	}
	tests := []struct {
		content  string
		expected string
	}{
		{"hello world", "hello world"},
		{"<@bot123> hello", "hello"},
		{"<@!bot123> hello", "hello"},
		{"hello <@bot123>", "hello"},
	}
	for _, tt := range tests {
		got := h.cleanMentions(session, tt.content)
		if got != tt.expected {
			t.Errorf("cleanMentions(%q) = %q, want %q", tt.content, got, tt.expected)
		}
	}
}
