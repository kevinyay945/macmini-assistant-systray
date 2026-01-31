# Copilot Instructions: MacMini Assistant Systray

## Project Overview

A Go-based macOS system tray application that orchestrates AI-powered tool execution via GitHub Copilot SDK, accessible through LINE and Discord messaging platforms. The app enables remote task automation (video downloads via Downie, Google Drive uploads) with self-updating capability.

**Status**: Phase 0 Bootstrap - Project infrastructure setup in progress.

## Architecture

```
Core Orchestrator (cmd/orchestrator/)
├── Copilot SDK Integration (internal/copilot/)
├── Tool Registry - YAML config-based (internal/registry/)
├── Message Handlers (internal/handlers/)
│   ├── LINE Bot - webhook at /webhook/line
│   └── Discord Bot - mentions, DMs, slash commands
├── Tool Plugins (internal/tools/)
│   ├── Downie - deep link URL scheme
│   └── Google Drive - OAuth2/service account
├── System Tray (internal/systray/)
└── Auto-updater (internal/updater/)
```

Config location: `~/.macmini-assistant/config.yaml`

## Tech Stack & Key Libraries

| Component | Library |
|-----------|---------|
| System Tray | `github.com/getlantern/systray` |
| HTTP Server | `github.com/gin-gonic/gin` |
| CLI | `github.com/spf13/cobra` |
| LINE Bot | `github.com/line/line-bot-sdk-go/v8/linebot` |
| Discord Bot | `github.com/bwmarrin/discordgo` |
| Auto-update | `github.com/inconshreveable/go-update` |

Use the skills in `.github/skills/` for library-specific guidance (cobra, gin, discordgo, linebot-go, go-update, copilot-sdk-go, google-drive-go).

## Development Conventions

### Testing Strategy (TDD)

Use build tags to separate test environments:

```go
// Standard tests - run in CI
func TestConfig_Load(t *testing.T) { ... }

// Local-only tests - require macOS tools like Downie
//go:build local

// Integration tests - require external services
//go:build integration
```

**Commands**:
- `make test` - CI-safe tests only
- `make test-local` - includes local-only tests
- `make test-integration` - includes integration tests
- `make test-all` - all tests

### Code Organization

- Entry point: `cmd/orchestrator/main.go`
- Internal packages: `internal/<package>/`
- Each package has `*_test.go` files alongside implementation
- Config-based tool registration (no recompilation for new tools)

### Linting

Use golangci-lint with `.golangci.yml` config. Key settings:
- Local import prefix: `github.com/kevinyay945/macmini-assistant-systray`
- Max function length: 120 lines (relaxed in tests)
- Package comments disabled

### Error Handling

- Return user-friendly errors to messaging platforms
- Log detailed errors with stack traces to Discord status channel
- Never log credentials, tokens, or full URLs with query params

## Key Patterns

### Tool Definition (YAML)

```yaml
tools:
  - name: youtube_download
    type: downie
    config:
      deep_link_scheme: downie://
      default_format: mp4
      default_resolution: 1080p
```

### Message Handler Interface

Both LINE and Discord handlers should implement unified message processing:
1. Receive user message → 2. Send to Copilot SDK → 3. Execute tool or return LLM response → 4. Reply to user

### Timeout Enforcement

LLM requests have a hard 10-minute timeout. Tool executions should respect this limit.

## Quick Reference

| Action | Command |
|--------|---------|
| Build | `make build` |
| Run | `make run` or `go run cmd/orchestrator/main.go` |
| Lint | `make lint` |
| Test | `make test` |
| Coverage | `make test-coverage` |
| Init dev env | `make init` |

## Documentation References

- [PRD](docs/PRD.md) - Feature requirements and specs
- [Development Plan](docs/DEVELOPMENT_PLAN.md) - Phase breakdown
- [Phase Documents](docs/phases/) - Detailed implementation tasks
- [Discussion Summary](docs/DISCUSSION_SUMMARY.md) - Design decisions
