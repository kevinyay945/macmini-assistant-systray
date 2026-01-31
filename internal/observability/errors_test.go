package observability_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/observability"
)

func TestAppError_Error(t *testing.T) {
	err := &observability.AppError{Code: "TEST_CODE", Message: "test message"}

	got := err.Error()
	expected := "[TEST_CODE] test message"
	if got != expected {
		t.Errorf("Error() = %q, want %q", got, expected)
	}
}

func TestAppError_ErrorWithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := &observability.AppError{Code: "TEST_CODE", Message: "test message", Cause: cause}

	got := err.Error()
	if !strings.Contains(got, "TEST_CODE") || !strings.Contains(got, "test message") || !strings.Contains(got, "underlying error") {
		t.Errorf("Error() = %q, should contain code, message, and cause", got)
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &observability.AppError{Code: "TEST_CODE", Message: "test", Cause: cause}

	unwrapped := err.Unwrap()
	if !errors.Is(unwrapped, cause) {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestAppError_WithCause(t *testing.T) {
	original := &observability.AppError{Code: "TEST_CODE", Message: "test"}
	cause := errors.New("new cause")

	wrapped := original.WithCause(cause)
	if !errors.Is(wrapped.Cause, cause) {
		t.Error("WithCause() should set the cause")
	}
	if wrapped == original {
		t.Error("WithCause() should return a new AppError")
	}
	if wrapped.Code != original.Code || wrapped.Message != original.Message {
		t.Error("WithCause() should preserve code and message")
	}
}

func TestAppError_WithMessage(t *testing.T) {
	original := &observability.AppError{Code: "TEST_CODE", Message: "original"}

	modified := original.WithMessage("new message")
	if modified.Message != "new message" {
		t.Errorf("WithMessage() Message = %q, want %q", modified.Message, "new message")
	}
	if modified.Code != original.Code {
		t.Error("WithMessage() should preserve code")
	}
}

func TestAppError_WithRequestID(t *testing.T) {
	original := &observability.AppError{Code: "TEST_CODE", Message: "test"}

	modified := original.WithRequestID("req-123")
	if modified.RequestID != "req-123" {
		t.Errorf("WithRequestID() RequestID = %q, want %q", modified.RequestID, "req-123")
	}
}

func TestAppError_WithExtra(t *testing.T) {
	original := &observability.AppError{Code: "TEST_CODE", Message: "test"}

	modified := original.WithExtra("key1", "value1").WithExtra("key2", 42)
	if modified.Extra["key1"] != "value1" {
		t.Error("WithExtra() should add key1")
	}
	if modified.Extra["key2"] != 42 {
		t.Error("WithExtra() should add key2")
	}
}

func TestAppError_Is(t *testing.T) {
	err1 := &observability.AppError{Code: "SAME_CODE", Message: "error 1"}
	err2 := &observability.AppError{Code: "SAME_CODE", Message: "error 2"}
	err3 := &observability.AppError{Code: "DIFFERENT_CODE", Message: "error 3"}

	if !err1.Is(err2) {
		t.Error("Is() should return true for same code")
	}
	if err1.Is(err3) {
		t.Error("Is() should return false for different code")
	}
}

func TestSentinelErrors(t *testing.T) {
	testCases := []struct {
		name string
		err  *observability.AppError
		code string
	}{
		{"ErrConfigNotFound", observability.ErrConfigNotFound, observability.CodeConfigNotFound},
		{"ErrToolNotFound", observability.ErrToolNotFound, observability.CodeToolNotFound},
		{"ErrToolTimeout", observability.ErrToolTimeout, observability.CodeToolTimeout},
		{"ErrInvalidParams", observability.ErrInvalidParams, observability.CodeInvalidParams},
		{"ErrCopilotConnection", observability.ErrCopilotConnection, observability.CodeCopilotConnection},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Code != tc.code {
				t.Errorf("Code = %q, want %q", tc.err.Code, tc.code)
			}
		})
	}
}

func TestNewAppError(t *testing.T) {
	err := observability.NewAppError("CUSTOM_CODE", "custom message")

	if err.Code != "CUSTOM_CODE" {
		t.Errorf("Code = %q, want %q", err.Code, "CUSTOM_CODE")
	}
	if err.Message != "custom message" {
		t.Errorf("Message = %q, want %q", err.Message, "custom message")
	}
}

func TestWrapError(t *testing.T) {
	cause := errors.New("original error")
	err := observability.WrapError("WRAP_CODE", "wrapped", cause)

	if err.Code != "WRAP_CODE" {
		t.Errorf("Code = %q, want %q", err.Code, "WRAP_CODE")
	}
	if !errors.Is(err.Cause, cause) {
		t.Error("Cause should be the original error")
	}
}

func TestIsAppError(t *testing.T) {
	err := observability.NewAppError("TEST_CODE", "test")

	if !observability.IsAppError(err, "TEST_CODE") {
		t.Error("IsAppError() should return true for matching code")
	}
	if observability.IsAppError(err, "OTHER_CODE") {
		t.Error("IsAppError() should return false for non-matching code")
	}
	if observability.IsAppError(errors.New("plain error"), "TEST_CODE") {
		t.Error("IsAppError() should return false for non-AppError")
	}
}

func TestGetAppError(t *testing.T) {
	appErr := observability.NewAppError("TEST_CODE", "test")

	got, ok := observability.GetAppError(appErr)
	if !ok {
		t.Error("GetAppError() should return true for AppError")
	}
	if got != appErr {
		t.Error("GetAppError() should return the same error")
	}

	_, ok = observability.GetAppError(errors.New("plain error"))
	if ok {
		t.Error("GetAppError() should return false for non-AppError")
	}
}

func TestAppError_UserMessage(t *testing.T) {
	testCases := []struct {
		code     string
		contains string
	}{
		{observability.CodeConfigNotFound, "Configuration"},
		{observability.CodeToolNotFound, "tool"},
		{observability.CodeToolTimeout, "too long"},
		{observability.CodeInvalidParams, "Invalid"},
		{observability.CodeCopilotConnection, "AI service"},
		{observability.CodeAuthFailed, "Authentication"},
		{observability.CodeMessageFailed, "message"},
		{"UNKNOWN_CODE", "unexpected"},
	}

	for _, tc := range testCases {
		t.Run(tc.code, func(t *testing.T) {
			err := observability.NewAppError(tc.code, "internal message")
			msg := err.UserMessage()
			if !strings.Contains(msg, tc.contains) {
				t.Errorf("UserMessage() = %q, should contain %q", msg, tc.contains)
			}
		})
	}
}
