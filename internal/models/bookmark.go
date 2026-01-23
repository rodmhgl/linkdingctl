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

// UserProfile represents a LinkDing user profile
type UserProfile struct {
	Theme         string `json:"theme"`
	BookmarkCount int    `json:"bookmark_count"`
	DisplayName   string `json:"display_name"`
	Username      string `json:"username"`
}
