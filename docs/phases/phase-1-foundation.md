# Phase 1: Core Foundation

**Duration**: Weeks 2-3
**Status**: 🟢 Completed
**Goal**: Build the foundational components without external integrations

---

## Overview

This phase establishes the core infrastructure components: configuration system, tool registry, and logging/error handling. These are foundational pieces that all other phases will depend on.

---

## 1.1 Configuration System

**Duration**: 2 days
**Status**: 🟢 Completed

### Tasks

- [x] Define config schema (YAML)
- [x] Implement config loader with validation
- [x] Add default config generation
- [x] Support environment variable overrides (optional)

### Implementation Details

**Config Location**: `~/.macmini-assistant/config.yaml`

**Config Schema**:
```yaml
# ~/.macmini-assistant/config.yaml
app:
  download_folder: /tmp/downloads
  auto_start: true
  auto_update: true
  log_level: info  # debug, info, warn, error

copilot:
  api_key: ${GITHUB_COPILOT_API_KEY}
  timeout_seconds: 600  # 10 minutes

line:
  channel_secret: ${LINE_CHANNEL_SECRET}
  access_token: ${LINE_ACCESS_TOKEN}
  webhook_port: 8080

discord:
  bot_token: ${DISCORD_BOT_TOKEN}
  status_channel_id: "1234567890"
  enable_slash_commands: true

tools:
  - name: youtube_download
    type: downie
    enabled: true
    config:
      deep_link_scheme: "downie://"
      default_format: mp4
      default_resolution: 1080p

  - name: gdrive_upload
    type: google_drive
    enabled: true
    config:
      credentials_path: ~/.macmini-assistant/gdrive-creds.json
      default_timeout: 300

updater:
  github_repo: username/macmini-assistant
  check_interval_hours: 6
  enabled: true
```

### Test Cases

```go
// internal/config/config_test.go
func TestLoadConfig_ValidFile(t *testing.T)
func TestLoadConfig_InvalidYAML(t *testing.T)
func TestLoadConfig_MissingRequired(t *testing.T)
func TestLoadConfig_DefaultValues(t *testing.T)
func TestValidateConfig_DownloadFolder(t *testing.T)
func TestGenerateDefaultConfig(t *testing.T)
func TestConfig_EnvironmentVariableExpansion(t *testing.T)
```

### Acceptance Criteria

- [x] Config loads from `~/.macmini-assistant/config.yaml`
- [x] Validation fails early with clear error messages
- [x] `orchestrator init` generates valid default config
- [x] All config fields have defaults or required validation
- [x] Environment variables are expanded (e.g., `${VAR_NAME}`)

### Notes

Implemented in `internal/config/config.go`. All tests passing.

---

## 1.2 Tool Registry

**Duration**: 3 days
**Status**: 🟢 Completed

### Tasks

- [x] Define tool interface
- [x] Implement registry with dynamic loading
- [x] Add tool metadata validation
- [x] Create tool execution wrapper (timeout, logging)

### Implementation Details

**Tool Interface**:
```go
// internal/registry/tool.go
package registry

import "context"

type Tool interface {
    // Name returns the unique identifier for this tool
    Name() string

    // Description returns a human-readable description
    Description() string

    // Execute runs the tool with the given parameters
    Execute(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error)

    // Schema returns the input/output schema for this tool
    Schema() ToolSchema
}

type ToolSchema struct {
    Inputs  []Parameter `json:"inputs"`
    Outputs []Parameter `json:"outputs"`
}

type Parameter struct {
    Name        string      `json:"name"`
    Type        string      `json:"type"` // string, integer, boolean, array
    Required    bool        `json:"required"`
    Default     interface{} `json:"default,omitempty"`
    Description string      `json:"description"`
    Allowed     []string    `json:"allowed,omitempty"` // for enum types
}
```

