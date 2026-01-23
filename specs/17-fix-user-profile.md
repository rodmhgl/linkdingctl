# Specification: Fix User Profile to Match API

## Problem

Spec 12 defined `UserProfile` with fields that do not exist in the LinkDing API (`username`, `display_name`, `bookmark_count`). The `GET /api/user/profile/` endpoint returns **user preferences**, not identity info. Only `theme` was correct.

## Actual API Response

`GET /api/user/profile/`

```json
{
  "theme": "auto",
  "bookmark_date_display": "relative",
  "bookmark_link_target": "_blank",
  "web_archive_integration": "enabled",
  "tag_search": "lax",
  "enable_sharing": true,
  "enable_public_sharing": true,
  "enable_favicons": false,
  "display_url": false,
  "permanent_notes": false,
  "search_preferences": {
    "sort": "title_asc",
    "shared": "off",
    "unread": "off"
  }
}
```

## Updated Model

Replace the `UserProfile` struct in `internal/models/bookmark.go`:

```go
type SearchPreferences struct {
    Sort   string `json:"sort"`
    Shared string `json:"shared"`
    Unread string `json:"unread"`
}

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
```

## Updated Command Output

`linkdingctl user profile`

Human-readable:
```
Theme:                  auto
Bookmark Date Display:  relative
Bookmark Link Target:   _blank
Web Archive:            enabled
Tag Search:             lax
Sharing:                enabled
Public Sharing:         enabled
Favicons:               disabled
Display URL:            disabled
Permanent Notes:        disabled
Search Sort:            title_asc
Search Shared:          off
Search Unread:          off
```

JSON: passes through the API response unchanged (via struct marshalling).

## Updated Jobs to Be Done

- User can inspect their LinkDing display preferences
- User can verify sharing and feature toggle settings
- User can confirm search preference configuration
- User can troubleshoot authentication (401/403 errors still apply)

## Implementation Notes

- Replace `UserProfile` struct in `internal/models/bookmark.go` with correct fields
- Add `SearchPreferences` struct in the same file
- Rewrite `userProfileCmd.RunE` in `cmd/linkdingctl/user.go` to display all real fields
- Use `"enabled"`/`"disabled"` rendering for boolean fields in human output
- Flatten `search_preferences` into `Search Sort`, `Search Shared`, `Search Unread` rows in human output
- Update `TestGetUserProfile_Success` in `internal/api/client_test.go` to use real API fields
- Update `TestUserProfileCommand` in `cmd/linkdingctl/commands_test.go` to assert real fields
- Remove all references to `Username`, `DisplayName`, `BookmarkCount`

## Success Criteria

- [ ] `UserProfile` struct matches the actual LinkDing API response exactly
- [ ] `linkdingctl user profile` displays all preference fields
- [ ] `linkdingctl user profile --json` outputs valid JSON matching API schema
- [ ] Boolean fields render as "enabled"/"disabled" in human output
- [ ] Nested `search_preferences` fields are displayed as flattened rows
- [ ] HTTP 401 still produces: "Authentication failed. Check your API token."
- [ ] HTTP 403 still produces: "Insufficient permissions for this operation."
- [ ] All existing tests updated and passing
- [ ] `make cover` passes with 70% threshold