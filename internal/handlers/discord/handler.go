// Package discord provides Discord bot event handling.
package discord

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
	"github.com/kevinyay945/macmini-assistant-systray/internal/observability"
	"github.com/kevinyay945/macmini-assistant-systray/internal/registry"
)

// Compile-time interface checks
var (
	_ handlers.Handler        = (*Handler)(nil)
	_ handlers.StatusReporter = (*Handler)(nil)
)

// Embed colors for status messages
const (
	ColorBlue   = 0x3498db // Tool started
	ColorGreen  = 0x2ecc71 // Tool completed
	ColorRed    = 0xe74c3c // Error
	ColorYellow = 0xf1c40f // Warning/progress
)

// Handler processes Discord bot events.
type Handler struct {
	token           string
	guildID         string
	statusChannelID string
	router          handlers.MessageRouter
	registry        *registry.Registry
	logger          *observability.Logger
	enableSlashCmds bool

	session            *discordgo.Session
	registeredCommands []*discordgo.ApplicationCommand

	mu      sync.RWMutex
	started bool
}

// Config holds Discord handler configuration.
type Config struct {
	Token               string
	GuildID             string
	StatusChannelID     string
	Router              handlers.MessageRouter
	Registry            *registry.Registry
	Logger              *observability.Logger
	EnableSlashCommands bool
}

// slashCommands defines available slash commands.
var slashCommands = []*discordgo.ApplicationCommand{
	{
		Name:        "status",
		Description: "Show bot health and uptime",
	},
	{
		Name:        "tools",
		Description: "List available tools",
	},
	{
		Name:        "help",
		Description: "Show usage instructions",
	},
}

// New creates a new Discord event handler.
func New(cfg Config) *Handler {
	logger := cfg.Logger
	if logger == nil {
		logger = observability.New(observability.WithLevel(observability.LevelInfo))
	}

	return &Handler{
		token:           cfg.Token,
		guildID:         cfg.GuildID,
		statusChannelID: cfg.StatusChannelID,
		router:          cfg.Router,
		registry:        cfg.Registry,
		logger:          logger.WithPlatform("discord"),
		enableSlashCmds: cfg.EnableSlashCommands,
	}
}

// Start begins listening for Discord events.
func (h *Handler) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.started {
		return nil
	}

	if h.token == "" {
		return errors.New("discord bot token is required")
	}

	// Create Discord session
	session, err := discordgo.New("Bot " + h.token)
	if err != nil {
		return fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Set intents
	session.Identify.Intents = discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsMessageContent

	// Register event handlers
	session.AddHandler(h.handleReady)
	session.AddHandler(h.handleMessageCreate)
	session.AddHandler(h.handleInteractionCreate)

	// Open websocket connection
	if err := session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord connection: %w", err)
	}

	h.session = session

	// Register slash commands if enabled
	if h.enableSlashCmds {
		if err := h.registerSlashCommands(); err != nil {
			h.logger.Error(context.Background(), "failed to register slash commands", "error", err)
			// Don't fail startup for slash command registration failure
		}
	}

	h.started = true
	h.logger.Info(context.Background(), "Discord handler started")
	return nil
}

// Stop gracefully shuts down the Discord handler.
func (h *Handler) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.started {
		return nil
	}

	// Unregister slash commands if they were registered
	if h.enableSlashCmds && len(h.registeredCommands) > 0 {
		h.unregisterSlashCommands()
	}

	// Close Discord session
	if h.session != nil {
		if err := h.session.Close(); err != nil {
			h.logger.Error(context.Background(), "error closing Discord session", "error", err)
		}
		h.session = nil
	}

	h.started = false
	h.logger.Info(context.Background(), "Discord handler stopped")
	return nil
}

// handleReady is called when the bot successfully connects to Discord.
func (h *Handler) handleReady(s *discordgo.Session, event *discordgo.Ready) {
	h.logger.Info(context.Background(), "Discord bot connected",
		"username", s.State.User.Username,
		"discriminator", s.State.User.Discriminator,
	)
}

