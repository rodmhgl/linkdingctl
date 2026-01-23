# Implementation Plan - LinkDing CLI

> Gap analysis based on specs 01-11 vs current codebase (2026-01-23).
> Specs 01-08 are fully implemented. Specs 09-11 introduce new defects and cleanup tasks.

## Current State Summary

| Spec | Status | Notes |
|------|--------|-------|
| 01 - Core CLI | Complete | Config init/show/test, env overrides, --json, --debug all work |
| 02 - Bookmark CRUD | Complete | add/list/get/update/delete all implemented with --json |
| 03 - Tags | Complete | tags/tags show/rename/delete all implemented |
| 04 - Import/Export | Complete | JSON/HTML/CSV export+import, backup, restore with --wipe |
| 05 - Security | Complete | 0700/0600 perms, token masking, safe JSON backup output |
| 06 - Tags Performance | Complete | FetchAllBookmarks on Client, client-side counting, paginated rename/delete |
| 07 - Test Coverage | **Partial** | cmd/ld=55.5% (target 70%), other packages meet threshold |
| 08 - Tags Show Pagination | Complete | Uses FetchAllBookmarks for all bookmarks |
| 09 - Makefile Portability | **Open** | `bc` dependency remains in `cover` target |
| 10 - Config Token Trim | **Open** | Redundant `strings.TrimSpace(token)` on line 56 of config.go |
| 11 - Test Robustness | **Open** | Unsafe string slicing in config_test.go; missing flag resets in commands_test.go |

## Coverage Status

```
Package              Current   Target   Status
cmd/ld               55.5%     70%      BELOW THRESHOLD (-14.5%)
internal/api         80.1%     70%      PASS
internal/config      70.3%     70%      PASS
internal/export      78.4%     70%      PASS
```

---

## Remaining Tasks

### Phase 1: Makefile Portability (spec 09)

- [x] **P1** | Replace `bc` with `awk` in Makefile `cover` target | ~small
  - Acceptance: `make cover` does not require `bc`; uses `awk` for float comparison; coverage validation behavior unchanged for all packages
  - Files: `Makefile` (line 66)
  - Details: Replace `result=$$(echo "$$cov < 70" | bc -l)` with `result=$$(echo "$$cov 70" | awk '{print ($$1 < $$2)}')` — `awk` is universally available, `bc` is not (Alpine, minimal Docker, some CI runners)

### Phase 2: Config Token Trim Cleanup (spec 10)

- [ ] **P1** | Remove redundant `strings.TrimSpace(token)` | ~small
  - Acceptance: Token trimmed exactly once per code path; no redundant `TrimSpace` call; TTY and non-TTY input still work; existing tests pass
  - Files: `cmd/ld/config.go` (line 56)
  - Details: The final `token = strings.TrimSpace(token)` on line 56 is redundant — the TTY branch (`term.ReadPassword`) never includes a trailing newline, and the non-TTY branch already calls `TrimSpace` on line 54. Remove line 56 and add `strings.TrimSpace()` around `string(tokenBytes)` on line 46 for defensive clarity in the TTY branch.

### Phase 3: Test Robustness (spec 11)

- [ ] **P1** | Fix unsafe string slicing in config test | ~small
  - Acceptance: `TestLoad_NonYAMLFile` uses `strings.Contains` or `strings.HasPrefix` (no direct slice); cannot panic on short error messages; all existing tests pass
  - Files: `internal/config/config_test.go` (line 251)
  - Details: Replace `err.Error()[:len(expectedErrSubstring)] != expectedErrSubstring` with `!strings.Contains(err.Error(), expectedErrSubstring)` — the current code panics with index-out-of-range if the error is shorter than the expected substring

- [ ] **P1** | Add missing flag resets to `executeCommand` helper | ~small
  - Acceptance: `executeCommand` resets `backupOutput`, `backupPrefix`, `tagsRenameForce`, `tagsDeleteForce`; no test pollution between runs; all existing tests pass
  - Files: `cmd/ld/commands_test.go` (after line 83)
  - Details: Add the following resets to the `executeCommand` function after the existing flag resets:
    ```go
    backupOutput = "."
    backupPrefix = "linkding-backup"
    tagsRenameForce = false
    tagsDeleteForce = false
    ```
    Note: `backupOutput` default is `"."` and `backupPrefix` default is `"linkding-backup"` (from flag definitions in `backup.go` init())

### Phase 4: cmd/ld Test Coverage to 70% (spec 07)

- [ ] **P2** | Increase `cmd/ld` coverage from 55.5% to 70%+ | ~large
  - Acceptance: `go test -cover ./cmd/ld/` reports >= 70%; `make cover` passes the 70% gate for all packages
  - Files: `cmd/ld/commands_test.go`
  - Details: The remaining 14.5% gap is in complex interactive/multi-step commands. Areas needing coverage:
    - `import` command: test actual import from each format file (JSON, HTML, CSV) with mock server handling create/update; test `--dry-run`, `--skip-duplicates`, `--add-tags`, error paths
    - `restore` command: test `--wipe` path with piped "yes\n" confirmation; test `--wipe` with bookmarks present; test `--dry-run` + `--wipe` combination
    - `tags rename` command: test with multiple bookmarks (progress output); test error during update (partial failure); test abort on "n" confirmation
    - `tags delete` command: test `--force` with bookmarks needing removal; test force with update errors; test confirmation abort
    - `backup` command: test `--prefix` flag changes filename; test error on invalid output directory
    - Error handling paths: test API error responses (401, 404, 500) for each command

---

## Dependency Graph

```
Phase 1 (Makefile):
  bc→awk replacement — standalone, no deps

Phase 2 (Config cleanup):
  Token trim fix — standalone, no deps

Phase 3 (Test robustness):
  Config test fix — standalone
  Flag reset fix — standalone

Phase 4 (Coverage):
  cmd/ld 70%+ — depends on Phase 3 (flag reset fix prevents test pollution)
                  depends on Phase 2 (token trim change may alter test behavior)
```

## Execution Order

1. **Phase 1** (P1): Makefile `bc` → `awk` (standalone, immediate portability win)
2. **Phase 2** (P1): Config token trim cleanup (standalone, simple code clarity fix)
3. **Phase 3** (P1): Test robustness fixes (blocks Phase 4 — must fix flag resets before adding more tests)
4. **Phase 4** (P2): cmd/ld coverage to 70% (depends on Phases 2+3 being complete)

---

## Notes

- Specs 01-08 are fully implemented — no code changes needed for those
- Specs 09-11 are all new items not previously tracked in the implementation plan
- The `cmd/` skip condition referenced in spec 09 does NOT exist in the current Makefile (it was already removed in a prior change). Only the `bc` dependency remains.
- The `backupOutput` and `backupPrefix` flags have non-zero defaults (`"."` and `"linkding-backup"`) — reset must use those defaults, not empty strings
- Phase 4 is the most labor-intensive task (~300+ lines of test code) but Phase 3's flag reset fix is a prerequisite to prevent flaky tests
- `golang.org/x/term` is already in go.mod — no dependency changes needed
