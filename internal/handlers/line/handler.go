// Package line provides LINE bot webhook handling.
package line

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"

	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
	"github.com/kevinyay945/macmini-assistant-systray/internal/observability"
)

// Compile-time interface check
var _ handlers.Handler = (*Handler)(nil)

// ErrInvalidSignature is returned when the LINE webhook signature is invalid.
var ErrInvalidSignature = errors.New("invalid LINE signature")

// ErrEmptyMessage is returned when a message event contains no content.
var ErrEmptyMessage = errors.New("empty message content")

// MaxMessageLength is the maximum length for reply messages (LINE limit is 5000).
const MaxMessageLength = 5000

// Handler processes LINE bot webhook events.
type Handler struct {
	channelSecret string
	channelToken  string
	bot           *messaging_api.MessagingApiAPI
	router        handlers.MessageRouter
	logger        *observability.Logger

	mu      sync.RWMutex
	started bool
}

// Config holds LINE handler configuration.
type Config struct {
	ChannelSecret string
	ChannelToken  string
	Router        handlers.MessageRouter
	Logger        *observability.Logger
}

// New creates a new LINE webhook handler.
func New(cfg Config) *Handler {
	logger := cfg.Logger
	if logger == nil {
		logger = observability.New(observability.WithLevel(observability.LevelInfo))
	}

	return &Handler{
		channelSecret: cfg.ChannelSecret,
		channelToken:  cfg.ChannelToken,
		router:        cfg.Router,
		logger:        logger.WithPlatform("line"),
	}
}

// Start begins the LINE webhook handler.
// This initializes the LINE Messaging API client.
// Note: LINE uses webhooks, so the actual HTTP server should be started
// separately and route requests to HandleWebhook.
func (h *Handler) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.started {
		return nil
	}

	// Initialize LINE Messaging API client if token is provided
	if h.channelToken != "" {
		bot, err := messaging_api.NewMessagingApiAPI(h.channelToken)
		if err != nil {
			return fmt.Errorf("failed to create LINE messaging API client: %w", err)
		}
		h.bot = bot
	}

	h.started = true
	h.logger.Info(context.Background(), "LINE handler started")
	return nil
}

// Stop gracefully shuts down the LINE handler.
func (h *Handler) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.started {
		return nil
	}

	h.started = false
	h.bot = nil
	h.logger.Info(context.Background(), "LINE handler stopped")
	return nil
}

