// Package export handles importing and exporting bookmarks in various formats.
package export

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/models"
)

// ImportResult tracks the outcome of an import operation
type ImportResult struct {
	Added   int
	Updated int
	Skipped int
	Failed  int
	Errors  []ImportError
}

// ImportError represents a single import failure
type ImportError struct {
	Line    int
	Message string
}

// ImportOptions configures the import behavior
type ImportOptions struct {
	Format         string // json, html, csv, or auto
	DryRun         bool
	SkipDuplicates bool
	AddTags        []string
}

// DetectFormat determines the import format from the file extension
func DetectFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return "json"
	case ".html", ".htm":
		return "html"
	case ".csv":
		return "csv"
	default:
		return ""
	}
}

// ImportBookmarks imports bookmarks from a file
func ImportBookmarks(client *api.Client, filename string, options ImportOptions) (*ImportResult, error) {
	// Auto-detect format if not specified
	format := options.Format
	if format == "" || format == "auto" {
		format = DetectFormat(filename)
		if format == "" {
			return nil, fmt.Errorf("cannot detect format from file extension. Use --format flag")
		}
	}

	// Open file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Import based on format
	switch format {
	case "json":
		return importJSON(client, file, options)
	case "html":
		return importHTML(client, file, options)
	case "csv":
		return importCSV(client, file, options)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// importJSON imports bookmarks from JSON format
func importJSON(client *api.Client, reader io.Reader, options ImportOptions) (*ImportResult, error) {
	var data ExportData
	if err := json.NewDecoder(reader).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	result := &ImportResult{}

	// Get existing bookmarks to check for duplicates
	existingURLs := make(map[string]int)
	if !options.DryRun {
		existing, err := client.FetchAllBookmarks(nil, true)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch existing bookmarks: %w", err)
		}
		for _, b := range existing {
			existingURLs[b.URL] = b.ID
		}
	}

	// Import each bookmark
	for i, exportBookmark := range data.Bookmarks {
		lineNum := i + 1

		// Validate required fields
		if exportBookmark.URL == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Line:    lineNum,
				Message: "Missing required field \"url\"",
			})
			continue
		}

		// Prepare bookmark for creation
		tags := exportBookmark.Tags
		if len(options.AddTags) > 0 {
			tags = append(tags, options.AddTags...)
		}

		bookmarkCreate := &models.BookmarkCreate{
			URL:         exportBookmark.URL,
			Title:       exportBookmark.Title,
			Description: exportBookmark.Description,
			TagNames:    tags,
			IsArchived:  exportBookmark.Archived,
			Unread:      exportBookmark.Unread,
			Shared:      exportBookmark.Shared,
		}

		// Check for duplicates
		existingID, exists := existingURLs[exportBookmark.URL]

		if exists && options.SkipDuplicates {
			result.Skipped++
			continue
		}

		if options.DryRun {
			if exists {
				result.Updated++
			} else {
				result.Added++
			}
			continue
		}

		// Create or update bookmark
		if exists {
			// Update existing bookmark
			update := &models.BookmarkUpdate{
				URL:         &bookmarkCreate.URL,
				Title:       &bookmarkCreate.Title,
				Description: &bookmarkCreate.Description,
				TagNames:    &bookmarkCreate.TagNames,
				IsArchived:  &bookmarkCreate.IsArchived,
				Unread:      &bookmarkCreate.Unread,
				Shared:      &bookmarkCreate.Shared,
			}
			_, err := client.UpdateBookmark(existingID, update)
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, ImportError{
					Line:    lineNum,
					Message: fmt.Sprintf("Failed to update: %v", err),
				})
				continue
			}
			result.Updated++
		} else {
			// Create new bookmark
			_, err := client.CreateBookmark(bookmarkCreate)
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, ImportError{
					Line:    lineNum,
					Message: fmt.Sprintf("Failed to create: %v", err),
				})
				continue
			}
			result.Added++
		}
	}

	return result, nil
}

// importHTML imports bookmarks from Netscape HTML bookmark format
func importHTML(client *api.Client, reader io.Reader, options ImportOptions) (*ImportResult, error) {
	result := &ImportResult{}

	// Get existing bookmarks to check for duplicates
	existingURLs := make(map[string]int)
	if !options.DryRun {
		existing, err := client.FetchAllBookmarks(nil, true)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch existing bookmarks: %w", err)
		}
		for _, b := range existing {
			existingURLs[b.URL] = b.ID
		}
	}

	// Regular expressions for parsing Netscape bookmark format
	linkPattern := regexp.MustCompile(`<DT><A[^>]+HREF="([^"]+)"([^>]*)>([^<]*)</A>`)
	tagsPattern := regexp.MustCompile(`TAGS="([^"]+)"`)
	descPattern := regexp.MustCompile(`<DD>([^\n<]+)`)

	scanner := bufio.NewScanner(reader)
	lineNum := 0
	var lastURL, lastTitle, lastTags string

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Check for link tag
		if matches := linkPattern.FindStringSubmatch(line); matches != nil {
			// Process previous bookmark if we found a new one
			if lastURL != "" {
				if err := processHTMLBookmark(client, result, existingURLs, lastURL, lastTitle, lastTags, "", lineNum-1, options); err != nil {
					// Error already recorded in result
				}
			}

			// Extract new bookmark data
			lastURL = matches[1]
			attrs := matches[2]
			lastTitle = matches[3]

			// Extract tags if present
			if tagMatches := tagsPattern.FindStringSubmatch(attrs); tagMatches != nil {
				lastTags = tagMatches[1]
			} else {
				lastTags = ""
			}
		} else if matches := descPattern.FindStringSubmatch(line); matches != nil && lastURL != "" {
			// Found description for current bookmark
			description := matches[1]
			if err := processHTMLBookmark(client, result, existingURLs, lastURL, lastTitle, lastTags, description, lineNum, options); err != nil {
				// Error already recorded in result
			}
			lastURL = ""
			lastTitle = ""
			lastTags = ""
		}
	}

	// Process last bookmark if it didn't have a description
	if lastURL != "" {
		if err := processHTMLBookmark(client, result, existingURLs, lastURL, lastTitle, lastTags, "", lineNum, options); err != nil {
			// Error already recorded in result
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read HTML: %w", err)
	}

	return result, nil
}

