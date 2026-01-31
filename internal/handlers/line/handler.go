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

// ErrEmptyMessage is returned when a message event contains no content.
var ErrEmptyMessage = errors.New("empty message content")

// MaxMessageLength is the maximum length for reply messages (LINE limit is 5000 characters).
const MaxMessageLength = 5000

// TruncationSuffix is appended to truncated messages.
const TruncationSuffix = "..."

// DefaultReplyTimeout is the default timeout for reply message operations.
const DefaultReplyTimeout = 30 * time.Second

// EventProcessingTimeout is the timeout for processing webhook events asynchronously.
const EventProcessingTimeout = 10 * time.Minute

// Retry configuration for API client initialization.
const (
	maxRetries = 3
	retryDelay = 2 * time.Second
)

// Handler processes LINE bot webhook events.
type Handler struct {
	channelSecret string
	channelToken  string
	bot           *messaging_api.MessagingApiAPI
	router        handlers.MessageRouter
	logger        *observability.Logger

	mu         sync.RWMutex
	started    bool
	shutdownCh chan struct{}
	wg         sync.WaitGroup
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
		var bot *messaging_api.MessagingApiAPI
		var lastErr error

		for i := 0; i < maxRetries; i++ {
			bot, lastErr = messaging_api.NewMessagingApiAPI(h.channelToken)
			if lastErr == nil {
				break
			}
			h.logger.Warn(context.Background(), "retrying LINE API client creation",
				"attempt", i+1,
				"error", lastErr,
			)
			if i < maxRetries-1 {
				time.Sleep(retryDelay * time.Duration(i+1))
			}
		}

		if lastErr != nil {
			return fmt.Errorf("failed to create LINE messaging API client after %d attempts: %w", maxRetries, lastErr)
		}
		h.bot = bot
	}

	h.shutdownCh = make(chan struct{})
	h.started = true
	h.logger.Info(context.Background(), "LINE handler started")
	return nil
}

// Stop gracefully shuts down the LINE handler.
// It waits for all in-flight webhook processing to complete.
func (h *Handler) Stop() error {
	h.mu.Lock()
	if !h.started {
		h.mu.Unlock()
		return nil
	}

	// Signal shutdown to prevent new goroutines
	close(h.shutdownCh)
	h.mu.Unlock()

	// Wait for all in-flight webhook processing to complete
	h.wg.Wait()

	h.mu.Lock()
	defer h.mu.Unlock()

	h.started = false
	h.bot = nil
	h.logger.Info(context.Background(), "LINE handler stopped")
	return nil
}

// HandleWebhook processes incoming LINE webhook requests.
// This is designed to be used with net/http or any HTTP framework.
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
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
			h.logger.Warn(r.Context(), "invalid LINE signature received")
			http.Error(w, "Invalid signature", http.StatusBadRequest)
			return
		}
		h.logger.Error(r.Context(), "failed to parse LINE webhook request", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Return 200 OK immediately to LINE (best practice)
	// LINE has a 1-second timeout, so we must respond quickly
	w.WriteHeader(http.StatusOK)

	// Check if we're shutting down before spawning a new goroutine
	h.mu.RLock()
	shutdownCh := h.shutdownCh
	h.mu.RUnlock()

	select {
	case <-shutdownCh:
		h.logger.Warn(r.Context(), "rejecting webhook during shutdown")
		return
	default:
	}

	// Process events asynchronously to avoid blocking the response
	// Create a new context since request context will be cancelled after response
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()

		ctx, cancel := context.WithTimeout(context.Background(), EventProcessingTimeout)
		defer cancel()

		for _, event := range cb.Events {
			// Check for shutdown signal between events
			select {
			case <-shutdownCh:
				h.logger.Warn(ctx, "stopping event processing due to shutdown")
				return
			default:
			}
			h.processEvent(ctx, event)
		}
	}()
}

