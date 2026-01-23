# Implementation Plan - LinkDing CLI

> Gap analysis based on specs 01-07 vs current codebase. All Phase 0-3 original items are complete.
> This plan covers remaining gaps identified from specs 05, 06, 07, and 03.

## Phase 5: Security Hardening (spec 05)

- [x] **P1** | Config file permissions hardening | ~small
  - Acceptance: `Save()` creates directory with `0700`, file with `0600`; existing configs not re-permissioned on read
  - Files: internal/config/config.go
  - Details: Change `os.MkdirAll(dir, 0755)` to `0700`; add `os.Chmod(configPath, 0600)` after `v.WriteConfig()`

- [x] **P1** | Token input masking in config init | ~small
  - Acceptance: `ld config init` does not echo token; works when stdin is piped (non-TTY fallback)
  - Files: cmd/ld/config.go, go.mod (add `golang.org/x/term`)
  - Details: Use `term.ReadPassword(int(os.Stdin.Fd()))` for token prompt; detect non-TTY with `term.IsTerminal()` and fall back to `bufio.Reader`

- [x] **P1** | Safe JSON output in backup command | ~small
  - Acceptance: `ld backup --json` produces valid JSON regardless of output path characters (quotes, backslashes)
  - Files: cmd/ld/backup.go
  - Details: Replace `fmt.Printf("{\"file\": \"%s\"}\n", fullPath)` with `json.NewEncoder(os.Stdout).Encode(map[string]string{"file": fullPath})`

## Phase 6: Tags Performance & Pagination (spec 06)

- [x] **P0** | Extract fetchAllBookmarks to shared location | ~small
  - Acceptance: `fetchAllBookmarks` is accessible from both `internal/export` and `cmd/ld` packages
  - Files: internal/api/client.go (add `FetchAllBookmarks` method), internal/export/json.go (update to call new method)
  - Details: Move pagination logic to `Client.FetchAllBookmarks(tags []string, includeArchived bool) ([]models.Bookmark, error)` as exported method on Client; update `internal/export/json.go` to delegate to it

- [x] **P1** | Fix N+1 query in `ld tags` | ~medium
  - Acceptance: `ld tags` makes at most O(N/page_size) API calls, not O(tags) calls; tag counts are correct
  - Files: cmd/ld/tags.go
  - Details: Replace per-tag `GetBookmarks` loop with single `FetchAllBookmarks(nil, true)` call, then count tags client-side in a `map[string]int`
  - Depends on: Extract fetchAllBookmarks

- [x] **P1** | Paginate tag rename operations | ~small
  - Acceptance: `ld tags rename` processes all matching bookmarks regardless of count (not just first 1000)
  - Files: cmd/ld/tags.go (runTagsRename around line 177)
  - Details: Replace `client.GetBookmarks("", []string{oldTag}, nil, nil, 1000, 0)` with `client.FetchAllBookmarks([]string{oldTag}, true)`; update progress display to use `len(allBookmarks)` instead of `bookmarkList.Count`
  - Depends on: Extract fetchAllBookmarks

- [ ] **P1** | Paginate tag delete operations | ~small
  - Acceptance: `ld tags delete --force` processes all matching bookmarks regardless of count
  - Files: cmd/ld/tags.go (runTagsDelete around line 271)
  - Details: Same pattern as rename — replace fixed-limit fetch with `FetchAllBookmarks`
  - Depends on: Extract fetchAllBookmarks

- [ ] **P1** | Paginate tags list | ~small
  - Acceptance: `ld tags` displays all tags even if count exceeds 1000
  - Files: cmd/ld/tags.go (runTags around line 60), internal/api/client.go
  - Details: Add `Client.FetchAllTags() ([]models.Tag, error)` that loops with pagination until `Next == nil`; replace `client.GetTags(1000, 0)` with `client.FetchAllTags()`

## Phase 7: Missing Commands (spec 03)

- [ ] **P2** | Tags show subcommand | ~small
  - Acceptance: `ld tags show <tag-name>` lists all bookmarks with that tag (delegates to list --tags behavior)
  - Files: cmd/ld/tags.go
  - Details: Add `tagsShowCmd` that calls `runList` equivalent with the specified tag filter; respects `--json` flag

## Phase 8: Test Coverage (spec 07)

- [ ] **P1** | API client pagination tests | ~medium
  - Acceptance: Tests cover multi-page `GetBookmarks` fetch, `GetTags` pagination, `FetchAllBookmarks` with multiple pages
  - Files: internal/api/client_test.go
  - Details: Mock multi-page responses with `Next` pointer set; verify all pages are fetched and combined

- [ ] **P1** | HTML export tests | ~medium
  - Acceptance: `ExportHTML` produces valid Netscape format; HTML-escapes URLs, titles, descriptions; includes TAGS attribute; omits DD when description empty
  - Files: internal/export/html_test.go (new)
  - Details: Use mock `httptest.Server` or refactor to accept `[]models.Bookmark` directly for testability

- [ ] **P1** | Import tests (all formats) | ~large
  - Acceptance: Round-trip tests for JSON/CSV; HTML import parses correctly; `DetectFormat` works; duplicate handling works; `--add-tags` appends correctly
  - Files: internal/export/import_test.go (new)
  - Details: Test `importJSON`, `importHTML`, `importCSV` with fixture data; test `DetectFormat` for all extensions and unknown; test skip-duplicates vs update behavior

- [ ] **P2** | Command-level tests | ~large
  - Acceptance: Core commands tested via cobra's `Execute()` with `httptest.NewServer`; `go test ./...` covers `cmd/ld`
  - Files: cmd/ld/commands_test.go (new)
  - Details: Pattern from spec: `executeCommand(args ...string)` helper; test `add`, `list`, `list --json`, `get`, `update --add-tags`, `delete --force`, `export -f csv`, `backup`, `tags`, `config show`

- [ ] **P2** | Integration test for stdin interaction | ~small
  - Acceptance: Delete confirmation and config init can be tested with piped stdin
  - Files: cmd/ld/commands_test.go (extend)
  - Details: Use `os.Pipe()` pattern from spec to provide "y\n" via stdin

- [ ] **P3** | Coverage target validation | ~small
  - Acceptance: `go test -cover ./...` reports >70% coverage for each package
  - Files: Makefile (add `cover` target)
  - Details: Add `make cover` that runs `go test -cover ./...` and fails if any package is below 70%

---

## Dependency Graph

```
Phase 6 (Performance):
  Extract fetchAllBookmarks (P0)
    -> Fix N+1 in ld tags (P1)
    -> Paginate tag rename (P1)
    -> Paginate tag delete (P1)
    -> Paginate tags list (P1)

Phase 5 (Security): Independent, can run in parallel with Phase 6
Phase 7 (Commands): Independent
Phase 8 (Tests): Should run last (tests validate all other changes)
```

## Execution Order

1. **P0**: Extract `fetchAllBookmarks` to `api.Client` (blocks all Phase 6 items)
2. **P1 Security**: Config permissions, token masking, backup JSON (independent batch)
3. **P1 Performance**: N+1 fix, pagination for rename/delete/list (depends on #1)
4. **P2**: Tags show command, command-level tests
5. **P3**: Coverage validation

---

## Notes

- `golang.org/x/term` dependency needed for token masking (spec 05)
- The existing `IMPLEMENTATION_PLAN.md` had all original phases marked complete; this plan covers only remaining gaps
- Specs 01, 02, 04 are fully implemented with no remaining gaps
- The `tags show` subcommand (spec 03) was never implemented despite tags.go being marked complete
