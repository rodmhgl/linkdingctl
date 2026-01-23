# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

LinkDing CLI (`linkdingctl`) — a Go CLI for managing bookmarks in a [LinkDing](https://github.com/sissbruecker/linkding) instance. Uses Cobra for commands, Viper for config, and stdlib `net/http` for API calls.

## Build & Development Commands

```bash
make build          # Build binary → ./linkdingctl
make test           # Run all tests (go test -v ./...)
make cover          # Tests with 70% per-package coverage gate
make test-coverage  # Tests + HTML coverage report
make vet            # go vet ./...
make fmt            # go fmt ./...
make lint           # golangci-lint (must be installed separately)
make check          # fmt + vet + test
make clean          # Remove binary and coverage artifacts
```

### Running a single test

```bash
go test -v -run TestFunctionName ./cmd/linkdingctl/
go test -v -run TestFunctionName ./internal/api/
```

### Build from scratch

```bash
go build -trimpath -ldflags "-s -w" -o linkdingctl ./cmd/linkdingctl
```

## Architecture

```
cmd/linkdingctl/             # Cobra commands (one file per command + root.go + main.go)
internal/
  api/              # LinkDing REST API client (Client struct, all HTTP logic)
  config/           # Viper-based config loading (~/.config/linkdingctl/config.yaml + env vars)
  models/           # Bookmark, Tag, and request/response structs
  export/           # Import/export logic (JSON, HTML/Netscape, CSV formats)
specs/              # Feature specification documents (numbered, sequential)
```

### Key Design Decisions

- **No local state** — LinkDing is the single source of truth; no local DB or cache.
- **Pagination** — API client handles cursor-based pagination via `FetchAllBookmarks`/`FetchAllTags` methods on `Client`. Commands fetch all pages transparently.
- **`--json` flag** — Every command supports JSON output for scripting. The global `jsonOutput` bool is set in `root.go`.
- **Exit codes** — 0=success, 1=error, 2=config error.
- **Auth** — Token-based: `Authorization: Token <token>` header. Token comes from config file or `LINKDING_TOKEN` env var.
- **Config precedence** — Environment variables (`LINKDING_URL`, `LINKDING_TOKEN`) override config file values.
- **Security** — Config directory created with `0700`, config file with `0600`. Token input masked during `config init`.

### Data Flow

1. Command parses flags → calls `loadConfig()` → gets `*config.Config`
2. Creates `api.NewClient(cfg.URL, cfg.Token)`
3. Client methods (`ListBookmarks`, `CreateBookmark`, etc.) handle HTTP + pagination
4. Results rendered as table (default) or JSON (`--json`)

## Testing Conventions

- Unit tests use mock HTTP servers (`httptest.NewServer`)
- Tests live alongside source files (`*_test.go`)
- `cmd/linkdingctl/commands_test.go` contains integration-style tests for all CLI commands
- Coverage threshold: 70% per package (enforced by `make cover`)

## Constraints (Do Not Add)

- No local database or caching layer
- No interactive/TUI mode — keep it scriptable
- No third-party HTTP clients — stdlib `net/http` only
- No browser integration