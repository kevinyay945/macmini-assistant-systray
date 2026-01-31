package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Build-time variables (set by goreleaser)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "orchestrator",
		Short: "MacMini Assistant Orchestrator",
		Long: `AI-powered tool execution orchestrator for macOS.

This application provides remote task automation through LINE and Discord
messaging platforms, powered by GitHub Copilot SDK.`,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("MacMini Assistant Orchestrator")
			fmt.Println("Status: Phase 0 Bootstrap - Under Development")
			fmt.Println("")
			fmt.Println("Use --help to see available commands.")
		},
	}

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
		os.Exit(1)
	}
}
