package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/config"
)

func TestConfig_Validate_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() returned error for valid config: %v", err)
	}
}

func TestConfig_Validate_InvalidPort(t *testing.T) {
	testCases := []struct {
		name string
		port int
	}{
		{"zero port", 0},
		{"negative port", -1},
		{"port too high", 65536},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				Server: config.ServerConfig{Port: tc.port},
			}

			if err := cfg.Validate(); err == nil {
				t.Error("Validate() should return error for invalid port")
			}
		})
	}
}

func TestConfig_Validate_LINERequiresToken(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		LINE: config.LINEConfig{
			ChannelSecret: "secret",
			ChannelToken:  "",
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error when LINE secret is set but token is missing")
	}
}

func TestConfig_Validate_LINERequiresSecret(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		LINE: config.LINEConfig{
			ChannelSecret: "",
			ChannelToken:  "token",
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error when LINE token is set but secret is missing")
	}
}

func TestConfig_Validate_DiscordRequiresGuildID(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		Discord: config.DiscordConfig{
			Token:   "token",
			GuildID: "",
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error when Discord token is set but guild_id is missing")
	}
}

func TestConfig_Validate_GoogleDriveRequiresCredentials(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		Tools: config.ToolsConfig{
			GoogleDrive: config.GoogleDriveConfig{
				Enabled:            true,
				CredentialsPath:    "",
				ServiceAccountPath: "",
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error when Google Drive is enabled but no credentials")
	}
}

func TestConfig_Load_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Load() should return error for nonexistent file")
	}
}

func TestConfig_Load_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
server:
  port: 8080
line:
  channel_secret: "test-secret"
  channel_token: "test-token"
discord:
  token: "test-token"
  guild_id: "123456789"
tools:
  downie:
    enabled: true
  google_drive:
    enabled: false
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.LINE.ChannelSecret != "test-secret" {
		t.Errorf("LINE.ChannelSecret = %q, want %q", cfg.LINE.ChannelSecret, "test-secret")
	}
	if cfg.Discord.GuildID != "123456789" {
		t.Errorf("Discord.GuildID = %q, want %q", cfg.Discord.GuildID, "123456789")
	}
}

func TestConfig_Load_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
server:
  port: "not a number"
  invalid yaml here
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	_, err := config.Load(configPath)
	if err == nil {
		t.Error("Load() should return error for invalid YAML")
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path, err := config.DefaultConfigPath()
	if err != nil {
		t.Fatalf("DefaultConfigPath() returned error: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Error("DefaultConfigPath() should return absolute path")
	}

	if filepath.Base(path) != "config.yaml" {
		t.Errorf("DefaultConfigPath() path should end with config.yaml, got %s", path)
	}
}
