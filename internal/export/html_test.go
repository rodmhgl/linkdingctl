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

func TestExportHTML_ValidNetscapeFormat(t *testing.T) {
	// Create test bookmarks
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	bookmarks := []models.Bookmark{
		{
			ID:          1,
			URL:         "https://example.com",
			Title:       "Example Site",
			Description: "A test site",
			TagNames:    []string{"test", "example"},
			DateAdded:   testTime,
		},
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	// Create client and export
	client := api.NewClient(server.URL, "test-token")
	var buf bytes.Buffer
	err := ExportHTML(client, &buf, ExportOptions{})

	if err != nil {
		t.Fatalf("ExportHTML() failed: %v", err)
	}

	output := buf.String()

	// Verify Netscape format header
	if !strings.Contains(output, "<!DOCTYPE NETSCAPE-Bookmark-file-1>") {
		t.Error("Expected Netscape DOCTYPE")
	}

	if !strings.Contains(output, "<META HTTP-EQUIV=\"Content-Type\" CONTENT=\"text/html; charset=UTF-8\">") {
		t.Error("Expected UTF-8 meta tag")
	}

	if !strings.Contains(output, "<TITLE>Bookmarks</TITLE>") {
		t.Error("Expected title tag")
	}

	if !strings.Contains(output, "<H1>Bookmarks</H1>") {
		t.Error("Expected H1 heading")
	}

	// Verify bookmark entry
	if !strings.Contains(output, "HREF=\"https://example.com\"") {
		t.Error("Expected bookmark URL")
	}

	if !strings.Contains(output, "ADD_DATE=\"1704110400\"") {
		t.Error("Expected Unix timestamp")
	}

	if !strings.Contains(output, ">Example Site</A>") {
		t.Error("Expected bookmark title")
	}
}

func TestExportHTML_HTMLEscaping(t *testing.T) {
	// Create bookmarks with special characters that need escaping
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	bookmarks := []models.Bookmark{
		{
			ID:          1,
			URL:         "https://example.com?foo=bar&baz=qux",
			Title:       "Test & <Special> \"Chars\"",
			Description: "Description with <HTML> & \"quotes\"",
			TagNames:    []string{"tag&special", "tag<html>"},
			DateAdded:   testTime,
		},
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	// Create client and export
	client := api.NewClient(server.URL, "test-token")
	var buf bytes.Buffer
	err := ExportHTML(client, &buf, ExportOptions{})

	if err != nil {
		t.Fatalf("ExportHTML() failed: %v", err)
	}

	output := buf.String()

	// Verify HTML escaping in URL
	if !strings.Contains(output, "https://example.com?foo=bar&amp;baz=qux") {
		t.Error("Expected escaped ampersand in URL")
	}

	// Verify HTML escaping in title
	if !strings.Contains(output, "Test &amp; &lt;Special&gt; &#34;Chars&#34;") {
		t.Error("Expected escaped special characters in title")
	}

	// Verify HTML escaping in description
	if !strings.Contains(output, "Description with &lt;HTML&gt; &amp; &#34;quotes&#34;") {
		t.Error("Expected escaped special characters in description")
	}

	// Verify HTML escaping in tags
	if !strings.Contains(output, "TAGS=\"tag&amp;special,tag&lt;html&gt;\"") {
		t.Error("Expected escaped special characters in tags")
	}
}

func TestExportHTML_TagsAttribute(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		tags        []string
		expectTags  bool
		expectedStr string
	}{
		{
			name:        "Multiple tags",
			tags:        []string{"web", "development", "tools"},
			expectTags:  true,
			expectedStr: "TAGS=\"web,development,tools\"",
		},
		{
			name:        "Single tag",
			tags:        []string{"testing"},
			expectTags:  true,
			expectedStr: "TAGS=\"testing\"",
		},
		{
			name:        "No tags",
			tags:        []string{},
			expectTags:  false,
			expectedStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmarks := []models.Bookmark{
				{
					ID:        1,
					URL:       "https://example.com",
					Title:     "Test",
					TagNames:  tt.tags,
					DateAdded: testTime,
				},
			}

			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			// Create client and export
			client := api.NewClient(server.URL, "test-token")
			var buf bytes.Buffer
			err := ExportHTML(client, &buf, ExportOptions{})

			if err != nil {
				t.Fatalf("ExportHTML() failed: %v", err)
			}

			output := buf.String()

			if tt.expectTags {
				if !strings.Contains(output, tt.expectedStr) {
					t.Errorf("Expected TAGS attribute with %q, output:\n%s", tt.expectedStr, output)
				}
			} else {
				if strings.Contains(output, "TAGS=") {
					t.Error("Expected no TAGS attribute for bookmark without tags")
				}
			}
		})
	}
}

