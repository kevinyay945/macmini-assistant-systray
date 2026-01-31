.PHONY: all build test test-local test-integration test-all test-coverage lint clean run help init

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
	@rm -rf bin/ coverage.out coverage.html dist/

# Run the application
run:
	@go run cmd/orchestrator/main.go

# Initialize development environment
init:
	@echo "Installing development dependencies..."
	@command -v golangci-lint > /dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@command -v goreleaser > /dev/null 2>&1 || go install github.com/goreleaser/goreleaser@latest
	@command -v pre-commit > /dev/null 2>&1 || (echo "Installing pre-commit..." && brew install pre-commit)
	@[ -f .git/hooks/pre-commit ] || pre-commit install
	@go mod tidy
	@echo "Development environment ready!"

# Show help
help:
	@echo "Available targets:"
	@echo "  all              - lint, test, and build"
	@echo "  build            - Build the application"
	@echo "  test             - Run standard tests (CI-safe)"
	@echo "  test-local       - Run tests including local-only"
	@echo "  test-integration - Run integration tests"
	@echo "  test-all         - Run all tests"
	@echo "  test-coverage    - Generate coverage report"
	@echo "  lint             - Run linter"
	@echo "  clean            - Remove build artifacts"
	@echo "  run              - Run the application"
	@echo "  init             - Initialize development environment"
	@echo "  help             - Show this help"
