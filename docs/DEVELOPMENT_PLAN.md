# Development Plan
# MacMini Assistant Systray Orchestrator

**Version:** 1.0
**Date:** January 31, 2026
**Methodology:** Test-Driven Development (TDD)

---

## Overview

This document outlines the phased development approach for the MacMini Assistant Systray Orchestrator. The project follows TDD principles with comprehensive test coverage and uses Go build tags to manage environment-specific tests.

### Development Principles

1. **Test-First Development**: Write tests before implementation
2. **Incremental Delivery**: Each phase produces a working deliverable
3. **Continuous Integration**: All commits pass automated tests (excluding local-only tests)
4. **Code Review**: All changes reviewed before merge
5. **Documentation**: Update docs alongside code

---

## Project Phases

### Phase 0: Project Bootstrap (Week 1)
**Goal**: Set up project infrastructure and development environment

#### Tasks
- [x] Initialize Go module
- [ ] Set up project structure
- [ ] Configure GitHub Actions CI/CD
  - [ ] Test workflow (excluding `local` and `integration` tags)
  - [ ] Build workflow for macOS
  - [ ] Release workflow with goreleaser
- [ ] Set up development tools
  - [ ] golangci-lint configuration
  - [ ] Pre-commit hooks
  - [ ] Makefile for common tasks
- [ ] Create base documentation
  - [x] PRD
  - [x] Development Plan
  - [ ] Contributing guidelines
  - [ ] Architecture Decision Records (ADR) template

#### Deliverables
- Working CI/CD pipeline
- Project structure skeleton
- Development environment documentation

#### Testing Strategy
```bash
# Run all tests except local-only
make test

# Run with coverage
make test-coverage

# Local testing with all tags
make test-local
```

---

### Phase 1: Core Foundation (Weeks 2-3)
**Goal**: Build the foundational components without external integrations

#### 1.1 Configuration System
**Duration**: 2 days

**Implementation**:
- [ ] Define config schema (YAML)
- [ ] Implement config loader with validation
- [ ] Add default config generation
- [ ] Support environment variable overrides (optional)

**Test Cases**:
```go
// internal/config/config_test.go
TestLoadConfig_ValidFile
TestLoadConfig_InvalidYAML
TestLoadConfig_MissingRequired
TestLoadConfig_DefaultValues
TestValidateConfig_DownloadFolder
TestGenerateDefaultConfig
```

**Acceptance Criteria**:
- ✅ Config loads from `~/.macmini-assistant/config.yaml`
- ✅ Validation fails early with clear error messages
- ✅ `orchestrator init` generates valid default config
- ✅ All config fields have defaults or required validation

---

#### 1.2 Tool Registry
**Duration**: 3 days

**Implementation**:
- [ ] Define tool interface
- [ ] Implement registry with dynamic loading
- [ ] Add tool metadata validation
- [ ] Create tool execution wrapper (timeout, logging)

**Tool Interface**:
```go
// internal/registry/tool.go
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error)
    Schema() ToolSchema
}

type ToolSchema struct {
    Inputs  []Parameter
    Outputs []Parameter
}
```

**Test Cases**:
```go
// internal/registry/registry_test.go
TestRegistry_RegisterTool
TestRegistry_GetTool
TestRegistry_ExecuteWithTimeout
TestRegistry_ExecuteWithInvalidParams
TestRegistry_LoadFromConfig
TestToolSchema_Validation
```

**Acceptance Criteria**:
- ✅ Tools registered via config file
- ✅ Tool execution respects 10-min timeout
- ✅ Invalid parameters rejected with clear errors
- ✅ Tool metadata accessible for Copilot SDK registration

---

#### 1.3 Logging & Error Handling
**Duration**: 2 days

**Implementation**:
- [ ] Set up structured logging (recommend: `log/slog`)
- [ ] Define error types and wrapping
- [ ] Create error reporter interface
- [ ] Add request ID tracing

