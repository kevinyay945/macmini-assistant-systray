package testutil

import (
	"context"
	"errors"
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
)

func TestMockRouter_Route_Basic(t *testing.T) {
	expectedResp := &handlers.Response{Text: "test response"}
	m := &MockRouter{Response: expectedResp}

	msg := &handlers.Message{Content: "test", Platform: handlers.PlatformDiscord}
	resp, err := m.Route(context.Background(), msg)

	if err != nil {
		t.Errorf("Route() error = %v", err)
	}
	if !m.Called {
		t.Error("Route() should set Called = true")
	}
	if m.LastMsg != msg {
		t.Error("Route() should store LastMsg")
	}
	if resp != expectedResp {
		t.Errorf("Route() response = %v, want %v", resp, expectedResp)
	}
}

func TestMockRouter_Route_NoResponse(t *testing.T) {
	m := &MockRouter{}
	resp, err := m.Route(context.Background(), &handlers.Message{Content: "test"})

	if err != nil {
		t.Errorf("Route() error = %v", err)
	}
	if resp != nil {
		t.Errorf("Route() response = %v, want nil", resp)
	}
}

func TestMockRouter_Route_WithError(t *testing.T) {
	expectedErr := errors.New("mock error")
	m := &MockRouter{Err: expectedErr}

	_, err := m.Route(context.Background(), &handlers.Message{})

	if err != expectedErr {
		t.Errorf("Route() error = %v, want %v", err, expectedErr)
	}
}

func TestMockRouter_Route_CapturesMessage(t *testing.T) {
	m := &MockRouter{}
	msg := &handlers.Message{
		Content:  "test content",
		Platform: handlers.PlatformLINE,
		UserID:   "user123",
	}

	_, _ = m.Route(context.Background(), msg)

	if m.LastMsg.Content != "test content" {
		t.Errorf("LastMsg.Content = %q, want %q", m.LastMsg.Content, "test content")
	}
	if m.LastMsg.Platform != handlers.PlatformLINE {
		t.Errorf("LastMsg.Platform = %q, want %q", m.LastMsg.Platform, handlers.PlatformLINE)
	}
	if m.LastMsg.UserID != "user123" {
		t.Errorf("LastMsg.UserID = %q, want %q", m.LastMsg.UserID, "user123")
	}
}

func TestMockRouter_Reset(t *testing.T) {
	m := &MockRouter{
		Response: &handlers.Response{Text: "test"},
	}
	_, _ = m.Route(context.Background(), &handlers.Message{Content: "test"})

	if !m.Called {
		t.Fatal("expected Called = true before reset")
	}
	if m.LastMsg == nil {
		t.Fatal("expected LastMsg != nil before reset")
	}

	m.Reset()

	if m.Called {
		t.Error("Reset() should set Called = false")
	}
	if m.LastMsg != nil {
		t.Error("Reset() should set LastMsg = nil")
	}
}

func TestMockRouter_Reset_PreservesConfig(t *testing.T) {
	expectedResp := &handlers.Response{Text: "preserved"}
	expectedErr := errors.New("preserved error")
	m := &MockRouter{
		Response: expectedResp,
		Err:      expectedErr,
	}

	_, _ = m.Route(context.Background(), &handlers.Message{})
	m.Reset()

	if m.Response != expectedResp {
		t.Error("Reset() should preserve Response")
	}
	if m.Err != expectedErr {
		t.Error("Reset() should preserve Err")
	}
}