func TestExportHTML_OmitsDescriptionWhenEmpty(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		description string
		expectDD    bool
	}{
		{
			name:        "With description",
			description: "This is a description",
			expectDD:    true,
		},
		{
			name:        "Empty description",
			description: "",
			expectDD:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmarks := []models.Bookmark{
				{
					ID:          1,
					URL:         "https://example.com",
					Title:       "Test",
					Description: tt.description,
					DateAdded:   testTime,
				},
			}

			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			// Create client and export
			client := api.NewClient(server.URL, "test-token")
			var buf bytes.Buffer
			err := ExportHTML(client, &buf, ExportOptions{})

			if err != nil {
				t.Fatalf("ExportHTML() failed: %v", err)
			}

			output := buf.String()

			// Count DD tags
			ddCount := strings.Count(output, "<DD>")

			if tt.expectDD && ddCount == 0 {
				t.Error("Expected DD tag for bookmark with description")
			}

			if !tt.expectDD && ddCount > 0 {
				t.Error("Expected no DD tag for bookmark without description")
			}

			// Verify description content if present
			if tt.expectDD && !strings.Contains(output, tt.description) {
				t.Errorf("Expected description %q in output", tt.description)
			}
		})
	}
}

func TestExportHTML_MultipleBookmarks(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	bookmarks := []models.Bookmark{
		{
			ID:          1,
			URL:         "https://example1.com",
			Title:       "Example 1",
			Description: "First bookmark",
			TagNames:    []string{"tag1"},
			DateAdded:   testTime,
		},
		{
			ID:          2,
			URL:         "https://example2.com",
			Title:       "Example 2",
			Description: "",
			TagNames:    []string{"tag2", "tag3"},
			DateAdded:   testTime,
		},
		{
			ID:        3,
			URL:       "https://example3.com",
			Title:     "Example 3",
			TagNames:  []string{},
			DateAdded: testTime,
		},
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := models.BookmarkList{
			Count:   3,
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
	err := ExportHTML(client, &buf, ExportOptions{})

	if err != nil {
		t.Fatalf("ExportHTML() failed: %v", err)
	}

	output := buf.String()

	// Verify all bookmarks are present
	for _, b := range bookmarks {
		if !strings.Contains(output, b.URL) {
			t.Errorf("Expected URL %s in output", b.URL)
		}
		if !strings.Contains(output, b.Title) {
			t.Errorf("Expected title %s in output", b.Title)
		}
	}

	// Count anchor tags (should be 3)
	anchorCount := strings.Count(output, "<A HREF=")
	if anchorCount != 3 {
		t.Errorf("Expected 3 anchor tags, got %d", anchorCount)
	}
}

func TestExportHTML_WithFilters(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Test includeArchived=true (should not set archived parameter, gets all bookmarks)
	t.Run("IncludeArchived", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()

			// When includeArchived=true, no archived filter is set (gets all bookmarks)
			if query.Get("archived") != "" {
				t.Error("Expected no archived parameter when includeArchived=true")
			}

			bookmarks := []models.Bookmark{
				{
					ID:        1,
					URL:       "https://example.com",
					Title:     "All Bookmarks",
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

		client := api.NewClient(server.URL, "test-token")
		var buf bytes.Buffer
		err := ExportHTML(client, &buf, ExportOptions{
			IncludeArchived: true,
		})

		if err != nil {
			t.Fatalf("ExportHTML() with includeArchived=true failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "All Bookmarks") {
			t.Error("Expected bookmark in output")
		}
	})

	// Test includeArchived=false (should set archived=false, gets only non-archived)
	t.Run("ExcludeArchived", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bookmarks := []models.Bookmark{
				{
					ID:        1,
					URL:       "https://example.com",
					Title:     "Non-Archived",
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

		client := api.NewClient(server.URL, "test-token")
		var buf bytes.Buffer
		err := ExportHTML(client, &buf, ExportOptions{
			IncludeArchived: false,
		})

		if err != nil {
			t.Fatalf("ExportHTML() with includeArchived=false failed: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Non-Archived") {
			t.Error("Expected non-archived bookmark in output")
		}
	})
}
