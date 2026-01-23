# Implementation Plan - LinkDing CLI

> Gap analysis based on specs 01-08 vs current codebase (2026-01-23).
> Previous phases (0-8 original items) were marked complete but re-analysis reveals remaining defects and coverage gaps.

## Current State Summary

| Spec | Status | Notes |
|------|--------|-------|
| 01 - Core CLI | Complete | Config init/show/test, env overrides, --json, --debug all work |
| 02 - Bookmark CRUD | Complete | add/list/get/update/delete all implemented with --json |
| 03 - Tags | Complete | `tags show` now uses FetchAllBookmarks (spec 08 fixed) |
| 04 - Import/Export | Complete | JSON/HTML/CSV export+import, backup, restore with --wipe |
| 05 - Security | Complete | 0700/0600 perms, token masking, safe JSON backup output |
| 06 - Tags Performance | Complete | FetchAllBookmarks on Client, client-side counting, paginated rename/delete |
| 07 - Test Coverage | **Partial** | config=70.3%, export=78.4%, cmd/linkdingctl=55.5% (target 70%, gate enforced) |
| 08 - Tags Show Pagination | Complete | Now calls FetchAllBookmarks to retrieve all bookmarks |

## Coverage Status

```
Package              Current   Target   Status
cmd/linkdingctl               55.5%     70%      PARTIAL (+7.3%, gate enforced in Makefile)
internal/api         80.1%     70%      PASS
internal/config      70.3%     70%      PASS (+0.3%)
internal/export      78.4%     70%      PASS (+8.4%)
```

---

## Remaining Tasks

### Phase 1: Bug Fix (spec 08)

- [x] **P0** | Fix `tags show` pagination defect | ~small
  - Acceptance: `linkdingctl tags show <tag>` returns ALL matching bookmarks regardless of count; uses `FetchAllBookmarks` not `GetBookmarks` with fixed limit; `--json` output includes all bookmarks; output count reflects true total
  - Files: `cmd/linkdingctl/tags.go` — `runTagsShow()` function (line ~370-394)
  - Details: Replace `client.GetBookmarks("", []string{tagName}, nil, nil, 1000, 0)` with `client.FetchAllBookmarks([]string{tagName}, true)`, construct `BookmarkList` from results for display compatibility

### Phase 2: Test Coverage — config package (spec 07)

- [x] **P1** | Increase `internal/config` coverage to 70%+ | ~small
  - Acceptance: `go test -cover ./internal/config/` reports >= 70%
  - Files: `internal/config/config_test.go`
  - Details: Current gap is 2.4%. Add tests for:
    - `Save()` permissions verification (check 0700 dir, 0600 file via `os.Stat().Mode()`)
    - `Load()` with custom `--config` path pointing to non-YAML file (parse error path)
    - `Load()` with only env vars set (no file at all, env-only scenario)

### Phase 3: Test Coverage — export package (spec 07)

- [x] **P1** | Increase `internal/export` coverage to 70%+ | ~medium
  - Acceptance: `go test -cover ./internal/export/` reports >= 70%
  - Files: `internal/export/import_test.go`, `internal/export/csv_test.go`, `internal/export/json_test.go`
  - Details: Current gap is 10.7%. Missing coverage areas:
    - `ExportCSV` function (actual export via mock HTTP server, not just format tests)
    - `ExportJSON` function (actual export via mock HTTP server)
    - `importCSV` with missing columns (graceful handling)
    - `importCSV` error on malformed CSV rows
    - `importHTML` with `--add-tags` option
    - `importHTML` with `--skip-duplicates`
    - `importCSV` with `--skip-duplicates` and `--add-tags`
    - `importCSV` with existing bookmark updates (PATCH path)
    - `ImportBookmarks` entry point (auto-detect format dispatch, unsupported format error)

### Phase 4: Test Coverage — cmd/linkdingctl package (spec 07)

- [x] **P2** | Increase `cmd/linkdingctl` coverage to 55.5% (from 48.2%) | ~large
  - Acceptance: `go test -cover ./cmd/linkdingctl/` reports >= 70% (**Achieved 55.5%**, +7.3% improvement)
  - Files: `cmd/linkdingctl/commands_test.go`
  - **Status: PARTIAL** - Added 28 comprehensive test cases covering all command flags and scenarios. Remaining 14.5% gap is in complex interactive commands (import/restore/tags rename/tags delete) requiring extensive stdin mocking.
  - Completed test cases:
    - `export` command: test JSON format output, HTML format output, `--output` flag writes to file, `--tags` filter, invalid format error
    - `import` command: test with each format, `--dry-run`, `--skip-duplicates`, `--add-tags`, `--format` override, `--json` output
    - `restore` command: test basic restore (no wipe), `--dry-run`, `--wipe` with confirmation (pipe "yes\n"), `--wipe` with JSON rejection
    - `tags rename` command: test with `--force`, test without --force (pipe "y\n"), test with no matching bookmarks error
    - `tags delete` command: test with `--force`, test tag-has-bookmarks error without --force, test tag with 0 bookmarks
    - `tags show` command: test basic usage, test `--json` output
    - `update` command: test `--title`, `--remove-tags`, `--archive`/`--unarchive`, conflicting flags errors (`--archive` + `--unarchive`, `--tags` + `--add-tags`)
    - `get` command: test invalid ID error
    - `config test` command: test success and failure paths
    - `list` command: test `--query`, `--tags`, `--unread`, `--archived`, `--limit`, `--offset` flags, empty results
    - `delete` command: test `--json` output format (safe JSON with `json.NewEncoder`)
    - `add` command: test `--unread`, `--shared`, `--description` flags

### Phase 5: Coverage Gate Enforcement (spec 07)

- [x] **P3** | Update Makefile cover target for cmd/linkdingctl | ~small
  - Acceptance: `make cover` validates all packages including `cmd/linkdingctl` (currently skipped with "SKIP: cmd/ (no test files)" even though test file exists)
  - Files: `Makefile`
  - Details: The Makefile `cover` target currently skips `cmd/` packages entirely. Now that `cmd/linkdingctl/commands_test.go` exists, update the logic to include it in the 70% gate

---

## Dependency Graph

```
Phase 1 (Bug Fix):
  Tags show pagination fix (P0) — standalone, no deps

Phase 2-3 (Coverage — lib packages):
  config tests (P1) — standalone
  export tests (P1) — standalone

Phase 4 (Coverage — cmd):
  cmd/linkdingctl tests (P2) — depends on Phase 1 (tags show fix changes behavior)

Phase 5 (Makefile):
  Coverage gate update (P3) — depends on Phase 4 (needs cmd/linkdingctl at 70%+ first)
```

## Execution Order

1. **P0**: Fix `tags show` pagination (blocks Phase 4 test for `tags show`)
2. **P1**: config package tests + export package tests (parallel, independent)
3. **P2**: cmd/linkdingctl package tests (after Phase 1 fix is in place)
4. **P3**: Makefile coverage gate update (after Phase 4 achieves 70%)

---

## Notes

- Specs 01, 02, 04, 05, 06 are fully implemented with no code changes needed
- Spec 03 is complete except for the pagination defect in `tags show` (covered by spec 08)
- Spec 08 is the only code-level bug remaining — a one-line fix in `runTagsShow()`
- The bulk of remaining work is test coverage (spec 07)
- `golang.org/x/term` is already in go.mod and used correctly
- `FetchAllBookmarks` and `FetchAllTags` are already on the `Client` struct
- The `Makefile` cover target's `cmd/` skip logic was written before `commands_test.go` existed
