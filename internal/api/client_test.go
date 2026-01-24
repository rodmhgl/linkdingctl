package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rodstewart/linkding-cli/internal/models"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://test.example.com", "test-token")

	if client.baseURL != "https://test.example.com" {
		t.Errorf("expected baseURL 'https://test.example.com', got '%s'", client.baseURL)
	}
	if client.token != "test-token" {
		t.Errorf("expected token 'test-token', got '%s'", client.token)
	}
}

func TestNewClient_TrimsTrailingSlash(t *testing.T) {
	client := NewClient("https://test.example.com/", "test-token")

	if client.baseURL != "https://test.example.com" {
		t.Errorf("expected baseURL without trailing slash, got '%s'", client.baseURL)
	}
}

func TestTestConnection_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/bookmarks/" {
			t.Errorf("expected path '/api/bookmarks/', got '%s'", r.URL.Path)
		}

		auth := r.Header.Get("Authorization")
		if auth != "Token test-token" {
			t.Errorf("expected Authorization 'Token test-token', got '%s'", auth)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(models.BookmarkList{})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.TestConnection()

	if err != nil {
		t.Fatalf("TestConnection() failed: %v", err)
	}
}

func TestTestConnection_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad-token")
	err := client.TestConnection()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedMsg := "authentication failed. Check your API token"
	if err.Error() != expectedMsg {
		t.Errorf("expected error '%s', got '%v'", expectedMsg, err)
	}
}

func TestGetBookmarks_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/bookmarks/" {
			t.Errorf("expected path '/api/bookmarks/', got '%s'", r.URL.Path)
		}

		response := models.BookmarkList{
			Count: 2,
			Results: []models.Bookmark{
				{ID: 1, URL: "https://example.com", Title: "Example"},
				{ID: 2, URL: "https://test.com", Title: "Test"},
			},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	bookmarks, err := client.GetBookmarks("", nil, nil, nil, 0, 0)

	if err != nil {
		t.Fatalf("GetBookmarks() failed: %v", err)
	}

	if bookmarks.Count != 2 {
		t.Errorf("expected count 2, got %d", bookmarks.Count)
	}

	if len(bookmarks.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(bookmarks.Results))
	}
}

func TestGetBookmarks_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if query.Get("q") == "" {
			t.Error("expected query parameter 'q' to be set")
		}

		if query.Get("limit") != "50" {
			t.Errorf("expected limit '50', got '%s'", query.Get("limit"))
		}

		if query.Get("offset") != "10" {
			t.Errorf("expected offset '10', got '%s'", query.Get("offset"))
		}

		response := models.BookmarkList{Count: 0, Results: []models.Bookmark{}}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	unread := true
	_, err := client.GetBookmarks("test", []string{"tag1", "tag2"}, &unread, nil, 50, 10)

	if err != nil {
		t.Fatalf("GetBookmarks() with filters failed: %v", err)
	}
}

func TestGetBookmark_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/bookmarks/123/" {
			t.Errorf("expected path '/api/bookmarks/123/', got '%s'", r.URL.Path)
		}

		bookmark := models.Bookmark{
			ID:    123,
			URL:   "https://example.com",
			Title: "Example",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(bookmark)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	bookmark, err := client.GetBookmark(123)

	if err != nil {
		t.Fatalf("GetBookmark() failed: %v", err)
	}

	if bookmark.ID != 123 {
		t.Errorf("expected ID 123, got %d", bookmark.ID)
	}
}

func TestGetBookmark_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	_, err := client.GetBookmark(999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedMsg := "bookmark with ID 999 not found"
	if err.Error() != expectedMsg {
		t.Errorf("expected error '%s', got '%v'", expectedMsg, err)
	}
}

