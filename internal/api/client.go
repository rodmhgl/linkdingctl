package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rodstewart/linkding-cli/internal/models"
)

// Client is the LinkDing API client
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new LinkDing API client
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// doRequest performs an HTTP request with auth and error handling
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	reqURL := c.baseURL + path
	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Token "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to %s. Is LinkDing running?", c.baseURL)
	}

	return resp, nil
}

// handleErrorResponse converts HTTP error responses into user-friendly messages
func (c *Client) handleErrorResponse(resp *http.Response) error {
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed. Check your API token")
	case http.StatusNotFound:
		return fmt.Errorf("LinkDing not found at %s. Check your URL", c.baseURL)
	case http.StatusBadRequest:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bad request: %s", string(body))
	default:
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}
}

// TestConnection tests the connection to LinkDing
func (c *Client) TestConnection() error {
	resp, err := c.doRequest("GET", "/api/bookmarks/", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// GetBookmarks retrieves a list of bookmarks with optional filters
func (c *Client) GetBookmarks(query string, tags []string, unread, archived *bool, limit, offset int) (*models.BookmarkList, error) {
	params := url.Values{}
	if query != "" {
		params.Set("q", query)
	}
	if len(tags) > 0 {
		// LinkDing expects space-separated tags for AND logic
		params.Set("q", params.Get("q")+" "+strings.Join(tags, " "))
	}
	if unread != nil && *unread {
		params.Set("unread", "yes")
	}
	if archived != nil && *archived {
		params.Set("archived", "yes")
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", offset))
	}

	path := "/api/bookmarks/"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var bookmarkList models.BookmarkList
	if err := json.NewDecoder(resp.Body).Decode(&bookmarkList); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &bookmarkList, nil
}

// GetBookmark retrieves a single bookmark by ID
func (c *Client) GetBookmark(id int) (*models.Bookmark, error) {
	path := fmt.Sprintf("/api/bookmarks/%d/", id)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("bookmark with ID %d not found", id)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var bookmark models.Bookmark
	if err := json.NewDecoder(resp.Body).Decode(&bookmark); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &bookmark, nil
}

// CreateBookmark creates a new bookmark
func (c *Client) CreateBookmark(bookmark *models.BookmarkCreate) (*models.Bookmark, error) {
	resp, err := c.doRequest("POST", "/api/bookmarks/", bookmark)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	var created models.Bookmark
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &created, nil
}

// UpdateBookmark updates an existing bookmark
func (c *Client) UpdateBookmark(id int, update *models.BookmarkUpdate) (*models.Bookmark, error) {
	path := fmt.Sprintf("/api/bookmarks/%d/", id)

	resp, err := c.doRequest("PATCH", path, update)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("bookmark with ID %d not found", id)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var updated models.Bookmark
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updated, nil
}

// DeleteBookmark deletes a bookmark
func (c *Client) DeleteBookmark(id int) error {
	path := fmt.Sprintf("/api/bookmarks/%d/", id)

	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("bookmark with ID %d not found", id)
	}

	if resp.StatusCode != http.StatusNoContent {
		return c.handleErrorResponse(resp)
	}

	return nil
}

// FetchAllBookmarks retrieves all bookmarks from the API, handling pagination automatically.
// If includeArchived is false, only non-archived bookmarks are fetched.
// Tags can be used to filter bookmarks (space-separated for AND logic).
func (c *Client) FetchAllBookmarks(tags []string, includeArchived bool) ([]models.Bookmark, error) {
	var allBookmarks []models.Bookmark
	limit := 100
	offset := 0

	// Determine archived filter
	var archivedPtr *bool
	if !includeArchived {
		archived := false
		archivedPtr = &archived
	}

	for {
		// Fetch a page of bookmarks
		bookmarkList, err := c.GetBookmarks("", tags, nil, archivedPtr, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch bookmarks: %w", err)
		}

		// Add to results
		allBookmarks = append(allBookmarks, bookmarkList.Results...)

		// Check if there are more pages
		if bookmarkList.Next == nil || len(bookmarkList.Results) == 0 {
			break
		}

		offset += limit
	}

	return allBookmarks, nil
}

// GetTags retrieves a list of all tags with optional filters
func (c *Client) GetTags(limit, offset int) (*models.TagList, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", offset))
	}

	path := "/api/tags/"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var tagList models.TagList
	if err := json.NewDecoder(resp.Body).Decode(&tagList); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tagList, nil
}

// FetchAllTags retrieves all tags from the API, handling pagination automatically.
func (c *Client) FetchAllTags() ([]models.Tag, error) {
	var allTags []models.Tag
	limit := 100
	offset := 0

	for {
		// Fetch a page of tags
		tagList, err := c.GetTags(limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tags: %w", err)
		}

		// Add to results
		allTags = append(allTags, tagList.Results...)

		// Check if there are more pages
		if tagList.Next == nil || len(tagList.Results) == 0 {
			break
		}

		offset += limit
	}

	return allTags, nil
}

// CreateTag creates a new tag
func (c *Client) CreateTag(name string) (*models.Tag, error) {
	body := map[string]string{"name": name}

	resp, err := c.doRequest("POST", "/api/tags/", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		// Handle duplicate tag error
		respBody, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(respBody), "already exists") || strings.Contains(string(respBody), "duplicate") {
			return nil, fmt.Errorf("tag '%s' already exists", name)
		}
		return nil, fmt.Errorf("invalid tag name: %s", string(respBody))
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	var created models.Tag
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &created, nil
}

// GetTag retrieves a single tag by ID
func (c *Client) GetTag(id int) (*models.Tag, error) {
	path := fmt.Sprintf("/api/tags/%d/", id)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("tag with ID %d not found", id)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var tag models.Tag
	if err := json.NewDecoder(resp.Body).Decode(&tag); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tag, nil
}

// GetUserProfile retrieves the user profile information
func (c *Client) GetUserProfile() (*models.UserProfile, error) {
	path := "/api/user/profile/"

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed. Check your API token")
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("Insufficient permissions for this operation.")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var profile models.UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &profile, nil
}
