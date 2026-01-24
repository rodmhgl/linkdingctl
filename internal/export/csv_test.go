package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/models"
)

func TestExportCSVFormat(t *testing.T) {
	// Create test bookmarks
	testBookmarks := []models.Bookmark{
		{
			ID:           123,
			URL:          "https://example.com",
			Title:        "Example",
			Description:  "Notes here",
			TagNames:     []string{"tag1", "tag2"},
			DateAdded:    time.Date(2025, 6, 15, 8, 0, 0, 0, time.UTC),
			DateModified: time.Date(2025, 6, 20, 12, 0, 0, 0, time.UTC),
			Unread:       false,
			Shared:       false,
			IsArchived:   false,
		},
		{
			ID:           124,
			URL:          "https://test.com",
			Title:        "Test Site",
			Description:  "Test description",
			TagNames:     []string{},
			DateAdded:    time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC),
			DateModified: time.Date(2025, 7, 2, 11, 0, 0, 0, time.UTC),
			Unread:       true,
			Shared:       true,
			IsArchived:   true,
		},
	}

	// Create a buffer to write CSV to
	var buf bytes.Buffer

	// Create CSV writer
	csvWriter := csv.NewWriter(&buf)

	// Write header
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
		t.Fatalf("Failed to write header: %v", err)
	}

	// Write test data
	for _, b := range testBookmarks {
		tags := strings.Join(b.TagNames, ",")
		dateAdded := b.DateAdded.Format("2006-01-02T15:04:05Z07:00")
		row := []string{
			b.URL,
			b.Title,
			b.Description,
			tags,
			dateAdded,
			"false", // unread
			"false", // shared
			"false", // archived
		}
		if b.Unread {
			row[5] = "true"
		}
		if b.Shared {
			row[6] = "true"
		}
		if b.IsArchived {
			row[7] = "true"
		}
		if err := csvWriter.Write(row); err != nil {
			t.Fatalf("Failed to write row: %v", err)
		}
	}
	csvWriter.Flush()

	// Parse the CSV back to verify format
	csvReader := csv.NewReader(&buf)
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV: %v", err)
	}

	// Verify we have header + 2 data rows
	if len(records) != 3 {
		t.Errorf("Expected 3 records (1 header + 2 data), got %d", len(records))
	}

	// Verify header
	expectedHeader := []string{"url", "title", "description", "tags", "date_added", "unread", "shared", "archived"}
	for i, h := range expectedHeader {
		if records[0][i] != h {
			t.Errorf("Header column %d: expected %s, got %s", i, h, records[0][i])
		}
	}

	// Verify first data row
	if records[1][0] != "https://example.com" {
		t.Errorf("Expected URL https://example.com, got %s", records[1][0])
	}
	if records[1][1] != "Example" {
		t.Errorf("Expected title Example, got %s", records[1][1])
	}
	if records[1][3] != "tag1,tag2" {
		t.Errorf("Expected tags 'tag1,tag2', got %s", records[1][3])
	}
	if records[1][5] != "false" {
		t.Errorf("Expected unread false, got %s", records[1][5])
	}

	// Verify second data row
	if records[2][0] != "https://test.com" {
		t.Errorf("Expected URL https://test.com, got %s", records[2][0])
	}
	if records[2][3] != "" {
		t.Errorf("Expected empty tags, got %s", records[2][3])
	}
	if records[2][5] != "true" {
		t.Errorf("Expected unread true, got %s", records[2][5])
	}
	if records[2][6] != "true" {
		t.Errorf("Expected shared true, got %s", records[2][6])
	}
	if records[2][7] != "true" {
		t.Errorf("Expected archived true, got %s", records[2][7])
	}
}

func TestCSVHandlesSpecialCharacters(t *testing.T) {
	testBookmarks := []models.Bookmark{
		{
			ID:          1,
			URL:         "https://example.com",
			Title:       "Title with, comma",
			Description: "Description with \"quotes\" and, commas",
			TagNames:    []string{"tag1", "tag2"},
			DateAdded:   time.Date(2025, 6, 15, 8, 0, 0, 0, time.UTC),
		},
	}

	var buf bytes.Buffer
	csvWriter := csv.NewWriter(&buf)

	// Write header
	header := []string{"url", "title", "description", "tags", "date_added", "unread", "shared", "archived"}
	if err := csvWriter.Write(header); err != nil {
		t.Fatalf("Failed to write CSV header: %v", err)
	}

	// Write data
	for _, b := range testBookmarks {
		row := []string{
			b.URL,
			b.Title,
			b.Description,
			strings.Join(b.TagNames, ","),
			b.DateAdded.Format("2006-01-02T15:04:05Z07:00"),
			"false",
			"false",
			"false",
		}
		if err := csvWriter.Write(row); err != nil {
			t.Fatalf("Failed to write CSV row: %v", err)
		}
	}
	csvWriter.Flush()

	// Parse back
	csvReader := csv.NewReader(&buf)
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV with special characters: %v", err)
	}

	// Verify special characters are properly escaped
	if records[1][1] != "Title with, comma" {
		t.Errorf("Expected title with comma to be preserved, got %s", records[1][1])
	}
	if records[1][2] != "Description with \"quotes\" and, commas" {
		t.Errorf("Expected description with quotes to be preserved, got %s", records[1][2])
	}
}

