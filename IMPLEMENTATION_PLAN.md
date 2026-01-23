# Implementation Plan - LinkDing CLI

> Gap analysis based on specs 01-17 vs current codebase (2026-01-23).
> Specs 01-16 are fully implemented. Spec 17 (Fix User Profile) introduces an API alignment fix.

## Current State Summary

| Spec | Status | Notes |
|------|--------|-------|
| 01 - Core CLI | Complete | Config init/show/test, env overrides, --json, --debug all work |
| 02 - Bookmark CRUD | Complete | add/list/get/update/delete all implemented with --json |
| 03 - Tags | Complete | tags/tags show/rename/delete all implemented |
| 04 - Import/Export | Complete | JSON/HTML/CSV export+import, backup, restore with --wipe |
| 05 - Security | Complete | 0700/0600 perms, token masking, safe JSON backup output |
| 06 - Tags Performance | Complete | FetchAllBookmarks/FetchAllTags on Client, client-side counting |
| 07 - Test Coverage | Complete | All packages meet 70% threshold |
| 08 - Tags Show Pagination | Complete | Uses FetchAllBookmarks for all bookmarks |
| 09 - Makefile Portability | Complete | `bc` replaced with `awk` in cover target |
| 10 - Config Token Trim | Complete | Redundant TrimSpace removed |
| 11 - Test Robustness | Complete | Unsafe slicing fixed, missing flag resets added |
| 12 - User Profile | Complete | `user profile` command exists (uses incorrect API model — see spec 17) |
| 13 - Tags CRUD Enhancements | Complete | `tags create` and `tags get` subcommands implemented |
| 14 - Per-Command Overrides | Complete | Global `--url` and `--token` flags implemented |
| 15 - Notes Field | Complete | `--notes` flag on add/update commands implemented |
| 16 - Rename to linkdingctl | Complete | Binary, package, config path, docs renamed |
| 17 - Fix User Profile | Complete | UserProfile model corrected to match real API; all tests updated and passing |

## Coverage Status

```
Package                    Current   Target   Status
cmd/linkdingctl            80.7%     70%      PASS
internal/api               79.5%     70%      PASS
internal/config            83.1%     70%      PASS
internal/export            78.4%     70%      PASS
```

---

## Remaining Tasks

### Phase 1: Fix UserProfile Model (spec 17)

- [x] **P0** | Replace `UserProfile` struct with correct API fields | ~small
  - Acceptance: `UserProfile` struct has fields: Theme, BookmarkDateDisplay, BookmarkLinkTarget, WebArchiveIntegration, TagSearch, EnableSharing, EnablePublicSharing, EnableFavicons, DisplayURL, PermanentNotes, SearchPreferences (nested struct); no Username/DisplayName/BookmarkCount fields remain
  - Files: `internal/models/bookmark.go`
  - Details: Add `SearchPreferences` struct with Sort, Shared, Unread string fields (all with json tags). Replace current 4-field `UserProfile` with the 11-field version matching the actual LinkDing `GET /api/user/profile/` response.

- [x] **P0** | Update `user profile` command output | ~medium
  - Acceptance: Human output shows all preference fields; boolean fields render as "enabled"/"disabled"; nested `search_preferences` displayed as flattened rows (Search Sort, Search Shared, Search Unread); `--json` outputs full struct matching API schema; removes all references to Username/DisplayName/BookmarkCount
  - Files: `cmd/linkdingctl/user.go`
  - Details: Rewrite `userProfileCmd.RunE` to display: Theme, Bookmark Date Display, Bookmark Link Target, Web Archive, Tag Search, Sharing (enabled/disabled), Public Sharing (enabled/disabled), Favicons (enabled/disabled), Display URL (enabled/disabled), Permanent Notes (enabled/disabled), Search Sort, Search Shared, Search Unread. JSON output uses `json.NewEncoder` with `SetIndent`.

- [x] **P0** | Update 403 error message to match spec | ~small
  - Acceptance: HTTP 403 from `GetUserProfile` returns "Insufficient permissions for this operation." (not "access forbidden")
  - Files: `internal/api/client.go`
  - Details: Change the 403 error string in `GetUserProfile()` from "access forbidden. You don't have permission to view this profile" to "Insufficient permissions for this operation."

### Phase 2: Fix User Profile Tests (spec 17)

- [x] **P1** | Update API client test for GetUserProfile | ~small
  - Acceptance: Mock server returns realistic API response with all real fields (theme, bookmark_date_display, bookmark_link_target, web_archive_integration, tag_search, enable_sharing, enable_public_sharing, enable_favicons, display_url, permanent_notes, search_preferences); test verifies correct deserialization including nested SearchPreferences; no references to Username/DisplayName/BookmarkCount
  - Files: `internal/api/client_test.go`

- [x] **P1** | Update command test for user profile | ~medium
  - Acceptance: Command test mock returns real API fields; human output test asserts presence of: "Theme:", "Bookmark Date Display:", "Bookmark Link Target:", "Web Archive:", "Tag Search:", "Sharing:", "Public Sharing:", "Favicons:", "Display URL:", "Permanent Notes:", "Search Sort:", "Search Shared:", "Search Unread:"; JSON output test asserts keys match API schema; boolean fields test "enabled"/"disabled" rendering; HTTP 401/403 error tests assert correct messages
  - Files: `cmd/linkdingctl/commands_test.go`

### Phase 3: Verify Coverage & Quality

- [x] **P2** | Run `make cover` and verify all packages pass 70% threshold | ~small
  - Acceptance: `make cover` passes; `go test ./...` all green; no regressions in any package
  - Files: (none — verification only)

---

## Dependency Graph

```
Phase 1 (Fix Model + Command + Error Msg) ── must be first (model change breaks existing tests)
        │
        ▼
Phase 2 (Fix Tests) ──────────────────────── depends on Phase 1 (tests reference new struct)
        │
        ▼
Phase 3 (Verify Coverage) ────────────────── depends on Phase 2 (needs passing tests)
```

## Notes

- Specs 01-16 are fully implemented — no code changes needed for those
- The `GetUserProfile` API client method needs only the 403 message update; the HTTP logic and endpoint are correct
- The `SearchPreferences` struct is new — needed for the nested JSON object in the API response
- No new dependencies needed — all required packages already in go.mod
- The change is backwards-incompatible at the struct level (fields removed/renamed) but this is intentional: the old fields never existed in the real API
- All existing tests that reference UserProfile fields will fail after Phase 1 — this is expected and fixed in Phase 2
- After all phases, `make cover` should continue to pass the 70% threshold
