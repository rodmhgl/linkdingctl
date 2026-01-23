package export

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/rodstewart/linkding-cli/internal/models"
)

func TestConvertToExportFormat(t *testing.T) {
	now := time.Now()
	bookmarks := []models.Bookmark{
		{
			ID:           1,
			URL:          "https://example.com",
			Title:        "Example",
			Description:  "Test bookmark",
			TagNames:     []string{"tag1", "tag2"},
			DateAdded:    now,
			DateModified: now,
			Unread:       false,
			Shared:       true,
			IsArchived:   false,
		},
		{
			ID:           2,
			URL:          "https://test.com",
			Title:        "Test",
			Description:  "Another bookmark",
			TagNames:     []string{},
			DateAdded:    now,
			DateModified: now,
			Unread:       true,
			Shared:       false,
			IsArchived:   true,
		},
	}

	exported := convertToExportFormat(bookmarks)

	if len(exported) != 2 {
		t.Errorf("Expected 2 exported bookmarks, got %d", len(exported))
	}

	// Check first bookmark
	if exported[0].ID != 1 {
		t.Errorf("Expected ID 1, got %d", exported[0].ID)
	}
	if exported[0].URL != "https://example.com" {
		t.Errorf("Expected URL https://example.com, got %s", exported[0].URL)
	}
	if len(exported[0].Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(exported[0].Tags))
	}
	if exported[0].Shared != true {
		t.Errorf("Expected Shared to be true")
	}
	if exported[0].Archived != false {
		t.Errorf("Expected Archived to be false")
	}

	// Check second bookmark
	if exported[1].ID != 2 {
		t.Errorf("Expected ID 2, got %d", exported[1].ID)
	}
	if exported[1].Unread != true {
		t.Errorf("Expected Unread to be true")
	}
	if exported[1].Archived != true {
		t.Errorf("Expected Archived to be true")
	}
}

func TestExportJSONFormat(t *testing.T) {
	// This is a basic format test without API client
	// We'll create a mock scenario by testing the data structure
	exportData := ExportData{
		Version:    "1",
		ExportedAt: time.Date(2026, 1, 22, 10, 30, 0, 0, time.UTC),
		Source:     "linkding",
		Bookmarks: []ExportBookmark{
			{
				ID:           1,
				URL:          "https://example.com",
				Title:        "Example",
				Description:  "Test",
				Tags:         []string{"tag1"},
				DateAdded:    time.Date(2025, 6, 15, 8, 0, 0, 0, time.UTC),
				DateModified: time.Date(2025, 6, 20, 12, 0, 0, 0, time.UTC),
				Unread:       false,
				Shared:       false,
				Archived:     false,
			},
		},
	}

	// Encode to JSON
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(exportData); err != nil {
		t.Fatalf("Failed to encode JSON: %v", err)
	}

	// Verify it's valid JSON by decoding it back
	var decoded ExportData
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	// Verify key fields
	if decoded.Version != "1" {
		t.Errorf("Expected version 1, got %s", decoded.Version)
	}
	if decoded.Source != "linkding" {
		t.Errorf("Expected source linkding, got %s", decoded.Source)
	}
	if len(decoded.Bookmarks) != 1 {
		t.Errorf("Expected 1 bookmark, got %d", len(decoded.Bookmarks))
	}
	if decoded.Bookmarks[0].URL != "https://example.com" {
		t.Errorf("Expected URL https://example.com, got %s", decoded.Bookmarks[0].URL)
	}
}
