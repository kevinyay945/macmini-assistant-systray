---
name: cobra
description: Expert guidance for using the spf13/cobra library with Go to build modern CLI applications, including command structure, flags, subcommands, argument validation, and shell completion. Use when building command-line tools, creating CLI commands, working with flags, implementing subcommands, or adding shell completions.
license: Complete terms in LICENSE.txt
---

# Cobra CLI Builder

Expert guidance for using the spf13/cobra library to create powerful CLI applications in Go.

## Quick Start

### Basic Command Structure

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
)

func main() {
    var rootCmd = &cobra.Command{
        Use:   "myapp",
        Short: "A brief description of your application",
        Long:  `A longer description of your application`,
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Hello from myapp!")
        },
    }

    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
```

### Command Fields

- **Use**: One-line usage message (e.g., `"server [flags]"`)
- **Short**: Brief description shown in help
- **Long**: Detailed description shown in command's help
- **Example**: Usage examples
- **Run**: Function to execute (receives `cmd` and `args`)
- **RunE**: Like Run but returns error
- **Aliases**: Alternative names for the command

## Flags

### Local Flags

Available only to the specific command:

```go
var port int
rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port number")
```

### Persistent Flags

Available to command and all subcommands:

```go
var verbose bool
rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
```

### Flag Types

```go
// String flag
cmd.Flags().StringVarP(&name, "name", "n", "default", "Description")

// Int flag
cmd.Flags().IntVarP(&count, "count", "c", 0, "Description")

// Bool flag
cmd.Flags().BoolVarP(&enabled, "enabled", "e", false, "Description")

// StringSlice flag
cmd.Flags().StringSliceVarP(&items, "items", "i", []string{}, "Description")

// Duration flag
cmd.Flags().DurationVarP(&timeout, "timeout", "t", 30*time.Second, "Description")
```

### Required Flags

```go
cmd.MarkFlagRequired("name")
```

### Mutually Exclusive Flags

```go
cmd.MarkFlagsMutuallyExclusive("json", "yaml")
```

### Flag Groups

```go
// At least one flag required
cmd.MarkFlagsOneRequired("config", "interactive")

// All flags must be used together
cmd.MarkFlagsRequiredTogether("username", "password")
```

## Subcommands

### Creating Subcommands

```go
var rootCmd = &cobra.Command{Use: "app"}

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the server",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Server starting...")
    },
}

var fetchCmd = &cobra.Command{
    Use:   "fetch",
    Short: "Fetch data",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Fetching data...")
    },
}

func init() {
    rootCmd.AddCommand(serveCmd)
    rootCmd.AddCommand(fetchCmd)
}
```

### Nested Subcommands

```go
var configCmd = &cobra.Command{Use: "config"}
var configGetCmd = &cobra.Command{Use: "get", Run: /* ... */}
var configSetCmd = &cobra.Command{Use: "set", Run: /* ... */}

configCmd.AddCommand(configGetCmd)
configCmd.AddCommand(configSetCmd)
rootCmd.AddCommand(configCmd)

// Usage: app config get
//        app config set
```

### Command Groups

```go
rootCmd.AddGroup(&cobra.Group{
    ID:    "management",
    Title: "Management Commands:",
})

serveCmd.GroupID = "management"
```

## Argument Validation

### Built-in Validators

```go
// No arguments allowed
cmd.Args = cobra.NoArgs

// Arbitrary arguments (no validation)
cmd.Args = cobra.ArbitraryArgs

// Exactly N arguments
cmd.Args = cobra.ExactArgs(2)

// Minimum N arguments
cmd.Args = cobra.MinimumNArgs(1)

// Maximum N arguments
cmd.Args = cobra.MaximumNArgs(3)

// Range of arguments
cmd.Args = cobra.RangeArgs(1, 3)
```

### Valid Arguments

```go
cmd.ValidArgs = []cobra.Completion{"pod", "service", "deployment"}
cmd.Args = cobra.OnlyValidArgs
```

### Combining Validators

```go
cmd.Args = cobra.MatchAll(
    cobra.ExactArgs(2),
    cobra.OnlyValidArgs,
)
```

### Custom Validation

```go
cmd.Args = func(cmd *cobra.Command, args []string) error {
    if len(args) < 1 {
        return fmt.Errorf("requires at least one arg")
    }
    if !isValidColor(args[0]) {
        return fmt.Errorf("invalid color: %s", args[0])
    }
    return nil
}
```

## Run Hooks

Execute functions at different stages:

```go
cmd := &cobra.Command{
    // Runs first for this command and all children
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        // Setup logging, load config, etc.
    },
    
    // Runs before Run, only for this command
    PreRun: func(cmd *cobra.Command, args []string) {
        // Command-specific setup
    },
    
    // Main command logic
    Run: func(cmd *cobra.Command, args []string) {
        // Do the work
    },
    
    // Runs after Run, only for this command
    PostRun: func(cmd *cobra.Command, args []string) {
        // Command-specific cleanup
    },
    
    // Runs last for this command and all children
    PersistentPostRun: func(cmd *cobra.Command, args []string) {
        // Cleanup, close connections, etc.
    },
}
```

**Execution order:**
1. PersistentPreRun (parent → child)
2. PreRun
3. Run
4. PostRun
5. PersistentPostRun (child → parent)

**Use `*RunE` variants to return errors:**

```go
cmd.RunE = func(cmd *cobra.Command, args []string) error {
    if err := doWork(); err != nil {
        return err
    }
    return nil
}
```

## Shell Completions

### Enable Completions

```go
// Add completion command automatically
rootCmd.CompletionOptions.DisableDefaultCmd = false

