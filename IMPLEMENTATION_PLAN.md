# Implementation Plan

> Gap analysis: specs 01–21 vs current codebase (2026-01-24).
> Specs 01–19 are **fully implemented and passing**. Specs 20–21 (pre-commit hooks) remain.

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
| 19 – Errcheck Lint Fixes | ✅ Complete | All errcheck violations resolved |
| 20 – Pre-Commit go vet | ❌ Not done | Lefthook + go vet hook |
| 21 – Pre-Commit golangci-lint | ❌ Not done | Lefthook + golangci-lint hook |

---

## Remaining Tasks

### Spec 20: Pre-Commit Hook — go vet

- [x] **P0** | Create `lefthook.yml` with go vet and beads hook | ~small
  - Acceptance: `lefthook.yml` exists at repo root; defines `pre-commit` section with `beads-sync` command (skip if `bd` not installed) and `go-vet` command running `go vet ./...`; file is valid YAML
  - Files: `lefthook.yml` (new)

- [ ] **P1** | Add `hooks` Makefile target | ~small
  - Acceptance: `make hooks` checks for lefthook availability, prints install instructions if missing, runs `lefthook install` if found; target is listed in `make help` output
  - Files: `Makefile`

### Spec 21: Pre-Commit Hook — golangci-lint

- [ ] **P1** | Add golangci-lint command to `lefthook.yml` | ~small
  - Acceptance: `lefthook.yml` includes `golangci-lint` command in `pre-commit` section; command warns and exits 0 if `golangci-lint` not installed; runs `golangci-lint run` if installed; all three hooks run in parallel
  - Files: `lefthook.yml`

### Validation

- [ ] **P2** | Verify hooks work end-to-end | ~small
  - Acceptance: After `lefthook install`, committing code with a `go vet` violation is blocked; committing code with a lint violation is blocked; committing clean code succeeds; `bd hook pre-commit` is still invoked when `bd` is available
  - Files: (validation only)

---

## Dependency Graph

```
P0 (lefthook.yml with go vet + beads) ─── creates the hook infrastructure
        │
        ├──▶ P1 (Makefile hooks target) ────── developer convenience for installing hooks
        │
        └──▶ P1 (golangci-lint hook) ──────── extends lefthook.yml with lint check
                │
                ▼
        P2 (end-to-end validation) ─────────── verify all hooks work together
```

## Implementation Notes

- Lefthook is a Go-based git hook manager (single binary, no runtime deps). It replaces `.git/hooks/pre-commit` with its own dispatcher that runs configured commands.
- The existing `.git/hooks/pre-commit` delegates to `bd hook pre-commit`. Lefthook will replace it, so `bd` must be chained as a command in `lefthook.yml` with a skip condition for environments where `bd` is not installed.
- Both `go vet` and `golangci-lint` run against the full module (`./...` scope) to match CI behavior in `.github/workflows/release.yaml`.
- `golangci-lint` uses a soft-fail pattern: if the tool is not installed, the hook prints a warning and exits 0 (non-blocking). This prevents blocking commits for developers who haven't installed it yet.
- All pre-commit commands run in parallel by default (Lefthook's default behavior), which reduces wall-clock time since both checks are read-only.
- No changes to existing source code are required — these specs only add tooling configuration.
