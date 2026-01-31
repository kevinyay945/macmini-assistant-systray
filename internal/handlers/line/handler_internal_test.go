package line

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"

	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers/testutil"
)

func TestTruncateMessage(t *testing.T) {
	tests := []struct {
		message  string
		maxLen   int
		expected string
	}{
		{"Hello", 100, "Hello"},
		{"12345", 5, "12345"},
		{"Hello, World!", 10, "Hello, ..."},
		{"你好世界測試", 6, "你好世界測試"},   // exactly 6 runes, no truncation
		{"你好世界測試更多", 6, "你好世..."}, // 8 runes, truncate to 6 (3 chars + ...)
		{"Hello", 3, "..."},
		{"", 10, ""},
	}
	for _, tt := range tests {
		got := truncateMessage(tt.message, tt.maxLen)
		if got != tt.expected {
			t.Errorf("truncateMessage(%q, %d) = %q, want %q", tt.message, tt.maxLen, got, tt.expected)
		}
	}
}

func TestGetUserIDFromSource_NilSource(t *testing.T) {
	h := New(Config{})
	result := h.getUserIDFromSource(nil)
	if result != "" {
		t.Errorf("getUserIDFromSource(nil) = %q, want empty string", result)
	}
}

func TestGetUserIDFromSource_UserSource(t *testing.T) {
	h := New(Config{})
	source := webhook.UserSource{UserId: "U123456"}
	result := h.getUserIDFromSource(source)
	if result != "U123456" {
		t.Errorf("getUserIDFromSource(UserSource) = %q, want %q", result, "U123456")
	}
}

func TestGetUserIDFromSource_GroupSource(t *testing.T) {
	h := New(Config{})
	source := webhook.GroupSource{UserId: "U789", GroupId: "G123"}
	result := h.getUserIDFromSource(source)
	if result != "U789" {
		t.Errorf("getUserIDFromSource(GroupSource) = %q, want %q", result, "U789")
	}
}

func TestGetUserIDFromSource_RoomSource(t *testing.T) {
	h := New(Config{})
	source := webhook.RoomSource{UserId: "UABC", RoomId: "R456"}
	result := h.getUserIDFromSource(source)
	if result != "UABC" {
		t.Errorf("getUserIDFromSource(RoomSource) = %q, want %q", result, "UABC")
	}
}

func TestGetUserIDFromSource_EmptyUserId(t *testing.T) {
	h := New(Config{})
	source := webhook.UserSource{UserId: ""}
	result := h.getUserIDFromSource(source)
	if result != "" {
		t.Errorf("getUserIDFromSource(empty user) = %q, want empty", result)
	}
}

func TestParseMessage_TextMessage(t *testing.T) {
	h := New(Config{})
	event := webhook.MessageEvent{
		ReplyToken: "reply-token-123",
		Source:     webhook.UserSource{UserId: "U123"},
		Message:    webhook.TextMessageContent{Id: "msg-123", Text: "Hello world"},
	}
	msg, err := h.ParseMessage(event)
	if err != nil {
		t.Errorf("ParseMessage() error = %v", err)
	}
	if msg.Content != "Hello world" {
		t.Errorf("ParseMessage() content = %q, want %q", msg.Content, "Hello world")
	}
	if msg.UserID != "U123" {
		t.Errorf("ParseMessage() userID = %q, want %q", msg.UserID, "U123")
	}
	if msg.Metadata["reply_token"] != "reply-token-123" {
		t.Errorf("ParseMessage() reply_token = %q, want %q", msg.Metadata["reply_token"], "reply-token-123")
	}
}

func TestParseMessage_EmptyText(t *testing.T) {
	h := New(Config{})
	event := webhook.MessageEvent{
		ReplyToken: "token",
		Source:     webhook.UserSource{UserId: "U123"},
		Message:    webhook.TextMessageContent{Id: "msg-123", Text: ""},
	}
	_, err := h.ParseMessage(event)
	if err != ErrEmptyMessage {
		t.Errorf("ParseMessage(empty) error = %v, want ErrEmptyMessage", err)
	}
}

