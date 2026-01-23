package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/spf13/cobra"
)

var forceDelete bool

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a bookmark by ID",
	Long: `Delete a bookmark by ID. Requires confirmation unless --force or --json flag is set.

Examples:
  ld delete 123
  ld delete 123 --force
  ld delete 123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "skip confirmation prompt")
}

func runDelete(cmd *cobra.Command, args []string) error {
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

	// Get bookmark details for confirmation (unless force or json mode)
	if !forceDelete && !jsonOutput {
		bookmark, err := client.GetBookmark(id)
		if err != nil {
			return err
		}

		// Show bookmark details and ask for confirmation
		fmt.Printf("About to delete bookmark:\n")
		fmt.Printf("  ID:    %d\n", bookmark.ID)
		fmt.Printf("  Title: %s\n", bookmark.Title)
		fmt.Printf("  URL:   %s\n", bookmark.URL)
		fmt.Printf("\nAre you sure? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Delete cancelled")
			return nil
		}
	}

	// Delete the bookmark
	err = client.DeleteBookmark(id)
	if err != nil {
		return err
	}

	// Output based on format
	if jsonOutput {
		fmt.Printf("{\"deleted\": true, \"id\": %d}\n", id)
	} else {
		fmt.Printf("✓ Bookmark %d deleted\n", id)
	}

	return nil
}
