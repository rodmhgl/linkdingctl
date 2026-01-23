# Implementation Plan - LinkDing CLI

> Auto-generated plan. Run `./loop.sh plan` to regenerate from specs.

## Phase 0: Foundation

- [x] **P0** | Project scaffolding | ~small
  - Acceptance: `go mod init`, directory structure created, main.go compiles
  - Files: go.mod, cmd/linkdingctl/main.go, internal/ directories

- [x] **P0** | Configuration system | ~medium
  - Acceptance: Loads from file and env vars, viper configured
  - Files: internal/config/config.go, internal/config/config_test.go

- [x] **P0** | LinkDing API client | ~medium
  - Acceptance: HTTP client with auth, error handling, can call /api/bookmarks/
  - Files: internal/api/client.go, internal/api/client_test.go

- [x] **P0** | Root command and help | ~small
  - Acceptance: `linkdingctl --help` shows usage, global flags work
  - Files: cmd/linkdingctl/root.go

## Phase 1: Core CRUD

- [x] **P1** | Config commands | ~medium
  - Acceptance: `linkdingctl config init`, `linkdingctl config show`, `linkdingctl config test` work
  - Files: cmd/linkdingctl/config.go

- [x] **P1** | Add bookmark command | ~medium
  - Acceptance: `linkdingctl add <url>` creates bookmark with all flags
  - Files: cmd/linkdingctl/add.go, internal/api/bookmarks.go

- [x] **P1** | List bookmarks command | ~medium
  - Acceptance: `linkdingctl list` with filters, table output, JSON output
  - Files: cmd/linkdingctl/list.go

- [x] **P1** | Get bookmark command | ~small
  - Acceptance: `linkdingctl get <id>` shows full details
  - Files: cmd/linkdingctl/get.go

- [x] **P1** | Update bookmark command | ~medium
  - Acceptance: `linkdingctl update <id>` with all flags, partial updates
  - Files: cmd/linkdingctl/update.go

- [x] **P1** | Delete bookmark command | ~small
  - Acceptance: `linkdingctl delete <id>` with confirmation, --force
  - Files: cmd/linkdingctl/delete.go

## Phase 2: Tags

- [x] **P2** | List tags command | ~small
  - Acceptance: `linkdingctl tags` lists with counts, sorting
  - Files: cmd/linkdingctl/tags.go, internal/api/tags.go

- [x] **P2** | Tag rename command | ~medium
  - Acceptance: `linkdingctl tags rename` updates all affected bookmarks
  - Files: cmd/linkdingctl/tags.go (extend)

- [x] **P2** | Tag delete command | ~small
  - Acceptance: `linkdingctl tags delete` with safety check and --force
  - Files: cmd/linkdingctl/tags.go (extend)

## Phase 3: Import/Export

- [x] **P2** | Export JSON | ~medium
  - Acceptance: `linkdingctl export` outputs valid JSON with metadata
  - Files: cmd/linkdingctl/export.go, internal/export/json.go

- [x] **P2** | Export HTML | ~medium
  - Acceptance: `linkdingctl export -f html` produces Netscape format
  - Files: internal/export/html.go

- [x] **P2** | Export CSV | ~small
  - Acceptance: `linkdingctl export -f csv` produces valid CSV
  - Files: internal/export/csv.go

- [x] **P2** | Import JSON | ~medium
  - Acceptance: `linkdingctl import file.json` with dry-run, progress
  - Files: cmd/linkdingctl/import.go, internal/export/import.go

- [x] **P2** | Import HTML | ~medium
  - Acceptance: `linkdingctl import file.html` parses Netscape format
  - Files: internal/export/import.go (extend)

- [x] **P2** | Import CSV | ~small
  - Acceptance: `linkdingctl import file.csv` with proper parsing
  - Files: internal/export/import.go (extend)

- [x] **P2** | Backup command | ~small
  - Acceptance: `linkdingctl backup` creates timestamped file
  - Files: cmd/linkdingctl/backup.go

- [x] **P2** | Restore command | ~medium
  - Acceptance: `linkdingctl restore` with --wipe safety
  - Files: cmd/linkdingctl/restore.go

## Phase 4: Polish

- [x] **P3** | README documentation | ~small
  - Acceptance: Installation, usage examples, all commands documented
  - Files: README.md

- [x] **P3** | Makefile | ~small
  - Acceptance: `make build`, `make test`, `make install`
  - Files: Makefile

---

## Notes

_Claude: Add notes about blockers or discoveries here_
