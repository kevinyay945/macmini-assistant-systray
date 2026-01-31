// Package config handles application configuration loading and validation.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultServerPort is the default HTTP server port.
const DefaultServerPort = 8080

// envVarPattern is a pre-compiled regex for environment variable substitution.
// Defined at package level to avoid recompilation on every call to expandEnvVars.
var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// DefaultConfigPath returns the default configuration file path.
func DefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".macmini-assistant", "config.yaml"), nil
}

// DefaultDownloadFolder returns the default download folder path.
func DefaultDownloadFolder() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, "Downloads", "macmini-assistant"), nil
}

// Load reads configuration from the specified path or default location.
func Load(path string) (*Config, error) {
	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	// Expand environment variables before parsing
	expanded := expandEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// expandEnvVars replaces ${VAR_NAME} and ${VAR_NAME:-default} patterns with environment variable values.
// Supports default values using the syntax ${VAR:-default_value}.
// Uses os.LookupEnv to distinguish between "not set" and "set to empty string".
// NOTE: Nested variable substitution (e.g., ${VAR1:-${VAR2}}) is NOT supported.
func expandEnvVars(content string) string {
	return envVarPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Use FindStringSubmatch to get the capture group directly
		// instead of TrimPrefix/TrimSuffix for better performance
		matches := envVarPattern.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		inner := matches[1]
		// Support ${VAR:-default} syntax
		if idx := strings.Index(inner, ":-"); idx != -1 {
			varName := inner[:idx]
			defaultVal := inner[idx+2:]
			if val, ok := os.LookupEnv(varName); ok {
				return val
			}
			return defaultVal
		}
		if val, ok := os.LookupEnv(inner); ok {
			return val
		}
		return ""
	})
}

