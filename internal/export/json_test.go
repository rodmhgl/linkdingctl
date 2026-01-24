package export

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rodstewart/linkding-cli/internal/api"
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

// TestExportJSON_WithMockServer tests ExportJSON with actual API client
func TestExportJSON_WithMockServer(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	bookmarks := []models.Bookmark{
		{
			ID:           1,
			URL:          "https://example.com",
			Title:        "Example",
			Description:  "Test bookmark",
			TagNames:     []string{"tag1", "tag2"},
			DateAdded:    testTime,
			DateModified: testTime,
			Unread:       false,
			Shared:       true,
			IsArchived:   false,
		},
		{
			ID:           2,
			URL:          "https://test.com",
			Title:        "Test",
			Description:  "",
			TagNames:     []string{},
			DateAdded:    testTime,
			DateModified: testTime,
			Unread:       true,
			Shared:       false,
			IsArchived:   true,
		},
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := models.BookmarkList{
			Count:   2,
			Next:    nil,
			Results: bookmarks,
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client and export
	client := api.NewClient(server.URL, "test-token")
	var buf bytes.Buffer
	err := ExportJSON(client, &buf, ExportOptions{})

	if err != nil {
		t.Fatalf("ExportJSON() failed: %v", err)
	}

	// Decode the exported JSON
	var exported ExportData
	if err := json.Unmarshal(buf.Bytes(), &exported); err != nil {
		t.Fatalf("Failed to decode exported JSON: %v", err)
	}

	// Verify export data
	if exported.Version != "1" {
		t.Errorf("Expected version 1, got %s", exported.Version)
	}
	if exported.Source != "linkding" {
		t.Errorf("Expected source linkding, got %s", exported.Source)
	}
	if len(exported.Bookmarks) != 2 {
		t.Errorf("Expected 2 bookmarks, got %d", len(exported.Bookmarks))
	}

	// Verify first bookmark
	if exported.Bookmarks[0].URL != "https://example.com" {
		t.Errorf("Expected URL https://example.com, got %s", exported.Bookmarks[0].URL)
	}
	if exported.Bookmarks[0].Title != "Example" {
		t.Errorf("Expected title Example, got %s", exported.Bookmarks[0].Title)
	}
	if len(exported.Bookmarks[0].Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(exported.Bookmarks[0].Tags))
	}
	if !exported.Bookmarks[0].Shared {
		t.Error("Expected Shared to be true")
	}

	// Verify second bookmark
	if exported.Bookmarks[1].URL != "https://test.com" {
		t.Errorf("Expected URL https://test.com, got %s", exported.Bookmarks[1].URL)
	}
	if !exported.Bookmarks[1].Unread {
		t.Error("Expected Unread to be true")
	}
	if !exported.Bookmarks[1].Archived {
		t.Error("Expected Archived to be true")
	}
}

// TestExportJSON_WithTags tests ExportJSON with tag filtering
func TestExportJSON_WithTags(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create mock server that checks for tag filter
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify tag filter is passed via query parameter (tags are added to "q" parameter)
		query := r.URL.Query()
		qParam := query.Get("q")
		if !strings.Contains(qParam, "test-tag") {
			t.Errorf("Expected 'test-tag' in q parameter, got %s", qParam)
		}

		bookmarks := []models.Bookmark{
			{
				ID:        1,
				URL:       "https://example.com",
				Title:     "Example",
				TagNames:  []string{"test-tag"},
				DateAdded: testTime,
			},
		}

		response := models.BookmarkList{
			Count:   1,
			Next:    nil,
			Results: bookmarks,
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client and export with tags filter
	client := api.NewClient(server.URL, "test-token")
	var buf bytes.Buffer
	err := ExportJSON(client, &buf, ExportOptions{
		Tags: []string{"test-tag"},
	})

	if err != nil {
		t.Fatalf("ExportJSON() with tags failed: %v", err)
	}

	// Verify export succeeded
	var exported ExportData
	if err := json.Unmarshal(buf.Bytes(), &exported); err != nil {
		t.Fatalf("Failed to decode exported JSON: %v", err)
	}

	if len(exported.Bookmarks) != 1 {
		t.Errorf("Expected 1 bookmark, got %d", len(exported.Bookmarks))
	}
}

// TestJSON_RoundTrip tests that JSON export→import preserves all bookmark data
func TestJSON_RoundTrip(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	// Original bookmarks to export
	originalBookmarks := []models.Bookmark{
		{
			ID:           1,
			URL:          "https://example.com",
			Title:        "Example Site",
			Description:  "A test bookmark with all fields",
			Notes:        "Some notes",
			TagNames:     []string{"test", "example", "golang"},
			DateAdded:    testTime,
			DateModified: testTime,
			Unread:       false,
			Shared:       true,
			IsArchived:   false,
		},
		{
			ID:           2,
			URL:          "https://test.org",
			Title:        "Test Organization",
			Description:  "Another bookmark",
			Notes:        "",
			TagNames:     []string{"test"},
			DateAdded:    testTime,
			DateModified: testTime,
			Unread:       true,
			Shared:       false,
			IsArchived:   true,
		},
		{
			ID:           3,
			URL:          "https://minimal.com",
			Title:        "Minimal",
			Description:  "",
			Notes:        "",
			TagNames:     []string{},
			DateAdded:    testTime,
			DateModified: testTime,
			Unread:       false,
			Shared:       false,
			IsArchived:   false,
		},
	}

	// Step 1: Export bookmarks to JSON
	var exportBuf bytes.Buffer

	// Create mock server for export (returns original bookmarks)
	exportServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := models.BookmarkList{
			Count:   len(originalBookmarks),
			Next:    nil,
			Results: originalBookmarks,
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode export response: %v", err)
		}
	}))
	defer exportServer.Close()

	exportClient := api.NewClient(exportServer.URL, "test-token")
	if err := ExportJSON(exportClient, &exportBuf, ExportOptions{}); err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Verify export succeeded
	var exportedData ExportData
	if err := json.Unmarshal(exportBuf.Bytes(), &exportedData); err != nil {
		t.Fatalf("Failed to unmarshal exported JSON: %v", err)
	}
	if len(exportedData.Bookmarks) != 3 {
		t.Fatalf("Expected 3 exported bookmarks, got %d", len(exportedData.Bookmarks))
	}

	// Step 2: Import bookmarks back
	createdBookmarks := make([]models.Bookmark, 0)
	importCallCount := 0

	// Create mock server for import
	importServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/api/bookmarks/" {
			// FetchAllBookmarks call to check for duplicates
			// Return empty list (no existing bookmarks)
			response := models.BookmarkList{
				Count:   0,
				Next:    nil,
				Results: []models.Bookmark{},
			}
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		} else if r.Method == "POST" && r.URL.Path == "/api/bookmarks/" {
			// CreateBookmark call
			var create models.BookmarkCreate
			if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
				t.Fatalf("Failed to decode create request: %v", err)
			}

			// Create response bookmark with assigned ID
			created := models.Bookmark{
				ID:           importCallCount + 1,
				URL:          create.URL,
				Title:        create.Title,
				Description:  create.Description,
				Notes:        create.Notes,
				TagNames:     create.TagNames,
				DateAdded:    testTime,
				DateModified: testTime,
				Unread:       create.Unread,
				Shared:       create.Shared,
				IsArchived:   create.IsArchived,
			}
			createdBookmarks = append(createdBookmarks, created)
			importCallCount++

			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(created); err != nil {
				t.Errorf("Failed to encode created bookmark: %v", err)
			}
		}
	}))
	defer importServer.Close()

	importClient := api.NewClient(importServer.URL, "test-token")
	importResult, err := importJSON(importClient, bytes.NewReader(exportBuf.Bytes()), ImportOptions{})
	if err != nil {
		t.Fatalf("importJSON failed: %v", err)
	}

	// Verify import results
	if importResult.Added != 3 {
		t.Errorf("Expected 3 bookmarks added, got %d", importResult.Added)
	}
	if importResult.Failed != 0 {
		t.Errorf("Expected 0 failures, got %d", importResult.Failed)
		for _, e := range importResult.Errors {
			t.Logf("  Line %d: %s", e.Line, e.Message)
		}
	}

	// Step 3: Verify round-trip data integrity
	if len(createdBookmarks) != 3 {
		t.Fatalf("Expected 3 created bookmarks, got %d", len(createdBookmarks))
	}

	// Compare each bookmark
	for i, original := range originalBookmarks {
		created := createdBookmarks[i]

		if created.URL != original.URL {
			t.Errorf("Bookmark %d: URL mismatch: expected %s, got %s", i, original.URL, created.URL)
		}
		if created.Title != original.Title {
			t.Errorf("Bookmark %d: Title mismatch: expected %s, got %s", i, original.Title, created.Title)
		}
		if created.Description != original.Description {
			t.Errorf("Bookmark %d: Description mismatch: expected %s, got %s", i, original.Description, created.Description)
		}
		if len(created.TagNames) != len(original.TagNames) {
			t.Errorf("Bookmark %d: Tag count mismatch: expected %d, got %d", i, len(original.TagNames), len(created.TagNames))
		} else {
			for j, tag := range original.TagNames {
				if created.TagNames[j] != tag {
					t.Errorf("Bookmark %d: Tag %d mismatch: expected %s, got %s", i, j, tag, created.TagNames[j])
				}
			}
		}
		if created.Unread != original.Unread {
			t.Errorf("Bookmark %d: Unread mismatch: expected %v, got %v", i, original.Unread, created.Unread)
		}
		if created.Shared != original.Shared {
			t.Errorf("Bookmark %d: Shared mismatch: expected %v, got %v", i, original.Shared, created.Shared)
		}
		if created.IsArchived != original.IsArchived {
			t.Errorf("Bookmark %d: IsArchived mismatch: expected %v, got %v", i, original.IsArchived, created.IsArchived)
		}
	}
}