// handleMessageCreate processes incoming messages.
func (h *Handler) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := context.Background()

	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if this is a DM or a mention
	isDM := m.GuildID == ""
	isMention := h.isBotMentioned(s, m)

	if !isDM && !isMention {
		// Ignore messages that aren't DMs or mentions
		return
	}

	// Extract content (remove mention if present)
	content := h.cleanMentions(s, m.Content)
	if content == "" {
		return
	}

	h.logger.Info(ctx, "received Discord message",
		"message_id", m.ID,
		"user_id", m.Author.ID,
		"channel_id", m.ChannelID,
		"is_dm", isDM,
	)

	// Create reply function
	replyFunc := func(response string) error {
		_, err := s.ChannelMessageSend(m.ChannelID, response)
		return err
	}

	// Create platform-agnostic message
	msg := handlers.NewMessage(m.ID, m.Author.ID, "discord", content, replyFunc)
	msg.Metadata["channel_id"] = m.ChannelID
	msg.Metadata["guild_id"] = m.GuildID
	msg.Metadata["author_username"] = m.Author.Username

	// Route message if router is configured
	if h.router != nil {
		resp, err := h.router.Route(ctx, msg)
		if err != nil {
			h.logger.Error(ctx, "failed to route message", "error", err)
			if _, sendErr := s.ChannelMessageSend(m.ChannelID, handlers.FormatUserFriendlyError(err)); sendErr != nil {
				h.logger.Error(ctx, "failed to send error reply",
					"message_id", m.ID,
					"error", sendErr,
				)
			}
			return
		}
		if resp != nil && resp.Text != "" {
			if _, sendErr := s.ChannelMessageSend(m.ChannelID, resp.Text); sendErr != nil {
				h.logger.Error(ctx, "failed to send reply after successful routing",
					"message_id", m.ID,
					"error", sendErr,
				)
			}
		}
	}
}

// handleInteractionCreate processes slash command and component interactions.
func (h *Handler) handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h.handleSlashCommand(ctx, s, i)
	case discordgo.InteractionMessageComponent:
		h.handleComponentInteraction(ctx, s, i)
	}
}

// handleSlashCommand processes slash command interactions.
func (h *Handler) handleSlashCommand(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmdName := i.ApplicationCommandData().Name

	// Get user ID safely - Member is nil for DM interactions
	userID := ""
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}

	h.logger.Info(ctx, "received slash command",
		"command", cmdName,
		"user_id", userID,
	)

	var response *discordgo.InteractionResponse

	switch cmdName {
	case "status":
		response = h.handleStatusCommand(ctx)
	case "tools":
		response = h.handleToolsCommand(ctx)
	case "help":
		response = h.handleHelpCommand(ctx)
	default:
		response = &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown command",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		}
	}

	if err := s.InteractionRespond(i.Interaction, response); err != nil {
		h.logger.Error(ctx, "failed to respond to slash command", "error", err)
	}
}

// handleStatusCommand handles the /status slash command.
func (h *Handler) handleStatusCommand(_ context.Context) *discordgo.InteractionResponse {
	h.mu.RLock()
	started := h.started
	h.mu.RUnlock()

	status := "🟢 Online"
	if !started {
		status = "🔴 Offline"
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Bot Status",
		Color:       ColorGreen,
		Description: status,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Platform",
				Value:  "Discord",
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	}
}

// handleToolsCommand handles the /tools slash command.
func (h *Handler) handleToolsCommand(_ context.Context) *discordgo.InteractionResponse {
	var toolsList strings.Builder
	toolsList.WriteString("**Available Tools:**\n")

	if h.registry != nil {
		tools := h.registry.ListTools()
		if len(tools) == 0 {
			toolsList.WriteString("No tools configured.")
		} else {
			for _, tool := range tools {
				toolsList.WriteString(fmt.Sprintf("- ✅ `%s` - %s\n", tool.Name(), tool.Description()))
			}
		}
	} else {
		toolsList.WriteString("No tools registry available.")
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: toolsList.String(),
		},
	}
}

// handleHelpCommand handles the /help slash command.
func (h *Handler) handleHelpCommand(_ context.Context) *discordgo.InteractionResponse {
	embed := &discordgo.MessageEmbed{
		Title:       "MacMini Assistant Help",
		Color:       ColorBlue,
		Description: "I'm your MacMini Assistant! Here's how to use me:",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "💬 Chat",
				Value: "Mention me or send a DM to chat and execute tasks.",
			},
			{
				Name:  "📋 Commands",
				Value: "`/status` - Check bot health\n`/tools` - List available tools\n`/help` - Show this help",
			},
			{
				Name:  "🎬 Download Videos",
				Value: "Send a video URL to download it using Downie.",
			},
			{
				Name:  "☁️ Upload to Drive",
				Value: "Request files to be uploaded to Google Drive.",
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "MacMini Assistant",
		},
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	}
}

// handleComponentInteraction processes button/select menu interactions.
func (h *Handler) handleComponentInteraction(ctx context.Context, _ *discordgo.Session, i *discordgo.InteractionCreate) {
	// Placeholder for future component interactions
	h.logger.Debug(ctx, "received component interaction",
		"custom_id", i.MessageComponentData().CustomID,
	)
}

// registerSlashCommands registers slash commands with Discord.
func (h *Handler) registerSlashCommands() error {
	if h.session == nil {
		return errors.New("discord session not initialized")
	}

	h.registeredCommands = make([]*discordgo.ApplicationCommand, 0, len(slashCommands))

	for _, cmd := range slashCommands {
		registered, err := h.session.ApplicationCommandCreate(
			h.session.State.User.ID,
			h.guildID, // Use guildID for guild-specific commands (faster), "" for global
			cmd,
		)
		if err != nil {
			h.logger.Error(context.Background(), "failed to register slash command",
				"command", cmd.Name,
				"error", err,
			)
			continue
		}
		h.registeredCommands = append(h.registeredCommands, registered)
		h.logger.Info(context.Background(), "registered slash command", "command", cmd.Name)
	}

	return nil
}

