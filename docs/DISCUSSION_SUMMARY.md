# Project Discussion Summary
# MacMini Assistant Systray Orchestrator

**Date:** January 31, 2026
**Participants:** Development Team
**Status:** Ready for Implementation

---

## Overview

This document summarizes the key decisions, considerations, and next steps for the MacMini Assistant Systray Orchestrator project.

## Key Design Decisions

### 1. Architecture: Core + Plugin System ✅
**Decision**: Implement a core orchestrator with config-based dynamic tool registration.

**Rationale**:
- Flexibility to add tools without recompilation
- Clear separation of concerns
- Easier testing and maintenance
- Future-proof for tool marketplace

**Implementation**:
```yaml
# config.yaml example
tools:
  - name: youtube_download
    type: downie
    config:
      deep_link_scheme: downie://
      default_format: mp4
      default_resolution: 1080p

  - name: gdrive_upload
    type: google_drive
    config:
      credentials_path: ~/.macmini-assistant/gdrive-creds.json
      default_timeout: 300
```

---

### 2. Configuration: YAML-based ✅
**Decision**: Use YAML configuration files with validation.

**Rationale**:
- Human-readable and editable
- Good Go library support (gopkg.in/yaml.v3)
- Supports complex nested structures
- Comments for documentation

**Location**: `~/.macmini-assistant/config.yaml`

**Structure**:
```yaml
app:
  download_folder: /tmp/downloads
  auto_start: true
  auto_update: true
  log_level: info

copilot:
  api_key: ${GITHUB_COPILOT_API_KEY}  # env var support
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
  # ... (as shown above)

updater:
  github_repo: user/macmini-assistant
  check_interval_hours: 6
  enabled: true
```

---

### 3. Status Panel: Execution Logs + Error Traces ✅
**Decision**: Discord status panel shows real-time tool execution logs and detailed error traces.

**Implementation**:
```go
// Status Panel Message Format
type StatusMessage struct {
    Type      string    // "execution_start", "execution_complete", "error"
    ToolName  string
    UserID    string
    Timestamp time.Time
    Duration  *time.Duration  // for completed executions
    Result    *ToolResult      // for successful completions
    Error     *ErrorDetails    // for failures
}

// Discord Embed Example
{
  "title": "🎬 YouTube Download Started",
  "color": 3447003,  // Blue
  "fields": [
    {"name": "Tool", "value": "youtube_download", "inline": true},
    {"name": "User", "value": "@username", "inline": true},
    {"name": "URL", "value": "https://youtube.com/...", "inline": false}
  ],
  "timestamp": "2026-01-31T10:30:00Z"
}
```

**Monitored Events**:
- ✅ Tool execution start
- ✅ Tool execution complete (with duration)
- ✅ Tool execution errors (with stack trace)
- ⚠️ System warnings (rate limits, low disk space)

---

## Open Questions & Recommendations

### Q1: OAuth2 Flow for Google Drive
**Question**: How should users authenticate Google Drive?

**Recommendation**: **Service Account (Primary) + Browser Flow (Fallback)**

**Rationale**:
1. **Service Account** (Recommended for v1.0):
   - Headless authentication (no browser needed)
   - Suitable for server/bot use case
   - One-time setup via JSON credentials file
   - Automatic token refresh

2. **Browser Flow** (Future enhancement):
   - Better UX for multiple users
   - Per-user Drive access
   - Requires web server for OAuth callback

**Implementation for v1.0**:
```go
// Use service account credentials
creds, err := google.CredentialsFromJSON(
    ctx,
    jsonKey,
    drive.DriveScope,
)
```

**Setup Instructions** (for users):
1. Create Google Cloud Project
2. Enable Drive API
3. Create Service Account
4. Download JSON credentials
5. Place at `~/.macmini-assistant/gdrive-creds.json`
6. Share target folder with service account email

---

### Q2: Rate Limiting
**Question**: Should we implement rate limits per user/tool?

**Recommendation**: **Yes, implement basic rate limiting in v1.0**

