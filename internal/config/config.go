// Package config handles application configuration loading and validation.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultConfigPath returns the default configuration file path.
func DefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".macmini-assistant", "config.yaml"), nil
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

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	cfg.applyDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// DefaultServerPort is the default HTTP server port.
const DefaultServerPort = 8080

// Config represents the application configuration loaded from config.yaml.
type Config struct {
	// Server configuration
	Server ServerConfig `yaml:"server"`
	// LINE bot configuration
	LINE LINEConfig `yaml:"line"`
	// Discord bot configuration
	Discord DiscordConfig `yaml:"discord"`
	// Copilot SDK configuration
	Copilot CopilotConfig `yaml:"copilot"`
	// Tool configurations
	Tools ToolsConfig `yaml:"tools"`
}

// CopilotConfig holds GitHub Copilot SDK settings.
type CopilotConfig struct {
	APIKey  string `yaml:"api_key"`
	Timeout int    `yaml:"timeout"` // Timeout in seconds, default 600 (10 minutes)
}

// applyDefaults sets default values for unset configuration options.
func (c *Config) applyDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = DefaultServerPort
	}
	if c.Copilot.Timeout == 0 {
		c.Copilot.Timeout = 600 // 10 minutes default
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	var errs []error

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Errorf("server.port must be between 1 and 65535, got %d", c.Server.Port))
	}

	// LINE credentials require both secret and token
	if c.LINE.ChannelSecret != "" && c.LINE.ChannelToken == "" {
		errs = append(errs, errors.New("line.channel_token is required when line.channel_secret is set"))
	}
	if c.LINE.ChannelToken != "" && c.LINE.ChannelSecret == "" {
		errs = append(errs, errors.New("line.channel_secret is required when line.channel_token is set"))
	}

	if c.Discord.Token != "" && c.Discord.GuildID == "" {
		errs = append(errs, errors.New("discord.guild_id is required when discord.token is set"))
	}
	if c.Discord.GuildID != "" && c.Discord.Token == "" {
		errs = append(errs, errors.New("discord.token is required when discord.guild_id is set"))
	}

	if c.Tools.GoogleDrive.Enabled {
		if c.Tools.GoogleDrive.CredentialsPath == "" && c.Tools.GoogleDrive.ServiceAccountPath == "" {
			errs = append(errs, errors.New("google_drive requires either credentials_path or service_account_path"))
		}
	}

	return errors.Join(errs...)
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port int `yaml:"port"`
}

// LINEConfig holds LINE bot credentials.
type LINEConfig struct {
	ChannelSecret string `yaml:"channel_secret"`
	ChannelToken  string `yaml:"channel_token"`
}

// DiscordConfig holds Discord bot credentials.
type DiscordConfig struct {
	Token   string `yaml:"token"`
	GuildID string `yaml:"guild_id"`
}

// ToolsConfig holds tool-specific configurations.
type ToolsConfig struct {
	Downie      DownieConfig      `yaml:"downie"`
	GoogleDrive GoogleDriveConfig `yaml:"google_drive"`
}

// DownieConfig holds Downie tool settings.
type DownieConfig struct {
	Enabled bool `yaml:"enabled"`
}

// GoogleDriveConfig holds Google Drive tool settings.
type GoogleDriveConfig struct {
	Enabled            bool   `yaml:"enabled"`
	CredentialsPath    string `yaml:"credentials_path"`
	ServiceAccountPath string `yaml:"service_account_path"`
}