// HandleWebhookGin processes incoming LINE webhook requests using Gin framework.
// This provides Gin-native integration for the webhook endpoint.
func (h *Handler) HandleWebhookGin(c *gin.Context) {
	// Parse and validate the webhook request
	cb, err := webhook.ParseRequest(h.channelSecret, c.Request)
	if err != nil {
		if errors.Is(err, webhook.ErrInvalidSignature) {
			h.logger.Warn(c.Request.Context(), "invalid LINE signature received")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		h.logger.Error(c.Request.Context(), "failed to parse LINE webhook request", "error", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Return 200 OK immediately to LINE
	// LINE has a 1-second timeout, so we must respond quickly
	c.Status(http.StatusOK)

	// Check if we're shutting down before spawning a new goroutine
	h.mu.RLock()
	shutdownCh := h.shutdownCh
	h.mu.RUnlock()

	select {
	case <-shutdownCh:
		h.logger.Warn(c.Request.Context(), "rejecting webhook during shutdown")
		return
	default:
	}

	// Process events asynchronously to avoid blocking the response
	// Create a new context since request context will be cancelled after response
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()

		ctx, cancel := context.WithTimeout(context.Background(), EventProcessingTimeout)
		defer cancel()

		for _, event := range cb.Events {
			// Check for shutdown signal between events
			select {
			case <-shutdownCh:
				h.logger.Warn(ctx, "stopping event processing due to shutdown")
				return
			default:
			}
			h.processEvent(ctx, event)
		}
	}()
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
	msg := handlers.NewMessage(messageID, userID, handlers.PlatformLINE, content, replyFunc)
	msg.Metadata["reply_token"] = e.ReplyToken

	// Route message if router is configured
	if h.router != nil {
		resp, err := h.router.Route(ctx, msg)
		if err != nil {
			h.logger.Error(ctx, "failed to route message", "error", err)
			if replyErr := h.sendReply(ctx, e.ReplyToken, handlers.FormatUserFriendlyError(err)); replyErr != nil {
				h.logger.Error(ctx, "failed to send error reply",
					"message_id", messageID,
					"error", replyErr,
				)
			}
			return
		}
		if resp != nil && resp.Text != "" {
			if replyErr := h.sendReply(ctx, e.ReplyToken, resp.Text); replyErr != nil {
				h.logger.Error(ctx, "failed to send reply after successful routing",
					"message_id", messageID,
					"error", replyErr,
				)
			}
		}
	}
}

// handleFollowEvent processes follow events (user adds the bot).
func (h *Handler) handleFollowEvent(ctx context.Context, e webhook.FollowEvent) {
	userID := h.getUserIDFromSource(e.Source)
	h.logger.Info(ctx, "user followed bot", "user_id", userID)

	// Send welcome message
	if err := h.sendReply(ctx, e.ReplyToken, "Welcome! I'm your MacMini Assistant. Send me a message to get started."); err != nil {
		h.logger.Error(ctx, "failed to send welcome message", "user_id", userID, "error", err)
	}
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
	var userID string
	var sourceType string

	switch s := source.(type) {
	case webhook.UserSource:
		userID = s.UserId
		sourceType = "user"
	case webhook.GroupSource:
		userID = s.UserId
		sourceType = "group"
	case webhook.RoomSource:
		userID = s.UserId
		sourceType = "room"
	default:
		h.logger.Debug(context.Background(), "unknown source type", "type", fmt.Sprintf("%T", source))
		return ""
	}

	if userID == "" {
		h.logger.Warn(context.Background(), "empty user ID from source", "source_type", sourceType)
	}

	return userID
}

// truncateMessage safely truncates a message to the maximum allowed length.
// It operates on runes to avoid cutting multi-byte Unicode characters.
func truncateMessage(message string, maxLen int) string {
	runes := []rune(message)
	if len(runes) <= maxLen {
		return message
	}
	// Reserve space for truncation suffix
	truncateAt := maxLen - len([]rune(TruncationSuffix))
	if truncateAt < 0 {
		truncateAt = 0
	}
	return string(runes[:truncateAt]) + TruncationSuffix
}

// sendReply sends a reply message using the reply token.
// TODO: Implement rate limiting to respect LINE API limits
// See https://developers.line.biz/en/docs/messaging-api/rate-limits/
func (h *Handler) sendReply(ctx context.Context, replyToken string, message string) error {
	h.mu.RLock()
	bot := h.bot
	h.mu.RUnlock()

	if bot == nil {
		return handlers.ErrBotNotInitialized
	}

	// Truncate message if it exceeds the limit (using rune-safe truncation)
	message = truncateMessage(message, MaxMessageLength)

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
// TODO: Implement rate limiting to respect LINE API limits
// See https://developers.line.biz/en/docs/messaging-api/rate-limits/
func (h *Handler) PushMessage(ctx context.Context, userID string, message string) error {
	h.mu.RLock()
	bot := h.bot
	h.mu.RUnlock()

	if bot == nil {
		return handlers.ErrBotNotInitialized
	}

	// Truncate message if it exceeds the limit (using rune-safe truncation)
	message = truncateMessage(message, MaxMessageLength)

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
		ctx, cancel := context.WithTimeout(context.Background(), DefaultReplyTimeout)
		defer cancel()
		return h.sendReply(ctx, e.ReplyToken, response)
	}

	msg := handlers.NewMessage(messageID, userID, handlers.PlatformLINE, content, replyFunc)
	msg.Metadata["reply_token"] = e.ReplyToken

	return msg, nil
}
