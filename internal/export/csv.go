package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/rodstewart/linkding-cli/internal/api"
)

// ExportCSV exports bookmarks to CSV format
func ExportCSV(client *api.Client, writer io.Writer, options ExportOptions) error {
	// Fetch all bookmarks using the Client's pagination method
	bookmarks, err := client.FetchAllBookmarks(options.Tags, options.IncludeArchived)
	if err != nil {
		return err
	}

	// Create CSV writer
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header row
	header := []string{
		"url",
		"title",
		"description",
		"tags",
		"date_added",
		"unread",
		"shared",
		"archived",
	}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write each bookmark as a row
	for _, b := range bookmarks {
		// Join tags with comma
		tags := strings.Join(b.TagNames, ",")

		// Format date as RFC3339
		dateAdded := b.DateAdded.Format("2006-01-02T15:04:05Z07:00")

		// Convert booleans to strings
		unread := strconv.FormatBool(b.Unread)
		shared := strconv.FormatBool(b.Shared)
		archived := strconv.FormatBool(b.IsArchived)

		// Create row
		row := []string{
			b.URL,
			b.Title,
			b.Description,
			tags,
			dateAdded,
			unread,
			shared,
			archived,
		}

		// Write row
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
