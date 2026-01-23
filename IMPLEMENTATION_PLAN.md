# Implementation Plan - LinkDing CLI

> Auto-generated plan. Run `./loop.sh plan` to regenerate from specs.

## Phase 0: Foundation

- [x] **P0** | Project scaffolding | ~small
  - Acceptance: `go mod init`, directory structure created, main.go compiles
  - Files: go.mod, cmd/ld/main.go, internal/ directories

- [x] **P0** | Configuration system | ~medium
  - Acceptance: Loads from file and env vars, viper configured
  - Files: internal/config/config.go, internal/config/config_test.go

- [x] **P0** | LinkDing API client | ~medium
  - Acceptance: HTTP client with auth, error handling, can call /api/bookmarks/
  - Files: internal/api/client.go, internal/api/client_test.go

- [x] **P0** | Root command and help | ~small
  - Acceptance: `ld --help` shows usage, global flags work
  - Files: cmd/ld/root.go

## Phase 1: Core CRUD

- [x] **P1** | Config commands | ~medium
  - Acceptance: `ld config init`, `ld config show`, `ld config test` work
  - Files: cmd/ld/config.go

- [x] **P1** | Add bookmark command | ~medium
  - Acceptance: `ld add <url>` creates bookmark with all flags
  - Files: cmd/ld/add.go, internal/api/bookmarks.go

- [x] **P1** | List bookmarks command | ~medium
  - Acceptance: `ld list` with filters, table output, JSON output
  - Files: cmd/ld/list.go

- [x] **P1** | Get bookmark command | ~small
  - Acceptance: `ld get <id>` shows full details
  - Files: cmd/ld/get.go

- [x] **P1** | Update bookmark command | ~medium
  - Acceptance: `ld update <id>` with all flags, partial updates
  - Files: cmd/ld/update.go

- [x] **P1** | Delete bookmark command | ~small
  - Acceptance: `ld delete <id>` with confirmation, --force
  - Files: cmd/ld/delete.go

## Phase 2: Tags

- [ ] **P2** | List tags command | ~small
  - Acceptance: `ld tags` lists with counts, sorting
  - Files: cmd/ld/tags.go, internal/api/tags.go

- [ ] **P2** | Tag rename command | ~medium
  - Acceptance: `ld tags rename` updates all affected bookmarks
  - Files: cmd/ld/tags.go (extend)

- [ ] **P2** | Tag delete command | ~small
  - Acceptance: `ld tags delete` with safety check and --force
  - Files: cmd/ld/tags.go (extend)

## Phase 3: Import/Export

- [ ] **P2** | Export JSON | ~medium
  - Acceptance: `ld export` outputs valid JSON with metadata
  - Files: cmd/ld/export.go, internal/export/json.go

- [ ] **P2** | Export HTML | ~medium
  - Acceptance: `ld export -f html` produces Netscape format
  - Files: internal/export/html.go

- [ ] **P2** | Export CSV | ~small
  - Acceptance: `ld export -f csv` produces valid CSV
  - Files: internal/export/csv.go

- [ ] **P2** | Import JSON | ~medium
  - Acceptance: `ld import file.json` with dry-run, progress
  - Files: cmd/ld/import.go, internal/export/import.go

- [ ] **P2** | Import HTML | ~medium
  - Acceptance: `ld import file.html` parses Netscape format
  - Files: internal/export/import.go (extend)

- [ ] **P2** | Import CSV | ~small
  - Acceptance: `ld import file.csv` with proper parsing
  - Files: internal/export/import.go (extend)

- [ ] **P2** | Backup command | ~small
  - Acceptance: `ld backup` creates timestamped file
  - Files: cmd/ld/backup.go

- [ ] **P2** | Restore command | ~medium
  - Acceptance: `ld restore` with --wipe safety
  - Files: cmd/ld/restore.go

## Phase 4: Polish

- [ ] **P3** | README documentation | ~small
  - Acceptance: Installation, usage examples, all commands documented
  - Files: README.md

- [ ] **P3** | Makefile | ~small
  - Acceptance: `make build`, `make test`, `make install`
  - Files: Makefile

---

## Notes

_Claude: Add notes about blockers or discoveries here_

