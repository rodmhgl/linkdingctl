# Specification: Staticcheck SA9003 — Empty Branches in HTML Import

## Jobs to Be Done
- Resolve three SA9003 (empty branch) violations reported by `staticcheck` in `internal/export/import.go`
- Remove the misleading `error` return from `processHTMLBookmark` since errors are already recorded in the `result` struct
- Prevent future misuse where a caller might double-count failures by handling the (redundant) returned error

## Problem: Redundant Error Return in `processHTMLBookmark`

`internal/export/import.go` — the `processHTMLBookmark` function records errors directly into the `*ImportResult` struct (incrementing `result.Failed`, appending to `result.Errors`), then also returns the error. All three call sites in `importHTML` discard the return value with empty `if err != nil {}` branches:

```go
// Line ~229
if err := processHTMLBookmark(client, result, existingURLs, lastURL, lastTitle, lastTags, "", lineNum-1, options); err != nil {
    // Error already recorded in result
}

// Line ~248
if err := processHTMLBookmark(client, result, existingURLs, lastURL, lastTitle, lastTags, description, lineNum, options); err != nil {
    // Error already recorded in result
}

// Line ~259
if err := processHTMLBookmark(client, result, existingURLs, lastURL, lastTitle, lastTags, "", lineNum, options); err != nil {
    // Error already recorded in result
}
```

This triggers staticcheck SA9003 at all three locations.

### Root Cause

The function's error return is architecturally redundant — errors are already captured in the `result` struct before the function returns. The return value exists only as a leftover from an earlier design.

### Implementation

**Step 1: Remove the `error` return type from `processHTMLBookmark` (line ~272)**

```go
// Before
func processHTMLBookmark(client *api.Client, result *ImportResult, existingURLs map[string]int, url, title string, tags []string, description string, lineNum int, options ImportOptions) error {

// After
func processHTMLBookmark(client *api.Client, result *ImportResult, existingURLs map[string]int, url, title string, tags []string, description string, lineNum int, options ImportOptions) {
```

**Step 2: Convert all `return err` / `return nil` statements inside the function to bare `return`**

The function has multiple return points where it returns `nil` on success or `err`/`fmt.Errorf(...)` on failure. Since errors are already recorded in `result.Errors` before these returns, change them all to `return`.

**Step 3: Simplify the three call sites in `importHTML`**

Remove the `if err := ...; err != nil {}` wrapping:

```go
// Before
if err := processHTMLBookmark(client, result, existingURLs, lastURL, lastTitle, lastTags, "", lineNum-1, options); err != nil {
    // Error already recorded in result
}

// After
processHTMLBookmark(client, result, existingURLs, lastURL, lastTitle, lastTags, "", lineNum-1, options)
```

Apply the same transformation at all three call sites (~lines 229, 248, 259).

Affected files:
- `internal/export/import.go` — `processHTMLBookmark` function and its three call sites in `importHTML`

## Success Criteria
- [ ] `processHTMLBookmark` no longer returns `error`
- [ ] All three call sites invoke the function without error-checking wrappers
- [ ] `golangci-lint run ./...` reports no SA9003 violations in `internal/export/import.go`
- [ ] All existing tests pass (`go test ./...`)
- [ ] Coverage gate passes (`make cover`)
