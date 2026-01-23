# Implementation Plan - LinkDing CLI

> Gap analysis based on specs 01-16 vs current codebase (2026-01-23).
> Specs 01-11 are fully implemented. Specs 12-16 introduce new features and a major rename.

## Current State Summary

| Spec | Status | Notes |
|------|--------|-------|
| 01 - Core CLI | Complete | Config init/show/test, env overrides, --json, --debug all work |
| 02 - Bookmark CRUD | Complete | add/list/get/update/delete all implemented with --json |
| 03 - Tags | Complete | tags/tags show/rename/delete all implemented |
| 04 - Import/Export | Complete | JSON/HTML/CSV export+import, backup, restore with --wipe |
| 05 - Security | Complete | 0700/0600 perms, token masking, safe JSON backup output |
| 06 - Tags Performance | Complete | FetchAllBookmarks on Client, client-side counting, paginated rename/delete |
| 07 - Test Coverage | Complete | cmd/ld=78.1%, all packages meet 70% threshold |
| 08 - Tags Show Pagination | Complete | Uses FetchAllBookmarks for all bookmarks |
| 09 - Makefile Portability | Complete | `bc` replaced with `awk` in cover target |
| 10 - Config Token Trim | Complete | Redundant TrimSpace removed |
| 11 - Test Robustness | Complete | Unsafe slicing fixed, missing flag resets added |
| 12 - User Profile | **Open** | New `user profile` command needed |
| 13 - Tags CRUD Enhancements | **Open** | `tags create` and `tags get` subcommands needed |
| 14 - Per-Command Overrides | **Open** | Global `--url` and `--token` flags needed |
| 15 - Notes Field | **Open** | `--notes` flag on add/update commands needed |
| 16 - Rename to linkdingctl | **Open** | Binary, package, config path, docs rename |

## Coverage Status

```
Package              Current   Target   Status
cmd/ld               78.1%     70%      PASS
internal/api         80.1%     70%      PASS
internal/config      70.3%     70%      PASS
internal/export      78.4%     70%      PASS
```

---

## Remaining Tasks

### Phase 5: Notes Field (spec 15)

- [x] **P1** | Add `--notes` flag to `add` command | ~small
  - Acceptance: `linkdingctl add --notes "..."` sets the notes field on creation; `-n` shorthand works; notes field included in API request body
  - Files: `cmd/linkdingctl/add.go`
  - Details: Add `addNotes string` var; register `-n`/`--notes` flag; map to `BookmarkCreate.Notes` field; update `--description` help text from "Description/notes" to "Description"

- [x] **P1** | Add `--notes` flag to `update` command | ~small
  - Acceptance: `linkdingctl update <id> --notes "..."` updates notes; `linkdingctl update <id> --notes ""` clears notes; notes and description settable independently and together
  - Files: `cmd/linkdingctl/update.go`
  - Details: Add `updateNotes string` var; register `-n`/`--notes` flag; use `cmd.Flags().Changed("notes")` to detect explicit set; map to `BookmarkUpdate.Notes` pointer field

- [x] **P2** | Add tests for `--notes` flag | ~small
  - Acceptance: Tests verify notes set on add, notes set on update, notes cleared on update; coverage maintained ≥70%
  - Files: `cmd/ld/commands_test.go`

### Phase 6: Tags CRUD Enhancements (spec 13)

- [x] **P1** | Add `CreateTag` API method | ~small
  - Acceptance: `CreateTag(name)` sends POST to `/api/tags/` with `{"name": "..."}` body; returns created Tag model; handles 400 (duplicate) with clear error
  - Files: `internal/api/client.go`

- [x] **P1** | Add `GetTag` API method | ~small
  - Acceptance: `GetTag(id)` sends GET to `/api/tags/{id}/`; returns Tag model; handles 404 with "Tag with ID X not found"
  - Files: `internal/api/client.go`

- [x] **P1** | Add `tags create` subcommand | ~small
  - Acceptance: `linkdingctl tags create <name>` creates tag and prints ID; duplicate name shows clear error; empty name shows validation error; respects `--json`
  - Files: `cmd/linkdingctl/tags.go`
  - Details: Add `tagsCreateCmd` cobra command; validate non-empty arg; call `client.CreateTag(name)`; output human/JSON based on flag

