package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kevinyay945/macmini-assistant-systray/internal/config"
)

func TestConfig_Validate_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			LogLevel:       "info",
			DownloadFolder: "/tmp/test",
		},
		LINE: config.LINEConfig{WebhookPort: 8080},
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() returned error for valid config: %v", err)
	}
}

func TestConfig_Validate_InvalidLogLevel(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			LogLevel: "invalid",
		},
		LINE: config.LINEConfig{WebhookPort: 8080},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error for invalid log level")
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
				App:  config.AppConfig{LogLevel: "info"},
				LINE: config.LINEConfig{WebhookPort: tc.port},
			}

			if err := cfg.Validate(); err == nil {
				t.Error("Validate() should return error for invalid port")
			}
		})
	}
}

func TestConfig_Validate_LINERequiresToken(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{LogLevel: "info"},
		LINE: config.LINEConfig{
			ChannelSecret: "secret",
			ChannelToken:  "",
			WebhookPort:   8080,
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error when LINE secret is set but token is missing")
	}
}

func TestConfig_Validate_LINERequiresSecret(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{LogLevel: "info"},
		LINE: config.LINEConfig{
			ChannelSecret: "",
			ChannelToken:  "token",
			WebhookPort:   8080,
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error when LINE token is set but secret is missing")
	}
}

