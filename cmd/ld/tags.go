package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/config"
	"github.com/rodstewart/linkding-cli/internal/models"
	"github.com/spf13/cobra"
)

// tagsCmd represents the tags command
var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all tags with bookmark counts",
	Long: `List all tags from LinkDing with their bookmark counts.

Examples:
  ld tags
  ld tags --sort count
  ld tags --unused
  ld tags --json`,
	RunE: runTags,
}

var (
	tagsSort        string
	tagsUnused      bool
	tagsRenameForce bool
	tagsDeleteForce bool
)

func init() {
	rootCmd.AddCommand(tagsCmd)
	tagsCmd.AddCommand(tagsRenameCmd)
	tagsCmd.AddCommand(tagsDeleteCmd)
	tagsCmd.AddCommand(tagsShowCmd)

	tagsCmd.Flags().StringVarP(&tagsSort, "sort", "s", "name", "Sort by: name, count")
	tagsCmd.Flags().BoolVar(&tagsUnused, "unused", false, "Show only tags with 0 bookmarks")

	tagsRenameCmd.Flags().BoolVarP(&tagsRenameForce, "force", "f", false, "Skip confirmation")
	tagsDeleteCmd.Flags().BoolVarP(&tagsDeleteForce, "force", "f", false, "Skip confirmation and remove tag from all bookmarks")
}

func runTags(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Fetch all tags to get complete list (including unused ones)
	allTagsList, err := client.FetchAllTags()
	if err != nil {
		return err
	}

	// Fetch all bookmarks (including archived) to count tags client-side
	// This is more efficient than making N API calls (one per tag)
	allBookmarks, err := client.FetchAllBookmarks(nil, true)
	if err != nil {
		return fmt.Errorf("failed to fetch bookmarks: %w", err)
	}

	// Build tag counts by iterating through bookmarks
	tagCounts := make(map[string]int)
	
	// Initialize all tags with 0 count
	for _, tag := range allTagsList {
		tagCounts[tag.Name] = 0
	}
	
	// Count tags from bookmarks
	for _, bookmark := range allBookmarks {
		for _, tag := range bookmark.TagNames {
			tagCounts[tag]++
		}
	}

	// Build list of TagWithCount
	var tagsWithCount []models.TagWithCount
	for name, count := range tagCounts {
		// Filter by unused if flag is set
		if tagsUnused && count > 0 {
			continue
		}
		tagsWithCount = append(tagsWithCount, models.TagWithCount{
			Name:  name,
			Count: count,
		})
	}

	// Sort based on flag
	switch tagsSort {
	case "name":
		sort.Slice(tagsWithCount, func(i, j int) bool {
			return tagsWithCount[i].Name < tagsWithCount[j].Name
		})
	case "count":
		sort.Slice(tagsWithCount, func(i, j int) bool {
			// Sort by count descending
			return tagsWithCount[i].Count > tagsWithCount[j].Count
		})
	default:
		return fmt.Errorf("invalid sort option: %s (use 'name' or 'count')", tagsSort)
	}

	// Output based on format
	if jsonOutput {
		return outputTagsJSON(tagsWithCount)
	}

	return outputTagsTable(tagsWithCount)
}

func outputTagsJSON(tags []models.TagWithCount) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tags)
}

func outputTagsTable(tags []models.TagWithCount) error {
	if len(tags) == 0 {
		fmt.Println("No tags found")
		return nil
	}

	// Create tabwriter for aligned columns
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintln(w, "TAG\tCOUNT")
	fmt.Fprintln(w, "---\t-----")

	// Rows
	for _, tag := range tags {
		fmt.Fprintf(w, "%s\t%d\n", tag.Name, tag.Count)
	}

	w.Flush()

	// Show summary
	fmt.Printf("\nTotal: %d tags\n", len(tags))

	return nil
}

// tagsRenameCmd represents the tags rename command
var tagsRenameCmd = &cobra.Command{
	Use:   "rename <old-name> <new-name>",
	Short: "Rename a tag across all bookmarks",
	Long: `Rename a tag by updating all bookmarks that use it.

This command will:
1. Find all bookmarks with the old tag
2. Update each bookmark to replace the old tag with the new tag
3. Show progress as bookmarks are updated

Examples:
  ld tags rename oldtag newtag
  ld tags rename "old tag" "new tag" --force`,
	Args: cobra.ExactArgs(2),
	RunE: runTagsRename,
}

