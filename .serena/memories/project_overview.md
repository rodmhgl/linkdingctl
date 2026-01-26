# LinkDing CLI Project Overview

## Purpose

A Go CLI tool (`linkdingctl`) for managing bookmarks in a LinkDing instance. Provides command-line access to create, read, update, delete bookmarks, manage tags, and import/export bookmark data.

## Tech Stack

- **Language**: Go 1.21+
- **CLI Framework**: cobra (github.com/spf13/cobra)
- **Configuration**: viper (github.com/spf13/viper) - supports config files and env vars
- **HTTP Client**: net/http (stdlib)
- **Output Format**: tablewriter for human-readable lists, JSON for scripting

## Project Structure

```
cmd/linkdingctl/           # Cobra commands (one file per command)
  - main.go       # Entry point
  - root.go       # Root command setup
  - config.go     # Config management commands
  - add.go        # Add bookmark command
internal/
  api/            # LinkDing API client
  config/         # Configuration loading
  models/         # Data structures
  export/         # Import/export logic
specs/            # Feature specifications with acceptance criteria
```

## Key Principles

- Single binary, no external dependencies
- Config via `~/.config/linkdingctl/config.yaml` or environment variables
- All commands support `--json` flag for scriptable output
- Exit codes: 0=success, 1=error, 2=config error
- LinkDing is the source of truth (no local caching)