func TestParseMessage_UnsupportedMessageType(t *testing.T) {
	h := New(Config{})
	event := webhook.MessageEvent{
		ReplyToken: "token",
		Source:     webhook.UserSource{UserId: "U123"},
		Message:    webhook.ImageMessageContent{Id: "img-123"},
	}
	_, err := h.ParseMessage(event)
	if err == nil {
		t.Error("ParseMessage(image) should return error")
	}
}

func TestSendReply_NilBot(t *testing.T) {
	h := New(Config{})
	err := h.sendReply(context.Background(), "token", "message")
	if err == nil {
		t.Error("sendReply should return error when bot is nil")
	}
}

func TestPushMessage_NilBot(t *testing.T) {
	h := New(Config{})
	err := h.PushMessage(context.Background(), "user", "message")
	if err == nil {
		t.Error("PushMessage should return error when bot is nil")
	}
}

func TestHandler_HandleWebhook_NilBody(t *testing.T) {
	h := New(Config{ChannelSecret: "test-secret"})
	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleWebhook() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_HandleWebhook_MissingSignature(t *testing.T) {
	h := New(Config{ChannelSecret: "test-secret"})
	body := strings.NewReader(`{"events":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleWebhook(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleWebhook() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_HandleWebhookGin_InvalidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := New(Config{ChannelSecret: "test-secret"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := strings.NewReader(`{"events":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Line-Signature", "invalid")
	c.Request = req
	h.HandleWebhookGin(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleWebhookGin() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_HandleWebhookGin_MissingSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := New(Config{ChannelSecret: "test-secret"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := strings.NewReader(`{"events":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/webhook", body)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	h.HandleWebhookGin(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleWebhookGin() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_StartWithToken(t *testing.T) {
	h := New(Config{ChannelSecret: "secret", ChannelToken: "token"})
	if err := h.Start(); err != nil {
		t.Errorf("Start() returned error: %v", err)
	}
	h.mu.RLock()
	hasBot := h.bot != nil
	h.mu.RUnlock()
	if !hasBot {
		t.Error("Start() should initialize bot when token is provided")
	}
	_ = h.Stop()
}

func TestHandler_StopClearsBot(t *testing.T) {
	h := New(Config{ChannelSecret: "secret", ChannelToken: "token"})
	_ = h.Start()
	_ = h.Stop()
	h.mu.RLock()
	hasBot := h.bot != nil
	h.mu.RUnlock()
	if hasBot {
		t.Error("Stop() should clear bot")
	}
}

func TestHandler_ProcessEvent_UnknownEventType(t *testing.T) {
	h := New(Config{})
	h.processEvent(context.Background(), nil) // Should not panic
}

func TestHandler_ProcessEvent_MessageEvent(t *testing.T) {
	h := New(Config{})
	event := webhook.MessageEvent{
		ReplyToken: "token",
		Source:     webhook.UserSource{UserId: "U123"},
		Message:    webhook.TextMessageContent{Id: "msg-1", Text: "hello"},
	}
	h.processEvent(context.Background(), event) // Should not panic, just log
}

func TestHandler_ProcessEvent_MessageEvent_NonText(t *testing.T) {
	h := New(Config{})
	event := webhook.MessageEvent{
		ReplyToken: "token",
		Source:     webhook.UserSource{UserId: "U123"},
		Message:    webhook.ImageMessageContent{Id: "img-1"},
	}
	h.processEvent(context.Background(), event) // Should not panic, just log
}

func TestHandler_ProcessEvent_MessageEvent_EmptyContent(t *testing.T) {
	h := New(Config{})
	event := webhook.MessageEvent{
		ReplyToken: "token",
		Source:     webhook.UserSource{UserId: "U123"},
		Message:    webhook.TextMessageContent{Id: "msg-1", Text: ""},
	}
	h.processEvent(context.Background(), event) // Should not panic, just log
}

func TestHandler_ProcessEvent_MessageEvent_WithRouter(t *testing.T) {
	mockRouter := &testutil.MockRouter{
		Response: &handlers.Response{Text: "processed"},
	}
	h := New(Config{Router: mockRouter})
	event := webhook.MessageEvent{
		ReplyToken: "token",
		Source:     webhook.UserSource{UserId: "U123"},
		Message:    webhook.TextMessageContent{Id: "msg-1", Text: "hello"},
	}
	h.processEvent(context.Background(), event)
	if !mockRouter.Called {
		t.Error("Router should be called for message event")
	}
}

func TestHandler_ProcessEvent_MessageEvent_RouterError(t *testing.T) {
	mockRouter := &testutil.MockRouter{
		Err: errors.New("routing failed"),
	}
	h := New(Config{Router: mockRouter})
	event := webhook.MessageEvent{
		ReplyToken: "token",
		Source:     webhook.UserSource{UserId: "U123"},
		Message:    webhook.TextMessageContent{Id: "msg-1", Text: "hello"},
	}
	h.processEvent(context.Background(), event) // Should not panic
	if !mockRouter.Called {
		t.Error("Router should be called even when error occurs")
	}
}

func TestHandler_ProcessEvent_FollowEvent(t *testing.T) {
	h := New(Config{})
	event := webhook.FollowEvent{
		ReplyToken: "token",
		Source:     webhook.UserSource{UserId: "U123"},
	}
	h.processEvent(context.Background(), event) // Should not panic
}

func TestHandler_HandleMessageEvent_RouteSuccessWithResponse(t *testing.T) {
	mockRouter := &testutil.MockRouter{
		Response: &handlers.Response{Text: "response text"},
	}
	h := New(Config{Router: mockRouter})
	event := webhook.MessageEvent{
		ReplyToken: "token",
		Source:     webhook.UserSource{UserId: "U123"},
		Message:    webhook.TextMessageContent{Id: "msg-1", Text: "test"},
	}
	h.handleMessageEvent(context.Background(), event)
	if !mockRouter.Called {
		t.Error("Router should be called")
	}
}

func TestHandler_HandleMessageEvent_RouteSuccessEmptyResponse(t *testing.T) {
	mockRouter := &testutil.MockRouter{
		Response: &handlers.Response{Text: ""},
	}
	h := New(Config{Router: mockRouter})
	event := webhook.MessageEvent{
		ReplyToken: "token",
		Source:     webhook.UserSource{UserId: "U123"},
		Message:    webhook.TextMessageContent{Id: "msg-1", Text: "test"},
	}
	h.handleMessageEvent(context.Background(), event)
	if !mockRouter.Called {
		t.Error("Router should be called")
	}
}

func TestHandler_HandleMessageEvent_RouteSuccessNilResponse(t *testing.T) {
	mockRouter := &testutil.MockRouter{
		Response: nil,
	}
	h := New(Config{Router: mockRouter})
	event := webhook.MessageEvent{
		ReplyToken: "token",
		Source:     webhook.UserSource{UserId: "U123"},
		Message:    webhook.TextMessageContent{Id: "msg-1", Text: "test"},
	}
	h.handleMessageEvent(context.Background(), event)
	if !mockRouter.Called {
		t.Error("Router should be called")
	}
}

func TestHandler_ProcessEvent_UnfollowEvent(t *testing.T) {
	h := New(Config{})
	event := webhook.UnfollowEvent{
		Source: webhook.UserSource{UserId: "U123"},
	}
	h.processEvent(context.Background(), event) // Should not panic
}

func TestHandler_ProcessEvent_PostbackEvent(t *testing.T) {
	h := New(Config{})
	event := webhook.PostbackEvent{
		Source:   webhook.UserSource{UserId: "U123"},
		Postback: &webhook.PostbackContent{Data: "action=test"},
	}
	h.processEvent(context.Background(), event) // Should not panic
}

func TestTruncationSuffix(t *testing.T) {
	if TruncationSuffix != "..." {
		t.Errorf("TruncationSuffix = %q, want %q", TruncationSuffix, "...")
	}
}
