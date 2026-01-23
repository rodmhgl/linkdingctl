package export

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/rodstewart/linkding-cli/internal/api"
)

// ExportHTML exports bookmarks to Netscape bookmark format (HTML)
func ExportHTML(client *api.Client, writer io.Writer, options ExportOptions) error {
	// Fetch all bookmarks using the Client's pagination method
	bookmarks, err := client.FetchAllBookmarks(options.Tags, options.IncludeArchived)
	if err != nil {
		return err
	}

	// Write HTML header
	if _, err := fmt.Fprintf(writer, "<!DOCTYPE NETSCAPE-Bookmark-file-1>\n"); err != nil {
		return fmt.Errorf("failed to write HTML header: %w", err)
	}
	if _, err := fmt.Fprintf(writer, "<META HTTP-EQUIV=\"Content-Type\" CONTENT=\"text/html; charset=UTF-8\">\n"); err != nil {
		return fmt.Errorf("failed to write HTML meta: %w", err)
	}
	if _, err := fmt.Fprintf(writer, "<TITLE>Bookmarks</TITLE>\n"); err != nil {
		return fmt.Errorf("failed to write HTML title: %w", err)
	}
	if _, err := fmt.Fprintf(writer, "<H1>Bookmarks</H1>\n"); err != nil {
		return fmt.Errorf("failed to write HTML heading: %w", err)
	}
	if _, err := fmt.Fprintf(writer, "<DL><p>\n"); err != nil {
		return fmt.Errorf("failed to write HTML list start: %w", err)
	}

	// Write each bookmark
	for _, b := range bookmarks {
		// Convert date_added to Unix timestamp
		addDate := b.DateAdded.Unix()

		// Escape HTML in title and URL
		escapedURL := html.EscapeString(b.URL)
		escapedTitle := html.EscapeString(b.Title)

		// Build tags string (comma-separated)
		tags := strings.Join(b.TagNames, ",")
		escapedTags := html.EscapeString(tags)

		// Write bookmark entry
		if _, err := fmt.Fprintf(writer, "    <DT><A HREF=\"%s\" ADD_DATE=\"%d\"", escapedURL, addDate); err != nil {
			return fmt.Errorf("failed to write bookmark entry: %w", err)
		}

		// Add tags attribute if there are tags
		if len(b.TagNames) > 0 {
			if _, err := fmt.Fprintf(writer, " TAGS=\"%s\"", escapedTags); err != nil {
				return fmt.Errorf("failed to write tags: %w", err)
			}
		}

		// Close the anchor tag and write title
		if _, err := fmt.Fprintf(writer, ">%s</A>\n", escapedTitle); err != nil {
			return fmt.Errorf("failed to write bookmark title: %w", err)
		}

		// Write description if present
		if b.Description != "" {
			escapedDesc := html.EscapeString(b.Description)
			if _, err := fmt.Fprintf(writer, "    <DD>%s\n", escapedDesc); err != nil {
				return fmt.Errorf("failed to write description: %w", err)
			}
		}
	}

	// Write HTML footer
	if _, err := fmt.Fprintf(writer, "</DL><p>\n"); err != nil {
		return fmt.Errorf("failed to write HTML list end: %w", err)
	}

	return nil
}
