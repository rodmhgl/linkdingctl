package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/config"
	"github.com/rodstewart/linkding-cli/internal/export"
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import bookmarks from a file",
	Long: `Import bookmarks from various formats (JSON, HTML, CSV).

Format is auto-detected from file extension:
  .json → JSON format
  .html, .htm → HTML/Netscape format
  .csv → CSV format

Examples:
  ld import bookmarks.json
  ld import bookmarks.html --add-tags "imported"
  ld import export.csv --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

var (
	importFormat         string
	importDryRun         bool
	importSkipDuplicates bool
	importAddTags        []string
)

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringVarP(&importFormat, "format", "f", "auto", "Input format: json, html, csv (default: auto-detect)")
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "Show what would be imported without making changes")
	importCmd.Flags().BoolVar(&importSkipDuplicates, "skip-duplicates", false, "Skip URLs that already exist (default: update them)")
	importCmd.Flags().StringSliceVarP(&importAddTags, "add-tags", "T", []string{}, "Add these tags to all imported bookmarks")
}

func runImport(cmd *cobra.Command, args []string) error {
	filename := args[0]

	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Create import options
	options := export.ImportOptions{
		Format:         importFormat,
		DryRun:         importDryRun,
		SkipDuplicates: importSkipDuplicates,
		AddTags:        importAddTags,
	}

	// Check if JSON output is requested
	if jsonOutput {
		return runImportJSON(client, filename, options)
	}

	// Perform import with progress display
	if importDryRun {
		fmt.Fprintln(os.Stderr, "Dry run - no changes will be made")
	}
	fmt.Fprintln(os.Stderr, "Importing bookmarks...")

	result, err := export.ImportBookmarks(client, filename, options)
	if err != nil {
		return err
	}

	// Display results
	displayImportResult(result)

	return nil
}

func runImportJSON(client *api.Client, filename string, options export.ImportOptions) error {
	result, err := export.ImportBookmarks(client, filename, options)
	if err != nil {
		return err
	}

	// Output as JSON
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

func displayImportResult(result *export.ImportResult) {
	// Display summary
	if result.Added > 0 {
		fmt.Fprintf(os.Stderr, "  ✓ %d new bookmarks added\n", result.Added)
	}
	if result.Updated > 0 {
		fmt.Fprintf(os.Stderr, "  ✓ %d existing bookmarks updated\n", result.Updated)
	}
	if result.Skipped > 0 {
		fmt.Fprintf(os.Stderr, "  ⊘ %d skipped (--skip-duplicates)\n", result.Skipped)
	}
	if result.Failed > 0 {
		fmt.Fprintf(os.Stderr, "  ✗ %d failed (see errors below)\n", result.Failed)
	}

	// Display errors
	if len(result.Errors) > 0 {
		fmt.Fprintln(os.Stderr, "\nErrors:")
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "  Line %d: %s\n", e.Line, e.Message)
		}
	}
}