**Test Cases**:
```go
// internal/observability/logger_test.go
TestLogger_StructuredOutput
TestLogger_LogLevels
TestErrorReporter_CaptureError
TestRequestID_Propagation
```

**Acceptance Criteria**:
- ✅ All logs in JSON format with timestamps
- ✅ Errors include context (stack trace, request ID)
- ✅ Log levels configurable (debug, info, warn, error)
- ✅ No sensitive data in logs

---

### Phase 2: Messaging Platform Integration (Weeks 4-5)
**Goal**: Implement LINE and Discord bot handlers

#### 2.1 LINE Bot Handler
**Duration**: 3 days

**Implementation**:
- [ ] Webhook endpoint with Gin
- [ ] Signature validation
- [ ] Message parsing and routing
- [ ] Reply message formatting
- [ ] Error handling and retry logic

**Test Cases**:
```go
// internal/handlers/line_test.go
TestLineHandler_WebhookValidation
TestLineHandler_TextMessage
TestLineHandler_InvalidSignature
TestLineHandler_ReplyFormatting
TestLineHandler_ErrorResponse
```

**Build Tag**: Standard (no special tags needed)

**Acceptance Criteria**:
- ✅ Webhook validates LINE signatures
- ✅ Bot responds only when messaged
- ✅ Messages parsed and forwarded to orchestrator
- ✅ Errors returned in user-friendly format

---

#### 2.2 Discord Bot Handler
**Duration**: 4 days

**Implementation**:
- [ ] Bot connection with intents
- [ ] Message event handling (mentions, DMs)
- [ ] Slash commands (`/status`, `/tools`, `/help`)
- [ ] Status panel integration (separate channel)
- [ ] Rich embed formatting

**Test Cases**:
```go
// internal/handlers/discord_test.go
TestDiscordHandler_MessageCreate
TestDiscordHandler_SlashCommand_Status
TestDiscordHandler_SlashCommand_Tools
TestDiscordHandler_StatusPanel_ToolExecution
TestDiscordHandler_StatusPanel_ErrorReport
TestDiscordHandler_RichEmbedFormatting
```

**Build Tag**: `integration` (requires Discord test bot)

**Acceptance Criteria**:
- ✅ Bot responds to mentions and DMs
- ✅ Slash commands functional
- ✅ Status panel posts tool execution events
- ✅ Errors posted to status channel with details
- ✅ Color-coded rich embeds

---

#### 2.3 Unified Message Interface
**Duration**: 2 days

**Implementation**:
- [ ] Abstract message interface for platforms
- [ ] Common message routing logic
- [ ] Platform-agnostic error formatting

**Test Cases**:
```go
// internal/handlers/interface_test.go
TestMessageInterface_Routing
TestMessageInterface_ErrorFormatting
TestMessageInterface_PlatformConversion
```

**Acceptance Criteria**:
- ✅ Single interface for LINE and Discord
- ✅ Platform-specific formatting abstracted
- ✅ Easy to add new platforms (Slack, etc.)

---

### Phase 3: GitHub Copilot SDK Integration (Week 6)
**Goal**: Connect orchestrator to Copilot SDK for AI-powered tool routing

#### 3.1 Copilot SDK Client
**Duration**: 4 days

**Implementation**:
- [ ] Initialize Copilot SDK client
- [ ] Register tools with SDK on startup
- [ ] Stream tool execution requests
- [ ] Handle tool responses and errors
- [ ] Implement timeout enforcement (10 min)

**Test Cases**:
```go
// internal/copilot/client_test.go
TestCopilotClient_Initialize
TestCopilotClient_RegisterTools
TestCopilotClient_StreamRequests
TestCopilotClient_ExecuteToolWithTimeout
TestCopilotClient_HandleLLMResponse
TestCopilotClient_ErrorHandling
```

**Build Tag**: `integration` (requires Copilot API credentials)

**Acceptance Criteria**:
- ✅ All registered tools available in Copilot
- ✅ Tool requests routed to correct handlers
- ✅ Hard timeout at 10 minutes
- ✅ Concurrent requests handled safely
- ✅ LLM responses returned to messaging platforms

