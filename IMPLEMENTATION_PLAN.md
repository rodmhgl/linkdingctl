# Implementation Plan

Gap analysis between specifications (`specs/01`–`specs/22`) and the current codebase.

---

## P0 — Foundational / Already Implemented

All P0 (foundational/blocking) items are complete. No new features need to be built from scratch.

---

## P1 — Staticcheck SA9003 (Spec 22)

- [x] **P1** | Fix SA9003: Remove error return from `processHTMLBookmark` | ~small
  - Acceptance: `processHTMLBookmark` has no `error` return type; all three call sites in `importHTML` invoke it as a bare statement (no `if err != nil {}` wrapper); `golangci-lint run ./...` reports no SA9003 in `internal/export/import.go`; all tests pass (`go test ./...`); coverage gate passes (`make cover`)
  - Files: `internal/export/import.go`

---

## P2 — Lint & Code Quality

### Uncommitted Changes (on `review` branch, need validation + commit)

- [x] **P2** | Validate and commit `//nolint:unused` on version var | ~small
  - Acceptance: `golangci-lint run` passes with no `unused` violation on `version` var
  - Files: `cmd/linkdingctl/main.go`

- [x] **P2** | Validate and commit removal of unused `updateUnread`/`updateShared` vars | ~small
  - Acceptance: No unused variable warnings; `go build ./...` and all tests pass
  - Files: `cmd/linkdingctl/update.go`

- [x] **P2** | Validate and commit error message lowercase fix in `GetUserProfile` | ~small
  - Acceptance: Error message starts lowercase per Go conventions; tests pass
  - Files: `internal/api/client.go`

- [x] **P2** | Validate and commit package-level doc comments | ~small
  - Acceptance: `golangci-lint run` reports no missing package comments for `api`, `config`, `export`, `models`
  - Files: `internal/api/client.go`, `internal/config/config.go`, `internal/export/import.go`, `internal/models/bookmark.go`

### Config Token Trim (Spec 10)

- [x] **P2** | Remove redundant `strings.TrimSpace(token)` in config init | ~small
  - Acceptance: Token is trimmed exactly once per branch (TTY and non-TTY); no redundant trim calls; existing tests pass
  - Files: `cmd/linkdingctl/config.go`

### Test Robustness (Spec 11)

- [x] **P2** | Replace unsafe string slicing with `strings.Contains` in config test | ~small
  - Acceptance: `TestLoad_NonYAMLFile` uses `strings.Contains` (no panic risk from index out-of-range); test passes
  - Files: `internal/config/config_test.go`

- [x] **P2** | Add missing flag resets to `executeCommand` test helper | ~small
  - Acceptance: `executeCommand` resets `backupOutput`, `backupPrefix`, `tagsRenameForce`, `tagsDeleteForce`; no test pollution
  - Files: `cmd/linkdingctl/commands_test.go`

### Errcheck Violations (Spec 19 — remaining items)

- [x] **P2** | Fix unchecked `w.Write` in `TestCreateTag_Duplicate` | ~small
  - Acceptance: `w.Write` error is checked with `t.Errorf`
  - Files: `internal/api/client_test.go`

- [x] **P2** | Fix unchecked `json.Decode` in import test mock handlers | ~small
  - Acceptance: All three `json.NewDecoder().Decode()` calls in mock handlers check errors
  - Files: `internal/export/import_test.go`

- [x] **P2** | Fix unchecked `json.Encode` in html_test and csv_test mock handlers | ~small
  - Acceptance: `json.NewEncoder().Encode()` calls check errors with `t.Errorf`
  - Files: `internal/export/html_test.go`, `internal/export/csv_test.go`

- [x] **P2** | Fix unchecked `csvWriter.Write` in csv_test helper | ~small
  - Acceptance: Both `csvWriter.Write` calls check errors with `t.Fatalf`
  - Files: `internal/export/csv_test.go`

### golangci-lint Config

- [x] **P2** | Verify `.golangci.yml` passes cleanly after all fixes | ~small
  - Acceptance: `golangci-lint run ./...` exits 0 with no violations
  - Files: `.golangci.yml`

---

## P3 — Polish / Optional

### Debug & Config UX (Spec 14 — partial gaps)

- [x] **P3** | Ensure `--token` value is never printed in debug output | ~small
  - Acceptance: With `--debug`, token is redacted; URL is shown normally
  - Files: `cmd/linkdingctl/root.go`

- [x] **P3** | `config show` indicates active CLI flag overrides | ~small
  - Acceptance: When `--url` or `--token` is provided, `config show` output marks them as overrides
  - Files: `cmd/linkdingctl/config.go`

### Additional Test Coverage (Spec 07 — gaps)

- [x] **P3** | Add multi-page pagination test for `GetBookmarks` | ~medium
  - Acceptance: Mock server returns paginated responses; test verifies all pages are collected
  - Files: `internal/api/client_test.go` or `internal/api/pagination_test.go`

- [x] **P3** | Add multi-page pagination test for `FetchAllTags` | ~small
  - Acceptance: Mock server returns paginated tag responses; test verifies all pages collected
  - Files: `internal/api/client_test.go` or `internal/api/pagination_test.go`

- [x] **P3** | Add `doRequest` timeout behavior test | ~small
  - Acceptance: Test verifies error is returned when server exceeds 30s timeout
  - Files: `internal/api/client_test.go`

- [ ] **P3** | Add JSON export→import round-trip test | ~small
  - Acceptance: Exported JSON can be re-imported with identical bookmark data
  - Files: `internal/export/json_test.go` or `internal/export/import_test.go`

- [ ] **P3** | Add CSV export→import round-trip test | ~small
  - Acceptance: Exported CSV can be re-imported with identical bookmark data
  - Files: `internal/export/csv_test.go` or `internal/export/import_test.go`

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
| 07 | Test coverage (>70% threshold per package) | Done (partial gaps in P3) |
| 08 | Tags show pagination | Done |
| 09 | Makefile portability (awk, no cmd/ skip) | Done |
| 10 | Config token trim | Done (see P2 for verification) |
| 11 | Test robustness | Done (see P2 for remaining items) |
| 12 | User profile command | Done (superseded by spec 17) |
| 13 | Tags CRUD (create, get) | Done |
| 14 | Per-command `--url`/`--token` overrides | Done (partial gaps in P3) |
| 15 | Notes field (`--notes`/`-n` on add and update) | Done |
| 16 | Rename to `linkdingctl` | Done |
| 17 | Fix user profile to match real API | Done |
| 18 | CI/Release workflow | Done |
| 19 | Errcheck lint fixes | Done (remaining items in P2) |
| 20 | Pre-commit: go vet via lefthook | Done |
| 21 | Pre-commit: golangci-lint via lefthook | Done |

---

## Summary

| Priority | Count | Focus |
|----------|-------|-------|
| P1 | 1 | SA9003 staticcheck fix |
| P2 | 12 | Lint fixes, test robustness, errcheck |
| P3 | 7 | Debug redaction, config UX, test coverage |
| **Total** | **20** | |

The codebase is feature-complete. Remaining work is lint compliance, test robustness, and minor UX polish.
