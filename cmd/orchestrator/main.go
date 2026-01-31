package main

import (
	"fmt"
	"os"
)

// Build-time variables (set by goreleaser)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("macmini-assistant %s (commit: %s, built: %s)\n", version, commit, date)
		return
	}

	fmt.Println("MacMini Assistant Orchestrator")
	fmt.Println("Status: Phase 0 Bootstrap - Under Development")
}
