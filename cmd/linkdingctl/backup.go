package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/export"
	"github.com/spf13/cobra"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a timestamped backup of all bookmarks",
	Long: `Create a timestamped JSON backup of all bookmarks.

The backup file is saved with a timestamp in the filename:
  linkding-backup-2026-01-22T103000.json

This is equivalent to running:
  ld export -f json -o <timestamped-file>

Examples:
  ld backup
  ld backup -o ~/backups/
  ld backup --prefix my-backup`,
	RunE: runBackup,
}

var (
	backupOutput string
	backupPrefix string
)

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().StringVarP(&backupOutput, "output", "o", ".", "Output directory (default: current directory)")
	backupCmd.Flags().StringVar(&backupPrefix, "prefix", "linkding-backup", "Filename prefix")
}

func runBackup(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Generate timestamped filename
	timestamp := time.Now().Format("2006-01-02T150405")
	filename := fmt.Sprintf("%s-%s.json", backupPrefix, timestamp)
	fullPath := filepath.Join(backupOutput, filename)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(backupOutput, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	// Export all bookmarks to JSON
	options := export.ExportOptions{
		Tags:            []string{},
		IncludeArchived: true,
	}

	if err := export.ExportJSON(client, file, options); err != nil {
		// Remove partial file on error
		os.Remove(fullPath)
		return fmt.Errorf("failed to export bookmarks: %w", err)
	}

	// Success message
	if !jsonOutput {
		fmt.Fprintf(os.Stderr, "Backup created: %s\n", fullPath)
	} else {
		// JSON output with proper escaping
		if err := json.NewEncoder(os.Stdout).Encode(map[string]string{"file": fullPath}); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
	}

	return nil
}
