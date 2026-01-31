// Package config handles application configuration loading and validation.
package config

// Config represents the application configuration loaded from config.yaml.
type Config struct {
	// Server configuration
	Server ServerConfig `yaml:"server"`
	// LINE bot configuration
	LINE LINEConfig `yaml:"line"`
	// Discord bot configuration
	Discord DiscordConfig `yaml:"discord"`
	// Tool configurations
	Tools ToolsConfig `yaml:"tools"`
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
