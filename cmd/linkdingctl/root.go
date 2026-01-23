package main

import (
	"fmt"
	"os"

	"github.com/rodstewart/linkding-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	jsonOutput bool
	debugMode  bool
	flagURL    string
	flagToken  string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "linkdingctl",
	Short: "LinkDing CLI - Manage your LinkDing bookmarks from the command line",
	Long: `linkdingctl is a command-line interface for managing bookmarks in your LinkDing instance.

Configure your LinkDing connection with 'linkdingctl config init', then use commands like
'linkdingctl add', 'linkdingctl list', and 'linkdingctl get' to manage your bookmarks.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default ~/.config/linkdingctl/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON instead of human-readable")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug logging")
	rootCmd.PersistentFlags().StringVar(&flagURL, "url", "", "LinkDing instance URL (overrides config and env)")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "API token (overrides config and env)")
}

// loadConfig loads the configuration from file and environment variables,
// then applies CLI flag overrides if provided.
func loadConfig() (*config.Config, error) {
	cfg, err := config.Load(cfgFile)

	// If config loading failed but we have both URL and token from CLI flags,
	// we can proceed without a config file
	if err != nil {
		if flagURL != "" && flagToken != "" {
			cfg = &config.Config{
				URL:   flagURL,
				Token: flagToken,
			}
			return cfg, nil
		}
		return nil, err
	}

	// Apply CLI flag overrides (highest precedence)
	if flagURL != "" {
		cfg.URL = flagURL
	}
	if flagToken != "" {
		cfg.Token = flagToken
	}

	return cfg, nil
}
