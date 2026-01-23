package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage LinkDing configuration",
	Long:  `Manage your LinkDing CLI configuration including connection settings and API token.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration interactively",
	Long:  `Create a new configuration file by prompting for LinkDing URL and API token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		// Get URL
		fmt.Print("LinkDing URL: ")
		url, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read URL: %w", err)
		}
		url = strings.TrimSpace(url)

		// Get token
		fmt.Print("API Token: ")
		var token string
		if term.IsTerminal(int(os.Stdin.Fd())) {
			// TTY: Use password masking
			tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return fmt.Errorf("failed to read token: %w", err)
			}
			token = string(tokenBytes)
			fmt.Println() // Print newline after password input
		} else {
			// Non-TTY: Fall back to regular reading (for piped input)
			tokenInput, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read token: %w", err)
			}
			token = strings.TrimSpace(tokenInput)
		}
		token = strings.TrimSpace(token)

		// Validate inputs
		if url == "" || token == "" {
			return fmt.Errorf("URL and token are required")
		}

		// Create config
		cfg := &config.Config{
			URL:   url,
			Token: token,
		}

		// Determine config path
		configPath := cfgFile
		if configPath == "" {
			defaultPath, err := config.DefaultConfigPath()
			if err != nil {
				return err
			}
			configPath = defaultPath
		}

		// Save config
		if err := config.Save(cfg, configPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		if jsonOutput {
			output := map[string]string{
				"status": "success",
				"path":   configPath,
			}
			return json.NewEncoder(os.Stdout).Encode(output)
		}

		fmt.Printf("✓ Configuration saved to %s\n", configPath)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long:  `Show the current configuration with API token redacted for security.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		if jsonOutput {
			output := map[string]string{
				"url":   cfg.URL,
				"token": redactToken(cfg.Token),
			}
			return json.NewEncoder(os.Stdout).Encode(output)
		}

		fmt.Printf("URL: %s\n", cfg.URL)
		fmt.Printf("Token: %s\n", redactToken(cfg.Token))
		return nil
	},
}

var configTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test connection to LinkDing",
	Long:  `Verify that the configured URL and token can successfully connect to LinkDing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.URL, cfg.Token)
		if err := client.TestConnection(); err != nil {
			if jsonOutput {
				output := map[string]string{
					"status": "failed",
					"error":  err.Error(),
				}
				json.NewEncoder(os.Stdout).Encode(output)
				return err
			}
			return fmt.Errorf("✗ Connection failed: %w", err)
		}

		if jsonOutput {
			output := map[string]string{
				"status": "success",
				"url":    cfg.URL,
			}
			return json.NewEncoder(os.Stdout).Encode(output)
		}

		fmt.Printf("✓ Successfully connected to %s\n", cfg.URL)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configTestCmd)
}

// redactToken masks most of the token for security
func redactToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
