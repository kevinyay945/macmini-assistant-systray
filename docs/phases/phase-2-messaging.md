# Phase 2: Messaging Platform Integration

**Duration**: Weeks 4-5
**Status**: ⚪ Not Started
**Goal**: Implement LINE and Discord bot handlers

---

## Overview

This phase implements the messaging platform integrations. Both LINE and Discord bots will handle incoming messages and route them to the orchestrator.

---

## 2.1 LINE Bot Handler

**Duration**: 3 days
**Status**: ⚪ Not Started

### Tasks

- [ ] Webhook endpoint with Gin
- [ ] Signature validation
- [ ] Message parsing and routing
- [ ] Reply message formatting
- [ ] Error handling and retry logic

### Implementation Details

**Webhook Endpoint**: `/webhook/line`

**Handler Structure**:
```go
// internal/handlers/line.go
package handlers

import (
    "github.com/gin-gonic/gin"
    "github.com/line/line-bot-sdk-go/v8/linebot"
)

type LineHandler struct {
    bot    *linebot.Client
    router MessageRouter
    logger *observability.Logger
}

func NewLineHandler(channelSecret, accessToken string, router MessageRouter) (*LineHandler, error)
func (h *LineHandler) HandleWebhook(c *gin.Context)
func (h *LineHandler) parseMessage(event *linebot.Event) (*Message, error)
func (h *LineHandler) sendReply(replyToken string, message string) error
```

### Test Cases

```go
// internal/handlers/line_test.go
func TestLineHandler_WebhookValidation(t *testing.T)
func TestLineHandler_TextMessage(t *testing.T)
func TestLineHandler_InvalidSignature(t *testing.T)
func TestLineHandler_ReplyFormatting(t *testing.T)
func TestLineHandler_ErrorResponse(t *testing.T)
func TestLineHandler_EmptyMessage(t *testing.T)
```

**Build Tag**: Standard (no special tags needed)

### Acceptance Criteria

- [ ] Webhook validates LINE signatures
- [ ] Bot responds only when messaged
- [ ] Messages parsed and forwarded to orchestrator
- [ ] Errors returned in user-friendly format
- [ ] Large messages handled correctly

### Notes

<!-- Add your notes here -->
please show me how to setup the cert

---

## 2.2 Discord Bot Handler

**Duration**: 4 days
**Status**: ⚪ Not Started

### Tasks

- [ ] Bot connection with intents
- [ ] Message event handling (mentions, DMs)
- [ ] Slash commands (`/status`, `/tools`, `/help`)
- [ ] Status panel integration (separate channel)
- [ ] Rich embed formatting

### Implementation Details

**Handler Structure**:
```go
// internal/handlers/discord.go
package handlers

import (
    "github.com/bwmarrin/discordgo"
)

type DiscordHandler struct {
    session        *discordgo.Session
    statusChannelID string
    router         MessageRouter
    logger         *observability.Logger
}

func NewDiscordHandler(token string, statusChannelID string, router MessageRouter) (*DiscordHandler, error)
func (h *DiscordHandler) Start() error
func (h *DiscordHandler) Stop() error
func (h *DiscordHandler) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate)
func (h *DiscordHandler) handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate)
func (h *DiscordHandler) PostStatus(msg StatusMessage) error
```

**Slash Commands**:
```go
var commands = []*discordgo.ApplicationCommand{
    {
        Name:        "status",
        Description: "Show bot health and uptime",
    },
    {
        Name:        "tools",
        Description: "List available tools",
    },
    {
        Name:        "help",
        Description: "Show usage instructions",
    },
}
```

**Status Panel Embeds**:
```go
// Execution started
&discordgo.MessageEmbed{
    Title: "🎬 YouTube Download Started",
    Color: 0x3498db, // Blue
    Fields: []*discordgo.MessageEmbedField{
        {Name: "Tool", Value: "youtube_download", Inline: true},
        {Name: "User", Value: "@username", Inline: true},
    },
    Timestamp: time.Now().Format(time.RFC3339),
}

// Execution complete
&discordgo.MessageEmbed{
    Title: "✅ YouTube Download Complete",
    Color: 0x2ecc71, // Green
    Fields: []*discordgo.MessageEmbedField{
        {Name: "Duration", Value: "32.5s", Inline: true},
        {Name: "File Size", Value: "125 MB", Inline: true},
    },
}

// Execution error
&discordgo.MessageEmbed{
    Title: "❌ Tool Execution Failed",
    Color: 0xe74c3c, // Red
    Description: "Error message here",
    Fields: []*discordgo.MessageEmbedField{
        {Name: "Stack Trace", Value: "```\n...\n```"},
    },
}
```

### Test Cases

