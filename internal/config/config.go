package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	URL   string
	Token string
}

// Load loads configuration from file and environment variables.
// Environment variables take precedence over config file values.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file path
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Default config location
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir := filepath.Join(homeDir, ".config", "ld")
		v.AddConfigPath(configDir)
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Environment variable bindings (higher priority)
	v.SetEnvPrefix("LINKDING")
	v.BindEnv("url")
	v.BindEnv("token")

	// Read config file (ignore if not found)
	if err := v.ReadInConfig(); err != nil {
		// Ignore "not found" errors - we'll validate required fields later
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Also ignore if the file simply doesn't exist
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read config file: %w", err)
			}
		}
	}

	cfg := &Config{
		URL:   v.GetString("url"),
		Token: v.GetString("token"),
	}

	// Validate that required fields are present
	if cfg.URL == "" || cfg.Token == "" {
		return nil, fmt.Errorf("no configuration found. Run 'ld config init' to set up")
	}

	return cfg, nil
}

// DefaultConfigPath returns the default configuration file path
func DefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "ld", "config.yaml"), nil
}

// Save writes configuration to the specified path
func Save(cfg *Config, configPath string) error {
	// Ensure directory exists with restricted permissions (owner-only)
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.Set("url", cfg.URL)
	v.Set("token", cfg.Token)

	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Set restrictive permissions on config file (owner read/write only)
	if err := os.Chmod(configPath, 0600); err != nil {
		return fmt.Errorf("failed to set config file permissions: %w", err)
	}

	return nil
}
