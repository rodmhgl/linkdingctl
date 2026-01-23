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
	tagsSort   string
	tagsUnused bool
)

func init() {
	rootCmd.AddCommand(tagsCmd)

	tagsCmd.Flags().StringVarP(&tagsSort, "sort", "s", "name", "Sort by: name, count")
	tagsCmd.Flags().BoolVar(&tagsUnused, "unused", false, "Show only tags with 0 bookmarks")
}

func runTags(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Fetch all tags (paginated)
	tagList, err := client.GetTags(1000, 0)
	if err != nil {
		return err
	}

	// Build tag counts by fetching bookmarks for each tag
	tagCounts := make(map[string]int)
	for _, tag := range tagList.Results {
		// Query bookmarks with this tag to get count
		bookmarkList, err := client.GetBookmarks("", []string{tag.Name}, nil, nil, 1, 0)
		if err != nil {
			return fmt.Errorf("failed to get count for tag '%s': %w", tag.Name, err)
		}
		tagCounts[tag.Name] = bookmarkList.Count
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