func TestCreateBookmark_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got '%s'", r.Method)
		}

		if r.URL.Path != "/api/bookmarks/" {
			t.Errorf("expected path '/api/bookmarks/', got '%s'", r.URL.Path)
		}

		var create models.BookmarkCreate
		if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if create.URL != "https://example.com" {
			t.Errorf("expected URL 'https://example.com', got '%s'", create.URL)
		}

		created := models.Bookmark{
			ID:    123,
			URL:   create.URL,
			Title: create.Title,
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(created)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	create := &models.BookmarkCreate{
		URL:   "https://example.com",
		Title: "Example",
	}

	bookmark, err := client.CreateBookmark(create)

	if err != nil {
		t.Fatalf("CreateBookmark() failed: %v", err)
	}

	if bookmark.ID != 123 {
		t.Errorf("expected ID 123, got %d", bookmark.ID)
	}
}

func TestUpdateBookmark_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH method, got '%s'", r.Method)
		}

		if r.URL.Path != "/api/bookmarks/123/" {
			t.Errorf("expected path '/api/bookmarks/123/', got '%s'", r.URL.Path)
		}

		updated := models.Bookmark{
			ID:    123,
			URL:   "https://example.com",
			Title: "Updated Title",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(updated)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	title := "Updated Title"
	update := &models.BookmarkUpdate{
		Title: &title,
	}

	bookmark, err := client.UpdateBookmark(123, update)

	if err != nil {
		t.Fatalf("UpdateBookmark() failed: %v", err)
	}

	if bookmark.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", bookmark.Title)
	}
}

func TestDeleteBookmark_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got '%s'", r.Method)
		}

		if r.URL.Path != "/api/bookmarks/123/" {
			t.Errorf("expected path '/api/bookmarks/123/', got '%s'", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.DeleteBookmark(123)

	if err != nil {
		t.Fatalf("DeleteBookmark() failed: %v", err)
	}
}

func TestDeleteBookmark_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	err := client.DeleteBookmark(999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedMsg := "bookmark with ID 999 not found"
	if err.Error() != expectedMsg {
		t.Errorf("expected error '%s', got '%v'", expectedMsg, err)
	}
}

// TestCreateTag_Success tests successful tag creation
func TestCreateTag_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags/" || r.Method != "POST" {
			t.Errorf("expected POST /api/tags/, got %s %s", r.Method, r.URL.Path)
		}

		var req map[string]string
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["name"] != "test-tag" {
			t.Errorf("expected name 'test-tag', got '%s'", req["name"])
		}

		tag := models.Tag{ID: 1, Name: "test-tag"}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(tag)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tag, err := client.CreateTag("test-tag")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.ID != 1 {
		t.Errorf("expected ID 1, got %d", tag.ID)
	}
	if tag.Name != "test-tag" {
		t.Errorf("expected name 'test-tag', got '%s'", tag.Name)
	}
}

// TestCreateTag_Duplicate tests creating a duplicate tag
func TestCreateTag_Duplicate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(`{"name":["Tag with this name already exists"]}`)); err != nil {
			t.Errorf("failed to write response: %v", err)
			return
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	_, err := client.CreateTag("duplicate")

	if err == nil {
		t.Fatal("expected error for duplicate tag, got nil")
	}
}

// TestGetTag_Success tests getting a tag by ID
func TestGetTag_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags/1/" || r.Method != "GET" {
			t.Errorf("expected GET /api/tags/1/, got %s %s", r.Method, r.URL.Path)
		}

		tag := models.Tag{ID: 1, Name: "test-tag"}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(tag)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	tag, err := client.GetTag(1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.ID != 1 {
		t.Errorf("expected ID 1, got %d", tag.ID)
	}
	if tag.Name != "test-tag" {
		t.Errorf("expected name 'test-tag', got '%s'", tag.Name)
	}
}

// TestGetTag_NotFound tests getting a non-existent tag
func TestGetTag_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	_, err := client.GetTag(999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedMsg := "tag with ID 999 not found"
	if err.Error() != expectedMsg {
		t.Errorf("expected error '%s', got '%v'", expectedMsg, err)
	}
}

