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
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "ld",
	Short: "LinkDing CLI - Manage your LinkDing bookmarks from the command line",
	Long: `ld is a command-line interface for managing bookmarks in your LinkDing instance.

Configure your LinkDing connection with 'ld config init', then use commands like
'ld add', 'ld list', and 'ld get' to manage your bookmarks.`,
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default ~/.config/ld/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON instead of human-readable")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug logging")
}

// loadConfig loads the configuration from file and environment variables
func loadConfig() (*config.Config, error) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