// HandleWebhook processes incoming LINE webhook requests.
// This is designed to be used with net/http or any HTTP framework.
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Always close the request body to prevent connection leaks
	if r.Body != nil {
		defer r.Body.Close()
	}

	// LINE webhooks only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse and validate the webhook request
	cb, err := webhook.ParseRequest(h.channelSecret, r)
	if err != nil {
		if errors.Is(err, webhook.ErrInvalidSignature) {
			h.logger.Warn(ctx, "invalid LINE signature received")
			http.Error(w, "Invalid signature", http.StatusBadRequest)
			return
		}
		h.logger.Error(ctx, "failed to parse LINE webhook request", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Return 200 OK immediately to LINE (best practice)
	// Process events asynchronously if needed
	w.WriteHeader(http.StatusOK)

	// Process each event
	for _, event := range cb.Events {
		h.processEvent(ctx, event)
	}
}

// HandleWebhookGin processes incoming LINE webhook requests using Gin framework.
// This provides Gin-native integration for the webhook endpoint.
func (h *Handler) HandleWebhookGin(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse and validate the webhook request
	cb, err := webhook.ParseRequest(h.channelSecret, c.Request)
	if err != nil {
		if errors.Is(err, webhook.ErrInvalidSignature) {
			h.logger.Warn(ctx, "invalid LINE signature received")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		h.logger.Error(ctx, "failed to parse LINE webhook request", "error", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Return 200 OK immediately to LINE
	c.Status(http.StatusOK)

	// Process each event
	for _, event := range cb.Events {
		h.processEvent(ctx, event)
	}
}

// processEvent handles a single webhook event.
func (h *Handler) processEvent(ctx context.Context, event webhook.EventInterface) {
	switch e := event.(type) {
	case webhook.MessageEvent:
		h.handleMessageEvent(ctx, e)
	case webhook.FollowEvent:
		h.handleFollowEvent(ctx, e)
	case webhook.UnfollowEvent:
		h.handleUnfollowEvent(ctx, e)
	case webhook.PostbackEvent:
		h.handlePostbackEvent(ctx, e)
	default:
		h.logger.Debug(ctx, "unhandled LINE event type", "type", fmt.Sprintf("%T", event))
	}
}

// handleMessageEvent processes incoming message events.
func (h *Handler) handleMessageEvent(ctx context.Context, e webhook.MessageEvent) {
	// Extract message content based on type
	var content string
	var messageID string

	switch msg := e.Message.(type) {
	case webhook.TextMessageContent:
		content = msg.Text
		messageID = msg.Id
	default:
		// For non-text messages, we might want to acknowledge but not process
		h.logger.Debug(ctx, "received non-text message", "type", fmt.Sprintf("%T", e.Message))
		return
	}

	if content == "" {
		h.logger.Debug(ctx, "received empty message")
		return
	}

	// Get user ID from source
	userID := h.getUserIDFromSource(e.Source)

	h.logger.Info(ctx, "received LINE message",
		"message_id", messageID,
		"user_id", userID,
		"content_length", len(content),
	)

	// Create reply function
	replyFunc := func(response string) error {
		return h.sendReply(ctx, e.ReplyToken, response)
	}

	// Create platform-agnostic message
	msg := handlers.NewMessage(messageID, userID, "line", content, replyFunc)
	msg.Metadata["reply_token"] = e.ReplyToken

	// Route message if router is configured
	if h.router != nil {
		resp, err := h.router.Route(ctx, msg)
		if err != nil {
			h.logger.Error(ctx, "failed to route message", "error", err)
			_ = h.sendReply(ctx, e.ReplyToken, formatErrorMessage(err))
			return
		}
		if resp != nil && resp.Text != "" {
			_ = h.sendReply(ctx, e.ReplyToken, resp.Text)
		}
	}
}

// handleFollowEvent processes follow events (user adds the bot).
func (h *Handler) handleFollowEvent(ctx context.Context, e webhook.FollowEvent) {
	userID := h.getUserIDFromSource(e.Source)
	h.logger.Info(ctx, "user followed bot", "user_id", userID)

	// Send welcome message
	_ = h.sendReply(ctx, e.ReplyToken, "Welcome! I'm your MacMini Assistant. Send me a message to get started.")
}

// handleUnfollowEvent processes unfollow events (user blocks the bot).
func (h *Handler) handleUnfollowEvent(ctx context.Context, e webhook.UnfollowEvent) {
	userID := h.getUserIDFromSource(e.Source)
	h.logger.Info(ctx, "user unfollowed bot", "user_id", userID)
}

// handlePostbackEvent processes postback events from buttons/quick replies.
func (h *Handler) handlePostbackEvent(ctx context.Context, e webhook.PostbackEvent) {
	userID := h.getUserIDFromSource(e.Source)
	h.logger.Info(ctx, "received postback",
		"user_id", userID,
		"data", e.Postback.Data,
	)
}

// getUserIDFromSource extracts the user ID from the event source.
func (h *Handler) getUserIDFromSource(source webhook.SourceInterface) string {
	switch s := source.(type) {
	case webhook.UserSource:
		return s.UserId
	case webhook.GroupSource:
		return s.UserId
	case webhook.RoomSource:
		return s.UserId
	default:
		return ""
	}
}

// sendReply sends a reply message using the reply token.
func (h *Handler) sendReply(ctx context.Context, replyToken string, message string) error {
	h.mu.RLock()
	bot := h.bot
	h.mu.RUnlock()

	if bot == nil {
		return errors.New("LINE bot client not initialized")
	}

	// Truncate message if it exceeds the limit
	if len(message) > MaxMessageLength {
		message = message[:MaxMessageLength-3] + "..."
	}

	_, err := bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
		ReplyToken: replyToken,
		Messages: []messaging_api.MessageInterface{
			messaging_api.TextMessage{
				Text: message,
			},
		},
	})
	if err != nil {
		h.logger.Error(ctx, "failed to send LINE reply", "error", err)
		return fmt.Errorf("failed to send LINE reply: %w", err)
	}

	return nil
}

// PushMessage sends a message to a specific user (not using reply token).
// This is useful for notifications or delayed responses.
func (h *Handler) PushMessage(ctx context.Context, userID string, message string) error {
	h.mu.RLock()
	bot := h.bot
	h.mu.RUnlock()

	if bot == nil {
		return errors.New("LINE bot client not initialized")
	}

	// Truncate message if it exceeds the limit
	if len(message) > MaxMessageLength {
		message = message[:MaxMessageLength-3] + "..."
	}

	_, err := bot.PushMessage(&messaging_api.PushMessageRequest{
		To: userID,
		Messages: []messaging_api.MessageInterface{
			messaging_api.TextMessage{
				Text: message,
			},
		},
	}, "")
	if err != nil {
		h.logger.Error(ctx, "failed to push LINE message", "user_id", userID, "error", err)
		return fmt.Errorf("failed to push LINE message: %w", err)
	}

	return nil
}

// formatErrorMessage formats an error into a user-friendly message.
func formatErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	// Check for specific error types
	if errors.Is(err, context.DeadlineExceeded) {
		return "Request timed out. Please try again."
	}
	if errors.Is(err, context.Canceled) {
		return "Request was cancelled."
	}

	return "An error occurred while processing your request. Please try again later."
}

// ParseMessage converts a LINE MessageEvent into a platform-agnostic Message.
// Exported for testing purposes.
func (h *Handler) ParseMessage(e webhook.MessageEvent) (*handlers.Message, error) {
	var content string
	var messageID string

	switch msg := e.Message.(type) {
	case webhook.TextMessageContent:
		content = msg.Text
		messageID = msg.Id
	default:
		return nil, fmt.Errorf("unsupported message type: %T", e.Message)
	}

	if content == "" {
		return nil, ErrEmptyMessage
	}

	userID := h.getUserIDFromSource(e.Source)

	replyFunc := func(response string) error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return h.sendReply(ctx, e.ReplyToken, response)
	}

	msg := handlers.NewMessage(messageID, userID, "line", content, replyFunc)
	msg.Metadata["reply_token"] = e.ReplyToken

	return msg, nil
}
