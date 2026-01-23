package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/models"
	"github.com/spf13/cobra"
)

var (
	addTitle       string
	addDescription string
	addTags        []string
	addUnread      bool
	addShared      bool
)

var addCmd = &cobra.Command{
	Use:   "add <url>",
	Short: "Add a new bookmark",
	Long:  `Add a new bookmark to your LinkDing instance with optional metadata.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		// Load config
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		// Create API client
		client := api.NewClient(cfg.URL, cfg.Token)

		// Create bookmark
		create := &models.BookmarkCreate{
			URL:         url,
			Title:       addTitle,
			Description: addDescription,
			TagNames:    addTags,
			Unread:      addUnread,
			Shared:      addShared,
		}

		bookmark, err := client.CreateBookmark(create)
		if err != nil {
			return err
		}

		// Output
		if jsonOutput {
			return json.NewEncoder(os.Stdout).Encode(bookmark)
		}

		fmt.Printf("✓ Bookmark added: %s\n", bookmark.Title)
		fmt.Printf("  ID: %d\n", bookmark.ID)
		fmt.Printf("  URL: %s\n", bookmark.URL)
		if len(bookmark.TagNames) > 0 {
			fmt.Printf("  Tags: %s\n", strings.Join(bookmark.TagNames, ", "))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringVarP(&addTitle, "title", "t", "", "Custom title (default: auto-fetch)")
	addCmd.Flags().StringVarP(&addDescription, "description", "d", "", "Description/notes")
	addCmd.Flags().StringSliceVarP(&addTags, "tags", "T", nil, "Comma-separated tags")
	addCmd.Flags().BoolVarP(&addUnread, "unread", "u", false, "Mark as unread")
	addCmd.Flags().BoolVarP(&addShared, "shared", "s", false, "Make publicly shared")
}