**Rationale**:
- Prevent abuse of expensive operations (video downloads)
- Protect against OpenAI API rate limits
- Ensure fair usage if multiple users

**Implementation**:
```go
// Per-user, per-tool rate limiting
type RateLimiter struct {
    MaxRequests int           // e.g., 10 requests
    Window      time.Duration // e.g., per hour
}

// Config
rate_limits:
  youtube_download:
    max_requests: 10
    window_minutes: 60

  gdrive_upload:
    max_requests: 20
    window_minutes: 60

  llm_requests:
    max_requests: 100
    window_minutes: 60
```

**Error Message**:
> "⏸️ Rate limit exceeded for youtube_download. You can make 10 requests per hour. Please try again in 23 minutes."

---

### Q3: File Cleanup
**Question**: Auto-delete downloaded files after N days?

**Recommendation**: **Yes, implement configurable auto-cleanup**

**Rationale**:
- Prevent disk space exhaustion
- User may forget to manually clean up
- Configurable for different use cases

**Implementation**:
```yaml
# config.yaml
app:
  download_folder: /tmp/downloads
  auto_cleanup:
    enabled: true
    retention_days: 7
    cleanup_schedule: "0 2 * * *"  # Daily at 2 AM (cron syntax)
```

**Features**:
- Scan download folder daily
- Delete files older than N days
- Skip files in use
- Log cleanup actions
- Manual cleanup via tray menu: "Clean Downloads..."

---

### Q4: macOS Notifications
**Question**: Should we send macOS notifications for tool completions?

**Recommendation**: **Yes, add optional notifications in v1.0**

**Rationale**:
- Better UX when user is on Mac
- Immediate feedback without checking Discord/LINE
- Can be disabled if annoying

**Implementation**:
```yaml
# config.yaml
notifications:
  enabled: true
  events:
    - tool_complete
    - tool_error
    - update_available
```

**Library**: Use `github.com/deckarep/gosx-notifier` (macOS-specific)

**Example Notification**:
```
Title: YouTube Download Complete
Body: video.mp4 (125 MB)
      /tmp/downloads/video.mp4
Actions: [Open Folder] [Dismiss]
```

---

## Technology Recommendations

### Logging Library: `log/slog` (Go 1.21+) ✅
**Rationale**:
- Standard library (no external dependency)
- Structured logging built-in
- High performance
- JSON output support
- Context-aware

**Alternative**: `github.com/rs/zerolog` (if more features needed)

---

### Config Format: YAML ✅
**Selected**: `gopkg.in/yaml.v3`

**Alternatives Considered**:
- ❌ TOML: Less familiar, more complex syntax
- ❌ JSON: No comments, less human-friendly
- ✅ YAML: Best balance of readability and features

---

### Versioning: Semantic Versioning (SemVer) ✅
Format: `vMAJOR.MINOR.PATCH` (e.g., `v1.0.0`)

**Git Tags**:
```bash
git tag -a v1.0.0 -m "Release v1.0.0 - Initial stable release"
```

---

### Metrics: None for v1.0, Optional for v1.1 📊
**Decision**: Skip metrics collection in v1.0 to reduce complexity.

**Future Enhancement** (v1.1+):
- Prometheus metrics endpoint
- Metrics: request count, latency, error rate, tool usage
- Grafana dashboard (optional)

---

## Risk Mitigation Strategies

### 1. Downie Deep Link API Changes
**Mitigation**:
- Document exact Downie version used
- Pin Downie version in documentation
- Create fallback plan: direct `youtube-dl` or `yt-dlp` integration
- Monitor Downie release notes

### 2. macOS Sandboxing & Permissions
**Mitigation**:
- Clear documentation of required permissions
- Startup permission checks with user-friendly prompts
- Graceful degradation if permissions denied

**Required Permissions**:
- Full Disk Access (for download folder access)
- Accessibility (if needed for system tray)
- Network (for webhooks and APIs)

**Implementation**:
```go
// Check permissions on startup
func CheckPermissions() error {
    if !hasFullDiskAccess() {
        return fmt.Errorf("Full Disk Access required. Please enable in System Preferences > Security & Privacy")
    }
    return nil
}
```