**Registry Interface**:
```go
// internal/registry/registry.go
package registry

import (
    "context"
    "time"
)

type Registry struct {
    tools   map[string]Tool
    timeout time.Duration
}

func NewRegistry(timeout time.Duration) *Registry

func (r *Registry) Register(tool Tool) error
func (r *Registry) Get(name string) (Tool, bool)
func (r *Registry) List() []Tool
func (r *Registry) Execute(ctx context.Context, name string, params map[string]interface{}) (map[string]interface{}, error)
func (r *Registry) LoadFromConfig(cfg []ToolConfig) error
```

### Test Cases

```go
// internal/registry/registry_test.go
func TestRegistry_RegisterTool(t *testing.T)
func TestRegistry_GetTool(t *testing.T)
func TestRegistry_ExecuteWithTimeout(t *testing.T)
func TestRegistry_ExecuteWithInvalidParams(t *testing.T)
func TestRegistry_LoadFromConfig(t *testing.T)
func TestToolSchema_Validation(t *testing.T)
func TestRegistry_DuplicateRegistration(t *testing.T)
func TestRegistry_ExecuteNonExistent(t *testing.T)
```

### Acceptance Criteria

- [x] Tools registered via config file
- [x] Tool execution respects 10-min timeout
- [x] Invalid parameters rejected with clear errors
- [x] Tool metadata accessible for Copilot SDK registration
- [x] Duplicate tool names are rejected
- [x] Registry is thread-safe for concurrent access

### Notes

Implemented in `internal/registry/registry.go`. Enhanced with:
- Parameter type validation (`string`, `integer`, `boolean`, `array`, `object`)
- Default value application for optional parameters
- Goroutine-safe execution with timeout handling (fixed potential goroutine leak)
- Factory pattern for tool creation (`RegisterFactory`, `MustRegisterFactory`)
- Duplicate detection for both tools (`ErrDuplicateTool`) and factories (`ErrDuplicateFactory`)
- Thread-safe `Timeout()` getter with read lock

---

## 1.3 Logging & Error Handling

**Duration**: 2 days
**Status**: 🟢 Completed

### Tasks

- [x] Set up structured logging (recommend: `log/slog`)
- [x] Define error types and wrapping
- [x] Create error reporter interface
- [x] Add request ID tracing

### Implementation Details

**Logger Setup**:
```go
// internal/observability/logger.go
package observability

import (
    "context"
    "log/slog"
    "os"
)

type Logger struct {
    *slog.Logger
}

func NewLogger(level string) *Logger {
    var l slog.Level
    switch level {
    case "debug":
        l = slog.LevelDebug
    case "info":
        l = slog.LevelInfo
    case "warn":
        l = slog.LevelWarn
    case "error":
        l = slog.LevelError
    default:
        l = slog.LevelInfo
    }

    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: l,
    })

    return &Logger{slog.New(handler)}
}

func (l *Logger) WithRequestID(ctx context.Context, requestID string) *Logger
func (l *Logger) WithTool(toolName string) *Logger
func (l *Logger) WithPlatform(platform string) *Logger
```

**Error Types**:
```go
// internal/observability/errors.go
package observability

type AppError struct {
    Code      string
    Message   string
    Cause     error
    RequestID string
}

var (
    ErrConfigNotFound    = &AppError{Code: "CONFIG_NOT_FOUND"}
    ErrToolNotFound      = &AppError{Code: "TOOL_NOT_FOUND"}
    ErrToolTimeout       = &AppError{Code: "TOOL_TIMEOUT"}
    ErrInvalidParams     = &AppError{Code: "INVALID_PARAMS"}
    ErrCopilotConnection = &AppError{Code: "COPILOT_CONNECTION"}
)

func (e *AppError) Error() string
func (e *AppError) Unwrap() error
func (e *AppError) WithCause(err error) *AppError
func (e *AppError) WithRequestID(id string) *AppError
```

**Error Reporter**:
```go
// internal/observability/reporter.go
package observability

type ErrorReporter interface {
    Report(ctx context.Context, err error)
    ReportWithContext(ctx context.Context, err error, extra map[string]interface{})
}

type StatusPanelReporter struct {
    // Discord status panel client
}
```

### Test Cases

