package main

import (
	"fmt"
	"os"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/config"
	"github.com/rodstewart/linkding-cli/internal/export"
	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export bookmarks",
	Long: `Export bookmarks to various formats (JSON, HTML, CSV).

Examples:
  ld export > bookmarks.json
  ld export -f html -o bookmarks.html
  ld export --tags homelab -f csv -o homelab.csv`,
	RunE: runExport,
}

var (
	exportFormat   string
	exportOutput   string
	exportTags     []string
	exportArchived bool
)

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "Output format: json, html, csv")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (default: stdout)")
	exportCmd.Flags().StringSliceVarP(&exportTags, "tags", "T", []string{}, "Export only bookmarks with these tags")
	exportCmd.Flags().BoolVar(&exportArchived, "archived", true, "Include archived bookmarks")
}

func runExport(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Validate format
	switch exportFormat {
	case "json":
		// JSON export is implemented
	case "html", "csv":
		return fmt.Errorf("export format '%s' is not yet implemented", exportFormat)
	default:
		return fmt.Errorf("invalid export format '%s'. Valid formats: json, html, csv", exportFormat)
	}

	// Determine output writer
	var writer *os.File
	if exportOutput == "" {
		writer = os.Stdout
	} else {
		file, err := os.Create(exportOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	// Create export options
	options := export.ExportOptions{
		Tags:            exportTags,
		IncludeArchived: exportArchived,
	}

	// Perform export based on format
	switch exportFormat {
	case "json":
		if err := export.ExportJSON(client, writer, options); err != nil {
			return err
		}
	}

	// Print success message to stderr if writing to file
	if exportOutput != "" {
		fmt.Fprintf(os.Stderr, "Exported bookmarks to %s\n", exportOutput)
	}

	return nil
}