- [x] **P1** | Add `tags get` subcommand | ~small
  - Acceptance: `linkdingctl tags get <id>` displays tag ID, Name, DateAdded; non-existent ID shows "not found"; respects `--json`
  - Files: `cmd/linkdingctl/tags.go`
  - Details: Add `tagsGetCmd` cobra command; parse int arg; call `client.GetTag(id)`; output human/JSON based on flag

- [x] **P2** | Add tests for tags create and get | ~medium
  - Acceptance: Tests for create success, create duplicate, create empty name, get success, get not found; all with human and JSON output; coverage maintained ≥70%
  - Files: `cmd/ld/commands_test.go`, `internal/api/client_test.go`

### Phase 7: User Profile Command (spec 12)

- [x] **P1** | Add `UserProfile` model | ~small
  - Acceptance: `UserProfile` struct with Theme, BookmarkCount, DisplayName, Username fields; proper JSON tags matching API response
  - Files: `internal/models/bookmark.go` (or new `internal/models/user.go`)

- [ ] **P1** | Add `GetUserProfile` API method | ~small
  - Acceptance: `GetUserProfile()` sends GET to `/api/user/profile/`; returns `*UserProfile`; handles 401/403 errors with clear messages
  - Files: `internal/api/client.go`

- [ ] **P1** | Add `user profile` command | ~medium
  - Acceptance: `ld user profile` displays Username, DisplayName, Theme, BookmarkCount; `--json` outputs full JSON; HTTP 401 shows auth error; HTTP 403 shows permission error; `ld user --help` lists subcommands
  - Files: `cmd/ld/user.go` (new)
  - Details: Create `userCmd` (parent, help-only), `userProfileCmd` (subcommand); register on rootCmd; load config, create client, call `GetUserProfile()`; format human/JSON output

- [ ] **P2** | Add tests for user profile command | ~small
  - Acceptance: Tests for successful profile display (human + JSON), 401 error, 403 error; coverage maintained ≥70%
  - Files: `cmd/ld/commands_test.go`, `internal/api/client_test.go`

### Phase 8: Per-Command Connection Overrides (spec 14)

- [ ] **P1** | Add `--url` and `--token` global flags | ~medium
  - Acceptance: `--url` and `--token` persistent flags on rootCmd; override config and env vars (highest precedence); partial override works (only `--url` or only `--token`); commands work with only CLI flags and no config file; `--token` never printed in debug output
  - Files: `cmd/ld/root.go`, `internal/config/config.go`
  - Details: Add `flagURL string` and `flagToken string` vars; register as PersistentFlags on rootCmd; modify `loadConfig()` to accept and apply overrides after `config.Load()`; if `config.Load()` fails but CLI flags provide both URL+token, succeed without config file

- [ ] **P1** | Update `config show` to reflect overrides | ~small
  - Acceptance: `ld config show --url X` shows "URL: X (--url flag)"; `ld config show --token Y` shows override indicator; JSON output includes override source
  - Files: `cmd/ld/config.go`

- [ ] **P1** | Update `config test` to use effective config | ~small
  - Acceptance: `ld config test --url X --token Y` tests the overridden connection, not the config file
  - Files: `cmd/ld/config.go`

- [ ] **P2** | Add tests for per-command overrides | ~medium
  - Acceptance: Tests for URL override, token override, both overrides, partial override, no-config-file-with-flags, token not in debug output; coverage maintained ≥70%
  - Files: `cmd/ld/commands_test.go`

### Phase 9: Rename to linkdingctl (spec 16)

- [x] **P0** | Rename `cmd/ld/` directory to `cmd/linkdingctl/` | ~small
  - Acceptance: All Go source files moved to `cmd/linkdingctl/`; package still compiles; `go build ./cmd/linkdingctl` succeeds
  - Files: `cmd/ld/*.go` → `cmd/linkdingctl/*.go`

- [x] **P0** | Update Makefile for new binary name | ~small
  - Acceptance: `BINARY_NAME=linkdingctl`; build path `./cmd/linkdingctl`; `make build` produces `./linkdingctl`; all targets work
  - Files: `Makefile`

