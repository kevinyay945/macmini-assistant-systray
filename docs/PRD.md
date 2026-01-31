# Product Requirements Document (PRD)
# MacMini Assistant Systray Orchestrator

**Version:** 1.0
**Date:** January 31, 2026
**Status:** Draft
**Platform:** macOS (M3)
**Language:** Go

---

## 1. Executive Summary

The MacMini Assistant Systray is an intelligent orchestration chatbot that runs as a macOS system tray application. It integrates with GitHub Copilot SDK to provide AI-powered tool execution through multiple messaging platforms (LINE and Discord), enabling users to perform complex tasks like video downloads and file uploads through natural language commands.

### Key Value Propositions
- **Multi-platform Access**: Interact via LINE or Discord from any device
- **AI-Powered Orchestration**: GitHub Copilot SDK handles intent recognition and tool routing
- **Extensible Tool System**: Dynamic tool registration without recompilation
- **Always Available**: Runs silently in system tray with auto-start capability
- **Self-Updating**: Automatic updates from GitHub releases

---

## 2. Project Goals & Success Metrics

### Primary Goals
1. Enable remote execution of macOS-specific tasks through chat interfaces
2. Provide seamless AI-powered tool orchestration via GitHub Copilot SDK
3. Support extensible tool ecosystem for future capabilities
4. Maintain high reliability with comprehensive error reporting

### Success Metrics
- **Uptime**: >99% availability when running
- **Response Time**: <2s for tool invocation (excluding tool execution time)
- **LLM Timeout**: Hard limit at 10 minutes
- **Update Success Rate**: >95% successful auto-updates
- **Error Visibility**: 100% of critical errors reported to status channel

---

## 3. User Personas & Use Cases

### Primary Persona: Remote Mac User
**Profile**: User with a Mac Mini running as a home server, needs to perform tasks remotely

**Use Cases**:
1. Download YouTube video from phone via LINE message
2. Upload large files to Google Drive and get shareable link
3. Monitor system status and tool execution via Discord
4. Extend functionality by adding new tools without server access

### Secondary Persona: Developer
**Profile**: Developer extending the tool ecosystem

**Use Cases**:
1. Create new tools using CLI
2. Test tools locally with build tags
3. Deploy tool updates via config changes

---

## 4. Technical Architecture

### 4.1 High-Level Architecture
```
┌─────────────────────────────────────────────┐
│         macOS System Tray App               │
│  ┌──────────────────────────────────────┐   │
│  │   Core Orchestrator                  │   │
│  │   - GitHub Copilot SDK Integration   │   │
│  │   - Tool Registry (Dynamic Config)   │   │
│  │   - LLM Request Handler              │   │
│  └──────────────────────────────────────┘   │
│           │              │                   │
│  ┌────────┴────┐    ┌────┴────────────┐     │
│  │ Message     │    │ Tool Plugins    │     │
│  │ Handlers    │    │ (Config-based)  │     │
│  │ - LINE      │    │ - Downie        │     │
│  │ - Discord   │    │ - Google Drive  │     │
│  └─────────────┘    └─────────────────┘     │
│           │                                  │
│  ┌────────┴───────────────────────────┐     │
│  │ Infrastructure                     │     │
│  │ - Config Manager (YAML)            │     │
│  │ - Auto-updater (GitHub Releases)   │     │
│  │ - Logging & Error Reporting        │     │
│  └────────────────────────────────────┘     │
└─────────────────────────────────────────────┘
         ↓                      ↓
    [LINE Platform]      [Discord Platform]
```

### 4.2 Component Breakdown

#### Core Components
1. **Orchestrator** (`cmd/orchestrator/`)
   - Main application entry point
   - System tray integration
   - Lifecycle management (start, stop, restart)

2. **Copilot SDK Integration** (`internal/copilot/`)
   - Tool registration with Copilot SDK
   - Request/response handling
   - Context management
   - Timeout enforcement (10 min)

3. **Tool Registry** (`internal/registry/`)
   - Dynamic tool loading from YAML config
   - Tool metadata management
   - Validation and error handling

4. **Message Handlers** (`internal/handlers/`)
   - LINE bot webhook handler
   - Discord bot event handler
   - Status panel publisher (Discord only)
   - Unified message interface

#### Tool Plugins
1. **Downie Integration** (`internal/tools/downie/`)
   - Deep link URL construction
   - Format and resolution validation
   - File path resolution

2. **Google Drive Uploader** (`internal/tools/gdrive/`)
   - OAuth2 authentication
   - Upload with progress tracking
   - Share link generation

#### Infrastructure
1. **Configuration** (`internal/config/`)
   - YAML-based config file
   - Default values and validation
   - Runtime reload support

2. **Auto-updater** (`internal/updater/`)
   - GitHub release polling
   - Binary replacement with go-update
   - Rollback on failure

3. **Observability** (`internal/observability/`)
   - Structured logging
   - Error capture and reporting
   - Metrics collection (optional)

