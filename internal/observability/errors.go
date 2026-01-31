package observability

import (
	"errors"
	"fmt"
)

// Error codes for the application.
const (
	CodeConfigNotFound    = "CONFIG_NOT_FOUND"
	CodeToolNotFound      = "TOOL_NOT_FOUND"
	CodeToolTimeout       = "TOOL_TIMEOUT"
	CodeInvalidParams     = "INVALID_PARAMS"
	CodeCopilotConnection = "COPILOT_CONNECTION"
	CodeAuthFailed        = "AUTH_FAILED"
	CodeMessageFailed     = "MESSAGE_FAILED"
	CodeInternal          = "INTERNAL_ERROR"
)

// AppError represents a structured application error with code and context.
type AppError struct {
	Code      string
	Message   string
	Cause     error
	RequestID string
	Extra     map[string]interface{}
}

// Sentinel errors for common error cases.
var (
	ErrConfigNotFound    = &AppError{Code: CodeConfigNotFound, Message: "configuration not found"}
	ErrToolNotFound      = &AppError{Code: CodeToolNotFound, Message: "tool not found"}
	ErrToolTimeout       = &AppError{Code: CodeToolTimeout, Message: "tool execution timed out"}
	ErrInvalidParams     = &AppError{Code: CodeInvalidParams, Message: "invalid parameters"}
	ErrCopilotConnection = &AppError{Code: CodeCopilotConnection, Message: "failed to connect to Copilot"}
	ErrAuthFailed        = &AppError{Code: CodeAuthFailed, Message: "authentication failed"}
	ErrMessageFailed     = &AppError{Code: CodeMessageFailed, Message: "failed to send message"}
)

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for use with errors.Is and errors.As.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is implements errors.Is support.
// Two AppErrors are considered equal if they have the same Code.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithCause returns a new AppError with the cause attached.
func (e *AppError) WithCause(err error) *AppError {
	return &AppError{
		Code:      e.Code,
		Message:   e.Message,
		Cause:     err,
		RequestID: e.RequestID,
		Extra:     e.Extra,
	}
}

// WithMessage returns a new AppError with a custom message.
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:      e.Code,
		Message:   msg,
		Cause:     e.Cause,
		RequestID: e.RequestID,
		Extra:     e.Extra,
	}
}

// WithRequestID returns a new AppError with the request ID attached.
func (e *AppError) WithRequestID(id string) *AppError {
	return &AppError{
		Code:      e.Code,
		Message:   e.Message,
		Cause:     e.Cause,
		RequestID: id,
		Extra:     e.Extra,
	}
}

// WithExtra returns a new AppError with additional context.
func (e *AppError) WithExtra(key string, value interface{}) *AppError {
	extra := make(map[string]interface{})
	for k, v := range e.Extra {
		extra[k] = v
	}
	extra[key] = value

	return &AppError{
		Code:      e.Code,
		Message:   e.Message,
		Cause:     e.Cause,
		RequestID: e.RequestID,
		Extra:     extra,
	}
}

// NewAppError creates a new AppError with the given code and message.
func NewAppError(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// WrapError wraps an error with an AppError.
func WrapError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// IsAppError checks if an error is an AppError with a specific code.
func IsAppError(err error, code string) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// GetAppError extracts an AppError from an error chain.
func GetAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// UserMessage returns a user-friendly message for the error.
// This filters out technical details that shouldn't be shown to end users.
func (e *AppError) UserMessage() string {
	switch e.Code {
	case CodeConfigNotFound:
		return "Configuration error. Please check your settings."
	case CodeToolNotFound:
		return "The requested tool is not available."
	case CodeToolTimeout:
		return "The operation took too long. Please try again."
	case CodeInvalidParams:
		return "Invalid input provided."
	case CodeCopilotConnection:
		return "Unable to connect to the AI service. Please try again later."
	case CodeAuthFailed:
		return "Authentication failed. Please check your credentials."
	case CodeMessageFailed:
		return "Failed to send the message. Please try again."
	default:
		return "An unexpected error occurred. Please try again."
	}
}
