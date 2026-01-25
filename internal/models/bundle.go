package models

import "time"

// Bundle represents a LinkDing bundle (saved search configuration)
type Bundle struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Search       string    `json:"search"`
	AnyTags      string    `json:"any_tags"`
	AllTags      string    `json:"all_tags"`
	ExcludedTags string    `json:"excluded_tags"`
	Order        int       `json:"order"`
	DateCreated  time.Time `json:"date_created"`
	DateModified time.Time `json:"date_modified"`
}

// BundleCreate represents the request to create a bundle
type BundleCreate struct {
	Name         string `json:"name"`
	Search       string `json:"search,omitempty"`
	AnyTags      string `json:"any_tags,omitempty"`
	AllTags      string `json:"all_tags,omitempty"`
	ExcludedTags string `json:"excluded_tags,omitempty"`
	Order        int    `json:"order,omitempty"`
}

// BundleUpdate represents the request to update a bundle with PATCH semantics
type BundleUpdate struct {
	Name         *string `json:"name,omitempty"`
	Search       *string `json:"search,omitempty"`
	AnyTags      *string `json:"any_tags,omitempty"`
	AllTags      *string `json:"all_tags,omitempty"`
	ExcludedTags *string `json:"excluded_tags,omitempty"`
	Order        *int    `json:"order,omitempty"`
}

// BundleList represents the paginated response from the bundles API
type BundleList struct {
	Count    int      `json:"count"`
	Next     *string  `json:"next"`
	Previous *string  `json:"previous"`
	Results  []Bundle `json:"results"`
}