// processHTMLBookmark processes a single bookmark from HTML import
func processHTMLBookmark(client *api.Client, result *ImportResult, existingURLs map[string]int,
	url, title, tagsStr, description string, lineNum int, options ImportOptions) error {

	// Parse tags
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	// Add custom tags
	if len(options.AddTags) > 0 {
		tags = append(tags, options.AddTags...)
	}

	bookmarkCreate := &models.BookmarkCreate{
		URL:         url,
		Title:       title,
		Description: description,
		TagNames:    tags,
	}

	// Check for duplicates
	existingID, exists := existingURLs[url]

	if exists && options.SkipDuplicates {
		result.Skipped++
		return nil
	}

	if options.DryRun {
		if exists {
			result.Updated++
		} else {
			result.Added++
		}
		return nil
	}

	// Create or update bookmark
	if exists {
		update := &models.BookmarkUpdate{
			URL:         &bookmarkCreate.URL,
			Title:       &bookmarkCreate.Title,
			Description: &bookmarkCreate.Description,
			TagNames:    &bookmarkCreate.TagNames,
		}
		_, err := client.UpdateBookmark(existingID, update)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Line:    lineNum,
				Message: fmt.Sprintf("Failed to update: %v", err),
			})
			return err
		}
		result.Updated++
	} else {
		_, err := client.CreateBookmark(bookmarkCreate)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Line:    lineNum,
				Message: fmt.Sprintf("Failed to create: %v", err),
			})
			return err
		}
		result.Added++
	}

	return nil
}

// importCSV imports bookmarks from CSV format
func importCSV(client *api.Client, reader io.Reader, options ImportOptions) (*ImportResult, error) {
	csvReader := csv.NewReader(reader)

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Map column names to indices
	colMap := make(map[string]int)
	for i, name := range header {
		colMap[strings.ToLower(strings.TrimSpace(name))] = i
	}

	result := &ImportResult{}

	// Get existing bookmarks to check for duplicates
	existingURLs := make(map[string]int)
	if !options.DryRun {
		existing, err := client.FetchAllBookmarks(nil, true)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch existing bookmarks: %w", err)
		}
		for _, b := range existing {
			existingURLs[b.URL] = b.ID
		}
	}

	lineNum := 1 // Start at 1 (header row)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Line:    lineNum + 1,
				Message: fmt.Sprintf("Failed to parse CSV: %v", err),
			})
			lineNum++
			continue
		}
		lineNum++

		// Extract fields
		url := getCSVField(record, colMap, "url")
		if url == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Line:    lineNum,
				Message: "Missing required field \"url\"",
			})
			continue
		}

		title := getCSVField(record, colMap, "title")
		description := getCSVField(record, colMap, "description")
		tagsStr := getCSVField(record, colMap, "tags")

		var tags []string
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
		}

		// Parse boolean fields
		unread := parseCSVBool(getCSVField(record, colMap, "unread"))
		shared := parseCSVBool(getCSVField(record, colMap, "shared"))
		archived := parseCSVBool(getCSVField(record, colMap, "archived"))

		// Add custom tags
		if len(options.AddTags) > 0 {
			tags = append(tags, options.AddTags...)
		}

		bookmarkCreate := &models.BookmarkCreate{
			URL:         url,
			Title:       title,
			Description: description,
			TagNames:    tags,
			Unread:      unread,
			Shared:      shared,
			IsArchived:  archived,
		}

		// Check for duplicates
		existingID, exists := existingURLs[url]

		if exists && options.SkipDuplicates {
			result.Skipped++
			continue
		}

		if options.DryRun {
			if exists {
				result.Updated++
			} else {
				result.Added++
			}
			continue
		}

		// Create or update bookmark
		if exists {
			update := &models.BookmarkUpdate{
				URL:         &bookmarkCreate.URL,
				Title:       &bookmarkCreate.Title,
				Description: &bookmarkCreate.Description,
				TagNames:    &bookmarkCreate.TagNames,
				Unread:      &bookmarkCreate.Unread,
				Shared:      &bookmarkCreate.Shared,
				IsArchived:  &bookmarkCreate.IsArchived,
			}
			_, err := client.UpdateBookmark(existingID, update)
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, ImportError{
					Line:    lineNum,
					Message: fmt.Sprintf("Failed to update: %v", err),
				})
				continue
			}
			result.Updated++
		} else {
			_, err := client.CreateBookmark(bookmarkCreate)
			if err != nil {
				result.Failed++
				result.Errors = append(result.Errors, ImportError{
					Line:    lineNum,
					Message: fmt.Sprintf("Failed to create: %v", err),
				})
				continue
			}
			result.Added++
		}
	}

	return result, nil
}

// getCSVField safely retrieves a field from a CSV record
func getCSVField(record []string, colMap map[string]int, fieldName string) string {
	if idx, ok := colMap[fieldName]; ok && idx < len(record) {
		return record[idx]
	}
	return ""
}

// parseCSVBool parses a boolean value from CSV
func parseCSVBool(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "true" || value == "1" || value == "yes"
}
