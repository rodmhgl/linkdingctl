# Specification: Tags Performance & Pagination

## Jobs to Be Done
- `ld tags` responds in reasonable time regardless of tag count
- Tag rename/delete operations process ALL matching bookmarks, not just the first page
- Tags list handles more than 1000 tags

## N+1 Query Problem in `ld tags`

Current behavior: makes 1 API call per tag to count bookmarks (N+1 pattern). With 100 tags, that's 101 HTTP requests.

### Implementation

Fetch all bookmarks once, count tags client-side:
```go
// Fetch all bookmarks (paginated)
allBookmarks, err := fetchAllBookmarks(client, nil, true)

// Count tags
tagCounts := make(map[string]int)
for _, b := range allBookmarks {
    for _, tag := range b.TagNames {
        tagCounts[tag] = tagCounts[tag] + 1
    }
}
```

This reduces N+1 API calls to a single paginated fetch.

**Trade-off**: For instances with thousands of bookmarks but few tags, this fetches more data than needed. For typical usage (more tags than bookmarks-per-tag), this is significantly faster.

## Pagination for Tag Rename & Delete

Current behavior: `client.GetBookmarks(..., 1000, 0)` — only processes the first 1000 bookmarks. Bookmarks beyond 1000 are silently skipped.

### Implementation

Use the same `fetchAllBookmarks` pagination pattern from `internal/export/json.go`:
```go
allBookmarks, err := fetchAllBookmarks(client, []string{tagName}, true)
```

This loops until `Next == nil`, handling any number of bookmarks.

Affected files:
- `cmd/ld/tags.go` — `runTagsRename()` line 177
- `cmd/ld/tags.go` — `runTagsDelete()` line 271

**Note**: `fetchAllBookmarks` is currently in `internal/export/json.go`. It should be either:
- Moved to `internal/api/client.go` as a method on `Client`, or
- Extracted to a shared utility in `internal/api/`

## Tags List Pagination

Current behavior: `client.GetTags(1000, 0)` — only shows the first 1000 tags.

### Implementation

Loop with pagination like `fetchAllBookmarks`:
```go
var allTags []models.Tag
offset := 0
limit := 100
for {
    tagList, err := client.GetTags(limit, offset)
    // ...
    allTags = append(allTags, tagList.Results...)
    if tagList.Next == nil || len(tagList.Results) == 0 {
        break
    }
    offset += limit
}
```

## Success Criteria
- [ ] `ld tags` makes at most O(N/page_size) API calls, not O(tags) calls
- [ ] `ld tags rename` processes all matching bookmarks regardless of count
- [ ] `ld tags delete --force` processes all matching bookmarks regardless of count
- [ ] `ld tags` displays all tags even if count exceeds 1000
- [ ] `fetchAllBookmarks` is accessible from both `export` and `cmd` packages
- [ ] Progress output shows correct totals (e.g., "Updating 1/1500...")