---

#### 3.2 Request/Response Pipeline
**Duration**: 2 days

**Implementation**:
- [ ] Message → Copilot request transformation
- [ ] Copilot response → Platform message transformation
- [ ] Context preservation across requests
- [ ] Error propagation

**Test Cases**:
```go
// internal/copilot/pipeline_test.go
TestPipeline_MessageToCopilot
TestPipeline_CopilotToMessage
TestPipeline_ContextPreservation
TestPipeline_ErrorPropagation
```

---

### Phase 4: Tool Implementation (Weeks 7-8)
**Goal**: Implement YouTube download and Google Drive upload tools

#### 4.1 Downie Tool (YouTube Download)
**Duration**: 3 days

**Implementation**:
- [ ] Deep link URL construction
- [ ] Format/resolution validation
- [ ] File path resolution
- [ ] Download completion detection
- [ ] Error handling (invalid URL, unsupported format)

**Test Cases**:
```go
// internal/tools/downie/downie_test.go
//go:build local

TestDownie_DeepLinkConstruction
TestDownie_DownloadMP4_1080p
TestDownie_InvalidURL
TestDownie_UnsupportedFormat
TestDownie_FilePathResolution
TestDownie_FileSizeCalculation
```

**Build Tag**: `local` (requires Downie installed)

**Acceptance Criteria**:
- ✅ Downloads YouTube videos via Downie
- ✅ Supports all specified formats and resolutions
- ✅ Returns absolute file path and size
- ✅ Handles errors gracefully
- ✅ Files saved to configured download folder

---

#### 4.2 Google Drive Upload Tool
**Duration**: 4 days

**Implementation**:
- [ ] OAuth2 authentication flow
- [ ] File upload with resumable upload
- [ ] Progress tracking (optional for v1)
- [ ] Share link generation
- [ ] Timeout handling

**Test Cases**:
```go
// internal/tools/gdrive/gdrive_test.go
//go:build integration

TestGDrive_OAuth2Flow
TestGDrive_UploadFile
TestGDrive_CustomFileName
TestGDrive_ShareLinkGeneration
TestGDrive_UploadTimeout
TestGDrive_LargeFileUpload
```

**Build Tag**: `integration` (requires Google Cloud credentials)

**Acceptance Criteria**:
- ✅ OAuth2 authentication with browser flow
- ✅ Files uploaded successfully
- ✅ Share links generated and public-accessible
- ✅ Custom file names supported
- ✅ Timeout respected
- ✅ Handles network interruptions

---

### Phase 5: System Tray & Auto-start (Week 9)
**Goal**: Create macOS system tray interface and startup integration

#### 5.1 System Tray Application
**Duration**: 3 days

**Implementation**:
- [ ] System tray icon and menu
- [ ] Menu items: Start/Stop, Settings, Check Updates, Quit
- [ ] Icon state indicators (active, error, idle)
- [ ] Graceful shutdown handling

**Test Cases**:
```go
// internal/systray/tray_test.go
//go:build local

TestSysTray_Initialize
TestSysTray_MenuItems
TestSysTray_IconStates
TestSysTray_GracefulShutdown
```

**Build Tag**: `local` (requires macOS GUI)

**Acceptance Criteria**:
- ✅ Icon appears in macOS menu bar
- ✅ Menu items functional
- ✅ Icon reflects application state
- ✅ App shuts down gracefully on Quit

---

#### 5.2 Auto-start on Login
**Duration**: 2 days

**Implementation**:
- [ ] LaunchAgent plist generation
- [ ] Installation to `~/Library/LaunchAgents/`
- [ ] Enable/disable via config
- [ ] Uninstall functionality

**Test Cases**:
```go
// internal/systray/autostart_test.go
//go:build local

TestAutoStart_PlistGeneration
TestAutoStart_Install
TestAutoStart_Uninstall
TestAutoStart_ConfigToggle
```

**Build Tag**: `local` (requires macOS)