```go
// internal/observability/logger_test.go
func TestLogger_StructuredOutput(t *testing.T)
func TestLogger_LogLevels(t *testing.T)
func TestLogger_WithRequestID(t *testing.T)
func TestLogger_WithTool(t *testing.T)
func TestLogger_SensitiveDataFiltering(t *testing.T)

// internal/observability/errors_test.go
func TestAppError_Error(t *testing.T)
func TestAppError_Unwrap(t *testing.T)
func TestAppError_WithCause(t *testing.T)

// internal/observability/reporter_test.go
func TestErrorReporter_CaptureError(t *testing.T)
func TestRequestID_Propagation(t *testing.T)
```

### Acceptance Criteria

- [x] All logs in JSON format with timestamps
- [x] Errors include context (stack trace, request ID)
- [x] Log levels configurable (debug, info, warn, error)
- [x] No sensitive data in logs (API keys, tokens filtered)
- [x] Request IDs propagate through entire request lifecycle

### Notes

Implemented in `internal/observability/`. Enhanced with:
- `AppError` with fluent API: `WithCause()`, `WithMessage()`, `WithRequestID()`, `WithExtra()`
- `UserMessage()` method for user-friendly error messages
- `LogReporter` with proper request ID priority (AppError > context)
- `MultiReporter` with parallel execution using `sync.WaitGroup`
- Sentinel errors for common cases (`ErrConfigNotFound`, `ErrToolNotFound`, etc.)

---

## Deliverables

By the end of Phase 1:

- [x] Configuration system fully functional
- [x] Tool registry with timeout enforcement
- [x] Structured logging throughout codebase
- [x] Error types and reporter interface

---

## Testing Strategy

```bash
# Run Phase 1 tests
go test ./internal/config/... ./internal/registry/... ./internal/observability/... -v

# With coverage
go test ./internal/config/... ./internal/registry/... ./internal/observability/... -coverprofile=phase1.out
```

---

## Dependencies

```go
// go.mod additions
require (
    gopkg.in/yaml.v3 v3.0.1
)
```

---

## Notes & Discoveries

### Code Review Improvements (2026-01-31)

**Performance & Memory:**
1. **Logger Regex Optimization**: Moved sensitive data filtering regex patterns to package-level variables to avoid recompilation on every logger initialization
2. **Goroutine Leak Fix**: `Registry.Execute()` now uses `<-timeoutCtx.Done()` instead of `default` case to properly handle context cancellation and prevent goroutine leaks
3. **Immutable Params**: `Execute()` copies params map to avoid mutating caller's data

**Type System & Validation:**
4. **Number Type Support**: Added `number` type validation for float32/float64 in addition to existing `integer` type
5. **Allowed Values Validation**: Implemented validation for `Parameter.Allowed` field to support enum-style constraints
6. **Parameter Type Validation**: Enhanced `validateParamType()` to validate string enum values against allowed list

**Error Handling:**
7. **AppError.Is Fix**: Simplified `AppError.Is()` implementation to use direct type assertion instead of `errors.As`, providing more predictable behavior
8. **Request ID Priority**: `LogReporter` now correctly prioritizes AppError's RequestID over context RequestID

**Configuration & Flexibility:**
9. **MultiReporter Configurable Timeout**: Added `WithTimeout()` method to allow custom timeout configuration (default: 5s)
10. **Environment Variable Documentation**: Added documentation note that nested variable substitution (e.g., `${VAR1:-${VAR2}}`) is not supported

**Testing:**
11. **Test Timing Constants**: Replaced magic numbers in timeout tests with named constants (`testShortTimeout`, `testLongOperation`) for reliability across different environments
12. **Enhanced Test Coverage**: Added tests for number type validation, allowed values validation, and MultiReporter timeout configuration

---

## Time Tracking

| Task | Estimated | Actual | Notes |
|------|-----------|--------|-------|
| 1.1 Config System | 2 days | | |
| 1.2 Tool Registry | 3 days | | |
| 1.3 Logging | 2 days | | |
| **Total** | **7 days** | | |

---

**Previous**: [Phase 0: Project Bootstrap](./phase-0-bootstrap.md)
**Next**: [Phase 2: Messaging Platform Integration](./phase-2-messaging.md)
