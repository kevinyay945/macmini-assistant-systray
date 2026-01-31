# Phase 7: Integration & Testing

**Duration**: Week 11
**Status**: ⚪ Not Started
**Goal**: End-to-end testing and bug fixes

---

## Overview

This phase focuses on integration testing, performance testing, and bug fixes. All components should be working together by this point.

---

## 7.1 Integration Tests

**Duration**: 3 days
**Status**: ⚪ Not Started

### Tasks

- [ ] End-to-end workflow tests
  - [ ] LINE → Copilot → Downie → Response
  - [ ] Discord → Copilot → GDrive → Response
  - [ ] Status panel logging
  - [ ] Error handling paths
- [ ] Performance testing
  - [ ] Concurrent requests
  - [ ] Memory usage under load
  - [ ] Startup time

### End-to-End Test Scenarios

#### Scenario 1: LINE YouTube Download
```go
// test/integration/e2e_test.go
//go:build integration

func TestE2E_LineYouTubeDownload(t *testing.T) {
    // 1. Simulate LINE webhook with YouTube URL
    // 2. Verify Copilot SDK receives message
    // 3. Verify youtube_download tool is called
    // 4. Verify Downie is triggered
    // 5. Verify file is downloaded
    // 6. Verify response sent to LINE
    // 7. Verify status panel updated
}
```

#### Scenario 2: Discord Google Drive Upload
```go
func TestE2E_DiscordGoogleDriveUpload(t *testing.T) {
    // 1. Simulate Discord message with file path
    // 2. Verify Copilot SDK receives message
    // 3. Verify gdrive_upload tool is called
    // 4. Verify file is uploaded
    // 5. Verify share link generated
    // 6. Verify response sent to Discord
    // 7. Verify status panel updated
}
```

#### Scenario 3: Error Handling
```go
func TestE2E_ErrorHandling(t *testing.T) {
    // Test various error scenarios:
    // - Invalid YouTube URL
    // - File not found for upload
    // - Network timeout
    // - LLM timeout (10 min)
    // - Tool execution failure
}
```

#### Scenario 4: Concurrent Requests
```go
func TestE2E_ConcurrentRequests(t *testing.T) {
    // 1. Send 10 requests simultaneously
    // 2. Verify all are processed correctly
    // 3. Verify no race conditions
    // 4. Verify status panel logs all events
}
```

### Test Cases

```go
// test/integration/e2e_test.go
//go:build integration

func TestE2E_LineYouTubeDownload(t *testing.T)
func TestE2E_DiscordGoogleDriveUpload(t *testing.T)
func TestE2E_StatusPanelLogging(t *testing.T)
func TestE2E_ConcurrentRequests(t *testing.T)
func TestE2E_LLMTimeout(t *testing.T)
func TestE2E_ToolChaining(t *testing.T)
func TestE2E_ErrorRecovery(t *testing.T)
func TestE2E_GracefulShutdown(t *testing.T)
```

### Performance Test Cases

```go
// test/integration/performance_test.go
//go:build integration

func TestPerf_StartupTime(t *testing.T) {
    // Startup should complete in <5 seconds
}

func TestPerf_MemoryUsage(t *testing.T) {
    // Idle memory should be <100MB
    // Under load should be <500MB
}

func TestPerf_RequestThroughput(t *testing.T) {
    // Should handle 10+ concurrent requests
}

func TestPerf_ToolInvocationLatency(t *testing.T) {
    // Tool invocation overhead should be <2 seconds
}
```

### Acceptance Criteria

- [ ] All user stories functional end-to-end
- [ ] No memory leaks
- [ ] Startup time <5 seconds
- [ ] Handles 10+ concurrent requests
- [ ] All error paths tested

### Notes

<!-- Add your notes here -->

---

## 7.2 Bug Fixes & Polish

**Duration**: 2 days
**Status**: ⚪ Not Started

### Tasks

- [ ] Address failing tests
- [ ] Fix edge cases
- [ ] Improve error messages
- [ ] Performance optimizations
- [ ] Code cleanup

### Common Issues to Check

1. **Race Conditions**
   - Concurrent map access
   - Channel operations
   - State management

2. **Memory Leaks**
   - Goroutine leaks
   - Unclosed connections
   - Accumulated buffers

3. **Error Handling**
   - Nil pointer dereferences
   - Unhandled errors
   - Context cancellation

4. **Timeouts**
   - HTTP client timeouts
   - Context deadlines
   - Tool execution limits

### Code Quality Checklist

- [ ] All exported functions have doc comments
- [ ] Error messages are clear and actionable
- [ ] No hardcoded values (use config)
- [ ] Consistent naming conventions
- [ ] No duplicate code
- [ ] Tests cover edge cases
- [ ] Logging is structured and useful

### Performance Optimization Checklist

- [ ] HTTP client reuse (connection pooling)
- [ ] Minimize allocations in hot paths
- [ ] Use context.Context everywhere
- [ ] Buffer sizes are appropriate
- [ ] Goroutines are properly managed

### Notes

<!-- Add your notes here -->

---

## Testing Infrastructure

### Test Environment Setup

```bash
# Required environment variables for integration tests
export GITHUB_COPILOT_API_KEY="test-key"
export LINE_CHANNEL_SECRET="test-secret"
export LINE_ACCESS_TOKEN="test-token"
export DISCORD_BOT_TOKEN="test-token"
export DISCORD_TEST_CHANNEL_ID="123456789"
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/test-creds.json"
```

### Test Data (Fixtures)

```
test/
├── fixtures/
│   ├── config/
│   │   ├── valid_config.yaml
│   │   ├── invalid_config.yaml
│   │   └── minimal_config.yaml
│   ├── webhooks/
│   │   ├── line_text_message.json
│   │   ├── line_invalid_signature.json
│   │   └── discord_message.json
│   └── files/
│       └── test_upload.txt
```

### CI Configuration for Integration Tests

```yaml
# .github/workflows/integration.yml
name: Integration Tests

on:
  workflow_dispatch:  # Manual trigger only
  schedule:
    - cron: '0 0 * * *'  # Nightly

jobs:
  integration:
    runs-on: macos-latest
    environment: integration-test
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run integration tests
        env:
          GITHUB_COPILOT_API_KEY: ${{ secrets.COPILOT_API_KEY }}
          DISCORD_BOT_TOKEN: ${{ secrets.DISCORD_BOT_TOKEN }}
          # ... other secrets
        run: go test ./... -v -tags=integration -timeout=30m
```

---

## Deliverables

By the end of Phase 7:

- [ ] All integration tests passing
- [ ] Performance benchmarks met
- [ ] Bug backlog cleared
- [ ] Code quality verified

---

## Bug Tracking Template

| ID | Description | Severity | Status | Resolution |
|----|-------------|----------|--------|------------|
| B001 | Example bug | High | Open | |
| | | | | |

---

## Time Tracking

| Task | Estimated | Actual | Notes |
|------|-----------|--------|-------|
| 7.1 Integration Tests | 3 days | | |
| 7.2 Bug Fixes | 2 days | | |
| **Total** | **5 days** | | |

---

**Previous**: [Phase 6: Auto-updater](./phase-6-updater.md)
**Next**: [Phase 8: Documentation & Release](./phase-8-release.md)
