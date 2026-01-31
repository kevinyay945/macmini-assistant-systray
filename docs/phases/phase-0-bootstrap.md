# Phase 0: Project Bootstrap

**Duration**: Week 1
**Status**: ✅ Complete
**Goal**: Set up project infrastructure and development environment

---

## Overview

This phase establishes the foundation for all subsequent development work. By the end of this phase, we should have a fully functional development environment with CI/CD, linting, and project structure in place.

---

## Tasks

### 0.1 Initialize Go Module

**Status**: ✅ Complete

- [x] Create `go.mod` with module path
- [x] Add initial dependencies
- [x] Pin dependency versions (go.mod created with go 1.24)

**Commands**:
```bash
go mod init github.com/kevinyay945/macmini-assistant-systray
go mod tidy
```

**Notes**:
<!-- Add your notes here -->

---

### 0.2 Set Up Project Structure

**Status**: ✅ Complete

- [x] Create directory structure
- [x] Add `.gitignore`
- [x] Create placeholder files

**Directory Structure**:
```
macmini-assistant-systray/
├── cmd/
│   └── orchestrator/
│       └── main.go
├── internal/
│   ├── config/
│   ├── registry/
│   ├── copilot/
│   ├── handlers/
│   ├── tools/
│   │   ├── downie/
│   │   └── gdrive/
│   ├── systray/
│   ├── updater/
│   └── observability/
├── test/
│   ├── integration/
│   └── fixtures/
├── docs/
├── .github/
│   └── workflows/
├── go.mod
├── go.sum
├── Makefile
├── .golangci.yml
├── .goreleaser.yml
└── README.md
```

**Notes**:
<!-- Add your notes here -->

---

### 0.3 Configure GitHub Actions CI/CD

**Status**: ✅ Complete

#### 0.3.1 Test Workflow

- [x] Create `.github/workflows/test.yml`
- [x] Configure Go version (1.22+)
- [x] Run tests (excluding `local` and `integration` tags)
- [x] Upload coverage reports

**File**: `.github/workflows/test.yml`
```yaml
name: Test

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run tests
        run: go test ./... -v -race -coverprofile=coverage.out

      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $coverage%"
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "Coverage $coverage% is below 80%"
            exit 1
          fi

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out
```

**Notes**:
<!-- Add your notes here -->

---

#### 0.3.2 Build Workflow

- [x] Create `.github/workflows/build.yml`
- [x] Build for macOS arm64 (M3)
- [x] Upload build artifacts

**File**: `.github/workflows/build.yml`
```yaml
name: Build

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build
        run: |
          GOOS=darwin GOARCH=arm64 go build -o bin/orchestrator-darwin-arm64 cmd/orchestrator/main.go

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: orchestrator-darwin-arm64
          path: bin/orchestrator-darwin-arm64
```

**Notes**:
<!-- Add your notes here -->

---

#### 0.3.3 Release Workflow

- [x] Create `.github/workflows/release.yml`
- [x] Configure goreleaser
- [x] Trigger on version tags

**File**: `.github/workflows/release.yml`
```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Notes**:
<!-- Add your notes here -->

---

### 0.4 Set Up Development Tools

**Status**: ✅ Complete

#### 0.4.1 golangci-lint Configuration

- [x] Create `.golangci.yml`
- [x] Configure linters
- [x] Set lint rules

**File**: `.golangci.yml`
```yaml
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - gofmt
    - goimports
    - misspell
    - unconvert
    - gocritic
    - revive

linters-settings:
  gofmt:
    simplify: true
  goimports:
    local-prefixes: github.com/kevinyay945/macmini-assistant-systray
  revive:
    rules:
      - name: blank-imports
      - name: context-as-argument
      - name: dot-imports
      - name: error-return
      - name: error-strings
      - name: exported
      - name: if-return
      - name: increment-decrement
      - name: var-declaration

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

**Notes**:
<!-- Add your notes here -->

---

#### 0.4.2 Pre-commit Hooks

- [x] Install pre-commit
- [x] Create `.pre-commit-config.yaml`
- [x] Configure hooks (lint, test, format)

**File**: `.pre-commit-config.yaml`
```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.5.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files

  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: go fmt ./...
        language: system
        pass_filenames: false

      - id: go-lint
        name: golangci-lint
        entry: golangci-lint run
        language: system
        pass_filenames: false

      - id: go-test
        name: go test
        entry: go test ./... -v -short
        language: system
        pass_filenames: false
```