**Acceptance Criteria**:
- ✅ App starts on login when enabled
- ✅ LaunchAgent correctly configured
- ✅ Can be disabled via config
- ✅ Uninstall removes LaunchAgent

---

### Phase 6: Auto-updater (Week 10)
**Goal**: Implement self-updating from GitHub releases

#### 6.1 Update Checker
**Duration**: 2 days

**Implementation**:
- [ ] GitHub Releases API polling
- [ ] Version comparison (SemVer)
- [ ] Periodic check (every 6 hours)
- [ ] Manual trigger from tray menu

**Test Cases**:
```go
// internal/updater/checker_test.go
TestUpdateChecker_CheckLatestRelease
TestUpdateChecker_VersionComparison
TestUpdateChecker_PeriodicPolling
TestUpdateChecker_ManualTrigger
```

**Acceptance Criteria**:
- ✅ Polls releases every 6 hours
- ✅ Detects newer versions correctly
- ✅ Manual check available
- ✅ No API rate limit issues

---

#### 6.2 Binary Updater
**Duration**: 3 days

**Implementation**:
- [ ] Download release binary
- [ ] Checksum verification
- [ ] Atomic binary replacement
- [ ] Graceful restart
- [ ] Rollback on failure

**Test Cases**:
```go
// internal/updater/updater_test.go
TestUpdater_DownloadBinary
TestUpdater_ChecksumVerification
TestUpdater_BinaryReplacement
TestUpdater_GracefulRestart
TestUpdater_RollbackOnFailure
```

**Acceptance Criteria**:
- ✅ Binary downloads and verifies checksums
- ✅ Replacement is atomic (no partial updates)
- ✅ App restarts automatically
- ✅ Rolls back if new version crashes
- ✅ User notified of update status

---

### Phase 7: Integration & Testing (Week 11)
**Goal**: End-to-end testing and bug fixes

#### 7.1 Integration Tests
**Duration**: 3 days

**Implementation**:
- [ ] End-to-end workflow tests
  - [ ] LINE → Copilot → Downie → Response
  - [ ] Discord → Copilot → GDrive → Response
  - [ ] Status panel logging
  - [ ] Error handling paths
- [ ] Performance testing
  - [ ] Concurrent requests
  - [ ] Memory usage under load
  - [ ] Startup time

**Test Cases**:
```go
// test/integration/e2e_test.go
//go:build integration

TestE2E_LineYouTubeDownload
TestE2E_DiscordGoogleDriveUpload
TestE2E_StatusPanelLogging
TestE2E_ConcurrentRequests
TestE2E_LLMTimeout
```

**Acceptance Criteria**:
- ✅ All user stories functional end-to-end
- ✅ No memory leaks
- ✅ Startup time <5 seconds
- ✅ Handles 10+ concurrent requests

---

#### 7.2 Bug Fixes & Polish
**Duration**: 2 days

**Tasks**:
- [ ] Address failing tests
- [ ] Fix edge cases
- [ ] Improve error messages
- [ ] Performance optimizations
- [ ] Code cleanup

---

### Phase 8: Documentation & Release (Week 12)
**Goal**: Prepare for v1.0 release

#### 8.1 Documentation
**Duration**: 2 days

**Tasks**:
- [ ] User guide (installation, configuration, usage)
- [ ] Developer guide (adding tools, testing)
- [ ] API documentation (godoc)
- [ ] Troubleshooting guide
- [ ] Release notes

---

#### 8.2 Release Preparation
**Duration**: 1 day

**Tasks**:
- [ ] Version tagging (v1.0.0)
- [ ] Build release binaries (goreleaser)
- [ ] Create GitHub release with notes
- [ ] Test auto-updater with release

---

#### 8.3 Post-release Monitoring
**Duration**: 2 days

**Tasks**:
- [ ] Monitor for crash reports
- [ ] Quick bug fix releases if needed
- [ ] Gather user feedback

---

## Development Workflow

