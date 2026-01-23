# LinkDing CLI - Agent Operations Guide

## Project Overview
A Go CLI tool (`ld`) for managing bookmarks in a LinkDing instance.

## Technology Stack
- Language: Go 1.21+
- CLI Framework: cobra
- HTTP Client: net/http (stdlib)
- Config: viper (for config file + env vars)
- Output: tablewriter for lists, JSON for scripting

## Architecture Principles
1. Single binary, no external dependencies
2. Config via `~/.config/ld/config.yaml` or environment variables
3. All commands support `--json` flag for scriptable output
4. Exit codes: 0=success, 1=error, 2=config error

## LinkDing API
- Base URL from config: `LINKDING_URL` or config file
- Auth: Token-based via `Authorization: Token <api_token>`
- API docs: `{base_url}/api/docs/`

## File Organization
```
cmd/           # Cobra commands (one file per command)
internal/
  api/         # LinkDing API client
  config/      # Configuration loading
  models/      # Data structures
  export/      # Import/export logic
```

## Testing Requirements
- Unit tests for API client (mock HTTP)
- Integration tests marked with build tag `//go:build integration`
- Run tests: `go test ./...`

## DO NOT
- Add database/local caching (LinkDing is the source of truth)
- Add interactive/TUI mode (keep it scriptable)
- Add browser integration (out of scope)
- Use third-party HTTP clients (stdlib is sufficient)