// TestGetUserProfile_Success tests successfully retrieving a user profile
func TestGetUserProfile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user/profile/" {
			t.Errorf("expected path /api/user/profile/, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"theme":                   "auto",
			"bookmark_date_display":   "relative",
			"bookmark_link_target":    "_blank",
			"web_archive_integration": "enabled",
			"tag_search":              "lax",
			"enable_sharing":          true,
			"enable_public_sharing":   true,
			"enable_favicons":         false,
			"display_url":             false,
			"permanent_notes":         false,
			"search_preferences": map[string]interface{}{
				"sort":   "title_asc",
				"shared": "off",
				"unread": "off",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	profile, err := client.GetUserProfile()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if profile == nil {
		t.Fatal("expected profile, got nil")
	}
	if profile.Theme != "auto" {
		t.Errorf("expected theme 'auto', got '%s'", profile.Theme)
	}
	if profile.BookmarkDateDisplay != "relative" {
		t.Errorf("expected bookmark_date_display 'relative', got '%s'", profile.BookmarkDateDisplay)
	}
	if profile.BookmarkLinkTarget != "_blank" {
		t.Errorf("expected bookmark_link_target '_blank', got '%s'", profile.BookmarkLinkTarget)
	}
	if profile.WebArchiveIntegration != "enabled" {
		t.Errorf("expected web_archive_integration 'enabled', got '%s'", profile.WebArchiveIntegration)
	}
	if profile.TagSearch != "lax" {
		t.Errorf("expected tag_search 'lax', got '%s'", profile.TagSearch)
	}
	if !profile.EnableSharing {
		t.Error("expected enable_sharing to be true")
	}
	if !profile.EnablePublicSharing {
		t.Error("expected enable_public_sharing to be true")
	}
	if profile.EnableFavicons {
		t.Error("expected enable_favicons to be false")
	}
	if profile.DisplayURL {
		t.Error("expected display_url to be false")
	}
	if profile.PermanentNotes {
		t.Error("expected permanent_notes to be false")
	}
	if profile.SearchPreferences.Sort != "title_asc" {
		t.Errorf("expected search_preferences.sort 'title_asc', got '%s'", profile.SearchPreferences.Sort)
	}
	if profile.SearchPreferences.Shared != "off" {
		t.Errorf("expected search_preferences.shared 'off', got '%s'", profile.SearchPreferences.Shared)
	}
	if profile.SearchPreferences.Unread != "off" {
		t.Errorf("expected search_preferences.unread 'off', got '%s'", profile.SearchPreferences.Unread)
	}
}

// TestGetUserProfile_Unauthorized tests 401 response
func TestGetUserProfile_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, "bad-token")
	_, err := client.GetUserProfile()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedMsg := "authentication failed. Check your API token"
	if err.Error() != expectedMsg {
		t.Errorf("expected error '%s', got '%v'", expectedMsg, err)
	}
}

// TestGetUserProfile_Forbidden tests 403 response
func TestGetUserProfile_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	_, err := client.GetUserProfile()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedMsg := "insufficient permissions for this operation"
	if err.Error() != expectedMsg {
		t.Errorf("expected error '%s', got '%v'", expectedMsg, err)
	}
}