func TestConfig_Validate_ToolsRequireName(t *testing.T) {
	cfg := &config.Config{
		App:  config.AppConfig{LogLevel: "info"},
		LINE: config.LINEConfig{WebhookPort: 8080},
		Tools: []config.ToolConfig{
			{Name: "", Type: "downie", Enabled: true},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error when tool name is missing")
	}
}

func TestConfig_Validate_ToolsRequireType(t *testing.T) {
	cfg := &config.Config{
		App:  config.AppConfig{LogLevel: "info"},
		LINE: config.LINEConfig{WebhookPort: 8080},
		Tools: []config.ToolConfig{
			{Name: "test_tool", Type: "", Enabled: true},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error when tool type is missing")
	}
}

func TestConfig_Validate_DuplicateToolNames(t *testing.T) {
	cfg := &config.Config{
		App:  config.AppConfig{LogLevel: "info"},
		LINE: config.LINEConfig{WebhookPort: 8080},
		Tools: []config.ToolConfig{
			{Name: "duplicate", Type: "downie", Enabled: true},
			{Name: "duplicate", Type: "downie", Enabled: true},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error for duplicate tool names")
	}
}

func TestConfig_Validate_GoogleDriveRequiresCredentials(t *testing.T) {
	cfg := &config.Config{
		App:  config.AppConfig{LogLevel: "info"},
		LINE: config.LINEConfig{WebhookPort: 8080},
		Tools: []config.ToolConfig{
			{
				Name:    "gdrive",
				Type:    "google_drive",
				Enabled: true,
				Config:  map[string]interface{}{},
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error when Google Drive is enabled but no credentials")
	}
}

func TestConfig_Validate_GoogleDriveNilConfig(t *testing.T) {
	cfg := &config.Config{
		App:  config.AppConfig{LogLevel: "info"},
		LINE: config.LINEConfig{WebhookPort: 8080},
		Tools: []config.ToolConfig{
			{
				Name:    "gdrive",
				Type:    "google_drive",
				Enabled: true,
				Config:  nil, // nil config should also error
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() should return error when Google Drive config is nil")
	}
}

func TestConfig_Validate_UpdaterRequiresRepo(t *testing.T) {
	cfg := &config.Config{
		App:     config.AppConfig{LogLevel: "info"},
		LINE:    config.LINEConfig{WebhookPort: 8080},
		Updater: config.UpdaterConfig{Enabled: true, GitHubRepo: ""},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error when updater is enabled but repo is missing")
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
app:
  download_folder: /tmp/downloads
  auto_start: true
  auto_update: false
  log_level: debug
copilot:
  api_key: "test-api-key"
  timeout_seconds: 300
line:
  channel_secret: "test-secret"
  channel_token: "test-token"
  webhook_port: 9000
discord:
  bot_token: "test-discord-token"
  status_channel_id: "123456789"
  enable_slash_commands: true
tools:
  - name: youtube_download
    type: downie
    enabled: true
    config:
      deep_link_scheme: "downie://"
updater:
  github_repo: "test/repo"
  check_interval_hours: 12
  enabled: true
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.App.LogLevel != "debug" {
		t.Errorf("App.LogLevel = %q, want %q", cfg.App.LogLevel, "debug")
	}
	if cfg.App.DownloadFolder != "/tmp/downloads" {
		t.Errorf("App.DownloadFolder = %q, want %q", cfg.App.DownloadFolder, "/tmp/downloads")
	}
	if cfg.LINE.WebhookPort != 9000 {
		t.Errorf("LINE.WebhookPort = %d, want 9000", cfg.LINE.WebhookPort)
	}
	if cfg.Copilot.APIKey != "test-api-key" {
		t.Errorf("Copilot.APIKey = %q, want %q", cfg.Copilot.APIKey, "test-api-key")
	}
	if cfg.Copilot.TimeoutSeconds != 300 {
		t.Errorf("Copilot.TimeoutSeconds = %d, want 300", cfg.Copilot.TimeoutSeconds)
	}
	if len(cfg.Tools) != 1 {
		t.Errorf("len(Tools) = %d, want 1", len(cfg.Tools))
	}
	if cfg.Updater.GitHubRepo != "test/repo" {
		t.Errorf("Updater.GitHubRepo = %q, want %q", cfg.Updater.GitHubRepo, "test/repo")
	}
}

func TestConfig_Load_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
app:
  log_level: "not a number"
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

func TestConfig_Load_AppliesDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Minimal config without port, log level, or copilot timeout
	content := `
line:
  channel_secret: ""
  channel_token: ""
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Should apply default port
	if cfg.LINE.WebhookPort != config.DefaultServerPort {
		t.Errorf("LINE.WebhookPort = %d, want default %d", cfg.LINE.WebhookPort, config.DefaultServerPort)
	}

	// Should apply default copilot timeout
	if cfg.Copilot.TimeoutSeconds != 600 {
		t.Errorf("Copilot.TimeoutSeconds = %d, want default 600", cfg.Copilot.TimeoutSeconds)
	}

	// Should apply default log level
	if cfg.App.LogLevel != "info" {
		t.Errorf("App.LogLevel = %q, want default %q", cfg.App.LogLevel, "info")
	}

	// Should apply default updater check interval
	if cfg.Updater.CheckIntervalHours != 6 {
		t.Errorf("Updater.CheckIntervalHours = %d, want default 6", cfg.Updater.CheckIntervalHours)
	}
}

func TestConfig_Load_EnvironmentVariableExpansion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Set environment variables
	t.Setenv("TEST_API_KEY", "expanded-api-key")
	t.Setenv("TEST_LINE_SECRET", "expanded-secret")

	content := `
app:
  log_level: info
copilot:
  api_key: "${TEST_API_KEY}"
line:
  channel_secret: "${TEST_LINE_SECRET}"
  channel_token: "${TEST_LINE_SECRET}"
  webhook_port: 8080
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Copilot.APIKey != "expanded-api-key" {
		t.Errorf("Copilot.APIKey = %q, want %q", cfg.Copilot.APIKey, "expanded-api-key")
	}
	if cfg.LINE.ChannelSecret != "expanded-secret" {
		t.Errorf("LINE.ChannelSecret = %q, want %q", cfg.LINE.ChannelSecret, "expanded-secret")
	}
}

func TestConfig_Load_EnvironmentVariableWithDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Ensure the env var is NOT set
	t.Setenv("UNSET_VAR", "")

	content := `
app:
  log_level: "${LOG_LEVEL_UNSET:-debug}"
  download_folder: "/tmp/test"
line:
  webhook_port: 8080
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.App.LogLevel != "debug" {
		t.Errorf("App.LogLevel = %q, want %q (default value)", cfg.App.LogLevel, "debug")
	}
}

func TestConfig_Load_EnvironmentVariableWithDefaultUsesEnvWhenSet(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Set the env var
	t.Setenv("LOG_LEVEL_SET", "warn")

	content := `
app:
  log_level: "${LOG_LEVEL_SET:-debug}"
  download_folder: "/tmp/test"
line:
  webhook_port: 8080
`
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.App.LogLevel != "warn" {
		t.Errorf("App.LogLevel = %q, want %q (env value over default)", cfg.App.LogLevel, "warn")
	}
}

func TestGenerateDefault(t *testing.T) {
	cfg, err := config.GenerateDefault()
	if err != nil {
		t.Fatalf("GenerateDefault() returned error: %v", err)
	}

	if cfg.App.LogLevel != "info" {
		t.Errorf("App.LogLevel = %q, want %q", cfg.App.LogLevel, "info")
	}
	if cfg.Copilot.TimeoutSeconds != 600 {
		t.Errorf("Copilot.TimeoutSeconds = %d, want 600", cfg.Copilot.TimeoutSeconds)
	}
	if len(cfg.Tools) != 2 {
		t.Errorf("len(Tools) = %d, want 2", len(cfg.Tools))
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	if err := config.WriteDefaultConfig(configPath); err != nil {
		t.Fatalf("WriteDefaultConfig() returned error: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify it's valid YAML
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	if len(data) == 0 {
		t.Error("Config file is empty")
	}
}

func TestConfig_GetToolConfig(t *testing.T) {
	cfg := &config.Config{
		Tools: []config.ToolConfig{
			{Name: "tool1", Type: "downie", Enabled: true},
			{Name: "tool2", Type: "google_drive", Enabled: false},
		},
	}

	tool, found := cfg.GetToolConfig("tool1")
	if !found {
		t.Error("GetToolConfig() should find existing tool")
	}
	if tool.Name != "tool1" {
		t.Errorf("GetToolConfig() returned wrong tool: %q", tool.Name)
	}

	_, found = cfg.GetToolConfig("nonexistent")
	if found {
		t.Error("GetToolConfig() should return false for nonexistent tool")
	}
}

func TestConfig_GetEnabledTools(t *testing.T) {
	cfg := &config.Config{
		Tools: []config.ToolConfig{
			{Name: "tool1", Type: "downie", Enabled: true},
			{Name: "tool2", Type: "google_drive", Enabled: false},
			{Name: "tool3", Type: "downie", Enabled: true},
		},
	}

	enabled := cfg.GetEnabledTools()
	if len(enabled) != 2 {
		t.Errorf("GetEnabledTools() returned %d tools, want 2", len(enabled))
	}
}