### 3. GitHub Copilot SDK Rate Limits
**Mitigation**:
- Implement request queuing
- User feedback: "Request queued, position #3"
- Exponential backoff on rate limit errors
- Monitor usage and adjust tier if needed

---

## Testing Strategy Summary

### Test Categories

1. **Unit Tests** (No build tag)
   - Run in CI: ✅
   - Coverage target: >80%
   - Fast execution (<1s per test)

2. **Integration Tests** (`//go:build integration`)
   - Run in CI: ❌ (optional, with credentials)
   - Require external services (Discord, LINE, Copilot)
   - Use test accounts/channels

3. **Local Tests** (`//go:build local`)
   - Run in CI: ❌
   - Require local tools (Downie, macOS-specific)
   - Manual execution on development machine

### CI Configuration
```yaml
# .github/workflows/test.yml
- name: Run unit tests
  run: go test ./... -v -coverprofile=coverage.out

- name: Check coverage
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$coverage < 80" | bc -l) )); then
      echo "Coverage $coverage% is below 80%"
      exit 1
    fi
```

---

## Security Considerations

### Credential Management
**Best Practice**: Use environment variables + macOS Keychain

**Implementation**:
```go
// Try keychain first, fall back to env vars
func GetCredential(key string) (string, error) {
    // Try macOS Keychain
    if val, err := keychain.Get(key); err == nil {
        return val, nil
    }

    // Fall back to environment variable
    if val := os.Getenv(key); val != "" {
        return val, nil
    }

    return "", fmt.Errorf("credential %s not found", key)
}
```

**Library**: `github.com/keybase/go-keychain`

### Webhook Signature Validation
**Critical**: Always validate LINE and Discord webhook signatures

```go
// LINE signature validation
func ValidateLineSignature(body []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

### Log Sanitization
**Never log**:
- API keys or tokens
- User passwords
- Webhook signatures (beyond validation)
- Full URLs (may contain sensitive query params)

```go
// Sanitize URLs before logging
func SanitizeURL(u string) string {
    parsed, _ := url.Parse(u)
    parsed.RawQuery = ""  // Remove query params
    return parsed.String()
}
```

---

## Next Steps

### Immediate Actions (This Week)
1. **Finalize PRD**: Review and approve [PRD.md](./PRD.md)
2. **Set up repository**:
   - Initialize Go module
   - Create project structure (as per Development Plan)
   - Configure GitHub Actions
3. **Create ADR document template**:
   - Document all architectural decisions
   - Format: `docs/adr/001-architecture.md`
4. **Environment setup**:
   - Install required tools (golangci-lint, goreleaser)
   - Set up test accounts (LINE, Discord)
   - Create GitHub Copilot SDK account

### Week 1 (Phase 0)
- [ ] Complete project bootstrap
- [ ] Set up CI/CD pipeline
- [ ] Create development environment guide
- [ ] Prepare task tracking (GitHub Projects or Issues)

### Week 2-3 (Phase 1)
- [ ] Implement configuration system
- [ ] Build tool registry
- [ ] Set up logging and error handling

---

## Discussion Points for Team Meeting

### For Stakeholder Review
1. **Budget & Resources**:
   - GitHub Copilot SDK costs (per-request pricing)
   - Google Cloud costs (for Drive API)
   - Developer time allocation (12 weeks)

2. **Scope Confirmation**:
   - Is v1.0 scope acceptable?
   - Any critical features missing?
   - Priority of future enhancements (v1.1+)

3. **Timeline**:
   - Is 12-week timeline realistic?
   - Any hard deadlines?
   - Beta testing period needed?

### For Development Team
1. **Tooling Preferences**:
   - IDE setup (VS Code recommended)
   - Prefer make or task runner?
   - Code review process (PR-based?)

2. **Communication**:
   - Daily standups needed?
   - Preferred communication channel (Slack, Discord)?
   - Weekly demo format?

3. **Testing**:
   - Who manages test accounts/credentials?
   - Integration test execution frequency?
   - Manual testing checklist?

---

## Resources & References

### Documentation
- [PRD.md](./PRD.md) - Complete product requirements
- [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md) - Detailed development phases
- [spec.md](../requirement/spec.md) - Original specification

### External Documentation
- [GitHub Copilot SDK Docs](https://github.com/copilot)
- [LINE Messaging API](https://developers.line.biz/en/docs/messaging-api/)
- [Discord Developer Docs](https://discord.com/developers/docs)
- [Downie Deep Link](https://software.charliemonroe.net/downie/)
- [Google Drive API v3](https://developers.google.com/drive/api/v3/reference)

### Libraries to Review
- [systray](https://github.com/getlantern/systray)
- [gin](https://github.com/gin-gonic/gin)
- [cobra](https://github.com/spf13/cobra)
- [line-bot-sdk-go](https://github.com/line/line-bot-sdk-go)
- [discordgo](https://github.com/bwmarrin/discordgo)
- [go-update](https://github.com/inconshreveable/go-update)

---

## Appendix: Example User Flows

### Flow 1: First-Time Setup
```bash
# User downloads binary from GitHub releases
$ brew install macmini-assistant  # or manual download