// TestExportCSV_WithMockServer tests ExportCSV with actual API client
func TestExportCSV_WithMockServer(t *testing.T) {
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
			ID:          2,
			URL:         "https://test.com",
			Title:       "Test",
			Description: "",
			TagNames:    []string{},
			DateAdded:   testTime,
			Unread:      true,
			Shared:      false,
			IsArchived:  true,
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
	err := ExportCSV(client, &buf, ExportOptions{})

	if err != nil {
		t.Fatalf("ExportCSV() failed: %v", err)
	}

	// Parse the exported CSV
	csvReader := csv.NewReader(&buf)
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read exported CSV: %v", err)
	}

	// Verify header + 2 rows
	if len(records) != 3 {
		t.Errorf("Expected 3 records (1 header + 2 data), got %d", len(records))
	}

	// Verify first data row
	if records[1][0] != "https://example.com" {
		t.Errorf("Expected URL https://example.com, got %s", records[1][0])
	}
	if records[1][1] != "Example" {
		t.Errorf("Expected title Example, got %s", records[1][1])
	}
	if records[1][3] != "tag1,tag2" {
		t.Errorf("Expected tags 'tag1,tag2', got %s", records[1][3])
	}
	if records[1][6] != "true" {
		t.Errorf("Expected shared true, got %s", records[1][6])
	}

	// Verify second data row
	if records[2][0] != "https://test.com" {
		t.Errorf("Expected URL https://test.com, got %s", records[2][0])
	}
	if records[2][5] != "true" {
		t.Errorf("Expected unread true, got %s", records[2][5])
	}
	if records[2][7] != "true" {
		t.Errorf("Expected archived true, got %s", records[2][7])
	}
}

// TestExportCSV_WithFilters tests ExportCSV with tag filtering
func TestExportCSV_WithFilters(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create mock server that checks for filters
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

	// Create client and export with filters
	client := api.NewClient(server.URL, "test-token")
	var buf bytes.Buffer
	err := ExportCSV(client, &buf, ExportOptions{
		Tags: []string{"test-tag"},
	})

	if err != nil {
		t.Fatalf("ExportCSV() with filters failed: %v", err)
	}

	// Verify export succeeded
	csvReader := csv.NewReader(&buf)
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read exported CSV: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("Expected 2 records (1 header + 1 data), got %d", len(records))
	}
}

// TestCSV_RoundTrip tests that CSV export→import preserves all bookmark data
func TestCSV_RoundTrip(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	// Original bookmarks to export
	originalBookmarks := []models.Bookmark{
		{
			ID:           1,
			URL:          "https://example.com",
			Title:        "Example Site",
			Description:  "A test bookmark with all fields",
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
			TagNames:     []string{},
			DateAdded:    testTime,
			DateModified: testTime,
			Unread:       false,
			Shared:       false,
			IsArchived:   false,
		},
	}

	// Step 1: Export bookmarks to CSV
	var exportBuf bytes.Buffer

	// Create mock server for export
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
	if err := ExportCSV(exportClient, &exportBuf, ExportOptions{}); err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}

	// Verify export succeeded and save the exported CSV
	exportedCSV := exportBuf.Bytes()
	csvReader := csv.NewReader(bytes.NewReader(exportedCSV))
	exportedRecords, err := csvReader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read exported CSV: %v", err)
	}
	if len(exportedRecords) != 4 { // header + 3 data rows
		t.Fatalf("Expected 4 CSV records (header + 3 data), got %d", len(exportedRecords))
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
	importResult, err := importCSV(importClient, bytes.NewReader(exportedCSV), ImportOptions{})
	if err != nil {
		t.Fatalf("importCSV failed: %v", err)
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
