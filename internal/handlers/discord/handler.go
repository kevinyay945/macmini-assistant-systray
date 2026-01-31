// Package discord provides Discord bot event handling.
package discord

import (
	"github.com/kevinyay945/macmini-assistant-systray/internal/copilot"
	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
	"github.com/kevinyay945/macmini-assistant-systray/internal/registry"
)

// Compile-time interface check
var _ handlers.Handler = (*Handler)(nil)

// Handler processes Discord bot events.
type Handler struct {
	token    string
	guildID  string
	copilot  *copilot.Client
	registry *registry.Registry
}

// Config holds Discord handler configuration.
type Config struct {
	Token    string
	GuildID  string
	Copilot  *copilot.Client
	Registry *registry.Registry
}

// New creates a new Discord event handler.
func New(cfg Config) *Handler {
	return &Handler{
		token:    cfg.Token,
		guildID:  cfg.GuildID,
		copilot:  cfg.Copilot,
		registry: cfg.Registry,
	}
}

// Start begins listening for Discord events.
func (h *Handler) Start() error {
	// TODO: Implement Discord event handling
	// 1. Create a new Discord session with token
	// 2. Register event handlers (message, interaction)
	// 3. Open websocket connection
	// 4. Register slash commands if guildID is specified
	return nil
}

// Stop gracefully shuts down the Discord handler.
func (h *Handler) Stop() error {
	// TODO: Implement graceful shutdown
	// 1. Unregister slash commands if needed
	// 2. Close Discord session
	return nil
}