- [x] **P0** | Update Cobra root command | ~small
  - Acceptance: `rootCmd.Use = "linkdingctl"`; Long description references `linkdingctl`; error messages say `"Run 'linkdingctl config init'"`
  - Files: `cmd/linkdingctl/root.go`

- [x] **P0** | Update config paths to `~/.config/linkdingctl/` | ~small
  - Acceptance: `DefaultConfigPath()` returns `~/.config/linkdingctl/config.yaml`; `Load()` uses `linkdingctl` directory; error messages reference `linkdingctl`
  - Files: `internal/config/config.go`

- [x] **P0** | Implement config migration from old path | ~medium
  - Acceptance: On startup, if `~/.config/linkdingctl/config.yaml` missing but `~/.config/ld/config.yaml` exists: copy with 0600 perms, create dir with 0700 perms, print notice to stderr; does NOT delete old config; runs only once (subsequent runs skip migration)
  - Files: `internal/config/config.go`

- [x] **P0** | Update `.gitignore` | ~small
  - Acceptance: `/linkdingctl` entry replaces or accompanies `/ld`
  - Files: `.gitignore`

- [x] **P1** | Update all test expected strings | ~medium
  - Acceptance: `config_test.go` references `linkdingctl`; `commands_test.go` uses correct binary name; all tests pass with `go test ./...`
  - Files: `internal/config/config_test.go`, `cmd/linkdingctl/commands_test.go`

- [x] **P1** | Add config migration tests | ~small
  - Acceptance: Test migration from old path; test skip-if-already-migrated; test old config not deleted
  - Files: `internal/config/config_test.go`

- [ ] **P2** | Update documentation | ~medium
  - Acceptance: README.md, CLAUDE.md, AGENTS.md all reference `linkdingctl` binary name and config path; command examples updated
  - Files: `README.md`, `CLAUDE.md`, `AGENTS.md`

- [ ] **P3** | Update specification files | ~medium
  - Acceptance: All spec files in `specs/` reference `linkdingctl` in examples and file paths
  - Files: `specs/01-core-cli.md` through `specs/16-rename-to-linkdingctl.md`

---

## Dependency Graph

```
Phase 5 (Notes Field) ─────────────────────── standalone, no deps
Phase 6 (Tags CRUD) ───────────────────────── standalone, no deps
Phase 7 (User Profile) ────────────────────── standalone, no deps
Phase 8 (Per-Command Overrides) ────────────── standalone, no deps
Phase 9 (Rename to linkdingctl) ────────────── depends on Phases 5-8
                                               (rename AFTER all features land,
                                               so rename applies to final code)
```

## Execution Order

Spec 16 states the rename has **HIGHEST** priority and must be done first. However, this creates a practical challenge: doing the rename first means all subsequent feature work targets `cmd/linkdingctl/` paths, but any in-flight branches targeting `cmd/ld/` will conflict.

**Recommended order (minimize merge conflicts):**

1. **Phase 9** (P0): Rename to `linkdingctl` — do this first as spec mandates
2. **Phase 5** (P1): Notes field — small, self-contained flag additions
3. **Phase 6** (P1): Tags create/get — new API methods + subcommands
4. **Phase 7** (P1): User profile — new command group, new API endpoint
5. **Phase 8** (P1): Per-command overrides — modifies config loading (touches root.go, config.go)

**Alternative order (if rename is deferred to avoid disruption):**

1. **Phases 5-8** (P1): All features on `cmd/ld/` paths
2. **Phase 9** (P0): Rename everything at once (single large commit)

---

## Notes

- Specs 01-11 are fully implemented — no code changes needed
- The `BookmarkCreate` and `BookmarkUpdate` models already include `Notes` field — only CLI flag wiring needed
- The `Tag` model already has ID, Name, DateAdded — sufficient for `tags get` output
- `golang.org/x/term` is already in go.mod — no new dependencies needed for any phase
- Phase 8 modifies `loadConfig()` which is called by every command — requires careful testing
- Phase 9 is the largest change by file count but each individual change is mechanical (find-replace)
- Config migration in Phase 9 must be idempotent (safe to run multiple times)
- The `-n` shorthand for `--notes` does not conflict with any existing flag shorthand
