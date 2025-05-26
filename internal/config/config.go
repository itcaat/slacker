package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/itcaat/slacker/models"
	"github.com/spf13/viper"
)

const (
	ConfigFileName = ".slacker"
	ConfigFileType = "yaml"
)

// Config represents the application configuration
type Config struct {
	Slack  models.SlackConfig `mapstructure:"slack"`
	Debug  bool               `mapstructure:"debug"`
	Export ExportConfig       `mapstructure:"export"`
}

// ExportConfig represents export-specific configuration
type ExportConfig struct {
	DefaultOutputDir string `mapstructure:"default_output_dir"`
	IncludeThreads   bool   `mapstructure:"include_threads"`
	IncludeUsers     bool   `mapstructure:"include_users"`
	MaxMessages      int    `mapstructure:"max_messages"`
}

// Manager handles configuration loading and saving
type Manager struct {
	configPath string
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{}
}

// Load loads the configuration from file
func (m *Manager) Load() (*Config, error) {
	// Set default values
	viper.SetDefault("debug", false)
	viper.SetDefault("export.default_output_dir", "./exports")
	viper.SetDefault("export.include_threads", true)
	viper.SetDefault("export.include_users", true)
	viper.SetDefault("export.max_messages", 0) // 0 = no limit

	// Find home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Set config file path
	configPath := filepath.Join(home, ConfigFileName+"."+ConfigFileType)
	m.configPath = configPath

	// Configure viper
	viper.SetConfigName(ConfigFileName)
	viper.SetConfigType(ConfigFileType)
	viper.AddConfigPath(home)
	viper.AddConfigPath(".")

	// Enable environment variable support
	viper.AutomaticEnv()
	viper.SetEnvPrefix("SLACKER")

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default config
			return m.createDefaultConfig()
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// Save saves the configuration to file
func (m *Manager) Save(config *Config) error {
	// Set all config values in viper
	viper.Set("slack.token", config.Slack.Token)
	viper.Set("slack.app_token", config.Slack.AppToken)
	viper.Set("slack.bot_token", config.Slack.BotToken)
	viper.Set("slack.user_token", config.Slack.UserToken)
	viper.Set("debug", config.Debug)
	viper.Set("export.default_output_dir", config.Export.DefaultOutputDir)
	viper.Set("export.include_threads", config.Export.IncludeThreads)
	viper.Set("export.include_users", config.Export.IncludeUsers)
	viper.Set("export.max_messages", config.Export.MaxMessages)

	// Write config file
	if err := viper.WriteConfig(); err != nil {
		// If config file doesn't exist, create it
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return viper.SafeWriteConfig()
		}
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SetToken sets the Slack token and saves the configuration
func (m *Manager) SetToken(token string) error {
	config, err := m.Load()
	if err != nil {
		// If config doesn't exist, create a new one
		config = &Config{}
	}

	config.Slack.Token = token
	return m.Save(config)
}

// GetToken retrieves the Slack token from configuration or environment
func (m *Manager) GetToken() (string, error) {
	// First check environment variable
	if token := os.Getenv("SLACKER_SLACK_TOKEN"); token != "" {
		return token, nil
	}

	// Then check config file
	config, err := m.Load()
	if err != nil {
		return "", err
	}

	if config.Slack.Token == "" {
		return "", fmt.Errorf("no Slack token found. Set SLACKER_SLACK_TOKEN environment variable or run 'slacker auth'")
	}

	return config.Slack.Token, nil
}

// createDefaultConfig creates a default configuration file
func (m *Manager) createDefaultConfig() (*Config, error) {
	config := &Config{
		Debug: false,
		Export: ExportConfig{
			DefaultOutputDir: "./exports",
			IncludeThreads:   true,
			IncludeUsers:     true,
			MaxMessages:      0,
		},
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return config, nil
}

// GetConfigPath returns the path to the configuration file
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// IsConfigured checks if the application is properly configured
func (m *Manager) IsConfigured() bool {
	token, err := m.GetToken()
	return err == nil && token != ""
}
