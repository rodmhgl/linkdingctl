package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/config"
	"github.com/rodstewart/linkding-cli/internal/export"
	"github.com/spf13/cobra"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore <backup-file>",
	Short: "Restore bookmarks from a backup file",
	Long: `Restore bookmarks from a backup file.

Without --wipe: Equivalent to 'ld import <file>'
  - Existing bookmarks are updated
  - New bookmarks are added

With --wipe: Deletes ALL existing bookmarks before importing (DANGEROUS)
  - Requires interactive confirmation
  - Cannot be undone

Examples:
  ld restore backup.json
  ld restore backup.json --dry-run
  ld restore backup.json --wipe`,
	Args: cobra.ExactArgs(1),
	RunE: runRestore,
}

var (
	restoreDryRun bool
	restoreWipe   bool
)

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().BoolVar(&restoreDryRun, "dry-run", false, "Show what would be restored without making changes")
	restoreCmd.Flags().BoolVar(&restoreWipe, "wipe", false, "Delete all existing bookmarks before restore (DANGEROUS)")
}

func runRestore(cmd *cobra.Command, args []string) error {
	filename := args[0]

	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// If --wipe is specified, handle deletion with confirmation
	if restoreWipe {
		if err := handleWipe(client); err != nil {
			return err
		}
	}

	// Import the backup file
	options := export.ImportOptions{
		Format:         "auto",
		DryRun:         restoreDryRun,
		SkipDuplicates: false,
		AddTags:        []string{},
	}

	if !jsonOutput {
		if restoreDryRun {
			fmt.Fprintln(os.Stderr, "Dry run - no changes will be made")
		}
		if restoreWipe && !restoreDryRun {
			fmt.Fprintln(os.Stderr, "Restoring bookmarks...")
		} else if !restoreDryRun {
			fmt.Fprintln(os.Stderr, "Importing bookmarks...")
		}
	}

	result, err := export.ImportBookmarks(client, filename, options)
	if err != nil {
		return err
	}

	// Display results
	if !jsonOutput {
		displayImportResult(result)
	} else {
		// JSON output - reuse import command's JSON output logic
		return outputImportResultJSON(result)
	}

	return nil
}

// handleWipe deletes all existing bookmarks with user confirmation
func handleWipe(client *api.Client) error {
	// Get count of existing bookmarks
	bookmarks, err := client.GetBookmarks("", []string{}, nil, nil, 1, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch bookmarks: %w", err)
	}

	count := bookmarks.Count
	if count == 0 {
		if !jsonOutput {
			fmt.Fprintln(os.Stderr, "No existing bookmarks to delete.")
		}
		return nil
	}

	// For dry-run, just show what would happen
	if restoreDryRun {
		if !jsonOutput {
			fmt.Fprintf(os.Stderr, "Dry run: Would delete %d existing bookmarks\n", count)
		}
		return nil
	}

	// JSON mode - don't prompt interactively
	if jsonOutput {
		return fmt.Errorf("--wipe requires interactive confirmation. Cannot use with --json flag")
	}

	// Prompt for confirmation
	fmt.Fprintf(os.Stderr, "WARNING: This will delete ALL %d existing bookmarks before restoring.\n", count)
	fmt.Fprint(os.Stderr, "Type 'yes' to confirm: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "yes" {
		return fmt.Errorf("restore cancelled")
	}

	// Fetch and delete all bookmarks
	fmt.Fprintln(os.Stderr, "Deleting existing bookmarks...")

	// Fetch all bookmarks
	allBookmarks := []int{}
	offset := 0
	limit := 100

	for {
		bookmarkList, err := client.GetBookmarks("", []string{}, nil, nil, limit, offset)
		if err != nil {
			return fmt.Errorf("failed to fetch bookmarks: %w", err)
		}

		for _, b := range bookmarkList.Results {
			allBookmarks = append(allBookmarks, b.ID)
		}

		if bookmarkList.Next == nil || len(bookmarkList.Results) == 0 {
			break
		}

		offset += limit
	}

	// Delete each bookmark
	deleted := 0
	failed := 0
	for _, id := range allBookmarks {
		if err := client.DeleteBookmark(id); err != nil {
			failed++
		} else {
			deleted++
		}
	}

	if failed > 0 {
		fmt.Fprintf(os.Stderr, "Deleted %d bookmarks, %d failed\n", deleted, failed)
	} else {
		fmt.Fprintf(os.Stderr, "Deleted %d bookmarks\n", deleted)
	}

	return nil
}

// outputImportResultJSON outputs the import result as JSON
func outputImportResultJSON(result *export.ImportResult) error {
	output := map[string]interface{}{
		"added":   result.Added,
		"updated": result.Updated,
		"skipped": result.Skipped,
		"failed":  result.Failed,
	}

	if len(result.Errors) > 0 {
		errors := make([]map[string]interface{}, len(result.Errors))
		for i, e := range result.Errors {
			errors[i] = map[string]interface{}{
				"line":    e.Line,
				"message": e.Message,
			}
		}
		output["errors"] = errors
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
