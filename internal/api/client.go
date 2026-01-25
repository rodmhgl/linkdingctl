// Package api provides a client for the LinkDing API.
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

// Client is the LinkDing API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new LinkDing API client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// doRequest performs an HTTP request with authentication headers.
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
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

// decodeResponse checks the status code and decodes the JSON response body into dest.
// If the status code does not match expectedStatus, it returns an appropriate error.
func (c *Client) decodeResponse(resp *http.Response, expectedStatus int, dest interface{}) error {
	if resp.StatusCode != expectedStatus {
		return c.handleErrorResponse(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

// handleErrorResponse converts HTTP error responses into user-friendly messages.
func (c *Client) handleErrorResponse(resp *http.Response) error {
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed. Check your API token")
	case http.StatusNotFound:
		return fmt.Errorf("LinkDing not found at %s. Check your URL", c.baseURL)
	default:
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusBadRequest {
			return fmt.Errorf("bad request: %s", string(body))
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}
}

// TestConnection tests the connection to LinkDing.
func (c *Client) TestConnection() error {
	resp, err := c.doRequest("GET", "/api/bookmarks/", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}
	return nil
}

// GetBookmarks retrieves a list of bookmarks with optional filters.
func (c *Client) GetBookmarks(query string, tags []string, unread, archived *bool, limit, offset int) (*models.BookmarkList, error) {
	params := url.Values{}
	if query != "" {
		params.Set("q", query)
	}
	if len(tags) > 0 {
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
	defer func() { _ = resp.Body.Close() }()

	var bookmarkList models.BookmarkList
	if err := c.decodeResponse(resp, http.StatusOK, &bookmarkList); err != nil {
		return nil, err
	}
	return &bookmarkList, nil
}

// GetBookmark retrieves a single bookmark by ID.
func (c *Client) GetBookmark(id int) (*models.Bookmark, error) {
	path := fmt.Sprintf("/api/bookmarks/%d/", id)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("bookmark with ID %d not found", id)
	}

	var bookmark models.Bookmark
	if err := c.decodeResponse(resp, http.StatusOK, &bookmark); err != nil {
		return nil, err
	}
	return &bookmark, nil
}

// CreateBookmark creates a new bookmark.
func (c *Client) CreateBookmark(bookmark *models.BookmarkCreate) (*models.Bookmark, error) {
	resp, err := c.doRequest("POST", "/api/bookmarks/", bookmark)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var created models.Bookmark
	if err := c.decodeResponse(resp, http.StatusCreated, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateBookmark updates an existing bookmark.
func (c *Client) UpdateBookmark(id int, update *models.BookmarkUpdate) (*models.Bookmark, error) {
	path := fmt.Sprintf("/api/bookmarks/%d/", id)

	resp, err := c.doRequest("PATCH", path, update)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("bookmark with ID %d not found", id)
	}

	var updated models.Bookmark
	if err := c.decodeResponse(resp, http.StatusOK, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

// DeleteBookmark deletes a bookmark by ID.
func (c *Client) DeleteBookmark(id int) error {
	path := fmt.Sprintf("/api/bookmarks/%d/", id)

	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("bookmark with ID %d not found", id)
	}
	if resp.StatusCode != http.StatusNoContent {
		return c.handleErrorResponse(resp)
	}
	return nil
}

// FetchAllBookmarks retrieves all bookmarks, handling pagination automatically.
// If includeArchived is false, only non-archived bookmarks are fetched.
func (c *Client) FetchAllBookmarks(tags []string, includeArchived bool) ([]models.Bookmark, error) {
	var allBookmarks []models.Bookmark
	limit := 100
	offset := 0

	var archivedPtr *bool
	if !includeArchived {
		archived := false
		archivedPtr = &archived
	}

	for {
		bookmarkList, err := c.GetBookmarks("", tags, nil, archivedPtr, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch bookmarks: %w", err)
		}

		allBookmarks = append(allBookmarks, bookmarkList.Results...)

		if bookmarkList.Next == nil || len(bookmarkList.Results) == 0 {
			break
		}
		offset += limit
	}

	return allBookmarks, nil
}

// GetTags retrieves a list of tags with optional pagination.
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
	defer func() { _ = resp.Body.Close() }()

	var tagList models.TagList
	if err := c.decodeResponse(resp, http.StatusOK, &tagList); err != nil {
		return nil, err
	}
	return &tagList, nil
}

// FetchAllTags retrieves all tags, handling pagination automatically.
func (c *Client) FetchAllTags() ([]models.Tag, error) {
	var allTags []models.Tag
	limit := 100
	offset := 0

	for {
		tagList, err := c.GetTags(limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tags: %w", err)
		}

		allTags = append(allTags, tagList.Results...)

		if tagList.Next == nil || len(tagList.Results) == 0 {
			break
		}
		offset += limit
	}

	return allTags, nil
}

// CreateTag creates a new tag with the given name.
func (c *Client) CreateTag(name string) (*models.Tag, error) {
	body := map[string]string{"name": name}

	resp, err := c.doRequest("POST", "/api/tags/", body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusBadRequest {
		respBody, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(respBody), "already exists") || strings.Contains(string(respBody), "duplicate") {
			return nil, fmt.Errorf("tag '%s' already exists", name)
		}
		return nil, fmt.Errorf("invalid tag name: %s", string(respBody))
	}

	var created models.Tag
	if err := c.decodeResponse(resp, http.StatusCreated, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// GetTag retrieves a single tag by ID.
func (c *Client) GetTag(id int) (*models.Tag, error) {
	path := fmt.Sprintf("/api/tags/%d/", id)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("tag with ID %d not found", id)
	}

	var tag models.Tag
	if err := c.decodeResponse(resp, http.StatusOK, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetUserProfile retrieves the user profile information.
func (c *Client) GetUserProfile() (*models.UserProfile, error) {
	resp, err := c.doRequest("GET", "/api/user/profile/", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("authentication failed. Check your API token")
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("insufficient permissions for this operation")
	}

	var profile models.UserProfile
	if err := c.decodeResponse(resp, http.StatusOK, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// GetBundles retrieves a list of bundles with optional pagination.
func (c *Client) GetBundles(limit, offset int) (*models.BundleList, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", offset))
	}

	path := "/api/bundles/"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var bundleList models.BundleList
	if err := c.decodeResponse(resp, http.StatusOK, &bundleList); err != nil {
		return nil, err
	}
	return &bundleList, nil
}

// FetchAllBundles retrieves all bundles, handling pagination automatically.
func (c *Client) FetchAllBundles() ([]models.Bundle, error) {
	var allBundles []models.Bundle
	limit := 100
	offset := 0

	for {
		bundleList, err := c.GetBundles(limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch bundles: %w", err)
		}

		allBundles = append(allBundles, bundleList.Results...)

		if bundleList.Next == nil || len(bundleList.Results) == 0 {
			break
		}
		offset += limit
	}

	return allBundles, nil
}

// GetBundle retrieves a single bundle by ID.
func (c *Client) GetBundle(id int) (*models.Bundle, error) {
	path := fmt.Sprintf("/api/bundles/%d/", id)

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("bundle with ID %d not found", id)
	}

	var bundle models.Bundle
	if err := c.decodeResponse(resp, http.StatusOK, &bundle); err != nil {
		return nil, err
	}
	return &bundle, nil
}

// CreateBundle creates a new bundle.
func (c *Client) CreateBundle(bundle *models.BundleCreate) (*models.Bundle, error) {
	resp, err := c.doRequest("POST", "/api/bundles/", bundle)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var created models.Bundle
	if err := c.decodeResponse(resp, http.StatusCreated, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateBundle updates an existing bundle.
func (c *Client) UpdateBundle(id int, update *models.BundleUpdate) (*models.Bundle, error) {
	path := fmt.Sprintf("/api/bundles/%d/", id)

	resp, err := c.doRequest("PATCH", path, update)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("bundle with ID %d not found", id)
	}

	var updated models.Bundle
	if err := c.decodeResponse(resp, http.StatusOK, &updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

// DeleteBundle deletes a bundle by ID.
func (c *Client) DeleteBundle(id int) error {
	path := fmt.Sprintf("/api/bundles/%d/", id)

	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("bundle with ID %d not found", id)
	}
	if resp.StatusCode != http.StatusNoContent {
		return c.handleErrorResponse(resp)
	}
	return nil
}
