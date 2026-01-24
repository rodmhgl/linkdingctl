# Implementation Plan

> Gap analysis: specs 01–19 vs current codebase (2026-01-24).
> Specs 01–18 are **fully implemented and passing**. Only Spec 19 (Errcheck Lint Fixes) remains.

## Status Summary

| Spec | Status | Notes |
|------|--------|-------|
| 01 – Core CLI | ✅ Complete | Config init/show/test, env overrides, --json, --debug |
| 02 – Bookmark CRUD | ✅ Complete | add/list/get/update/delete with --json support |
| 03 – Tags | ✅ Complete | tags list/show/rename/delete |
| 04 – Import/Export | ✅ Complete | JSON/HTML/CSV export+import, backup, restore |
| 05 – Security | ✅ Complete | 0700/0600 perms, token masking, safe JSON output |
| 06 – Tags Performance | ✅ Complete | FetchAllBookmarks/FetchAllTags, client-side counting |
| 07 – Test Coverage | ✅ Complete | All packages >70% threshold |
| 08 – Tags Show Pagination | ✅ Complete | Uses FetchAllBookmarks |
| 09 – Makefile Portability | ✅ Complete | awk-based coverage, no cmd/ skip |
| 10 – Config Token Trim | ✅ Complete | Clean per-branch trim |
| 11 – Test Robustness | ✅ Complete | strings.Contains, full flag reset |
| 12 – User Profile | ✅ Complete | Superseded by spec 17 |
| 13 – Tags CRUD | ✅ Complete | tags create/get subcommands |
| 14 – Per-Command Overrides | ✅ Complete | --url/--token global flags |
| 15 – Notes Field | ✅ Complete | --notes on add/update commands |
| 16 – Rename | ✅ Complete | All references updated to linkdingctl |
| 17 – Fix User Profile | ✅ Complete | Correct API model, all fields rendered |
| 18 – CI/Release Workflow | ✅ Complete | release.yaml with lint-test, release, build-binaries |
| 19 – Errcheck Lint Fixes | ❌ Not done | Unchecked error returns in prod + test code |

---

## Remaining Tasks (Spec 19)

### Production Code

- [x] **P0** | Fix unchecked `v.BindEnv` calls in config loader | ~small
  - Acceptance: `internal/config/config.go` wraps both `v.BindEnv("url")` and `v.BindEnv("token")` with error checks; returns `(nil, error)` if either fails; existing config tests still pass
  - Files: `internal/config/config.go` (lines 89–90)

### Test Code — API Client

- [x] **P1** | Fix unchecked `w.Write` in `TestCreateTag_Duplicate` | ~small
  - Acceptance: Mock handler checks `w.Write` return value; calls `t.Errorf` and returns on failure
  - Files: `internal/api/client_test.go` (line 351)

### Test Code — Import Tests

- [x] **P1** | Fix unchecked `json.NewDecoder().Decode()` in import test handlers | ~medium
  - Acceptance: All `json.NewDecoder(r.Body).Decode(...)` calls in mock handlers check error, call `t.Errorf`, respond with HTTP 400, and return on failure
  - Files: `internal/export/import_test.go` (lines 91, 259, 308, 381, 530, 622, 724)

### Test Code — Export Tests (Encode)

- [x] **P1** | Fix unchecked `json.NewEncoder().Encode()` in HTML test handlers | ~medium
  - Acceptance: All `json.NewEncoder(w).Encode(...)` calls in mock handlers check error and call `t.Errorf` on failure
  - Files: `internal/export/html_test.go` (lines 38, 106, 191, 259, 329, 389, 427)

- [x] **P1** | Fix unchecked `json.NewEncoder().Encode()` in CSV test handlers | ~small
  - Acceptance: Both `json.NewEncoder(w).Encode(...)` calls in mock handlers check error and call `t.Errorf` on failure
  - Files: `internal/export/csv_test.go` (lines 236, 316)

- [x] **P1** | Fix unchecked `json.NewEncoder().Encode()` in JSON test handlers | ~small
  - Acceptance: Both `json.NewEncoder(w).Encode(...)` calls in mock handlers check error and call `t.Errorf` on failure
  - Files: `internal/export/json_test.go` (lines 170, 255)

- [x] **P1** | Fix unchecked `json.NewEncoder().Encode()` in import test non-handler code | ~medium
  - Acceptance: All `json.NewEncoder(...).Encode(...)` calls outside of the decode handlers (buffer writes, mock responses) check error and call `t.Fatalf`/`t.Errorf` on failure
  - Files: `internal/export/import_test.go` (numerous lines — buffer encodes + handler encodes)

### Test Code — CSV Writer

- [x] **P1** | Fix unchecked `csvWriter.Write` in CSV test helper | ~small
  - Acceptance: `csvWriter.Write(header)` and `csvWriter.Write(row)` (lines 165, 179) check error; test calls `t.Fatalf` on failure
  - Files: `internal/export/csv_test.go`

### Validation

- [ ] **P2** | Run `golangci-lint` and verify zero errcheck violations | ~small
  - Acceptance: `golangci-lint run ./...` passes with no `errcheck` findings
  - Files: (validation only)

- [ ] **P2** | Verify all tests pass and coverage holds | ~small
  - Acceptance: `go test ./...` passes; `make cover` meets 70% per-package threshold
  - Files: (validation only)

---

## Dependency Graph

```
P0 (production code fix) ─── should be first (actual runtime risk)
        │
        ▼
P1 (test code fixes) ──────── independent of each other, can be done in any order
        │
        ▼
P2 (validation) ────────────── run after all fixes applied
```

## Implementation Notes

- The `v.BindEnv` fix is the only production code change. It's unlikely to fail in practice (BindEnv only errors when called with 0 args), but checking it satisfies errcheck and establishes good precedent.
- All test fixes follow the same pattern: check the error return, call `t.Errorf`/`t.Fatalf`, and short-circuit the handler. This prevents tests from silently operating on zero-value data.
- The scope of unchecked `json.NewEncoder().Encode()` in import_test.go is larger than spec 19 originally listed — there are ~40+ occurrences across buffer writes and mock handlers.
- No new test files need to be created. All changes are edits to existing files.
- The fixes should not change test behavior when everything works correctly — they only add safety nets for failure scenarios.
