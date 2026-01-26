package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/models"
	"github.com/spf13/cobra"
)

// bundlesCmd represents the bundles command
var bundlesCmd = &cobra.Command{
	Use:   "bundles",
	Short: "Manage LinkDing bundles (saved search configurations)",
	Long: `Manage bundles in LinkDing. Bundles are saved search configurations
that allow you to quickly filter bookmarks by specific criteria.

Examples:
  linkdingctl bundles list
  linkdingctl bundles get 1
  linkdingctl bundles create "Work" --search "work" --any-tags "project,task"`,
}

var (
	bundleName         string
	bundleSearch       string
	bundleAnyTags      string
	bundleAllTags      string
	bundleExcludedTags string
	bundleOrder        int
)

func init() {
	rootCmd.AddCommand(bundlesCmd)
	bundlesCmd.AddCommand(bundlesListCmd)
	bundlesCmd.AddCommand(bundlesGetCmd)
	bundlesCmd.AddCommand(bundlesCreateCmd)
	bundlesCmd.AddCommand(bundlesUpdateCmd)
	bundlesCmd.AddCommand(bundlesDeleteCmd)

	// Create command flags
	bundlesCreateCmd.Flags().StringVar(&bundleSearch, "search", "", "Search query for the bundle")
	bundlesCreateCmd.Flags().StringVar(&bundleAnyTags, "any-tags", "", "Comma-separated list of tags (any match)")
	bundlesCreateCmd.Flags().StringVar(&bundleAllTags, "all-tags", "", "Comma-separated list of tags (all required)")
	bundlesCreateCmd.Flags().StringVar(&bundleExcludedTags, "excluded-tags", "", "Comma-separated list of tags to exclude")
	bundlesCreateCmd.Flags().IntVar(&bundleOrder, "order", 0, "Display order")

	// Update command flags
	bundlesUpdateCmd.Flags().StringVar(&bundleName, "name", "", "New name for the bundle")
	bundlesUpdateCmd.Flags().StringVar(&bundleSearch, "search", "", "Search query for the bundle")
	bundlesUpdateCmd.Flags().StringVar(&bundleAnyTags, "any-tags", "", "Comma-separated list of tags (any match)")
	bundlesUpdateCmd.Flags().StringVar(&bundleAllTags, "all-tags", "", "Comma-separated list of tags (all required)")
	bundlesUpdateCmd.Flags().StringVar(&bundleExcludedTags, "excluded-tags", "", "Comma-separated list of tags to exclude")
	bundlesUpdateCmd.Flags().IntVar(&bundleOrder, "order", -1, "Display order")
}

// bundlesListCmd represents the bundles list command
var bundlesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all bundles",
	Long: `List all bundles from LinkDing.

Examples:
  linkdingctl bundles list
  linkdingctl bundles list --json`,
	RunE: runBundlesList,
}

func runBundlesList(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Fetch all bundles
	bundles, err := client.FetchAllBundles()
	if err != nil {
		return err
	}

	// Output based on format
	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(bundles)
	}

	return outputBundlesTable(bundles)
}

func outputBundlesTable(bundles []models.Bundle) error {
	if len(bundles) == 0 {
		fmt.Println("No bundles found")
		return nil
	}

	// Create tabwriter for aligned columns
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() { _ = w.Flush() }()

	// Header
	_, _ = fmt.Fprintln(w, "ID\tNAME\tSEARCH\tORDER")
	_, _ = fmt.Fprintln(w, "--\t----\t------\t-----")

	// Rows
	for _, bundle := range bundles {
		search := bundle.Search
		if search == "" {
			search = "-"
		}
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\t%d\n", bundle.ID, bundle.Name, search, bundle.Order)
	}

	_ = w.Flush()

	// Show summary
	fmt.Printf("\nTotal: %d bundles\n", len(bundles))

	return nil
}

// bundlesGetCmd represents the bundles get command
var bundlesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a bundle by ID",
	Long: `Retrieve a single bundle by its ID with full details.

Examples:
  linkdingctl bundles get 1
  linkdingctl bundles get 1 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runBundlesGet,
}

func runBundlesGet(cmd *cobra.Command, args []string) error {
	// Parse bundle ID
	var bundleID int
	if _, err := fmt.Sscanf(args[0], "%d", &bundleID); err != nil {
		return fmt.Errorf("invalid bundle ID: %s (must be a number)", args[0])
	}

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Get the bundle
	bundle, err := client.GetBundle(bundleID)
	if err != nil {
		return err
	}

	// Output based on format
	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(bundle)
	}

	fmt.Printf("Bundle: %s\n", bundle.Name)
	fmt.Printf("  ID: %d\n", bundle.ID)
	fmt.Printf("  Search: %s\n", bundle.Search)
	fmt.Printf("  Any Tags: %s\n", bundle.AnyTags)
	fmt.Printf("  All Tags: %s\n", bundle.AllTags)
	fmt.Printf("  Excluded Tags: %s\n", bundle.ExcludedTags)
	fmt.Printf("  Order: %d\n", bundle.Order)
	fmt.Printf("  Date Created: %s\n", bundle.DateCreated.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Date Modified: %s\n", bundle.DateModified.Format("2006-01-02 15:04:05"))

	return nil
}

// bundlesCreateCmd represents the bundles create command
var bundlesCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new bundle",
	Long: `Create a new bundle (saved search configuration) in LinkDing.

