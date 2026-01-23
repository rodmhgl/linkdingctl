package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/config"
	"github.com/rodstewart/linkding-cli/internal/models"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List bookmarks",
	Long: `List bookmarks with optional filtering.

Examples:
  ld list
  ld list --tags k8s,platform
  ld list -q "kubernetes" --unread
  ld list --limit 10`,
	RunE: runList,
}

var (
	listQuery    string
	listTags     []string
	listUnread   bool
	listArchived bool
	listLimit    int
	listOffset   int
)

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(&listQuery, "query", "q", "", "Search query")
	listCmd.Flags().StringSliceVarP(&listTags, "tags", "T", []string{}, "Filter by tags (AND logic)")
	listCmd.Flags().BoolVarP(&listUnread, "unread", "u", false, "Show only unread")
	listCmd.Flags().BoolVarP(&listArchived, "archived", "a", false, "Show only archived")
	listCmd.Flags().IntVarP(&listLimit, "limit", "l", 100, "Max results")
	listCmd.Flags().IntVarP(&listOffset, "offset", "o", 0, "Pagination offset")
}

func runList(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Convert bool flags to pointers for API call
	var unreadPtr, archivedPtr *bool
	if cmd.Flags().Changed("unread") {
		unreadPtr = &listUnread
	}
	if cmd.Flags().Changed("archived") {
		archivedPtr = &listArchived
	}

	// Fetch bookmarks
	bookmarkList, err := client.GetBookmarks(listQuery, listTags, unreadPtr, archivedPtr, listLimit, listOffset)
	if err != nil {
		return err
	}

	// Output based on format
	if jsonOutput {
		return outputJSON(bookmarkList)
	}

	return outputTable(bookmarkList)
}

func outputJSON(bookmarkList interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(bookmarkList)
}

func outputTable(bookmarkList *models.BookmarkList) error {
	if len(bookmarkList.Results) == 0 {
		fmt.Println("No bookmarks found")
		return nil
	}

	// Create tabwriter for aligned columns
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "ID\tTITLE\tTAGS\tDATE")
	fmt.Fprintln(w, "--\t-----\t----\t----")

	// Rows
	for _, bookmark := range bookmarkList.Results {
		title := truncate(bookmark.Title, 50)
		tags := strings.Join(bookmark.TagNames, ", ")
		if tags == "" {
			tags = "-"
		}
		tags = truncate(tags, 30)
		date := bookmark.DateAdded.Format("2006-01-02")

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", bookmark.ID, title, tags, date)
	}

	w.Flush()

	// Show pagination info
	fmt.Printf("\nShowing %d of %d total bookmarks\n", len(bookmarkList.Results), bookmarkList.Count)
	if bookmarkList.Next != nil {
		fmt.Printf("Use --offset %d to see more\n", listOffset+listLimit)
	}

	return nil
}

// truncate truncates a string to maxLen characters, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