### 4.3 Technology Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| System Tray | github.com/getlantern/systray | macOS menu bar integration |
| HTTP Server | github.com/gin-gonic/gin | Webhook endpoints |
| CLI | github.com/spf13/cobra | Command-line interface |
| LINE Bot | github.com/line/line-bot-sdk-go/v8 | LINE messaging API |
| Discord Bot | github.com/bwmarrin/discordgo | Discord API integration |
| Auto-update | github.com/inconshreveable/go-update | Binary self-updating |
| Copilot SDK | GitHub Copilot SDK (Go) | AI orchestration |

---

## 5. Functional Requirements

### 5.1 Core Features

#### FR-1: System Tray Application
- **FR-1.1**: Application runs as macOS system tray icon
- **FR-1.2**: Menu provides: Start/Stop, Settings, Check Updates, Quit
- **FR-1.3**: Icon shows status: Active (green), Error (red), Idle (gray)
- **FR-1.4**: Launches on system startup (configurable)

#### FR-2: Configuration Management
- **FR-2.1**: YAML config file at `~/.macmini-assistant/config.yaml`
- **FR-2.2**: Configurable settings:
  - Download folder path (default: `/tmp/downloads`)
  - LINE bot credentials (channel secret, access token)
  - Discord bot token and status channel ID
  - Tool definitions and parameters
  - GitHub Copilot API key
  - Auto-update enabled/disabled
- **FR-2.3**: Config validation on startup with clear error messages
- **FR-2.4**: CLI command to generate default config: `orchestrator init`

#### FR-3: GitHub Copilot SDK Integration
- **FR-3.1**: Register all tools with Copilot SDK on startup
- **FR-3.2**: Stream tool requests from Copilot
- **FR-3.3**: Execute tool with timeout enforcement (10 min hard limit)
- **FR-3.4**: Return structured responses (success/failure with data)
- **FR-3.5**: Handle concurrent tool executions safely

#### FR-4: LINE Bot Interface
- **FR-4.1**: Webhook endpoint: `/webhook/line`
- **FR-4.2**: Reply mode: Bot responds only when mentioned
- **FR-4.3**: Message flow:
  1. Receive user message
  2. Send to Copilot SDK for intent analysis
  3. Execute tool if requested
  4. Reply with result or direct LLM response
- **FR-4.4**: Support text messages and file attachments
- **FR-4.5**: Error messages in user-friendly format

#### FR-5: Discord Bot Interface
- **FR-5.1**: Bot responds to mentions and DMs
- **FR-5.2**: Reply mode identical to LINE
- **FR-5.3**: Status panel channel:
  - Post tool execution start events
  - Post tool completion with results
  - Post errors with stack traces
  - Format: Rich embeds with color coding
- **FR-5.4**: Commands:
  - `/status` - Show bot health and uptime
  - `/tools` - List available tools
  - `/help` - Usage instructions

#### FR-6: Auto-update System
- **FR-6.1**: Check GitHub releases every 6 hours
- **FR-6.2**: Compare current version with latest release
- **FR-6.3**: Download and verify binary (checksum validation)
- **FR-6.4**: Replace binary and restart gracefully
- **FR-6.5**: Rollback if new version fails to start
- **FR-6.6**: Manual update trigger via tray menu

### 5.2 Tool Specifications

#### Tool 1: YouTube Download (via Downie)
```yaml
name: youtube_download
description: Download YouTube video using Downie
inputs:
  - name: youtube_url
    type: string
    required: true
    description: YouTube video URL
  - name: file_format
    type: string
    required: false
    default: mp4
    allowed: [mp4, mkv, webm]
  - name: resolution
    type: string
    required: false
    default: 1080p
    allowed: [4320p, 2160p, 1440p, 1080p, 720p, 480p, 360p]
outputs:
  - name: file_path
    type: string
    description: Absolute path to downloaded file
  - name: file_size
    type: integer
    description: File size in bytes
```

**Implementation**: Use Downie deep link: `downie://[url]?format=[format]&quality=[resolution]`

#### Tool 2: Google Drive Upload
```yaml
name: gdrive_upload
description: Upload file to Google Drive and generate share link
inputs:
  - name: file_path
    type: string
    required: true
    description: Absolute path to file
  - name: upload_name
    type: string
    required: false
    description: Name for uploaded file (defaults to original filename)
  - name: timeout
    type: integer
    required: false
    default: 300
    description: Upload timeout in seconds
outputs:
  - name: share_link
    type: string
    description: Public share link to uploaded file
  - name: file_id
    type: string
    description: Google Drive file ID
```

**Implementation**: Google Drive API v3 with OAuth2

---

## 6. Non-Functional Requirements

### NFR-1: Performance
- Tool invocation overhead: <2 seconds
- LLM request hard timeout: 10 minutes
- Memory footprint: <100MB idle, <500MB under load
- Startup time: <5 seconds

### NFR-2: Reliability
- Graceful degradation when services unavailable
- Automatic reconnection for Discord/LINE websockets
- Transaction logging for all tool executions
- No data loss on unexpected shutdown

### NFR-3: Security
- Credentials stored securely (keychain integration recommended)
- HTTPS-only for webhook endpoints
- Signature validation for LINE/Discord webhooks
- No credential logging