func runTagsRename(cmd *cobra.Command, args []string) error {
	oldTag := args[0]
	newTag := args[1]

	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Get all bookmarks with the old tag (including archived)
	allBookmarks, err := client.FetchAllBookmarks([]string{oldTag}, true)
	if err != nil {
		return fmt.Errorf("failed to fetch bookmarks with tag '%s': %w", oldTag, err)
	}

	if len(allBookmarks) == 0 {
		return fmt.Errorf("no bookmarks found with tag '%s'", oldTag)
	}

	// Ask for confirmation unless --force is used
	if !tagsRenameForce {
		fmt.Printf("This will rename tag '%s' to '%s' on %d bookmark(s).\n", oldTag, newTag, len(allBookmarks))
		fmt.Print("Continue? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted")
			return nil
		}
	}

	// Update each bookmark
	successCount := 0
	errorCount := 0

	for i, bookmark := range allBookmarks {
		// Show progress
		fmt.Printf("Updating bookmark %d/%d (ID: %d)...\n", i+1, len(allBookmarks), bookmark.ID)

		// Build new tag list: replace old tag with new tag
		newTags := make([]string, 0, len(bookmark.TagNames))
		for _, tag := range bookmark.TagNames {
			if tag == oldTag {
				newTags = append(newTags, newTag)
			} else {
				newTags = append(newTags, tag)
			}
		}

		// Update the bookmark
		update := &models.BookmarkUpdate{
			TagNames: &newTags,
		}

		_, err := client.UpdateBookmark(bookmark.ID, update)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			errorCount++
			continue
		}

		successCount++
	}

	// Show summary
	fmt.Printf("\nCompleted: %d successful, %d errors\n", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("some bookmarks failed to update")
	}

	return nil
}

// tagsDeleteCmd represents the tags delete command
var tagsDeleteCmd = &cobra.Command{
	Use:   "delete <tag-name>",
	Short: "Delete a tag, optionally removing it from all bookmarks",
	Long: `Delete a tag from LinkDing.

By default, this command only works if the tag has 0 bookmarks.
Use --force to remove the tag from all bookmarks first.

Examples:
  ld tags delete unused-tag
  ld tags delete "old tag" --force`,
	Args: cobra.ExactArgs(1),
	RunE: runTagsDelete,
}

func runTagsDelete(cmd *cobra.Command, args []string) error {
	tagName := args[0]

	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Get all bookmarks with the tag to check count
	allBookmarks, err := client.FetchAllBookmarks([]string{tagName}, true)
	if err != nil {
		return fmt.Errorf("failed to fetch bookmarks with tag '%s': %w", tagName, err)
	}

	bookmarkCount := len(allBookmarks)

	// If tag has bookmarks and --force is not set, error
	if bookmarkCount > 0 && !tagsDeleteForce {
		return fmt.Errorf("tag '%s' has %d bookmark(s). Remove tag from bookmarks first or use --force to remove from all", tagName, bookmarkCount)
	}

	// If tag has no bookmarks, we can just confirm deletion
	if bookmarkCount == 0 {
		fmt.Printf("Tag '%s' has no bookmarks and will be removed.\n", tagName)
		return nil
	}

	// If we get here, --force is set and tag has bookmarks
	// Ask for confirmation
	fmt.Printf("This will remove tag '%s' from %d bookmark(s).\n", tagName, bookmarkCount)
	fmt.Print("Continue? (y/N): ")

	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("Aborted")
		return nil
	}

	// Remove tag from all bookmarks
	successCount := 0
	errorCount := 0

	for i, bookmark := range allBookmarks {
		// Show progress
		fmt.Printf("Updating bookmark %d/%d (ID: %d)...\n", i+1, bookmarkCount, bookmark.ID)

		// Build new tag list: remove the tag
		newTags := make([]string, 0, len(bookmark.TagNames)-1)
		for _, tag := range bookmark.TagNames {
			if tag != tagName {
				newTags = append(newTags, tag)
			}
		}

		// Update the bookmark
		update := &models.BookmarkUpdate{
			TagNames: &newTags,
		}

		_, err := client.UpdateBookmark(bookmark.ID, update)
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
			errorCount++
			continue
		}

		successCount++
	}

	// Show summary
	fmt.Printf("\nCompleted: %d successful, %d errors\n", successCount, errorCount)
	fmt.Printf("Tag '%s' has been removed from all bookmarks.\n", tagName)

	if errorCount > 0 {
		return fmt.Errorf("some bookmarks failed to update")
	}

	return nil
}

// tagsShowCmd represents the tags show command
var tagsShowCmd = &cobra.Command{
	Use:   "show <tag-name>",
	Short: "Show all bookmarks with a specific tag",
	Long: `List all bookmarks that have the specified tag.

This is equivalent to: ld list --tags <tag-name>

Examples:
  ld tags show kubernetes
  ld tags show "web dev" --json`,
	Args: cobra.ExactArgs(1),
	RunE: runTagsShow,
}

func runTagsShow(cmd *cobra.Command, args []string) error {
	tagName := args[0]

	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Fetch bookmarks with the specified tag
	bookmarkList, err := client.GetBookmarks("", []string{tagName}, nil, nil, 1000, 0)
	if err != nil {
		return err
	}

	// Output based on format
	if jsonOutput {
		return outputJSON(bookmarkList)
	}

	return outputTable(bookmarkList)
}
