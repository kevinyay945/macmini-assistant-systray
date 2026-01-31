package discord_test

import (
	"testing"

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

func TestHandler_Start(t *testing.T) {
	h := discord.New(discord.Config{})
	if err := h.Start(); err != nil {
		t.Errorf("Start() returned error: %v", err)
	}
}

func TestHandler_Stop(t *testing.T) {
	h := discord.New(discord.Config{})
	if err := h.Stop(); err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}
}