# Generate default config
$ macmini-assistant init
✓ Created config directory: ~/.macmini-assistant
✓ Generated default config: ~/.macmini-assistant/config.yaml

Please edit the config file and add your credentials:
  - LINE_CHANNEL_SECRET
  - LINE_ACCESS_TOKEN
  - DISCORD_BOT_TOKEN
  - GITHUB_COPILOT_API_KEY

# User edits config with their credentials
$ vim ~/.macmini-assistant/config.yaml

# Start the application
$ macmini-assistant start
✓ System tray icon appeared
✓ Webhook server started on :8080
✓ LINE bot connected
✓ Discord bot connected
✓ Registered 2 tools with Copilot SDK
✓ Ready to receive requests

# Enable auto-start
$ macmini-assistant autostart enable
✓ LaunchAgent installed
✓ App will start on next login
```

### Flow 2: YouTube Download via LINE
```
User (LINE): "Hey bot, download this video for me:
              https://www.youtube.com/watch?v=dQw4w9WgXcQ
              in 1080p please"

Bot (thinking): [Sends to Copilot SDK]

Copilot SDK: [Analyzes intent] -> Call youtube_download tool

Bot: [Executes downie tool]

Discord Status Channel:
  🎬 YouTube Download Started
  Tool: youtube_download
  User: @john (LINE)
  URL: https://youtube.com/...
  Format: mp4
  Resolution: 1080p
  Started: 2026-01-31 10:30:00

[30 seconds later]

Discord Status Channel:
  ✅ YouTube Download Complete
  Duration: 32.5 seconds
  File: /tmp/downloads/Rick_Astley_Never_Gonna_Give_You_Up.mp4
  Size: 125 MB

Bot (LINE reply):
  "✅ Video downloaded successfully!

  📁 File: Rick_Astley_Never_Gonna_Give_You_Up.mp4
  📊 Size: 125 MB
  💾 Location: /tmp/downloads/Rick_Astley_Never_Gonna_Give_You_Up.mp4

  The file will be automatically deleted in 7 days."
```

### Flow 3: Error Handling
```
User (Discord DM): "@bot upload /invalid/path.txt to Drive"

Bot: [Sends to Copilot SDK]

Copilot: [Calls gdrive_upload tool]

Bot: [Attempts to upload]
     [Error: file not found]

Discord Status Channel:
  ❌ Google Drive Upload Failed
  Tool: gdrive_upload
  User: @sarah (Discord)
  Error: File not found: /invalid/path.txt

  Stack Trace:
  internal/tools/gdrive/gdrive.go:45
  internal/registry/registry.go:120
  ...

Bot (Discord reply):
  "❌ Upload failed: File not found

  The file `/invalid/path.txt` does not exist.

  Please check the file path and try again.
  You can only upload files from your Mac Mini."
```

---

**Document Status**: Approved
**Next Review**: After Phase 0 completion
**Questions?** Create a GitHub issue or contact the team lead.