// GenerateDefault creates a default configuration.
func GenerateDefault() (*Config, error) {
	downloadFolder, err := DefaultDownloadFolder()
	if err != nil {
		return nil, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return &Config{
		App: AppConfig{
			DownloadFolder: downloadFolder,
			AutoStart:      true,
			AutoUpdate:     true,
			LogLevel:       "info",
		},
		Copilot: CopilotConfig{
			APIKey:         "${GITHUB_COPILOT_API_KEY}",
			TimeoutSeconds: 600,
		},
		LINE: LINEConfig{
			ChannelSecret: "${LINE_CHANNEL_SECRET}",
			ChannelToken:  "${LINE_ACCESS_TOKEN}",
			WebhookPort:   8080,
		},
		Discord: DiscordConfig{
			Token:               "${DISCORD_BOT_TOKEN}",
			StatusChannelID:     "",
			EnableSlashCommands: true,
		},
		Tools: []ToolConfig{
			{
				Name:    "youtube_download",
				Type:    "downie",
				Enabled: true,
				Config: map[string]interface{}{
					"deep_link_scheme":   "downie://",
					"default_format":     "mp4",
					"default_resolution": "1080p",
				},
			},
			{
				Name:    "gdrive_upload",
				Type:    "google_drive",
				Enabled: true,
				Config: map[string]interface{}{
					"credentials_path": filepath.Join(homeDir, ".macmini-assistant", "gdrive-creds.json"),
					"default_timeout":  300,
				},
			},
		},
		Updater: UpdaterConfig{
			GitHubRepo:         "username/macmini-assistant",
			CheckIntervalHours: 6,
			Enabled:            true,
		},
	}, nil
}

// WriteDefaultConfig generates and writes a default config file to the given path.
func WriteDefaultConfig(path string) error {
	cfg, err := GenerateDefault()
	if err != nil {
		return fmt.Errorf("failed to generate default config: %w", err)
	}

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comment
	header := "# MacMini Assistant Configuration\n# See docs/phases/phase-1-foundation.md for full schema documentation\n\n"

	if err := os.WriteFile(path, []byte(header+string(data)), 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Config represents the application configuration loaded from config.yaml.
type Config struct {
	App     AppConfig     `yaml:"app"`
	Copilot CopilotConfig `yaml:"copilot"`
	LINE    LINEConfig    `yaml:"line"`
	Discord DiscordConfig `yaml:"discord"`
	Tools   []ToolConfig  `yaml:"tools"`
	Updater UpdaterConfig `yaml:"updater"`
}

// AppConfig holds general application settings.
type AppConfig struct {
	DownloadFolder string `yaml:"download_folder"`
	AutoStart      bool   `yaml:"auto_start"`
	AutoUpdate     bool   `yaml:"auto_update"`
	LogLevel       string `yaml:"log_level"` // debug, info, warn, error
}

// CopilotConfig holds GitHub Copilot SDK settings.
type CopilotConfig struct {
	APIKey         string `yaml:"api_key"`
	TimeoutSeconds int    `yaml:"timeout_seconds"` // Timeout in seconds, default 600 (10 minutes)
}

// LINEConfig holds LINE bot credentials.
type LINEConfig struct {
	ChannelSecret string `yaml:"channel_secret"`
	ChannelToken  string `yaml:"channel_token"`
	WebhookPort   int    `yaml:"webhook_port"`
}

// DiscordConfig holds Discord bot credentials.
type DiscordConfig struct {
	Token               string `yaml:"bot_token"`
	StatusChannelID     string `yaml:"status_channel_id"`
	EnableSlashCommands bool   `yaml:"enable_slash_commands"`
}

// ToolConfig represents a single tool configuration.
type ToolConfig struct {
	Name    string                 `yaml:"name"`
	Type    string                 `yaml:"type"` // downie, google_drive, etc.
	Enabled bool                   `yaml:"enabled"`
	Config  map[string]interface{} `yaml:"config"`
}

// UpdaterConfig holds auto-updater settings.
type UpdaterConfig struct {
	GitHubRepo         string `yaml:"github_repo"`
	CheckIntervalHours int    `yaml:"check_interval_hours"`
	Enabled            bool   `yaml:"enabled"`
}

// applyDefaults sets default values for unset configuration options.
func (c *Config) applyDefaults() {
	if c.App.LogLevel == "" {
		c.App.LogLevel = "info"
	}
	if c.App.DownloadFolder == "" {
		if folder, err := DefaultDownloadFolder(); err == nil {
			c.App.DownloadFolder = folder
		} else {
			c.App.DownloadFolder = "/tmp/downloads"
		}
	}
	if c.Copilot.TimeoutSeconds == 0 {
		c.Copilot.TimeoutSeconds = 600 // 10 minutes default
	}
	if c.LINE.WebhookPort == 0 {
		c.LINE.WebhookPort = DefaultServerPort
	}
	if c.Updater.CheckIntervalHours == 0 {
		c.Updater.CheckIntervalHours = 6
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	var errs []error

	// Validate log level
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[c.App.LogLevel] {
		errs = append(errs, fmt.Errorf("app.log_level must be one of debug, info, warn, error; got %q", c.App.LogLevel))
	}

	// Validate webhook port
	if c.LINE.WebhookPort < 1 || c.LINE.WebhookPort > 65535 {
		errs = append(errs, fmt.Errorf("line.webhook_port must be between 1 and 65535, got %d", c.LINE.WebhookPort))
	}

	// LINE credentials require both secret and token
	if c.LINE.ChannelSecret != "" && c.LINE.ChannelToken == "" {
		errs = append(errs, errors.New("line.channel_token is required when line.channel_secret is set"))
	}
	if c.LINE.ChannelToken != "" && c.LINE.ChannelSecret == "" {
		errs = append(errs, errors.New("line.channel_secret is required when line.channel_token is set"))
	}

	// Validate download folder is accessible (or can be created)
	if c.App.DownloadFolder != "" {
		if info, err := os.Stat(c.App.DownloadFolder); err == nil {
			if !info.IsDir() {
				errs = append(errs, fmt.Errorf("app.download_folder %q exists but is not a directory", c.App.DownloadFolder))
			}
		}
		// It's okay if the folder doesn't exist - we'll create it when needed
	}

	// Validate tool configurations
	toolNames := make(map[string]bool)
	for i, tool := range c.Tools {
		switch {
		case tool.Name == "":
			errs = append(errs, fmt.Errorf("tools[%d].name is required", i))
		case toolNames[tool.Name]:
			errs = append(errs, fmt.Errorf("duplicate tool name: %q", tool.Name))
		default:
			toolNames[tool.Name] = true
		}

		if tool.Type == "" {
			errs = append(errs, fmt.Errorf("tools[%d].type is required", i))
		}

		// Validate google_drive tools have credentials
		if tool.Type == "google_drive" && tool.Enabled {
			if tool.Config == nil {
				errs = append(errs, fmt.Errorf("tool %q requires config section", tool.Name))
				continue
			}
			credPath, _ := tool.Config["credentials_path"].(string)
			svcPath, _ := tool.Config["service_account_path"].(string)
			if credPath == "" && svcPath == "" {
				errs = append(errs, fmt.Errorf("tool %q requires credentials_path or service_account_path", tool.Name))
			}
		}
	}

	// Validate Copilot timeout
	if c.Copilot.TimeoutSeconds < 0 {
		errs = append(errs, errors.New("copilot.timeout_seconds cannot be negative"))
	}
	const maxTimeoutSeconds = 3600 // 1 hour maximum
	if c.Copilot.TimeoutSeconds > maxTimeoutSeconds {
		errs = append(errs, fmt.Errorf("copilot.timeout_seconds exceeds maximum (%d), got %d", maxTimeoutSeconds, c.Copilot.TimeoutSeconds))
	}

	// Validate updater config
	if c.Updater.Enabled && c.Updater.GitHubRepo == "" {
		errs = append(errs, errors.New("updater.github_repo is required when updater is enabled"))
	}

	return errors.Join(errs...)
}

// deepCopyMap recursively copies a map[string]interface{} to prevent shared mutations.
// Handles nested maps and slices. Other types are copied by value.
func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		switch typed := v.(type) {
		case map[string]interface{}:
			result[k] = deepCopyMap(typed)
		case []interface{}:
			cp := make([]interface{}, len(typed))
			for i, item := range typed {
				if nested, ok := item.(map[string]interface{}); ok {
					cp[i] = deepCopyMap(nested)
				} else {
					cp[i] = item
				}
			}
			result[k] = cp
		default:
			result[k] = v
		}
	}
	return result
}

// GetToolConfig returns a copy of the configuration for a specific tool by name.
// Returns a deep copy to prevent callers from accidentally modifying the original config
// or holding a dangling pointer if the config's Tools slice is reallocated.
// The Config map is recursively deep-copied, including nested maps and slices.
func (c *Config) GetToolConfig(name string) (ToolConfig, bool) {
	for _, tool := range c.Tools {
		if tool.Name == name {
			// Deep copy the tool config including the Config map
			toolCopy := ToolConfig{
				Name:    tool.Name,
				Type:    tool.Type,
				Enabled: tool.Enabled,
				Config:  deepCopyMap(tool.Config),
			}
			return toolCopy, true
		}
	}
	return ToolConfig{}, false
}

// GetEnabledTools returns only the enabled tool configurations.
func (c *Config) GetEnabledTools() []ToolConfig {
	// Count enabled tools first for pre-allocation
	count := 0
	for _, tool := range c.Tools {
		if tool.Enabled {
			count++
		}
	}

	enabled := make([]ToolConfig, 0, count)
	for _, tool := range c.Tools {
		if tool.Enabled {
			enabled = append(enabled, tool)
		}
	}
	return enabled
}
