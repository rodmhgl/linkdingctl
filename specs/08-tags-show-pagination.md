# Specification: Tags Show Pagination

## Jobs to Be Done
- `ld tags show <tag-name>` displays ALL bookmarks with that tag, not just the first 1000
- Tags show output is consistent with the pagination behavior of `tags rename` and `tags delete`

## Problem

Current behavior: `runTagsShow` calls `client.GetBookmarks("", []string{tagName}, nil, nil, 1000, 0)` — only displays the first 1000 bookmarks. Users with more than 1000 bookmarks for a given tag see silently truncated results.

This is the same pagination defect fixed in `tags rename` and `tags delete` (spec 06), but `tags show` was not updated.

## Implementation

Replace the hardcoded-limit `GetBookmarks` call with `FetchAllBookmarks`:

```go
// Fetch ALL bookmarks with the specified tag (paginated)
allBookmarks, err := client.FetchAllBookmarks([]string{tagName}, true)
if err != nil {
    return err
}

bookmarkList := &models.BookmarkList{
    Count:   len(allBookmarks),
    Results: allBookmarks,
}
```

Affected files:
- `cmd/ld/tags.go` — `runTagsShow()` function

## Success Criteria
- [ ] `ld tags show <tag>` returns all matching bookmarks regardless of count
- [ ] `ld tags show <tag>` uses `FetchAllBookmarks` (not `GetBookmarks` with a fixed limit)
- [ ] Output count reflects the true total, not capped at 1000
- [ ] `--json` output includes all bookmarks