package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
		json.NewEncoder(w).Encode(models.BookmarkList{})
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(response)
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
		json.NewEncoder(w).Encode(bookmark)
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
		json.NewEncoder(w).Encode(created)
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
		json.NewEncoder(w).Encode(updated)
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