```go
// internal/handlers/discord_test.go
//go:build integration

func TestDiscordHandler_MessageCreate(t *testing.T)
func TestDiscordHandler_SlashCommand_Status(t *testing.T)
func TestDiscordHandler_SlashCommand_Tools(t *testing.T)
func TestDiscordHandler_StatusPanel_ToolExecution(t *testing.T)
func TestDiscordHandler_StatusPanel_ErrorReport(t *testing.T)
func TestDiscordHandler_RichEmbedFormatting(t *testing.T)
func TestDiscordHandler_MentionOnly(t *testing.T)
func TestDiscordHandler_DirectMessage(t *testing.T)
```

**Build Tag**: `integration` (requires Discord test bot)

### Acceptance Criteria

- [ ] Bot responds to mentions and DMs
- [ ] Slash commands functional
- [ ] Status panel posts tool execution events
- [ ] Errors posted to status channel with details
- [ ] Color-coded rich embeds

### Notes

<!-- Add your notes here -->
please show me how to setup the cert

---

## 2.3 Unified Message Interface

**Duration**: 2 days
**Status**: ✅ Complete

### Tasks

- [x] Abstract message interface for platforms
- [x] Common message routing logic
- [x] Platform-agnostic error formatting
- [x] Exported platform constants for type safety

### Implementation Details

```go
// internal/handlers/interface.go
package handlers

import (
    "context"
    "errors"
    "time"
)

// Platform constants for message sources
const (
    PlatformDiscord = "discord"
    PlatformLINE    = "line"
)

// Message represents a platform-agnostic message
type Message struct {
    ID        string
    UserID    string
    Platform  string // Use PlatformDiscord or PlatformLINE constants
    Content   string
    Timestamp time.Time
    ReplyFunc func(response string) error
    Metadata  map[string]interface{}
}

// MessageRouter routes messages to the orchestrator
type MessageRouter interface {
    Route(ctx context.Context, msg *Message) (*Response, error)
}

// Response represents the response to send back
type Response struct {
    Text   string
    Data   map[string]interface{}
    Error  error
}

// StatusMessage for status panel
type StatusMessage struct {
    Type      string // "start", "progress", "complete", "error"
    ToolName  string
    UserID    string
    Platform  string
    Duration  time.Duration
    Result    map[string]interface{}
    Error     error
    Message   string
}

// StatusReporter defines the interface for posting status updates
type StatusReporter interface {
    PostStatus(ctx context.Context, msg StatusMessage) error
}

// FormatUserFriendlyError formats an error into a user-friendly message
// with emoji support (canonical error formatter for all platforms)
func FormatUserFriendlyError(err error) string {
    if err == nil {
        return ""
    }
    if errors.Is(err, context.DeadlineExceeded) {
        return "⏱️ Request timed out. Please try again."
    }
    if errors.Is(err, context.Canceled) {
        return "🚫 Request was cancelled."
    }
    return "❌ An error occurred while processing your request. Please try again later."
}
```

### Test Cases

```go
// internal/handlers/interface_test.go
func TestPlatformConstants(t *testing.T)
func TestNewMessage(t *testing.T)
func TestNewMessage_FromLINE(t *testing.T)
func TestNewMessage_FromDiscord(t *testing.T)
func TestNewResponse(t *testing.T)
func TestNewErrorResponse(t *testing.T)
func TestNewStatusMessage(t *testing.T)
func TestStatusMessage_WithDuration(t *testing.T)
func TestStatusMessage_WithError(t *testing.T)
func TestDefaultErrorFormatter_FormatError(t *testing.T)
func TestMessage_MetadataUsage(t *testing.T)
func TestResponse_DataUsage(t *testing.T)
func TestStatusMessage_AllTypes(t *testing.T)
func TestFormatUserFriendlyError(t *testing.T)
```

### Acceptance Criteria

- [x] Single interface for LINE and Discord
- [x] Platform-specific formatting abstracted
- [x] Exported platform constants for type safety
- [x] Shared error formatter with emoji support
- [x] Easy to add new platforms (Slack, etc.)

### Notes

<!-- Add your notes here -->

---

## Deliverables

By the end of Phase 2:

- [ ] LINE bot receiving and responding to messages
- [ ] Discord bot with mentions, DMs, and slash commands
- [ ] Status panel posting execution events
- [ ] Unified message interface

---

## Dependencies

```go
// go.mod additions
require (
    github.com/gin-gonic/gin v1.9.1
    github.com/line/line-bot-sdk-go/v8 v8.x.x
    github.com/bwmarrin/discordgo v0.27.x
)
```

---

## Time Tracking

| Task | Estimated | Actual | Notes |
|------|-----------|--------|-------|
| 2.1 LINE Handler | 3 days | | |
| 2.2 Discord Handler | 4 days | | |
| 2.3 Unified Interface | 2 days | | |
| **Total** | **9 days** | | |

---

**Previous**: [Phase 1: Core Foundation](./phase-1-foundation.md)
**Next**: [Phase 3: GitHub Copilot SDK Integration](./phase-3-copilot.md)
