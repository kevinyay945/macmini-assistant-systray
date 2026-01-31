package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kevinyay945/macmini-assistant-systray/internal/config"
	"github.com/kevinyay945/macmini-assistant-systray/internal/observability"
)

// Build-time variables (set by goreleaser)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(run())
}

// run executes the main application logic and returns an exit code.
func run() int {
	// Set up context with signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	rootCmd := &cobra.Command{
		Use:   "orchestrator",
		Short: "MacMini Assistant Orchestrator",
		Long: `AI-powered tool execution orchestrator for macOS.

This application provides remote task automation through LINE and Discord
messaging platforms, powered by GitHub Copilot SDK.`,
		Run: func(cmd *cobra.Command, _ []string) {
			runOrchestrator(cmd.Context())
		},
	}

	// Inject context into cobra command
	rootCmd.SetContext(ctx)

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("macmini-assistant %s (commit: %s, built: %s)\n", version, commit, date)
		},
	}

	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// runOrchestrator starts the main application loop with context support.
func runOrchestrator(ctx context.Context) {
	// Initialize logger
	logger := observability.New(
		observability.WithLevel(observability.LevelInfo),
	)

	logger.Info(ctx, "MacMini Assistant Orchestrator starting",
		"version", version,
		"commit", commit,
		"status", "Phase 0 Bootstrap - Under Development",
	)

	// Attempt to load configuration
	cfg, err := config.Load("")
	if err != nil {
		logger.Warn(ctx, "could not load config, using defaults",
			"error", err,
			"hint", "Create ~/.macmini-assistant/config.yaml to configure the application",
		)
	} else {
		logger.Info(ctx, "configuration loaded successfully",
			"webhook_port", cfg.LINE.WebhookPort,
			"copilot_timeout", cfg.Copilot.TimeoutSeconds,
			"log_level", cfg.App.LogLevel,
		)
	}

	logger.Info(ctx, "Use --help to see available commands. Press Ctrl+C to exit.")

	// Wait for context cancellation (signal received)
	<-ctx.Done()
	logger.Info(ctx, "Shutting down gracefully...")
}
