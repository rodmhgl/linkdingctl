package export

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/rodstewart/linkding-cli/internal/api"
	"github.com/rodstewart/linkding-cli/internal/models"
)

// TestDetectFormat tests format detection from file extensions
func TestDetectFormat(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"bookmarks.json", "json"},
		{"bookmarks.JSON", "json"},
		{"bookmarks.html", "html"},
		{"bookmarks.HTML", "html"},
		{"bookmarks.htm", "html"},
		{"bookmarks.HTM", "html"},
		{"bookmarks.csv", "csv"},
		{"bookmarks.CSV", "csv"},
		{"bookmarks.txt", ""},
		{"bookmarks", ""},
		{"unknown.xyz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := DetectFormat(tt.filename)
			if result != tt.expected {
				t.Errorf("DetectFormat(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestImportJSON_RoundTrip tests JSON export -> import round trip
func TestImportJSON_RoundTrip(t *testing.T) {
	// Create test data
	exportData := ExportData{
		Version: "1.0",
		Source:  "linkding-cli",
		Bookmarks: []ExportBookmark{
			{
				URL:         "https://example.com",
				Title:       "Example",
				Description: "Test bookmark",
				Tags:        []string{"tag1", "tag2"},
				Archived:    false,
				Unread:      false,
				Shared:      true,
			},
			{
				URL:         "https://test.com",
				Title:       "Test",
				Description: "",
				Tags:        []string{},
				Archived:    true,
				Unread:      true,
				Shared:      false,
			},
		},
	}

	// Serialize to JSON
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(exportData); err != nil {
		t.Fatalf("Failed to encode JSON: %v", err)
	}

	// Create mock server that captures created bookmarks
	var created []models.BookmarkCreate
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/api/bookmarks/":
			// Return empty list (no existing bookmarks)
			response := models.BookmarkList{Count: 0, Results: []models.Bookmark{}}
			json.NewEncoder(w).Encode(response)

		case r.Method == "POST" && r.URL.Path == "/api/bookmarks/":
			// Capture created bookmark
			var bookmark models.BookmarkCreate
			if err := json.NewDecoder(r.Body).Decode(&bookmark); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			created = append(created, bookmark)

			// Return created bookmark
			w.WriteHeader(http.StatusCreated)
			response := models.Bookmark{
				ID:          len(created),
				URL:         bookmark.URL,
				Title:       bookmark.Title,
				Description: bookmark.Description,
				TagNames:    bookmark.TagNames,
				IsArchived:  bookmark.IsArchived,
				Unread:      bookmark.Unread,
				Shared:      bookmark.Shared,
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	// Import
	client := api.NewClient(server.URL, "test-token")
	result, err := importJSON(client, &buf, ImportOptions{})

	if err != nil {
		t.Fatalf("importJSON() failed: %v", err)
	}

	// Verify results
	if result.Added != 2 {
		t.Errorf("Expected 2 added, got %d", result.Added)
	}

	if result.Updated != 0 {
		t.Errorf("Expected 0 updated, got %d", result.Updated)
	}

	if result.Failed != 0 {
		t.Errorf("Expected 0 failed, got %d", result.Failed)
	}

	// Verify created bookmarks match original data
	if len(created) != 2 {
		t.Fatalf("Expected 2 created bookmarks, got %d", len(created))
	}

	// Check first bookmark
	if created[0].URL != exportData.Bookmarks[0].URL {
		t.Errorf("URL mismatch: got %q, want %q", created[0].URL, exportData.Bookmarks[0].URL)
	}

	if created[0].Title != exportData.Bookmarks[0].Title {
		t.Errorf("Title mismatch: got %q, want %q", created[0].Title, exportData.Bookmarks[0].Title)
	}

	if len(created[0].TagNames) != len(exportData.Bookmarks[0].Tags) {
		t.Errorf("Tags mismatch: got %v, want %v", created[0].TagNames, exportData.Bookmarks[0].Tags)
	}
}

// TestImportJSON_SkipDuplicates tests duplicate handling with skip option
func TestImportJSON_SkipDuplicates(t *testing.T) {
	exportData := ExportData{
		Bookmarks: []ExportBookmark{
			{URL: "https://example.com", Title: "Example"},
			{URL: "https://test.com", Title: "Test"},
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(exportData)

	// Mock server with one existing bookmark
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Return one existing bookmark
			response := models.BookmarkList{
				Count: 1,
				Results: []models.Bookmark{
					{ID: 1, URL: "https://example.com", Title: "Existing"},
				},
			}
			json.NewEncoder(w).Encode(response)
		} else if r.Method == "POST" {
			// Create new bookmark
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(models.Bookmark{ID: 2, URL: "https://test.com"})
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	result, err := importJSON(client, &buf, ImportOptions{SkipDuplicates: true})

	if err != nil {
		t.Fatalf("importJSON() failed: %v", err)
	}

	if result.Skipped != 1 {
		t.Errorf("Expected 1 skipped, got %d", result.Skipped)
	}

	if result.Added != 1 {
		t.Errorf("Expected 1 added, got %d", result.Added)
	}
}

// TestImportJSON_UpdateDuplicates tests duplicate handling with update behavior
func TestImportJSON_UpdateDuplicates(t *testing.T) {
	exportData := ExportData{
		Bookmarks: []ExportBookmark{
			{URL: "https://example.com", Title: "Updated Title"},
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(exportData)

	// Mock server with one existing bookmark
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			response := models.BookmarkList{
				Count: 1,
				Results: []models.Bookmark{
					{ID: 1, URL: "https://example.com", Title: "Old Title"},
				},
			}
			json.NewEncoder(w).Encode(response)
		} else if r.Method == "PATCH" {
			// Update existing bookmark
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(models.Bookmark{ID: 1, URL: "https://example.com", Title: "Updated Title"})
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	result, err := importJSON(client, &buf, ImportOptions{SkipDuplicates: false})

	if err != nil {
		t.Fatalf("importJSON() failed: %v", err)
	}

	if result.Updated != 1 {
		t.Errorf("Expected 1 updated, got %d", result.Updated)
	}

	if result.Added != 0 {
		t.Errorf("Expected 0 added, got %d", result.Added)
	}
}

// TestImportJSON_AddTags tests that --add-tags appends correctly
func TestImportJSON_AddTags(t *testing.T) {
	exportData := ExportData{
		Bookmarks: []ExportBookmark{
			{URL: "https://example.com", Title: "Example", Tags: []string{"original"}},
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(exportData)

	var createdBookmark models.BookmarkCreate
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(models.BookmarkList{Count: 0, Results: []models.Bookmark{}})
		} else if r.Method == "POST" {
			if err := json.NewDecoder(r.Body).Decode(&createdBookmark); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(models.Bookmark{ID: 1})
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	_, err := importJSON(client, &buf, ImportOptions{
		AddTags: []string{"imported", "new"},
	})

	if err != nil {
		t.Fatalf("importJSON() failed: %v", err)
	}

	// Verify tags were appended
	expectedTags := []string{"original", "imported", "new"}
	if len(createdBookmark.TagNames) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(createdBookmark.TagNames))
	}

	for i, tag := range expectedTags {
		if i >= len(createdBookmark.TagNames) || createdBookmark.TagNames[i] != tag {
			t.Errorf("Tag mismatch at index %d: got %v, want %v", i, createdBookmark.TagNames, expectedTags)
			break
		}
	}
}

// TestImportHTML_ParsesCorrectly tests HTML format parsing
func TestImportHTML_ParsesCorrectly(t *testing.T) {
	htmlInput := `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
    <DT><A HREF="https://example.com" ADD_DATE="1704110400" TAGS="web,dev">Example Site</A>
    <DD>Test description
    <DT><A HREF="https://test.com" ADD_DATE="1704110400">Test Site</A>
</DL><p>
`

	var created []models.BookmarkCreate
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(models.BookmarkList{Count: 0, Results: []models.Bookmark{}})
		} else if r.Method == "POST" {
			var bookmark models.BookmarkCreate
			if err := json.NewDecoder(r.Body).Decode(&bookmark); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			created = append(created, bookmark)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(models.Bookmark{ID: len(created)})
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	reader := strings.NewReader(htmlInput)
	result, err := importHTML(client, reader, ImportOptions{})

	if err != nil {
		t.Fatalf("importHTML() failed: %v", err)
	}

	if result.Added != 2 {
		t.Errorf("Expected 2 added, got %d", result.Added)
	}

	// Verify first bookmark (with tags and description)
	if len(created) < 1 {
		t.Fatal("Expected at least 1 created bookmark")
	}

	if created[0].URL != "https://example.com" {
		t.Errorf("URL mismatch: got %q, want %q", created[0].URL, "https://example.com")
	}

	if created[0].Title != "Example Site" {
		t.Errorf("Title mismatch: got %q, want %q", created[0].Title, "Example Site")
	}

	if created[0].Description != "Test description" {
		t.Errorf("Description mismatch: got %q, want %q", created[0].Description, "Test description")
	}

	expectedTags := []string{"web", "dev"}
	if len(created[0].TagNames) != len(expectedTags) {
		t.Errorf("Tags mismatch: got %v, want %v", created[0].TagNames, expectedTags)
	}

	// Verify second bookmark (no tags or description)
	if len(created) < 2 {
		t.Fatal("Expected at least 2 created bookmarks")
	}

	if created[1].URL != "https://test.com" {
		t.Errorf("URL mismatch: got %q, want %q", created[1].URL, "https://test.com")
	}

	if created[1].Title != "Test Site" {
		t.Errorf("Title mismatch: got %q, want %q", created[1].Title, "Test Site")
	}

	if created[1].Description != "" {
		t.Errorf("Expected empty description, got %q", created[1].Description)
	}
}

// TestImportCSV_RoundTrip tests CSV export -> import round trip
func TestImportCSV_RoundTrip(t *testing.T) {
	csvInput := `url,title,description,tags,archived,unread,shared
https://example.com,Example,Test description,"tag1,tag2",false,false,true
https://test.com,Test,,tag3,true,true,false
`

	var created []models.BookmarkCreate
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(models.BookmarkList{Count: 0, Results: []models.Bookmark{}})
		} else if r.Method == "POST" {
			var bookmark models.BookmarkCreate
			if err := json.NewDecoder(r.Body).Decode(&bookmark); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			created = append(created, bookmark)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(models.Bookmark{ID: len(created)})
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	reader := strings.NewReader(csvInput)
	result, err := importCSV(client, reader, ImportOptions{})

	if err != nil {
		t.Fatalf("importCSV() failed: %v", err)
	}

	if result.Added != 2 {
		t.Errorf("Expected 2 added, got %d", result.Added)
	}

	// Verify first bookmark
	if len(created) < 1 {
		t.Fatal("Expected at least 1 created bookmark")
	}

	if created[0].URL != "https://example.com" {
		t.Errorf("URL mismatch: got %q, want %q", created[0].URL, "https://example.com")
	}

	if created[0].Shared != true {
		t.Errorf("Shared mismatch: got %v, want true", created[0].Shared)
	}

	if len(created[0].TagNames) != 2 {
		t.Errorf("Expected 2 tags, got %d: %v", len(created[0].TagNames), created[0].TagNames)
	}

	// Verify second bookmark
	if len(created) < 2 {
		t.Fatal("Expected at least 2 created bookmarks")
	}

	if created[1].IsArchived != true {
		t.Errorf("Archived mismatch: got %v, want true", created[1].IsArchived)
	}

	if created[1].Unread != true {
		t.Errorf("Unread mismatch: got %v, want true", created[1].Unread)
	}
}

// TestImportJSON_ValidationError tests error handling for missing required fields
func TestImportJSON_ValidationError(t *testing.T) {
	exportData := ExportData{
		Bookmarks: []ExportBookmark{
			{URL: "", Title: "Missing URL"}, // Invalid: no URL
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(exportData)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(models.BookmarkList{Count: 0, Results: []models.Bookmark{}})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	result, err := importJSON(client, &buf, ImportOptions{})

	if err != nil {
		t.Fatalf("importJSON() should not fail on validation errors: %v", err)
	}

	if result.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", result.Failed)
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}

	if len(result.Errors) > 0 && !strings.Contains(result.Errors[0].Message, "Missing required field") {
		t.Errorf("Expected validation error, got: %s", result.Errors[0].Message)
	}
}

// TestImportJSON_DryRun tests dry-run mode
func TestImportJSON_DryRun(t *testing.T) {
	exportData := ExportData{
		Bookmarks: []ExportBookmark{
			{URL: "https://example.com", Title: "Example"},
			{URL: "https://test.com", Title: "Test"},
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(exportData)

	apiCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In dry-run mode, no API calls should be made except possibly fetching existing bookmarks
		// But since the implementation skips fetching in dry-run, no calls should be made at all
		if r.Method == "POST" || r.Method == "PATCH" {
			apiCalled = true
			t.Error("CreateBookmark/UpdateBookmark should not be called in dry-run mode")
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	result, err := importJSON(client, &buf, ImportOptions{DryRun: true})

	if err != nil {
		t.Fatalf("importJSON() failed: %v", err)
	}

	// In dry-run mode without fetching existing bookmarks, all bookmarks are treated as new
	if result.Added != 2 {
		t.Errorf("Expected 2 would be added, got %d", result.Added)
	}

	if result.Updated != 0 {
		t.Errorf("Expected 0 would be updated, got %d", result.Updated)
	}

	if result.Failed != 0 {
		t.Errorf("Expected 0 failed, got %d", result.Failed)
	}

	if apiCalled {
		t.Error("Create/Update API should not be called in dry-run mode")
	}
}

// TestImportCSV_MissingColumns tests graceful handling of CSV with missing columns
func TestImportCSV_MissingColumns(t *testing.T) {
	// CSV with only required "url" column
	csvInput := `url
https://example.com
https://test.com
`

	var created []models.BookmarkCreate
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(models.BookmarkList{Count: 0, Results: []models.Bookmark{}})
		} else if r.Method == "POST" {
			var bookmark models.BookmarkCreate
			if err := json.NewDecoder(r.Body).Decode(&bookmark); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			created = append(created, bookmark)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(models.Bookmark{ID: len(created)})
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	reader := strings.NewReader(csvInput)
	result, err := importCSV(client, reader, ImportOptions{})

	if err != nil {
		t.Fatalf("importCSV() failed: %v", err)
	}

	if result.Added != 2 {
		t.Errorf("Expected 2 added, got %d", result.Added)
	}

	// Verify bookmarks were created with empty optional fields
	if len(created) < 1 {
		t.Fatal("Expected at least 1 created bookmark")
	}

	if created[0].URL != "https://example.com" {
		t.Errorf("URL mismatch: got %q, want %q", created[0].URL, "https://example.com")
	}

	if created[0].Title != "" {
		t.Errorf("Expected empty title, got %q", created[0].Title)
	}

	if len(created[0].TagNames) != 0 {
		t.Errorf("Expected no tags, got %v", created[0].TagNames)
	}
}

// TestImportCSV_MalformedRows tests error handling for malformed CSV rows
func TestImportCSV_MalformedRows(t *testing.T) {
	// CSV with malformed row (unclosed quote)
	csvInput := `url,title,description
https://example.com,Example,Test
https://test.com,"Unclosed quote
https://valid.com,Valid,After error
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(models.BookmarkList{Count: 0, Results: []models.Bookmark{}})
		} else if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(models.Bookmark{ID: 1})
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	reader := strings.NewReader(csvInput)
	result, err := importCSV(client, reader, ImportOptions{})

	if err != nil {
		t.Fatalf("importCSV() should not fail on parse errors: %v", err)
	}

	// At least one row should fail due to malformed CSV
	if result.Failed == 0 {
		t.Error("Expected at least 1 failed row due to malformed CSV")
	}

	// Should have at least one error
	if len(result.Errors) == 0 {
		t.Error("Expected at least 1 error for malformed row")
	}
}

// TestImportHTML_WithAddTags tests HTML import with --add-tags option
func TestImportHTML_WithAddTags(t *testing.T) {
	htmlInput := `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
    <DT><A HREF="https://example.com" ADD_DATE="1704110400" TAGS="original">Example Site</A>
</DL><p>
`

	var createdBookmark models.BookmarkCreate
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(models.BookmarkList{Count: 0, Results: []models.Bookmark{}})
		} else if r.Method == "POST" {
			if err := json.NewDecoder(r.Body).Decode(&createdBookmark); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(models.Bookmark{ID: 1})
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	reader := strings.NewReader(htmlInput)
	result, err := importHTML(client, reader, ImportOptions{
		AddTags: []string{"imported", "new"},
	})

	if err != nil {
		t.Fatalf("importHTML() failed: %v", err)
	}

	if result.Added != 1 {
		t.Errorf("Expected 1 added, got %d", result.Added)
	}

	// Verify tags were appended
	expectedTags := []string{"original", "imported", "new"}
	if len(createdBookmark.TagNames) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(createdBookmark.TagNames))
	}

	for i, tag := range expectedTags {
		if i >= len(createdBookmark.TagNames) || createdBookmark.TagNames[i] != tag {
			t.Errorf("Tag mismatch at index %d: got %v, want %v", i, createdBookmark.TagNames, expectedTags)
			break
		}
	}
}

// TestImportHTML_SkipDuplicates tests HTML import with --skip-duplicates option
func TestImportHTML_SkipDuplicates(t *testing.T) {
	htmlInput := `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
    <DT><A HREF="https://example.com" ADD_DATE="1704110400">Example Site</A>
    <DT><A HREF="https://test.com" ADD_DATE="1704110400">Test Site</A>
</DL><p>
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Return one existing bookmark
			response := models.BookmarkList{
				Count: 1,
				Results: []models.Bookmark{
					{ID: 1, URL: "https://example.com", Title: "Existing"},
				},
			}
			json.NewEncoder(w).Encode(response)
		} else if r.Method == "POST" {
			// Create new bookmark
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(models.Bookmark{ID: 2, URL: "https://test.com"})
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	reader := strings.NewReader(htmlInput)
	result, err := importHTML(client, reader, ImportOptions{SkipDuplicates: true})

	if err != nil {
		t.Fatalf("importHTML() failed: %v", err)
	}

	if result.Skipped != 1 {
		t.Errorf("Expected 1 skipped, got %d", result.Skipped)
	}

	if result.Added != 1 {
		t.Errorf("Expected 1 added, got %d", result.Added)
	}
}

// TestImportCSV_SkipDuplicatesAndAddTags tests CSV import with combined options
func TestImportCSV_SkipDuplicatesAndAddTags(t *testing.T) {
	csvInput := `url,title
https://example.com,Example
https://test.com,Test
`

	var created []models.BookmarkCreate
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Return one existing bookmark
			response := models.BookmarkList{
				Count: 1,
				Results: []models.Bookmark{
					{ID: 1, URL: "https://example.com", Title: "Existing"},
				},
			}
			json.NewEncoder(w).Encode(response)
		} else if r.Method == "POST" {
			var bookmark models.BookmarkCreate
			if err := json.NewDecoder(r.Body).Decode(&bookmark); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			created = append(created, bookmark)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(models.Bookmark{ID: len(created) + 1})
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	reader := strings.NewReader(csvInput)
	result, err := importCSV(client, reader, ImportOptions{
		SkipDuplicates: true,
		AddTags:        []string{"imported"},
	})

	if err != nil {
		t.Fatalf("importCSV() failed: %v", err)
	}

	if result.Skipped != 1 {
		t.Errorf("Expected 1 skipped, got %d", result.Skipped)
	}

	if result.Added != 1 {
		t.Errorf("Expected 1 added, got %d", result.Added)
	}

	// Verify the created bookmark has the added tag
	if len(created) < 1 {
		t.Fatal("Expected at least 1 created bookmark")
	}

	if len(created[0].TagNames) != 1 || created[0].TagNames[0] != "imported" {
		t.Errorf("Expected tags [imported], got %v", created[0].TagNames)
	}
}

// TestImportCSV_UpdateExisting tests CSV import updating existing bookmarks
func TestImportCSV_UpdateExisting(t *testing.T) {
	csvInput := `url,title,description
https://example.com,Updated Title,Updated description
`

	var updated bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Return existing bookmark
			response := models.BookmarkList{
				Count: 1,
				Results: []models.Bookmark{
					{ID: 1, URL: "https://example.com", Title: "Old Title", Description: "Old description"},
				},
			}
			json.NewEncoder(w).Encode(response)
		} else if r.Method == "PATCH" {
			// Update existing bookmark
			updated = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(models.Bookmark{
				ID:          1,
				URL:         "https://example.com",
				Title:       "Updated Title",
				Description: "Updated description",
			})
		}
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")
	reader := strings.NewReader(csvInput)
	result, err := importCSV(client, reader, ImportOptions{SkipDuplicates: false})

	if err != nil {
		t.Fatalf("importCSV() failed: %v", err)
	}

	if result.Updated != 1 {
		t.Errorf("Expected 1 updated, got %d", result.Updated)
	}

	if result.Added != 0 {
		t.Errorf("Expected 0 added, got %d", result.Added)
	}

	if !updated {
		t.Error("Expected PATCH request to update bookmark")
	}
}

// TestImportBookmarks_AutoDetect tests ImportBookmarks entry point with auto-detection
func TestImportBookmarks_AutoDetect(t *testing.T) {
	// Create a temporary JSON file
	jsonData := ExportData{
		Version: "1",
		Source:  "test",
		Bookmarks: []ExportBookmark{
			{URL: "https://example.com", Title: "Example"},
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(jsonData)

	// Write to temp file
	tmpfile := "/tmp/test-bookmarks.json"
	if err := os.WriteFile(tmpfile, buf.Bytes(), 0600); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile)

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(models.BookmarkList{Count: 0, Results: []models.Bookmark{}})
		} else if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(models.Bookmark{ID: 1})
		}
	}))
	defer server.Close()

	// Test auto-detection (format="")
	client := api.NewClient(server.URL, "test-token")
	result, err := ImportBookmarks(client, tmpfile, ImportOptions{Format: ""})

	if err != nil {
		t.Fatalf("ImportBookmarks() with auto-detect failed: %v", err)
	}

	if result.Added != 1 {
		t.Errorf("Expected 1 added, got %d", result.Added)
	}
}

// TestImportBookmarks_UnsupportedFormat tests error handling for unsupported formats
func TestImportBookmarks_UnsupportedFormat(t *testing.T) {
	// Create a temporary file with unsupported extension
	tmpfile := "/tmp/test-bookmarks.xyz"
	if err := os.WriteFile(tmpfile, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	client := api.NewClient(server.URL, "test-token")

	// Test with auto-detect on unsupported format
	_, err := ImportBookmarks(client, tmpfile, ImportOptions{Format: ""})
	if err == nil {
		t.Error("Expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "cannot detect format") {
		t.Errorf("Expected 'cannot detect format' error, got: %v", err)
	}

	// Test with explicit unsupported format
	_, err = ImportBookmarks(client, tmpfile, ImportOptions{Format: "xml"})
	if err == nil {
		t.Error("Expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected 'unsupported format' error, got: %v", err)
	}
}

// TestImportBookmarks_FormatOverride tests explicit format override
func TestImportBookmarks_FormatOverride(t *testing.T) {
	// Create a temporary file with wrong extension but JSON content
	jsonData := ExportData{
		Version: "1",
		Source:  "test",
		Bookmarks: []ExportBookmark{
			{URL: "https://example.com", Title: "Example"},
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(jsonData)

	// Write to file with .txt extension
	tmpfile := "/tmp/test-bookmarks.txt"
	if err := os.WriteFile(tmpfile, buf.Bytes(), 0600); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile)

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(models.BookmarkList{Count: 0, Results: []models.Bookmark{}})
		} else if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(models.Bookmark{ID: 1})
		}
	}))
	defer server.Close()

	// Test with explicit format override
	client := api.NewClient(server.URL, "test-token")
	result, err := ImportBookmarks(client, tmpfile, ImportOptions{Format: "json"})

	if err != nil {
		t.Fatalf("ImportBookmarks() with format override failed: %v", err)
	}

	if result.Added != 1 {
		t.Errorf("Expected 1 added, got %d", result.Added)
	}
}
