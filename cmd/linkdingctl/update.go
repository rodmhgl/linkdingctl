package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	updateTitle       string
	updateDescription string
	updateNotes       string
	updateTags        []string
	updateAddTags     []string
	updateRemoveTags  []string
	updateArchive     bool
	updateUnarchive   bool
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update <id> [flags]",
	Short: "Update a bookmark",
	Long: `Update a bookmark's metadata. Only specified fields are modified.

Examples:
  ld update 123 --title "New Title"
  ld update 123 --add-tags "reviewed"
  ld update 123 --title "New Title" --archive
  ld update 123 --remove-tags "outdated" --add-tags "current"`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New title")
	updateCmd.Flags().StringVarP(&updateDescription, "description", "d", "", "New description")
	updateCmd.Flags().StringVarP(&updateNotes, "notes", "n", "", "New notes")
	updateCmd.Flags().StringSliceVarP(&updateTags, "tags", "T", nil, "Replace tags (comma-separated)")
	updateCmd.Flags().StringSliceVar(&updateAddTags, "add-tags", nil, "Add tags to existing (comma-separated)")
	updateCmd.Flags().StringSliceVar(&updateRemoveTags, "remove-tags", nil, "Remove specific tags (comma-separated)")
	updateCmd.Flags().BoolVarP(&updateArchive, "archive", "a", false, "Archive the bookmark")
	updateCmd.Flags().BoolVar(&updateUnarchive, "unarchive", false, "Unarchive the bookmark")

	// Create bool pointers for flags that need to detect if they were set
	updateCmd.Flags().BoolP("unread", "u", false, "Set unread status")
	updateCmd.Flags().BoolP("shared", "s", false, "Set shared status")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Parse bookmark ID
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid bookmark ID: %s (must be a number)", args[0])
	}

	// Validate conflicting flags
	if updateArchive && updateUnarchive {
		return fmt.Errorf("cannot use both --archive and --unarchive")
	}

	if len(updateTags) > 0 && (len(updateAddTags) > 0 || len(updateRemoveTags) > 0) {
		return fmt.Errorf("cannot use --tags with --add-tags or --remove-tags (use one approach)")
	}

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Build update request
	update := &models.BookmarkUpdate{}

	// Handle simple string fields
	if cmd.Flags().Changed("title") {
		update.Title = &updateTitle
	}
	if cmd.Flags().Changed("description") {
		update.Description = &updateDescription
	}
	if cmd.Flags().Changed("notes") {
		update.Notes = &updateNotes
	}

	// Handle boolean flags - only set if explicitly provided
	if cmd.Flags().Changed("unread") {
		val, _ := cmd.Flags().GetBool("unread")
		update.Unread = &val
	}
	if cmd.Flags().Changed("shared") {
		val, _ := cmd.Flags().GetBool("shared")
		update.Shared = &val
	}

	// Handle archive/unarchive
	if updateArchive {
		archived := true
		update.IsArchived = &archived
	} else if updateUnarchive {
		archived := false
		update.IsArchived = &archived
	}

	// Handle tag operations
	if len(updateTags) > 0 {
		// Replace all tags
		update.TagNames = &updateTags
	} else if len(updateAddTags) > 0 || len(updateRemoveTags) > 0 {
		// Need to fetch current bookmark to merge tags
		currentBookmark, err := client.GetBookmark(id)
		if err != nil {
			return err
		}

		// Start with current tags
		tagSet := make(map[string]bool)
		for _, tag := range currentBookmark.TagNames {
			tagSet[tag] = true
		}

		// Add new tags
		for _, tag := range updateAddTags {
			tagSet[tag] = true
		}

		// Remove tags
		for _, tag := range updateRemoveTags {
			delete(tagSet, tag)
		}

		// Convert back to slice
		newTags := make([]string, 0, len(tagSet))
		for tag := range tagSet {
			newTags = append(newTags, tag)
		}

		update.TagNames = &newTags
	}

	// Perform update
	bookmark, err := client.UpdateBookmark(id, update)
	if err != nil {
		return err
	}

	// Output based on format
	if jsonOutput {
		return outputBookmarkJSON(bookmark)
	}

	fmt.Printf("✓ Bookmark updated: %s\n", bookmark.Title)
	fmt.Printf("  ID: %d\n", bookmark.ID)
	fmt.Printf("  URL: %s\n", bookmark.URL)
	if len(bookmark.TagNames) > 0 {
		fmt.Printf("  Tags: %s\n", strings.Join(bookmark.TagNames, ", "))
	}
	if bookmark.IsArchived {
		fmt.Printf("  Status: Archived\n")
	}

	return nil
}
