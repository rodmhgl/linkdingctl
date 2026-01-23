package export

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/models"
)

// ExportBookmark represents a bookmark in the export format
type ExportBookmark struct {
	ID           int       `json:"id"`
	URL          string    `json:"url"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Tags         []string  `json:"tags"`
	DateAdded    time.Time `json:"date_added"`
	DateModified time.Time `json:"date_modified"`
	Unread       bool      `json:"unread"`
	Shared       bool      `json:"shared"`
	Archived     bool      `json:"archived"`
}

// ExportData represents the complete export data structure
type ExportData struct {
	Version    string           `json:"version"`
	ExportedAt time.Time        `json:"exported_at"`
	Source     string           `json:"source"`
	Bookmarks  []ExportBookmark `json:"bookmarks"`
}

// ExportOptions configures the export behavior
type ExportOptions struct {
	Tags            []string
	IncludeArchived bool
}

// convertToExportFormat converts internal bookmark models to export format
func convertToExportFormat(bookmarks []models.Bookmark) []ExportBookmark {
	exported := make([]ExportBookmark, len(bookmarks))
	for i, b := range bookmarks {
		exported[i] = ExportBookmark{
			ID:           b.ID,
			URL:          b.URL,
			Title:        b.Title,
			Description:  b.Description,
			Tags:         b.TagNames,
			DateAdded:    b.DateAdded,
			DateModified: b.DateModified,
			Unread:       b.Unread,
			Shared:       b.Shared,
			Archived:     b.IsArchived,
		}
	}
	return exported
}

// ExportJSON exports bookmarks to JSON format
func ExportJSON(client *api.Client, writer io.Writer, options ExportOptions) error {
	// Fetch all bookmarks using the Client's pagination method
	bookmarks, err := client.FetchAllBookmarks(options.Tags, options.IncludeArchived)
	if err != nil {
		return err
	}

	// Convert to export format
	exportBookmarks := convertToExportFormat(bookmarks)

	// Create export data structure
	data := ExportData{
		Version:    "1",
		ExportedAt: time.Now().UTC(),
		Source:     "linkding",
		Bookmarks:  exportBookmarks,
	}

	// Encode to JSON with indentation
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