// unregisterSlashCommands removes slash commands from Discord.
func (h *Handler) unregisterSlashCommands() {
	if h.session == nil {
		return
	}

	for _, cmd := range h.registeredCommands {
		if err := h.session.ApplicationCommandDelete(
			h.session.State.User.ID,
			h.guildID,
			cmd.ID,
		); err != nil {
			h.logger.Error(context.Background(), "failed to delete slash command",
				"command", cmd.Name,
				"error", err,
			)
		}
	}

	h.registeredCommands = nil
}

// PostStatus sends a status message to the configured status channel.
// Implements handlers.StatusReporter interface.
func (h *Handler) PostStatus(ctx context.Context, msg handlers.StatusMessage) error {
	h.mu.RLock()
	session := h.session
	statusChannelID := h.statusChannelID
	h.mu.RUnlock()

	if session == nil {
		return errors.New("discord session not initialized")
	}

	if statusChannelID == "" {
		return nil // No status channel configured, silently skip
	}

	embed := h.createStatusEmbed(msg)

	_, err := session.ChannelMessageSendEmbed(statusChannelID, embed)
	if err != nil {
		h.logger.Error(ctx, "failed to post status message", "error", err)
		return fmt.Errorf("failed to post status message: %w", err)
	}

	return nil
}

// createStatusEmbed creates a Discord embed for a status message.
func (h *Handler) createStatusEmbed(msg handlers.StatusMessage) *discordgo.MessageEmbed {
	var title string
	var color int
	var description string

	switch msg.Type {
	case "start":
		title = fmt.Sprintf("🎬 %s Started", msg.ToolName)
		color = ColorBlue
	case "progress":
		title = fmt.Sprintf("⏳ %s In Progress", msg.ToolName)
		color = ColorYellow
		description = msg.Message
	case "complete":
		title = fmt.Sprintf("✅ %s Complete", msg.ToolName)
		color = ColorGreen
	case "error":
		title = fmt.Sprintf("❌ %s Failed", msg.ToolName)
		color = ColorRed
		if msg.Error != nil {
			description = msg.Error.Error()
		}
	default:
		title = fmt.Sprintf("ℹ️ %s", msg.ToolName)
		color = ColorBlue
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Color:       color,
		Description: description,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	// Add fields
	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Tool",
			Value:  msg.ToolName,
			Inline: true,
		},
	}

	if msg.UserID != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "User",
			Value:  fmt.Sprintf("<@%s>", msg.UserID),
			Inline: true,
		})
	}

	if msg.Platform != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Platform",
			Value:  msg.Platform,
			Inline: true,
		})
	}

	if msg.Duration > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Duration",
			Value:  msg.Duration.Round(time.Millisecond).String(),
			Inline: true,
		})
	}

	// Add result fields
	for key, value := range msg.Result {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   key,
			Value:  fmt.Sprintf("%v", value),
			Inline: true,
		})
	}

	embed.Fields = fields

	return embed
}

// isBotMentioned checks if the bot was mentioned in the message.
func (h *Handler) isBotMentioned(s *discordgo.Session, m *discordgo.MessageCreate) bool {
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			return true
		}
	}
	return false
}

// cleanMentions removes bot mentions from the message content.
func (h *Handler) cleanMentions(s *discordgo.Session, content string) string {
	// Remove <@botID> and <@!botID> patterns
	botMention := fmt.Sprintf("<@%s>", s.State.User.ID)
	botMentionNick := fmt.Sprintf("<@!%s>", s.State.User.ID)

	content = strings.ReplaceAll(content, botMention, "")
	content = strings.ReplaceAll(content, botMentionNick, "")

	return strings.TrimSpace(content)
}

// SendMessage sends a message to a specific channel.
// TODO: Implement rate limiting to respect Discord API limits
// See https://discord.com/developers/docs/topics/rate-limits
func (h *Handler) SendMessage(_ context.Context, channelID string, message string) error {
	h.mu.RLock()
	session := h.session
	h.mu.RUnlock()

	if session == nil {
		return errors.New("discord session not initialized")
	}

	_, err := session.ChannelMessageSend(channelID, message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// SendEmbed sends an embed message to a specific channel.
// TODO: Implement rate limiting to respect Discord API limits
// See https://discord.com/developers/docs/topics/rate-limits
func (h *Handler) SendEmbed(_ context.Context, channelID string, embed *discordgo.MessageEmbed) error {
	h.mu.RLock()
	session := h.session
	h.mu.RUnlock()

	if session == nil {
		return errors.New("discord session not initialized")
	}

	_, err := session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		return fmt.Errorf("failed to send embed: %w", err)
	}

	return nil
}
