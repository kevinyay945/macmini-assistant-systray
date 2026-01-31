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