### Daily Workflow
```bash
# 1. Pull latest changes
git pull origin main

# 2. Create feature branch
git checkout -b feature/tool-registry

# 3. Write tests first
vim internal/registry/registry_test.go

# 4. Run tests (should fail)
make test

# 5. Implement feature
vim internal/registry/registry.go

# 6. Run tests (should pass)
make test

# 7. Run linter
make lint

# 8. Commit and push
git add .
git commit -m "feat: implement tool registry"
git push origin feature/tool-registry

# 9. Create PR on GitHub
```

### Testing Commands

```makefile
# Makefile

.PHONY: test test-local test-integration test-coverage lint

# Run standard tests (CI-safe)
test:
	go test ./... -v

# Run all tests including local-only
test-local:
	go test ./... -v -tags=local

# Run integration tests
test-integration:
	go test ./... -v -tags=integration

# Run all tests
test-all:
	go test ./... -v -tags=local,integration

# Coverage report
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# Lint
lint:
	golangci-lint run

# Build
build:
	go build -o bin/orchestrator cmd/orchestrator/main.go

# Run locally
run:
	go run cmd/orchestrator/main.go

# Clean
clean:
	rm -rf bin/ coverage.out
```

---

## Project Structure

```
macmini-assistant-systray/
├── cmd/
│   └── orchestrator/           # Main application entry point
│       └── main.go
├── internal/
│   ├── config/                 # Configuration management
│   │   ├── config.go
│   │   └── config_test.go
│   ├── registry/               # Tool registry
│   │   ├── registry.go
│   │   ├── tool.go
│   │   └── registry_test.go
│   ├── copilot/                # Copilot SDK integration
│   │   ├── client.go
│   │   ├── pipeline.go
│   │   └── client_test.go
│   ├── handlers/               # Message platform handlers
│   │   ├── line.go
│   │   ├── discord.go
│   │   ├── interface.go
│   │   └── discord_test.go
│   ├── tools/                  # Tool implementations
│   │   ├── downie/
│   │   │   ├── downie.go
│   │   │   └── downie_test.go  # +build local
│   │   └── gdrive/
│   │       ├── gdrive.go
│   │       └── gdrive_test.go  # +build integration
│   ├── systray/                # System tray integration
│   │   ├── tray.go
│   │   ├── autostart.go
│   │   └── tray_test.go        # +build local
│   ├── updater/                # Auto-updater
│   │   ├── checker.go
│   │   ├── updater.go
│   │   └── updater_test.go
│   └── observability/          # Logging and monitoring
│       ├── logger.go
│       └── logger_test.go
├── test/
│   ├── integration/            # E2E integration tests
│   │   └── e2e_test.go         # +build integration
│   └── fixtures/               # Test data
├── docs/
│   ├── PRD.md                  # Product Requirements Document
│   ├── DEVELOPMENT_PLAN.md     # This file
│   ├── USER_GUIDE.md           # User documentation
│   └── DEVELOPER_GUIDE.md      # Developer documentation
├── .github/
│   └── workflows/
│       ├── test.yml            # CI testing
│       ├── build.yml           # Build workflow
│       └── release.yml         # Release automation
├── go.mod
├── go.sum
├── Makefile
├── .golangci.yml               # Linter configuration
├── .goreleaser.yml             # Release configuration
└── README.md
```

---

## Risk Management

### Technical Risks

| Risk | Mitigation | Status |
|------|------------|--------|
| **Downie API instability** | Version pinning, fallback to ffmpeg | Monitoring |
| **Copilot SDK breaking changes** | Version locking, thorough testing | Monitoring |
| **macOS permission issues** | Early testing, clear documentation | In progress |
| **Auto-update failures** | Rollback mechanism, checksum validation | Planned |

### Timeline Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| **Underestimated complexity** | 1-2 week delay | Buffer in Phase 7 |
| **External API downtime** | Test delays | Mock services for testing |
| **Scope creep** | Timeline slip | Strict adherence to PRD v1.0 |

---

## Success Criteria