**Installation**:
```bash
brew install pre-commit
pre-commit install
```

**Notes**:
<!-- Add your notes here -->

---

#### 0.4.3 Makefile

- [x] Create `Makefile`
- [x] Add common tasks

**File**: `Makefile`
```makefile
.PHONY: all build test test-local test-integration test-all test-coverage lint clean run help

# Default target
all: lint test build

# Build the application
build:
	@echo "Building..."
	@go build -o bin/orchestrator cmd/orchestrator/main.go

# Run standard tests (CI-safe)
test:
	@echo "Running tests..."
	@go test ./... -v -race

# Run all tests including local-only
test-local:
	@echo "Running local tests..."
	@go test ./... -v -race -tags=local

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	@go test ./... -v -race -tags=integration

# Run all tests
test-all:
	@echo "Running all tests..."
	@go test ./... -v -race -tags=local,integration

# Coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lint
lint:
	@echo "Linting..."
	@golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ coverage.out coverage.html

# Run the application
run:
	@go run cmd/orchestrator/main.go

# Initialize development environment
init:
	@echo "Installing development dependencies..."
	@command -v golangci-lint > /dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@command -v goreleaser > /dev/null 2>&1 || go install github.com/goreleaser/goreleaser@latest
	@command -v pre-commit > /dev/null 2>&1 || brew install pre-commit
	@pre-commit install
	@echo "Development environment ready!"

# Show help
help:
	@echo "Available targets:"
	@echo "  all            - lint, test, and build"
	@echo "  build          - Build the application"
	@echo "  test           - Run standard tests (CI-safe)"
	@echo "  test-local     - Run tests including local-only"
	@echo "  test-integration - Run integration tests"
	@echo "  test-all       - Run all tests"
	@echo "  test-coverage  - Generate coverage report"
	@echo "  lint           - Run linter"
	@echo "  clean          - Remove build artifacts"
	@echo "  run            - Run the application"
	@echo "  init           - Initialize development environment"
	@echo "  help           - Show this help"
```

**Notes**:
<!-- Add your notes here -->

---

### 0.5 Create Base Documentation

**Status**: ✅ Complete

- [x] PRD (`docs/PRD.md`)
- [x] Development Plan (`docs/DEVELOPMENT_PLAN.md`)
- [ ] Contributing guidelines (`CONTRIBUTING.md`) - Deferred to later
- [ ] Architecture Decision Records template (`docs/adr/`) - Deferred to later
- [x] README (`README.md`)

**Notes**:
<!-- Add your notes here -->

---

### 0.6 goreleaser Configuration

**Status**: ✅ Complete

- [x] Create `.goreleaser.yml`
- [x] Configure macOS builds
- [x] Set up changelog generation

**File**: `.goreleaser.yml`
```yaml
project_name: macmini-assistant

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: orchestrator
    main: ./cmd/orchestrator
    binary: orchestrator
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
    goarch:
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: >-
      {{ .ProjectName }}_
      {{- .Version }}_
      {{- .Os }}_
      {{- .Arch }}
    files:
      - README.md
      - LICENSE

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'

release:
  github:
    owner: kevinyay945
    name: macmini-assistant-systray
  draft: false
  prerelease: auto
```

**Notes**:
<!-- Add your notes here -->

---

## Deliverables

By the end of Phase 0, the following should be complete:

- [x] Working CI/CD pipeline
  - [x] Tests run on every PR
  - [x] Build succeeds on every PR
  - [x] Releases triggered by version tags
- [x] Project structure skeleton
  - [x] All directories created
  - [x] Placeholder files in place
- [x] Development environment documentation
  - [x] How to set up local environment
  - [x] How to run tests
  - [ ] How to contribute (CONTRIBUTING.md deferred)

---

## Testing Strategy

```bash
# Run all tests except local-only
make test

# Run with coverage
make test-coverage

# Local testing with all tags
make test-local
```

---

## Acceptance Criteria

- [x] `make test` passes
- [x] `make lint` passes with no errors
- [x] `make build` produces a working binary
- [ ] GitHub Actions workflows are green (needs push to verify)
- [x] README contains setup instructions
- [x] All team members can clone and run `make init`

---

## Dependencies to Install

```bash
# Go 1.22+
brew install go

# golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# goreleaser
go install github.com/goreleaser/goreleaser@latest

# pre-commit
uv tool install pre-commit
```

