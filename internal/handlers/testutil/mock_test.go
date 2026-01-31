package testutil

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/handlers"
)

func TestMockRouter_Route_Basic(t *testing.T) {
	expectedResp := &handlers.Response{Text: "test response"}
	m := NewMockRouter()
	m.SetResponse(expectedResp)

	msg := &handlers.Message{Content: "test", Platform: handlers.PlatformDiscord}
	resp, err := m.Route(context.Background(), msg)

	if err != nil {
		t.Errorf("Route() error = %v", err)
	}
	if !m.Called() {
		t.Error("Route() should set Called = true")
	}
	if m.LastMsg() != msg {
		t.Error("Route() should store LastMsg")
	}
	if resp != expectedResp {
		t.Errorf("Route() response = %v, want %v", resp, expectedResp)
	}
}

func TestMockRouter_Route_NoResponse(t *testing.T) {
	m := NewMockRouter()
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
	m := NewMockRouter()
	m.SetError(expectedErr)

	_, err := m.Route(context.Background(), &handlers.Message{})

	if !errors.Is(err, expectedErr) {
		t.Errorf("Route() error = %v, want %v", err, expectedErr)
	}
}

func TestMockRouter_Route_CapturesMessage(t *testing.T) {
	m := NewMockRouter()
	msg := &handlers.Message{
		Content:  "test content",
		Platform: handlers.PlatformLINE,
		UserID:   "user123",
	}

	_, _ = m.Route(context.Background(), msg)

	lastMsg := m.LastMsg()
	if lastMsg.Content != "test content" {
		t.Errorf("LastMsg().Content = %q, want %q", lastMsg.Content, "test content")
	}
	if lastMsg.Platform != handlers.PlatformLINE {
		t.Errorf("LastMsg().Platform = %q, want %q", lastMsg.Platform, handlers.PlatformLINE)
	}
	if lastMsg.UserID != "user123" {
		t.Errorf("LastMsg().UserID = %q, want %q", lastMsg.UserID, "user123")
	}
}

func TestMockRouter_Reset(t *testing.T) {
	m := NewMockRouter()
	m.SetResponse(&handlers.Response{Text: "test"})
	_, _ = m.Route(context.Background(), &handlers.Message{Content: "test"})

	if !m.Called() {
		t.Fatal("expected Called() = true before reset")
	}
	if m.LastMsg() == nil {
		t.Fatal("expected LastMsg() != nil before reset")
	}

	m.Reset()

	if m.Called() {
		t.Error("Reset() should set Called = false")
	}
	if m.LastMsg() != nil {
		t.Error("Reset() should set LastMsg = nil")
	}
}

func TestMockRouter_Reset_PreservesConfig(t *testing.T) {
	expectedResp := &handlers.Response{Text: "preserved"}
	expectedErr := errors.New("preserved error")
	m := NewMockRouter()
	m.SetResponse(expectedResp)
	m.SetError(expectedErr)

	_, _ = m.Route(context.Background(), &handlers.Message{})
	m.Reset()

	if m.Response() != expectedResp {
		t.Error("Reset() should preserve Response")
	}
	if !errors.Is(m.Err(), expectedErr) {
		t.Error("Reset() should preserve Err")
	}
}

func TestMockRouter_ConcurrentAccess(t *testing.T) {
	m := NewMockRouter()
	m.SetResponse(&handlers.Response{Text: "ok"})
	var wg sync.WaitGroup

	// Spawn multiple goroutines to access the mock concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			msg := &handlers.Message{
				Content:  "test",
				Platform: handlers.PlatformDiscord,
				UserID:   "user",
			}
			_, _ = m.Route(context.Background(), msg)
			_ = m.Called()
			_ = m.LastMsg()
		}(i)
	}

	wg.Wait()

	if !m.Called() {
		t.Error("Expected Called() = true after concurrent access")
	}
	if m.LastMsg() == nil {
		t.Error("Expected LastMsg() != nil after concurrent access")
	}
}

func TestMockRouter_ConcurrentReset(t *testing.T) {
	m := NewMockRouter()
	m.SetResponse(&handlers.Response{Text: "ok"})
	var wg sync.WaitGroup

	// Test concurrent Route and Reset operations
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _ = m.Route(context.Background(), &handlers.Message{Content: "test"})
		}()
		go func() {
			defer wg.Done()
			m.Reset()
		}()
	}

	wg.Wait()
	// Should not panic or race - the test itself is a success if we get here
}