### Phase Completion Checklist
Each phase is considered complete when:
- [ ] All planned features implemented
- [ ] Tests passing (with appropriate build tags)
- [ ] Code reviewed and merged
- [ ] Documentation updated
- [ ] No critical bugs

### Project Completion Checklist
- [ ] All phases completed
- [ ] >80% test coverage
- [ ] All PRD requirements met
- [ ] Documentation complete
- [ ] v1.0.0 released on GitHub
- [ ] Auto-updater functional

---

## Key Milestones

| Milestone | Target Date | Deliverable |
|-----------|-------------|-------------|
| **M1: Foundation Complete** | End Week 3 | Config, registry, logging functional |
| **M2: Messaging Platforms Live** | End Week 5 | LINE and Discord bots operational |
| **M3: AI Integration** | End Week 6 | Copilot SDK routing requests |
| **M4: Tools Functional** | End Week 8 | Downie and GDrive tools working |
| **M5: System Tray Ready** | End Week 9 | macOS systray with auto-start |
| **M6: Self-updating** | End Week 10 | Auto-updater complete |
| **M7: Production Ready** | End Week 11 | All tests passing, bugs fixed |
| **M8: v1.0 Released** | End Week 12 | Public release on GitHub |

---

## Post-v1.0 Roadmap

### v1.1 (Month 2)
- [ ] Web dashboard for configuration
- [ ] Additional file format support in Downie
- [ ] Slack integration

### v1.2 (Month 3)
- [ ] Multi-user support
- [ ] Scheduled tasks (cron-like)
- [ ] Enhanced status panel with metrics

### v2.0 (Month 6)
- [ ] Tool marketplace
- [ ] Plugin system for external tools
- [ ] Mobile app for iOS

---

## Team & Responsibilities

| Role | Responsibilities | Contact |
|------|------------------|---------|
| **Lead Developer** | Overall architecture, code reviews | TBD |
| **Backend Engineer** | Copilot SDK, tools implementation | TBD |
| **Platform Engineer** | LINE/Discord integrations | TBD |
| **QA Engineer** | Test strategy, integration tests | TBD |
| **Technical Writer** | Documentation | TBD |

---

## Communication & Reporting

### Daily Standup (15 min)
- What did I complete yesterday?
- What will I work on today?
- Any blockers?

### Weekly Review (Friday)
- Demo completed features
- Review test coverage
- Plan next week's work

### Phase Review
- Formal phase completion review
- Retrospective: What went well? What to improve?
- Adjust timeline if needed

---

## Tools & Infrastructure

### Development Tools
- **IDE**: VS Code with Go extension
- **Version Control**: Git + GitHub
- **CI/CD**: GitHub Actions
- **Testing**: Go testing framework + testify
- **Linting**: golangci-lint
- **Release**: goreleaser

### Third-party Services
- **GitHub Copilot SDK**: AI orchestration
- **LINE Messaging API**: LINE bot
- **Discord API**: Discord bot
- **Google Drive API**: File uploads
- **GitHub Releases**: Binary distribution

---

## Appendix: Build Tags Reference

### Test Organization

```go
// Standard tests (run in CI)
// No build tag required

// Local-only tests (require local tools like Downie)
//go:build local

// Integration tests (require external services)
//go:build integration

// All tests (local development)
//go:build local && integration
```

### CI Configuration

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'
      - name: Run tests
        run: go test ./... -v
        # Excludes tests with 'local' and 'integration' tags
```

---

## Appendix: Versioning Strategy

### Semantic Versioning (SemVer)
- **MAJOR**: Breaking changes (e.g., 1.0.0 → 2.0.0)
- **MINOR**: New features, backward compatible (e.g., 1.0.0 → 1.1.0)
- **PATCH**: Bug fixes (e.g., 1.0.0 → 1.0.1)

### Git Tag Format
```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

---

## Questions?

For questions or clarifications:
- Create a GitHub issue
- Contact the lead developer
- Refer to [PRD.md](./PRD.md) for requirements

---

**Document Status**: Draft
**Last Updated**: January 31, 2026
**Next Review**: Week 3 (Phase 1 completion)
