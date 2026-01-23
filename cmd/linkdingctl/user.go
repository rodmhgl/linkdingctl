package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User profile operations",
	Long:  `Display and manage user profile information from your LinkDing instance.`,
}

var userProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Display user profile information",
	Long:  `Retrieve and display the profile information for the authenticated user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.URL, cfg.Token)
		profile, err := client.GetUserProfile()
		if err != nil {
			if jsonOutput {
				output := map[string]string{
					"status": "failed",
					"error":  err.Error(),
				}
				json.NewEncoder(os.Stdout).Encode(output)
				return err
			}
			return fmt.Errorf("failed to retrieve user profile: %w", err)
		}

		if jsonOutput {
			return json.NewEncoder(os.Stdout).Encode(profile)
		}

		fmt.Printf("Username: %s\n", profile.Username)
		fmt.Printf("Display Name: %s\n", profile.DisplayName)
		fmt.Printf("Theme: %s\n", profile.Theme)
		fmt.Printf("Bookmark Count: %d\n", profile.BookmarkCount)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userProfileCmd)
}