---

## Notes & Discoveries

<!-- Add any notes, discoveries, or decisions made during this phase -->
check tools exist before installing these tools

### Date: 2026-01-31

**Note**: Completed Phase 0 bootstrap:
- Created go.mod with module path github.com/kevinyay945/macmini-assistant-systray
- Set up full project structure with placeholder files for all packages
- Created GitHub Actions workflows (test.yml, build.yml, release.yml, lint.yml)
- Created Makefile, .pre-commit-config.yaml, .goreleaser.yml
- Updated .gitignore with comprehensive ignore patterns
- Created README.md with setup instructions
- Verified: `make build`, `make lint`, `make test` all pass locally

---

### Date: 2026-01-31 (Code Review Refactor)

**Note**: Major code quality improvements based on code review:

#### Fixed Issues:
1. **registry_test.go** - Fixed corrupted test file with double package declarations
2. **go.mod** - Changed Go version from non-existent `1.24.0` to `1.23.0`
3. **main.go** - Refactored to use Cobra CLI framework instead of manual `os.Args` parsing

#### Architecture Improvements:
4. **config/config.go** - Added `Load()`, `Validate()`, and `DefaultConfigPath()` functions
5. **observability/logger.go** - Replaced standard `log` with Go 1.21+ `log/slog` for structured logging
6. **registry/registry.go** - Added `sync.RWMutex` for concurrent access safety
7. **copilot/client.go** - Added `Config` struct and `context.Context` support
8. **tools/downie/downie.go** - Added `Config` struct, `context.Context`, parameter validation
9. **tools/gdrive/gdrive.go** - Added `Config` struct, `context.Context`, parameter validation
10. **handlers/line/handler.go** - Added `Config` struct and proper HTTP handler signature
11. **handlers/discord/handler.go** - Added `Config` struct with token/guild storage
12. **systray/systray.go** - Added functional options pattern for callbacks
13. **updater/updater.go** - Added `Config`, `UpdateInfo` struct, and `context.Context` support

#### New Test Files:
- `internal/config/config_test.go` - 10 tests for config loading and validation
- `internal/observability/logger_test.go` - 5 tests for structured logger
- `internal/tools/downie/downie_test.go` - 8 tests for Downie tool
- `internal/tools/gdrive/gdrive_test.go` - 8 tests for Google Drive tool

#### Dependencies Added:
- `github.com/spf13/cobra v1.8.1` - CLI framework
- `gopkg.in/yaml.v3 v3.0.1` - YAML config parsing

#### Verification:
- `make test` - All 31 tests pass
- `make lint` - 0 issues
- `make build` - Successful

---

### Date: 2026-01-31 (Code Review Round 2 - Gilfoyle Style)

**Note**: Second round of code quality improvements based on Gilfoyle-style code review:

#### Fixed Issues:

1. **config/config.go** - Added bidirectional LINE credential validation
   - Now validates both directions: secret→token AND token→secret

2. **registry/registry.go** - Added deterministic `List()` output
   - Added `slices.Sort()` to ensure consistent ordering
   - Added `TestRegistry_List_IsSorted` test to verify

3. **copilot/client.go** - Added YAML/JSON struct tags to `Config`
   - `APIKey` now has `yaml:"api_key" json:"api_key"` tags

4. **handlers/line/handler.go** - Proper HTTP body handling
   - Added `defer r.Body.Close()` to prevent connection leaks

5. **tools/downie/downie.go** & **tools/gdrive/gdrive.go**:
   - Context check moved to FIRST operation (fail fast)
   - Added sentinel errors: `ErrNotEnabled`, `ErrMissingURL`, `ErrMissingFilePath`

6. **registry_test.go** - Added concurrent access test
   - `TestRegistry_ConcurrentAccess` validates thread safety with 100 goroutines

7. **cmd/orchestrator/main.go** - Graceful shutdown support
   - Added `signal.NotifyContext` for SIGINT/SIGTERM handling
   - Refactored to `run()` function pattern to satisfy gocritic linter

8. **downie_test.go** & **gdrive_test.go** - Fixed flaky tests
   - Replaced `time.Sleep` race condition with `time.Now().Add(-time.Second)` deadline
   - Now uses `errors.Is()` for precise sentinel error verification

9. **config_test.go** - Added `TestConfig_Validate_LINERequiresSecret`
   - Tests the reverse validation (token set but secret missing)

