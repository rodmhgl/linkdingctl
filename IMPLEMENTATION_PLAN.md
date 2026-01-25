# Implementation Plan

Gap analysis between specifications (`specs/01`–`specs/23`) and the current codebase.

---

## P0 — Foundational / Already Implemented

All P0 (foundational/blocking) items are complete. No new features need to be built from scratch.

---

## P1 — Bundles Support (Spec 23) — NEW FEATURE

Bundle CRUD is the primary remaining feature. LinkDing supports bundles (saved search configurations), but `linkdingctl` currently has no bundle commands.

- [x] **P1** | Create Bundle model structs | ~small
  - Acceptance: `internal/models/bundle.go` exists with `Bundle`, `BundleCreate`, `BundleUpdate`, `BundleList` structs
  - Acceptance: `BundleUpdate` uses pointer fields for PATCH semantics (matching `BookmarkUpdate` pattern)
  - Acceptance: All fields match LinkDing API: id, name, search, any_tags, all_tags, excluded_tags, order, date_created, date_modified
  - Files: `internal/models/bundle.go` (new)

- [ ] **P1** | Add Bundle API client methods | ~medium
  - Acceptance: `GetBundles(limit, offset int)` returns paginated bundle list
  - Acceptance: `FetchAllBundles()` handles pagination transparently
  - Acceptance: `GetBundle(id int)` retrieves single bundle
  - Acceptance: `CreateBundle(*BundleCreate)` creates bundle via POST
  - Acceptance: `UpdateBundle(id int, *BundleUpdate)` updates bundle via PATCH
  - Acceptance: `DeleteBundle(id int)` deletes bundle via DELETE
  - Files: `internal/api/client.go`

- [ ] **P1** | Add unit tests for Bundle API client | ~medium
  - Acceptance: All 6 bundle API methods have tests using `httptest.NewServer`
  - Acceptance: Tests cover success cases and error cases (404, 400, 401)
  - Acceptance: Pagination test with multi-page mock response
  - Files: `internal/api/client_test.go`

- [ ] **P1** | Create `bundles` command with subcommands | ~medium
  - Acceptance: `linkdingctl bundles --help` shows available subcommands
  - Acceptance: `linkdingctl bundles list` displays all bundles in table (ID, Name, Search, Order)
  - Acceptance: `linkdingctl bundles list --json` outputs valid JSON array
  - Acceptance: `linkdingctl bundles get <id>` displays full bundle details (all fields)
  - Acceptance: `linkdingctl bundles create <name>` creates bundle (supports `--search`, `--any-tags`, `--all-tags`, `--excluded-tags`, `--order`)
  - Acceptance: `linkdingctl bundles update <id>` with flags updates only specified fields (PATCH semantics)
  - Acceptance: `linkdingctl bundles delete <id>` removes the bundle
  - Acceptance: All commands respect `--json` flag
  - Files: `cmd/linkdingctl/bundles.go` (new)

- [ ] **P1** | Add CLI command tests for bundles | ~medium
  - Acceptance: All bundle subcommands have tests using mock HTTP server
  - Acceptance: Tests cover JSON and table output modes
  - Acceptance: Tests verify error handling for invalid inputs
  - Acceptance: Coverage meets 70% threshold for new code
  - Files: `cmd/linkdingctl/commands_test.go`

---

## P2 — Staticcheck SA9003 (Spec 22)

- [x] **P2** | Fix SA9003: Remove error return from `processHTMLBookmark` | ~small
  - Acceptance: `processHTMLBookmark` has no `error` return type; all three call sites in `importHTML` invoke it as a bare statement (no `if err != nil {}` wrapper); `golangci-lint run ./...` reports no SA9003 in `internal/export/import.go`; all tests pass (`go test ./...`); coverage gate passes (`make cover`)
  - Files: `internal/export/import.go`

---

## P3 — Lint & Code Quality

### Previously Completed Items (Specs 10, 11, 14, 19)

All lint/code quality items from previous specs have been implemented:
- Config token trim (spec 10) ✓
- Test robustness (spec 11) ✓
- Per-command connection override debug redaction (spec 14) ✓
- Errcheck violations (spec 19) ✓
- Package doc comments ✓
- golangci-lint passing ✓

---

## Already Complete (No Action Needed)

| Spec | Feature | Status |
|------|---------|--------|
| 01 | Core CLI, config, env vars | Done |
| 02 | Bookmark CRUD (add, list, get, update, delete) | Done |
| 03 | Tags (list, show, rename, delete) | Done |
| 04 | Import/Export (JSON, HTML, CSV, backup, restore) | Done |
| 05 | Security hardening (0700/0600, token masking, safe JSON) | Done |
| 06 | Tags performance (client-side counting, full pagination) | Done |
| 07 | Test coverage (>70% threshold per package) | Done |
| 08 | Tags show pagination | Done |
| 09 | Makefile portability (awk, no cmd/ skip) | Done |
| 10 | Config token trim | Done |
| 11 | Test robustness | Done |
| 12 | User profile command | Done (superseded by spec 17) |
| 13 | Tags CRUD (create, get) | Done |
| 14 | Per-command `--url`/`--token` overrides | Done |
| 15 | Notes field (`--notes`/`-n` on add and update) | Done |
| 16 | Rename to `linkdingctl` | Done |
| 17 | Fix user profile to match real API | Done |
| 18 | CI/Release workflow | Done |
| 19 | Errcheck lint fixes | Done |
| 20 | Pre-commit: go vet via lefthook | Done |
| 21 | Pre-commit: golangci-lint via lefthook | Done |
| 22 | Staticcheck SA9003 fix | Done |

---

## Implementation Order for Bundles

1. **Bundle Models** (`internal/models/bundle.go`) — Create structs first (dependency for API methods)
2. **Bundle API Methods** (`internal/api/client.go`) — Implement CRUD (dependency for commands)
3. **Bundle API Tests** (`internal/api/client_test.go`) — Verify API layer works
4. **Bundle Commands** (`cmd/linkdingctl/bundles.go`) — Implement all subcommands
5. **Bundle CLI Tests** (`cmd/linkdingctl/commands_test.go`) — Verify full integration

---

## Summary

| Priority | Count | Focus |
|----------|-------|-------|
| P1 | 5 | Bundles feature (spec 23) |
| P2 | 1 | SA9003 staticcheck fix (spec 22) |
| P3 | 0 | All lint/polish items complete |
| **Total** | **6** | |

The codebase is nearly feature-complete. The primary remaining work is implementing bundle CRUD support (spec 23).

---

## Validation Checklist

Before marking bundles implementation complete:
- [ ] `make check` passes (fmt, vet, test)
- [ ] `make cover` passes (70% per package)
- [ ] `golangci-lint run` passes
- [ ] All bundle CRUD operations work against mock server
- [ ] JSON output is valid and matches API response format
- [ ] Table output is readable and consistent with other commands
- [ ] Error messages are user-friendly (404 → "Bundle not found", etc.)
