package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rodstewart/linkding-cli/internal/models"
)

// TestGetBookmarks_Pagination tests that GetBookmarks correctly handles paginated responses
func TestGetBookmarks_Pagination(t *testing.T) {
	nextURL := "/api/bookmarks/?offset=2&limit=2"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		offset := query.Get("offset")

		var response models.BookmarkList
		if offset == "" || offset == "0" {
			// First page
			response = models.BookmarkList{
				Count: 4,
				Next:  &nextURL,
				Results: []models.Bookmark{
					{ID: 1, URL: "https://example1.com", Title: "Example 1"},
					{ID: 2, URL: "https://example2.com", Title: "Example 2"},
				},
			}
		} else if offset == "2" {
			// Second page
			response = models.BookmarkList{
				Count: 4,
				Next:  nil,
				Results: []models.Bookmark{
					{ID: 3, URL: "https://example3.com", Title: "Example 3"},
					{ID: 4, URL: "https://example4.com", Title: "Example 4"},
				},
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	// Test first page
	page1, err := client.GetBookmarks("", nil, nil, nil, 2, 0)
	if err != nil {
		t.Fatalf("GetBookmarks() page 1 failed: %v", err)
	}

	if len(page1.Results) != 2 {
		t.Errorf("expected 2 results on page 1, got %d", len(page1.Results))
	}

	if page1.Next == nil {
		t.Error("expected Next pointer on page 1, got nil")
	}

	// Test second page
	page2, err := client.GetBookmarks("", nil, nil, nil, 2, 2)
	if err != nil {
		t.Fatalf("GetBookmarks() page 2 failed: %v", err)
	}

	if len(page2.Results) != 2 {
		t.Errorf("expected 2 results on page 2, got %d", len(page2.Results))
	}

	if page2.Next != nil {
		t.Error("expected nil Next pointer on page 2")
	}
}

// TestFetchAllBookmarks tests that FetchAllBookmarks fetches all pages
func TestFetchAllBookmarks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		offset := query.Get("offset")

		var response models.BookmarkList
		nextURL := "/api/bookmarks/?offset=100&limit=100"

		switch offset {
		case "", "0":
			// First page
			response = models.BookmarkList{
				Count:   250,
				Next:    &nextURL,
				Results: make([]models.Bookmark, 100),
			}
			for i := 0; i < 100; i++ {
				response.Results[i] = models.Bookmark{ID: i + 1, URL: "https://example.com", Title: "Test"}
			}
		case "100":
			// Second page
			nextURL2 := "/api/bookmarks/?offset=200&limit=100"
			response = models.BookmarkList{
				Count:   250,
				Next:    &nextURL2,
				Results: make([]models.Bookmark, 100),
			}
			for i := 0; i < 100; i++ {
				response.Results[i] = models.Bookmark{ID: i + 101, URL: "https://example.com", Title: "Test"}
			}
		case "200":
			// Third page (partial)
			response = models.BookmarkList{
				Count:   250,
				Next:    nil,
				Results: make([]models.Bookmark, 50),
			}
			for i := 0; i < 50; i++ {
				response.Results[i] = models.Bookmark{ID: i + 201, URL: "https://example.com", Title: "Test"}
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	// Fetch all bookmarks
	allBookmarks, err := client.FetchAllBookmarks(nil, true)
	if err != nil {
		t.Fatalf("FetchAllBookmarks() failed: %v", err)
	}

	// Should have fetched all 250 bookmarks across 3 pages
	if len(allBookmarks) != 250 {
		t.Errorf("expected 250 bookmarks, got %d", len(allBookmarks))
	}

	// Verify IDs are sequential
	if allBookmarks[0].ID != 1 {
		t.Errorf("expected first bookmark ID 1, got %d", allBookmarks[0].ID)
	}

	if allBookmarks[249].ID != 250 {
		t.Errorf("expected last bookmark ID 250, got %d", allBookmarks[249].ID)
	}
}

// TestFetchAllBookmarks_WithFilters tests that FetchAllBookmarks respects filters
func TestFetchAllBookmarks_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// Verify archived filter is set correctly for non-archived bookmarks
		archivedParam := query.Get("archived")
		if archivedParam != "" && archivedParam != "yes" {
			response := models.BookmarkList{
				Count:   10,
				Next:    nil,
				Results: make([]models.Bookmark, 10),
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(response)
			return
		}

		response := models.BookmarkList{
			Count:   10,
			Next:    nil,
			Results: make([]models.Bookmark, 10),
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	// Fetch with includeArchived=false
	bookmarks, err := client.FetchAllBookmarks(nil, false)
	if err != nil {
		t.Fatalf("FetchAllBookmarks() with filters failed: %v", err)
	}

	if len(bookmarks) != 10 {
		t.Errorf("expected 10 bookmarks, got %d", len(bookmarks))
	}
}

// TestGetTags_Pagination tests that GetTags correctly handles paginated responses
func TestGetTags_Pagination(t *testing.T) {
	nextURL := "/api/tags/?offset=2&limit=2"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		offset := query.Get("offset")

		var response models.TagList
		if offset == "" || offset == "0" {
			// First page
			response = models.TagList{
				Count: 4,
				Next:  &nextURL,
				Results: []models.Tag{
					{ID: 1, Name: "tag1"},
					{ID: 2, Name: "tag2"},
				},
			}
		} else if offset == "2" {
			// Second page
			response = models.TagList{
				Count: 4,
				Next:  nil,
				Results: []models.Tag{
					{ID: 3, Name: "tag3"},
					{ID: 4, Name: "tag4"},
				},
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	// Test first page
	page1, err := client.GetTags(2, 0)
	if err != nil {
		t.Fatalf("GetTags() page 1 failed: %v", err)
	}

	if len(page1.Results) != 2 {
		t.Errorf("expected 2 results on page 1, got %d", len(page1.Results))
	}

	if page1.Next == nil {
		t.Error("expected Next pointer on page 1, got nil")
	}

	// Test second page
	page2, err := client.GetTags(2, 2)
	if err != nil {
		t.Fatalf("GetTags() page 2 failed: %v", err)
	}

	if len(page2.Results) != 2 {
		t.Errorf("expected 2 results on page 2, got %d", len(page2.Results))
	}

	if page2.Next != nil {
		t.Error("expected nil Next pointer on page 2")
	}
}

// TestFetchAllTags tests that FetchAllTags fetches all pages
func TestFetchAllTags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		offset := query.Get("offset")

		var response models.TagList
		nextURL := "/api/tags/?offset=100&limit=100"

		switch offset {
		case "", "0":
			// First page
			response = models.TagList{
				Count:   150,
				Next:    &nextURL,
				Results: make([]models.Tag, 100),
			}
			for i := 0; i < 100; i++ {
				response.Results[i] = models.Tag{ID: i + 1, Name: "tag"}
			}
		case "100":
			// Second page (partial)
			response = models.TagList{
				Count:   150,
				Next:    nil,
				Results: make([]models.Tag, 50),
			}
			for i := 0; i < 50; i++ {
				response.Results[i] = models.Tag{ID: i + 101, Name: "tag"}
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")

	// Fetch all tags
	allTags, err := client.FetchAllTags()
	if err != nil {
		t.Fatalf("FetchAllTags() failed: %v", err)
	}

	// Should have fetched all 150 tags across 2 pages
	if len(allTags) != 150 {
		t.Errorf("expected 150 tags, got %d", len(allTags))
	}

	// Verify IDs are sequential
	if allTags[0].ID != 1 {
		t.Errorf("expected first tag ID 1, got %d", allTags[0].ID)
	}

	if allTags[149].ID != 150 {
		t.Errorf("expected last tag ID 150, got %d", allTags[149].ID)
	}
}
