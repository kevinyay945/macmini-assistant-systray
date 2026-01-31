# MacMini Assistant Systray

A Go-based macOS system tray application that orchestrates AI-powered tool execution via GitHub Copilot SDK, accessible through LINE and Discord messaging platforms.

## Features

- **System Tray Integration**: Native macOS menu bar application
- **Multi-Platform Messaging**: Support for LINE and Discord bots
- **AI-Powered**: GitHub Copilot SDK integration for intelligent tool selection
- **Extensible Tools**: Video downloads via Downie, Google Drive uploads
- **Self-Updating**: Automatic updates from GitHub releases

## Requirements

- macOS (ARM64 / Apple Silicon)
- Go 1.22+
- [Downie](https://software.charliemonroe.net/downie/) (for video downloads)

## Quick Start

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/kevinyay945/macmini-assistant-systray.git
cd macmini-assistant-systray

# Initialize development environment
make init

# Run tests
make test

# Build
make build

# Run
make run
```

### Configuration

Copy the sample configuration:

```bash
mkdir -p ~/.macmini-assistant
cp test/fixtures/config.sample.yaml ~/.macmini-assistant/config.yaml
```

Edit `~/.macmini-assistant/config.yaml` with your credentials.

## Development

### Available Commands

| Command | Description |
|---------|-------------|
| `make all` | Lint, test, and build |
| `make build` | Build the application |
| `make test` | Run standard tests (CI-safe) |
| `make test-local` | Run tests including local-only |
| `make test-integration` | Run integration tests |
| `make test-all` | Run all tests |
| `make test-coverage` | Generate coverage report |
| `make lint` | Run linter |
| `make clean` | Remove build artifacts |
| `make run` | Run the application |
| `make init` | Initialize development environment |

### Testing Strategy

Tests are organized with build tags:

- **Standard tests**: Run in CI, no external dependencies
- **Local tests** (`-tags=local`): Require macOS tools like Downie
- **Integration tests** (`-tags=integration`): Require external services

### Project Structure

```
macmini-assistant-systray/
├── cmd/orchestrator/          # Application entry point
├── internal/
│   ├── config/               # Configuration handling
│   ├── registry/             # Tool registry
│   ├── copilot/              # Copilot SDK integration
│   ├── handlers/
│   │   ├── line/             # LINE bot handler
│   │   └── discord/          # Discord bot handler
│   ├── tools/
│   │   ├── downie/           # Downie video download
│   │   └── gdrive/           # Google Drive upload
│   ├── systray/              # System tray UI
│   ├── updater/              # Self-update functionality
│   └── observability/        # Logging and metrics
├── test/
│   ├── integration/          # Integration tests
│   └── fixtures/             # Test fixtures
└── docs/                     # Documentation
```

## Documentation

- [PRD](docs/PRD.md) - Product Requirements Document
- [Development Plan](docs/DEVELOPMENT_PLAN.md) - Phase breakdown
- [Phase Documents](docs/phases/) - Detailed implementation tasks

## License

MIT
