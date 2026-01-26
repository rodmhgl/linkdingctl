package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/models"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a bookmark by ID",
	Long: `Get a bookmark by ID and display its full details.

Examples:
  linkdingctl get 123
  linkdingctl get 123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func init() {
	rootCmd.AddCommand(getCmd)
}

func runGet(cmd *cobra.Command, args []string) error {
	// Parse bookmark ID
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid bookmark ID: %s (must be a number)", args[0])
	}

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Fetch bookmark
	bookmark, err := client.GetBookmark(id)
	if err != nil {
		return err
	}

	// Output based on format
	if jsonOutput {
		return outputBookmarkJSON(bookmark)
	}

	return outputBookmarkHuman(bookmark)
}

func outputBookmarkJSON(bookmark interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(bookmark)
}

func outputBookmarkHuman(bookmark interface{}) error {
	// Type assert to access bookmark fields
	b, ok := bookmark.(*models.Bookmark)
	if !ok {
		return fmt.Errorf("unexpected bookmark type")
	}

	fmt.Printf("ID:          %d\n", b.ID)
	fmt.Printf("URL:         %s\n", b.URL)
	fmt.Printf("Title:       %s\n", b.Title)
	if b.Description != "" {
		fmt.Printf("Description: %s\n", b.Description)
	}
	if b.Notes != "" {
		fmt.Printf("Notes:       %s\n", b.Notes)
	}
	if len(b.TagNames) > 0 {
		fmt.Printf("Tags:        %s\n", joinTags(b.TagNames))
	} else {
		fmt.Printf("Tags:        -\n")
	}
	fmt.Printf("Added:       %s\n", b.DateAdded.Format("2006-01-02 15:04:05"))
	fmt.Printf("Modified:    %s\n", b.DateModified.Format("2006-01-02 15:04:05"))
	fmt.Printf("Unread:      %t\n", b.Unread)
	fmt.Printf("Shared:      %t\n", b.Shared)
	fmt.Printf("Archived:    %t\n", b.IsArchived)

	return nil
}

func joinTags(tags []string) string {
	return strings.Join(tags, ", ")
}
