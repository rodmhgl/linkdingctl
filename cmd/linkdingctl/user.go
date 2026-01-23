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
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(profile)
		}

		// Helper function to convert bool to enabled/disabled
		boolToStatus := func(b bool) string {
			if b {
				return "enabled"
			}
			return "disabled"
		}

		fmt.Printf("Theme:                  %s\n", profile.Theme)
		fmt.Printf("Bookmark Date Display:  %s\n", profile.BookmarkDateDisplay)
		fmt.Printf("Bookmark Link Target:   %s\n", profile.BookmarkLinkTarget)
		fmt.Printf("Web Archive:            %s\n", profile.WebArchiveIntegration)
		fmt.Printf("Tag Search:             %s\n", profile.TagSearch)
		fmt.Printf("Sharing:                %s\n", boolToStatus(profile.EnableSharing))
		fmt.Printf("Public Sharing:         %s\n", boolToStatus(profile.EnablePublicSharing))
		fmt.Printf("Favicons:               %s\n", boolToStatus(profile.EnableFavicons))
		fmt.Printf("Display URL:            %s\n", boolToStatus(profile.DisplayURL))
		fmt.Printf("Permanent Notes:        %s\n", boolToStatus(profile.PermanentNotes))
		fmt.Printf("Search Sort:            %s\n", profile.SearchPreferences.Sort)
		fmt.Printf("Search Shared:          %s\n", profile.SearchPreferences.Shared)
		fmt.Printf("Search Unread:          %s\n", profile.SearchPreferences.Unread)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userProfileCmd)
}