// TestFetchAllBookmarks_MultiPage tests that FetchAllBookmarks correctly handles pagination
func TestFetchAllBookmarks_MultiPage(t *testing.T) {
	pageNum := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/bookmarks/" {
			t.Errorf("expected path '/api/bookmarks/', got '%s'", r.URL.Path)
		}

		query := r.URL.Query()
		offset := query.Get("offset")
		limit := query.Get("limit")

		if limit != "100" {
			t.Errorf("expected limit '100', got '%s'", limit)
		}

		var response models.BookmarkList

		// Simulate 3 pages of results
		switch offset {
		case "", "0":
			// First page: 100 bookmarks
			pageNum = 1
			bookmarks := make([]models.Bookmark, 100)
			for i := 0; i < 100; i++ {
				bookmarks[i] = models.Bookmark{
					ID:    i + 1,
					URL:   fmt.Sprintf("https://example.com/page1/%d", i+1),
					Title: fmt.Sprintf("Bookmark %d", i+1),
				}
			}
			nextURL := "http://example.com/api/bookmarks/?limit=100&offset=100"
			response = models.BookmarkList{
				Count:   250,
				Next:    &nextURL,
				Results: bookmarks,
			}
		case "100":
			// Second page: 100 bookmarks
			pageNum = 2
			bookmarks := make([]models.Bookmark, 100)
			for i := 0; i < 100; i++ {
				bookmarks[i] = models.Bookmark{
					ID:    i + 101,
					URL:   fmt.Sprintf("https://example.com/page2/%d", i+101),
					Title: fmt.Sprintf("Bookmark %d", i+101),
				}
			}
			nextURL := "http://example.com/api/bookmarks/?limit=100&offset=200"
			response = models.BookmarkList{
				Count:   250,
				Next:    &nextURL,
				Results: bookmarks,
			}
		case "200":
			// Third page: 50 bookmarks (last page)
			pageNum = 3
			bookmarks := make([]models.Bookmark, 50)
			for i := 0; i < 50; i++ {
				bookmarks[i] = models.Bookmark{
					ID:    i + 201,
					URL:   fmt.Sprintf("https://example.com/page3/%d", i+201),
					Title: fmt.Sprintf("Bookmark %d", i+201),
				}
			}
			response = models.BookmarkList{
				Count:   250,
				Next:    nil, // No more pages
				Results: bookmarks,
			}
		default:
			t.Errorf("unexpected offset '%s'", offset)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-token")
	bookmarks, err := client.FetchAllBookmarks(nil, false)

	if err != nil {
		t.Fatalf("FetchAllBookmarks() failed: %v", err)
	}

	// Verify we got all 250 bookmarks from 3 pages
	if len(bookmarks) != 250 {
		t.Errorf("expected 250 bookmarks, got %d", len(bookmarks))
	}

	// Verify the server was called 3 times (for 3 pages)
	if pageNum != 3 {
		t.Errorf("expected 3 pages to be fetched, got %d", pageNum)
	}

	// Verify first bookmark from page 1
	if bookmarks[0].ID != 1 || bookmarks[0].Title != "Bookmark 1" {
		t.Errorf("first bookmark mismatch: got ID=%d, Title=%s", bookmarks[0].ID, bookmarks[0].Title)
	}

	// Verify first bookmark from page 2
	if bookmarks[100].ID != 101 || bookmarks[100].Title != "Bookmark 101" {
		t.Errorf("101st bookmark mismatch: got ID=%d, Title=%s", bookmarks[100].ID, bookmarks[100].Title)
	}

	// Verify first bookmark from page 3
	if bookmarks[200].ID != 201 || bookmarks[200].Title != "Bookmark 201" {
		t.Errorf("201st bookmark mismatch: got ID=%d, Title=%s", bookmarks[200].ID, bookmarks[200].Title)
	}

	// Verify last bookmark
	if bookmarks[249].ID != 250 || bookmarks[249].Title != "Bookmark 250" {
		t.Errorf("last bookmark mismatch: got ID=%d, Title=%s", bookmarks[249].ID, bookmarks[249].Title)
	}
}

// TestDoRequest_Timeout tests that doRequest returns an error when the server times out
func TestDoRequest_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block until client timeout cancels the request
		<-r.Context().Done()
	}))
	defer server.Close()

	// Create a client with a very short timeout for testing
	client := NewClient(server.URL, "test-token")
	// Replace the default 30-second timeout with a 100ms timeout for testing
	client.httpClient.Timeout = 100 * time.Millisecond

	// Try to make a request - should timeout
	err := client.TestConnection()

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	expectedMsg := fmt.Sprintf("cannot connect to %s. Is LinkDing running?", server.URL)
	if err.Error() != expectedMsg {
		t.Errorf("expected error '%s', got '%v'", expectedMsg, err)
	}
}