### NFR-4: Maintainability
- Comprehensive unit tests (>80% coverage)
- Integration tests with build tags for local-only tests
- Clear error messages with troubleshooting hints
- Structured logging (JSON format)

### NFR-5: Extensibility
- Tool addition without code changes (config-only)
- Plugin interface for future tool types
- Versioned config schema with migration support

---

## 7. Development Approach

### 7.1 Test-Driven Development (TDD)
- Write tests before implementation
- Use table-driven tests for comprehensive coverage
- Build tags for environment-specific tests:
  ```go
  //go:build integration
  // For tests requiring external services

  //go:build local
  // For tests requiring local tools (e.g., Downie)
  ```

### 7.2 Build Tags Strategy
```go
// Skip Downie tests in CI
//go:build local

// Skip integration tests by default
//go:build integration
```

**Test Execution**:
- All tests: `go test ./...`
- CI tests: `go test ./... --tags=!local,!integration`
- Local with integrations: `go test ./... --tags=local,integration`

---

## 8. User Stories

### Epic 1: Initial Setup
- **US-1.1**: As a user, I can run `orchestrator init` to generate a config file with prompts for credentials
- **US-1.2**: As a user, I can configure the download folder via config file
- **US-1.3**: As a user, I can enable auto-start on macOS login

### Epic 2: Video Downloads
- **US-2.1**: As a user, I can send a YouTube URL via LINE and receive the downloaded video file path
- **US-2.2**: As a user, I can specify video format and resolution in my message
- **US-2.3**: As a user, I receive progress updates for long downloads

### Epic 3: File Uploads
- **US-3.1**: As a user, I can upload a local file to Google Drive and get a share link
- **US-3.2**: As a user, I can customize the uploaded file name
- **US-3.3**: As a user, I receive an error if the upload times out

### Epic 4: Status Monitoring
- **US-4.1**: As a user, I can see all tool executions in a Discord status channel
- **US-4.2**: As a user, I receive detailed error reports when tools fail
- **US-4.3**: As a user, I can check bot health with `/status` command

### Epic 5: Maintenance
- **US-5.1**: As a user, I am notified when updates are available
- **US-5.2**: As a user, I can manually trigger updates from the system tray
- **US-5.3**: As a user, the app auto-recovers if an update fails

### Epic 6: Extensibility
- **US-6.1**: As a developer, I can add new tools via CLI command
- **US-6.2**: As a developer, I can test tools locally before deployment
- **US-6.3**: As a developer, tools are automatically registered with Copilot SDK on restart

---

## 9. Future Enhancements (Out of Scope for v1.0)

1. **Web Dashboard**: Browser-based configuration and monitoring
2. **Multi-user Support**: Multiple users with separate configs
3. **Scheduled Tasks**: Cron-like scheduled tool execution
4. **Tool Marketplace**: Public repository of community tools
5. **Mobile App**: Native iOS/Android system tray equivalents
6. **Analytics**: Usage patterns and optimization insights
7. **Voice Interface**: Integration with Siri/HomeKit
8. **Slack Integration**: Additional messaging platform

---

## 10. Open Questions & Decisions Needed

### Questions
1. **OAuth Flow**: How should users authenticate Google Drive? (Browser flow vs service account)
2. **Rate Limiting**: Should we implement rate limits per user/tool?
3. **File Cleanup**: Auto-delete downloaded files after N days?
4. **Notification System**: macOS notifications for tool completions?

### Decisions Needed
- [ ] Choose config format: YAML vs TOML vs JSON
- [ ] Define versioning scheme (SemVer recommended)
- [ ] Select logging library (zerolog vs zap vs slog)
- [ ] Decide on metrics collection (Prometheus vs statsd vs none)

---

## 11. Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Downie deep link API changes | High | Medium | Version pinning, fallback to direct ffmpeg |
| GitHub Copilot SDK rate limits | High | Low | Request queuing, user feedback |
| Discord/LINE platform changes | Medium | Medium | API client libraries, version monitoring |
| macOS permission issues | High | Medium | Clear documentation, permission checks on startup |
| Auto-update corruption | Critical | Low | Checksum validation, atomic replacement, rollback |

---

## 12. Timeline & Milestones

See [Development Plan](./DEVELOPMENT_PLAN.md) for detailed phases and timeline.

---

## Appendix A: Glossary

- **Orchestrator**: The core application managing tool execution
- **Tool**: A discrete capability exposed to the LLM (e.g., download, upload)
- **Status Panel**: Discord channel receiving execution logs
- **Reply Mode**: Bot only responds when explicitly messaged
- **Deep Link**: macOS URL scheme to trigger external applications

---

## Appendix B: References

- [GitHub Copilot SDK Documentation](https://github.com/copilot)
- [Downie Deep Link Documentation](https://software.charliemonroe.net/downie/)
- [LINE Messaging API](https://developers.line.biz/en/docs/messaging-api/)
- [Discord Developer Portal](https://discord.com/developers/docs)
- [go-update Library](https://github.com/inconshreveable/go-update)
