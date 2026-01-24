// Package models defines the data models used in the LinkDing application.
package models

import "time"

// Bookmark represents a LinkDing bookmark
type Bookmark struct {
	ID                 int       `json:"id"`
	URL                string    `json:"url"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	Notes              string    `json:"notes"`
	WebsiteTitle       string    `json:"website_title"`
	WebsiteDescription string    `json:"website_description"`
	IsArchived         bool      `json:"is_archived"`
	Unread             bool      `json:"unread"`
	Shared             bool      `json:"shared"`
	TagNames           []string  `json:"tag_names"`
	DateAdded          time.Time `json:"date_added"`
	DateModified       time.Time `json:"date_modified"`
}

// BookmarkCreate represents the request to create a bookmark
type BookmarkCreate struct {
	URL         string   `json:"url"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Notes       string   `json:"notes,omitempty"`
	IsArchived  bool     `json:"is_archived,omitempty"`
	Unread      bool     `json:"unread,omitempty"`
	Shared      bool     `json:"shared,omitempty"`
	TagNames    []string `json:"tag_names,omitempty"`
}

// BookmarkUpdate represents the request to update a bookmark
type BookmarkUpdate struct {
	URL         *string   `json:"url,omitempty"`
	Title       *string   `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
	IsArchived  *bool     `json:"is_archived,omitempty"`
	Unread      *bool     `json:"unread,omitempty"`
	Shared      *bool     `json:"shared,omitempty"`
	TagNames    *[]string `json:"tag_names,omitempty"`
}

// BookmarkList represents the paginated response from the bookmarks API
type BookmarkList struct {
	Count    int        `json:"count"`
	Next     *string    `json:"next"`
	Previous *string    `json:"previous"`
	Results  []Bookmark `json:"results"`
}

// Tag represents a LinkDing tag
type Tag struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	DateAdded time.Time `json:"date_added"`
}

// TagList represents the paginated response from the tags API
type TagList struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []Tag   `json:"results"`
}

// TagWithCount represents a tag with its bookmark count
type TagWithCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// SearchPreferences represents the user's search preference settings
type SearchPreferences struct {
	Sort   string `json:"sort"`
	Shared string `json:"shared"`
	Unread string `json:"unread"`
}

// UserProfile represents a LinkDing user profile preferences
type UserProfile struct {
	Theme                 string            `json:"theme"`
	BookmarkDateDisplay   string            `json:"bookmark_date_display"`
	BookmarkLinkTarget    string            `json:"bookmark_link_target"`
	WebArchiveIntegration string            `json:"web_archive_integration"`
	TagSearch             string            `json:"tag_search"`
	EnableSharing         bool              `json:"enable_sharing"`
	EnablePublicSharing   bool              `json:"enable_public_sharing"`
	EnableFavicons        bool              `json:"enable_favicons"`
	DisplayURL            bool              `json:"display_url"`
	PermanentNotes        bool              `json:"permanent_notes"`
	SearchPreferences     SearchPreferences `json:"search_preferences"`
}