Examples:
  linkdingctl bundles create "Work Projects"
  linkdingctl bundles create "Tech" --search "kubernetes" --any-tags "k8s,docker"
  linkdingctl bundles create "Important" --all-tags "urgent,todo" --order 1`,
	Args: cobra.ExactArgs(1),
	RunE: runBundlesCreate,
}

func runBundlesCreate(cmd *cobra.Command, args []string) error {
	bundleName := args[0]

	// Validate bundle name is not empty
	if bundleName == "" {
		return fmt.Errorf("bundle name cannot be empty")
	}

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Build bundle create request
	bundleCreate := &models.BundleCreate{
		Name:         bundleName,
		Search:       bundleSearch,
		AnyTags:      bundleAnyTags,
		AllTags:      bundleAllTags,
		ExcludedTags: bundleExcludedTags,
		Order:        bundleOrder,
	}

	// Create the bundle
	bundle, err := client.CreateBundle(bundleCreate)
	if err != nil {
		return err
	}

	// Output based on format
	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(bundle)
	}

	fmt.Printf("✓ Bundle created: %s\n", bundle.Name)
	fmt.Printf("  ID: %d\n", bundle.ID)

	return nil
}

// bundlesUpdateCmd represents the bundles update command
var bundlesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a bundle",
	Long: `Update an existing bundle. Only specified fields will be updated (PATCH semantics).

Examples:
  linkdingctl bundles update 1 --name "Renamed Bundle"
  linkdingctl bundles update 1 --search "new search"
  linkdingctl bundles update 1 --any-tags "tag1,tag2" --order 5
  linkdingctl bundles update 1 --excluded-tags "spam"`,
	Args: cobra.ExactArgs(1),
	RunE: runBundlesUpdate,
}

func runBundlesUpdate(cmd *cobra.Command, args []string) error {
	// Parse bundle ID
	var bundleID int
	if _, err := fmt.Sscanf(args[0], "%d", &bundleID); err != nil {
		return fmt.Errorf("invalid bundle ID: %s (must be a number)", args[0])
	}

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Build update request with only specified fields (PATCH semantics)
	update := &models.BundleUpdate{}
	hasUpdates := false

	// Check which flags were set
	if cmd.Flags().Changed("name") {
		update.Name = &bundleName
		hasUpdates = true
	}
	if cmd.Flags().Changed("search") {
		update.Search = &bundleSearch
		hasUpdates = true
	}
	if cmd.Flags().Changed("any-tags") {
		update.AnyTags = &bundleAnyTags
		hasUpdates = true
	}
	if cmd.Flags().Changed("all-tags") {
		update.AllTags = &bundleAllTags
		hasUpdates = true
	}
	if cmd.Flags().Changed("excluded-tags") {
		update.ExcludedTags = &bundleExcludedTags
		hasUpdates = true
	}
	if cmd.Flags().Changed("order") {
		update.Order = &bundleOrder
		hasUpdates = true
	}

	if !hasUpdates {
		return fmt.Errorf("no fields to update (use --name, --search, --any-tags, --all-tags, --excluded-tags, or --order)")
	}

	// Update the bundle
	bundle, err := client.UpdateBundle(bundleID, update)
	if err != nil {
		return err
	}

	// Output based on format
	if jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(bundle)
	}

	fmt.Printf("✓ Bundle updated: %s\n", bundle.Name)
	fmt.Printf("  ID: %d\n", bundle.ID)

	return nil
}

// bundlesDeleteCmd represents the bundles delete command
var bundlesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a bundle",
	Long: `Delete a bundle from LinkDing.

Examples:
  linkdingctl bundles delete 1`,
	Args: cobra.ExactArgs(1),
	RunE: runBundlesDelete,
}

func runBundlesDelete(cmd *cobra.Command, args []string) error {
	// Parse bundle ID
	var bundleID int
	if _, err := fmt.Sscanf(args[0], "%d", &bundleID); err != nil {
		return fmt.Errorf("invalid bundle ID: %s (must be a number)", args[0])
	}

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Create API client
	client := api.NewClient(cfg.URL, cfg.Token)

	// Delete the bundle
	err = client.DeleteBundle(bundleID)
	if err != nil {
		return err
	}

	if jsonOutput {
		fmt.Printf("{\"deleted\": true, \"id\": %d}\n", bundleID)
	} else {
		fmt.Printf("✓ Bundle %d deleted\n", bundleID)
	}

	return nil
}