// Generate completions: app completion bash
```

### Flag Completions

```go
// Fixed choices
cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"json", "yaml", "table"}, cobra.ShellCompDirectiveNoFileComp
})

// File completions
cmd.MarkFlagFilename("config", "yaml", "yml")

// Directory completions
cmd.MarkFlagDirname("output-dir")
```

### Argument Completions

```go
cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
    // Dynamic completions
    items := []string{"pod", "service", "deployment"}
    completions := make([]cobra.Completion, len(items))
    for i, item := range items {
        completions[i] = cobra.Completion(item)
    }
    return completions, cobra.ShellCompDirectiveNoFileComp
}
```

### Completions with Descriptions

```go
return []cobra.Completion{
    cobra.CompletionWithDesc("json", "JSON format"),
    cobra.CompletionWithDesc("yaml", "YAML format"),
    cobra.CompletionWithDesc("table", "Table format"),
}, cobra.ShellCompDirectiveNoFileComp
```

### Shell Directives

- `ShellCompDirectiveNoFileComp`: Don't suggest files
- `ShellCompDirectiveNoSpace`: Don't add space after completion
- `ShellCompDirectiveFilterFileExt`: Filter files by extension
- `ShellCompDirectiveFilterDirs`: Only show directories
- `ShellCompDirectiveKeepOrder`: Preserve completion order
- `ShellCompDirectiveError`: Error occurred, ignore completions

## Error Handling

### Error Modes

```go
// Don't show usage on error
cmd.SilenceUsage = true

// Don't print errors (handle manually)
cmd.SilenceErrors = true
```

### Custom Error Messages

```go
cmd.RunE = func(cmd *cobra.Command, args []string) error {
    return fmt.Errorf("operation failed: %w", err)
}

if err := rootCmd.Execute(); err != nil {
    // Custom error handling
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

## Advanced Features

### Version Flag

```go
rootCmd.Version = "1.0.0"
// Automatically adds --version and -v flags
```

### Help Customization

```go
cmd.SetHelpTemplate(customTemplate)
cmd.SetUsageTemplate(customUsageTemplate)

// Or override help function
cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
    // Custom help output
})
```

### Annotations

```go
cmd.Annotations = map[string]string{
    "category": "management",
    "priority": "high",
}
```

### PreRun Configuration Pattern

```go
var config Config

rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
    // Load configuration once for all commands
    cfg, err := loadConfig()
    if err != nil {
        return err
    }
    config = cfg
    return nil
}
```

### Accessing Parent Flags

```go
// In a subcommand
verbose, _ := cmd.Flags().GetBool("verbose")

// Or get from parent
verboseFlag := cmd.InheritedFlags().Lookup("verbose")
```

## Best Practices

1. **Use RunE for error handling**: Return errors instead of calling `os.Exit()` in Run functions
2. **Keep commands focused**: Each command should do one thing well
3. **Use persistent flags wisely**: Only make flags persistent if all subcommands need them
4. **Validate early**: Use Args validators and PreRun for validation
5. **Provide examples**: Set the Example field with realistic usage
6. **Add completions**: Enhance UX with shell completions for flags and args
7. **Group related commands**: Use command groups for better help organization
8. **Use Short descriptions**: Keep Short field brief (under 60 chars)
9. **Leverage hooks**: Use PreRun for setup, PostRun for cleanup
10. **Handle signals gracefully**: Use context and signal handling for long-running commands

## Common Patterns

### Config File Integration

```go
import "github.com/spf13/viper"

var cfgFile string

rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")

rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
        return viper.ReadInConfig()
    }
    return nil
}
```

### Output Format Flag

```go
var outputFormat string

cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (json|yaml|table)")
cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"json", "yaml", "table"}, cobra.ShellCompDirectiveNoFileComp
})
```

### Interactive vs Non-Interactive

```go
var nonInteractive bool

cmd.Flags().BoolVar(&nonInteractive, "yes", false, "Skip confirmations")

cmd.RunE = func(cmd *cobra.Command, args []string) error {
    if !nonInteractive {
        // Prompt user for confirmation
        if !confirm("Continue?") {
            return nil
        }
    }
    // Proceed with operation
    return doWork()
}
```

### Context with Cancellation

```go
cmd.RunE = func(cmd *cobra.Command, args []string) error {
    ctx, cancel := context.WithCancel(cmd.Context())
    defer cancel()
    
    // Handle interrupts
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
        <-sigCh
        cancel()
    }()
    
    return doWorkWithContext(ctx)
}
```

## Additional Resources

- **Complete API Reference**: See `references/cobra-api.txt` for full package documentation
- **Comprehensive User Guide**: See `references/user-guide.md` for detailed guides, advanced topics, and more examples
