# Development Plan
# MacMini Assistant Systray Orchestrator

**Version:** 1.0
**Date:** January 31, 2026
**Methodology:** Test-Driven Development (TDD)

---

## Overview

This document provides a high-level overview of the phased development approach. Each phase has its own detailed document in the [phases/](./phases/) folder where you can add more details as you work through implementation.

### Development Principles

1. **Test-First Development**: Write tests before implementation
2. **Incremental Delivery**: Each phase produces a working deliverable
3. **Continuous Integration**: All commits pass automated tests (excluding local-only tests)
4. **Code Review**: All changes reviewed before merge
5. **Documentation**: Update docs alongside code

---

## Phase Documents

Each phase has a dedicated document with detailed tasks, test cases, and implementation notes:

| Phase | Duration | Document | Focus |
|-------|----------|----------|-------|
| **Phase 0** | Week 1 | [phase-0-bootstrap.md](./phases/phase-0-bootstrap.md) | Project Bootstrap |
| **Phase 1** | Weeks 2-3 | [phase-1-foundation.md](./phases/phase-1-foundation.md) | Core Foundation |
| **Phase 2** | Weeks 4-5 | [phase-2-messaging.md](./phases/phase-2-messaging.md) | Messaging Platforms |
| **Phase 3** | Week 6 | [phase-3-copilot.md](./phases/phase-3-copilot.md) | Copilot SDK |
| **Phase 4** | Weeks 7-8 | [phase-4-tools.md](./phases/phase-4-tools.md) | Tool Implementation |
| **Phase 5** | Week 9 | [phase-5-systray.md](./phases/phase-5-systray.md) | System Tray |
| **Phase 6** | Week 10 | [phase-6-updater.md](./phases/phase-6-updater.md) | Auto-updater |
| **Phase 7** | Week 11 | [phase-7-integration.md](./phases/phase-7-integration.md) | Integration Testing |
| **Phase 8** | Week 12 | [phase-8-release.md](./phases/phase-8-release.md) | Release |

---

## Quick Testing Reference

```bash
# Run all tests except local-only
make test

# Run with coverage
make test-coverage

# Local testing with all tags
make test-local
```

---

## Development Workflow

### Daily Workflow
```bash
# 1. Pull latest changes
git pull origin main

# 2. Create feature branch
git checkout -b feature/tool-registry

# 3. Write tests first (TDD)
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

```

---

## Testing Commands

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