#### Test Summary:
- Total tests: **34** (was 31)
- `make lint` - 0 issues
- `make test` - All pass

---

### Date: 2026-01-31 (Code Review Round 3 - Comprehensive Fixes)

**Note**: Third round of comprehensive code quality improvements addressing all remaining Gilfoyle-style code review feedback:

#### New Files Created:

1. **internal/copilot/client_test.go** - 6 tests for Copilot client
   - Tests for `New()`, `ProcessMessage()` with various scenarios
   - Context cancellation and deadline tests

2. **internal/handlers/handler.go** - Common Handler interface
   - Defines `Handler` interface with `Start()` and `Stop()` methods
   - Both LINE and Discord handlers implement this interface

3. **internal/handlers/discord/handler_test.go** - 4 tests for Discord handler
   - Tests for `New()`, `Start()`, `Stop()`

4. **internal/handlers/line/handler_test.go** - 7 tests for LINE handler
   - Tests for `New()`, HTTP method validation, `Start()`, `Stop()`
   - Verifies only POST is accepted (GET, PUT, DELETE, PATCH rejected)

5. **internal/systray/systray_test.go** - 6 tests for systray
   - Tests for `New()`, `WithOnReady()`, `WithOnExit()`, `Run()`, `Quit()`

6. **internal/tools/params.go** - Shared parameter helpers
   - `GetRequiredString()` - extracts required string params
   - `GetOptionalString()` - extracts optional string with default
   - `GetOptionalInt()` - extracts optional int (handles float64 from JSON)
   - `GetOptionalBool()` - extracts optional bool with default

7. **internal/tools/params_test.go** - 12 tests for params helpers
   - Comprehensive tests for all helper functions

8. **internal/updater/updater_test.go** - 9 tests for updater
   - Tests for `New()`, `CurrentVersion()`, `IsNewerVersion()`
   - 12 version comparison scenarios including semver edge cases
   - Context cancellation tests

#### Code Fixes:

9. **internal/copilot/client.go**:
   - Added `ErrAPIKeyNotConfigured` sentinel error
   - Context check moved to FIRST operation (fail fast pattern)

10. **internal/config/config.go**:
    - Added Discord reverse validation (GuildID→Token)
    - Now validates both directions symmetrically

11. **internal/config/config_test.go**:
    - Added `TestConfig_Validate_DiscordRequiresToken` test

12. **internal/registry/registry.go**:
    - Added `Unregister(name string) bool` method
    - Allows removing tools from registry

13. **internal/registry/registry_test.go**:
    - Added `TestRegistry_Unregister` and `TestRegistry_Unregister_NotFound`

14. **internal/handlers/line/handler.go**:
    - Added HTTP method validation (only POST allowed)
    - Added `Start()` and `Stop()` methods for Handler interface
    - Returns 405 Method Not Allowed for non-POST requests

15. **internal/updater/updater.go**:
    - Added `CurrentVersion()` getter
    - Added `IsNewerVersion(newVersion string) bool` for semver comparison
    - Implemented pure Go version parsing without external dependencies
    - Handles versions with/without "v" prefix

16. **cmd/orchestrator/main.go**:
    - Added config loading attempt on startup
    - Shows config status or helpful error message
    - Fixed import grouping (goimports compliance)

#### Test Summary:
- Total test functions: **83** (was 34)
- `make lint` - 0 issues
- `make test` - All pass
- `make build` - Successful

---

### Date: ____

**Note**:

---

## Blockers & Issues

<!-- Track any blockers or issues encountered -->

| Issue | Status | Resolution |
|-------|--------|------------|
| | | |

---

## Time Tracking

| Task | Estimated | Actual | Notes |
|------|-----------|--------|-------|
| 0.1 Go Module | 0.5h | | |
| 0.2 Project Structure | 1h | | |
| 0.3 CI/CD | 4h | | |
| 0.4 Dev Tools | 2h | | |
| 0.5 Documentation | 2h | | |
| 0.6 goreleaser | 1h | | |
| **Total** | **10.5h** | | |

---

## References

- [Go Module Documentation](https://go.dev/doc/modules/gomod-ref)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint Configuration](https://golangci-lint.run/usage/configuration/)
- [goreleaser Documentation](https://goreleaser.com/quick-start/)
- [Pre-commit Documentation](https://pre-commit.com/)

---

**Next Phase**: [Phase 1: Core Foundation](./phase-1-foundation.md)
