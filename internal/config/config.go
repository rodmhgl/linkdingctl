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

// migrateFromOldPath attempts to migrate config from old path (~/.config/ld/)
// to new path (~/.config/linkdingctl/). Returns true if migration was performed.
func migrateFromOldPath(newConfigPath string) (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("failed to get user home directory: %w", err)
	}

	oldConfigPath := filepath.Join(homeDir, ".config", "ld", "config.yaml")
	
	// Check if new config already exists - if so, skip migration
	if _, err := os.Stat(newConfigPath); err == nil {
		return false, nil
	}

	// Check if old config exists - if not, nothing to migrate
	oldData, err := os.ReadFile(oldConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read old config file: %w", err)
	}

	// Create new config directory with restricted permissions
	newConfigDir := filepath.Dir(newConfigPath)
	if err := os.MkdirAll(newConfigDir, 0700); err != nil {
		return false, fmt.Errorf("failed to create new config directory: %w", err)
	}

	// Write config to new location with restricted permissions
	if err := os.WriteFile(newConfigPath, oldData, 0600); err != nil {
		return false, fmt.Errorf("failed to write new config file: %w", err)
	}

	// Print notice to stderr
	fmt.Fprintf(os.Stderr, "Notice: Configuration migrated from %s to %s\n", oldConfigPath, newConfigPath)
	fmt.Fprintf(os.Stderr, "The old config file has been preserved and can be safely deleted.\n")

	return true, nil
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
		configDir := filepath.Join(homeDir, ".config", "linkdingctl")
		defaultPath := filepath.Join(configDir, "config.yaml")
		
		// Attempt migration from old path if needed
		if _, err := migrateFromOldPath(defaultPath); err != nil {
			// Log the error but don't fail - config might exist elsewhere
			fmt.Fprintf(os.Stderr, "Warning: config migration failed: %v\n", err)
		}
		
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
		return nil, fmt.Errorf("no configuration found. Run 'linkdingctl config init' to set up")
	}

	return cfg, nil
}

// DefaultConfigPath returns the default configuration file path
func DefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "linkdingctl", "config.yaml"), nil
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
