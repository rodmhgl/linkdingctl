# LinkDing CLI (`ld`)

A fast, scriptable command-line interface for managing bookmarks in your [LinkDing](https://github.com/sissbruecker/linkding) instance.

## Features

- **Full CRUD operations**: Add, list, get, update, and delete bookmarks
- **Tag management**: List, rename, and delete tags
- **Import/Export**: Support for JSON, HTML (Netscape format), and CSV
- **Backup/Restore**: Timestamped backups with safe restore options
- **Scriptable**: All commands support `--json` output for automation
- **Fast**: Single Go binary, no external dependencies

## Installation

### From Source

```bash
git clone https://github.com/yourusername/linkding-cli
cd linkding-cli
go build -o ld ./cmd/ld
sudo mv ld /usr/local/bin/
```

### Binary Releases

Download the latest binary from the [Releases page](https://github.com/yourusername/linkding-cli/releases).

## Quick Start

### 1. Configure your LinkDing connection

```bash
# Interactive setup
ld config init

# Or set via environment variables
export LINKDING_URL="https://linkding.example.com"
export LINKDING_TOKEN="your-api-token"

# Test connection
ld config test
```

### 2. Add a bookmark

```bash
ld add https://example.com --title "Example Site" --tags "reference,tools"
```

### 3. List bookmarks

```bash
# List all bookmarks
ld list

# Filter by tags
ld list --tags "homelab,docker"

# Show only unread bookmarks
ld list --unread

# JSON output for scripting
ld list --json
```

## Commands

### Configuration

```bash
# Initialize configuration interactively
ld config init

# Show current configuration
ld config show

# Test connection to LinkDing
ld config test
```

Configuration is stored in `~/.config/ld/config.yaml` or can be set via environment variables:
- `LINKDING_URL`
- `LINKDING_TOKEN`

### Bookmarks

#### Add

```bash
ld add <url> [flags]

Flags:
  -t, --title string         Bookmark title
  -d, --description string   Bookmark description
  -T, --tags strings         Comma-separated tags
  --unread                   Mark as unread
  --shared                   Make publicly shared
  --archived                 Add to archive

Examples:
  ld add https://example.com
  ld add https://example.com --title "Example" --tags "dev,tools"
  ld add https://news.com --unread --tags "reading-list"
```

#### List

```bash
ld list [flags]

Flags:
  -q, --query string    Search query
  -T, --tags strings    Filter by tags
  --unread              Show only unread bookmarks
  --shared              Show only shared bookmarks
  --archived            Show only archived bookmarks
  --limit int           Number of results (default: all)

Examples:
  ld list
  ld list --tags "homelab"
  ld list --query "kubernetes" --limit 10
  ld list --unread --json
```

#### Get

```bash
ld get <id>

Examples:
  ld get 123
  ld get 123 --json
```

#### Update

```bash
ld update <id> [flags]

Flags:
  --url string              New URL
  -t, --title string        New title
  -d, --description string  New description
  -T, --tags strings        Replace tags (comma-separated)
  --add-tags strings        Add tags without removing existing
  --remove-tags strings     Remove specific tags
  --unread bool             Set unread status
  --shared bool             Set shared status
  --archived bool           Set archived status

Examples:
  ld update 123 --title "New Title"
  ld update 123 --add-tags "important"
  ld update 123 --archived=true
```

#### Delete

```bash
ld delete <id> [flags]

Flags:
  -f, --force   Skip confirmation prompt

Examples:
  ld delete 123
  ld delete 123 --force
```

### Tags

```bash
# List all tags with bookmark counts
ld tags

# Sort by name or count
ld tags --sort name
ld tags --sort count

# Rename a tag across all bookmarks
ld tags rename <old> <new>
ld tags rename "home-lab" "homelab"

# Delete a tag (shows affected bookmarks)
ld tags delete <name>
ld tags delete "obsolete"
ld tags delete "obsolete" --force  # Skip confirmation
```

### Import/Export

#### Export

```bash
ld export [flags]

Flags:
  -f, --format string    Output format: json, html, csv (default: json)
  -o, --output string    Output file (default: stdout)
  -T, --tags strings     Export only bookmarks with these tags
  --archived             Include archived bookmarks (default: true)

Examples:
  ld export > bookmarks.json
  ld export -f html -o bookmarks.html
  ld export --tags homelab -f csv -o homelab.csv
```

**Export Formats:**
- **JSON**: Full fidelity with all metadata
- **HTML**: Netscape bookmark format (browser-compatible)
- **CSV**: Simple tabular format

#### Import

```bash
ld import <file> [flags]

Flags:
  -f, --format string      Input format: json, html, csv (default: auto-detect)
  --dry-run                Show what would be imported without making changes
  --skip-duplicates        Skip URLs that already exist (default: update them)
  -T, --add-tags strings   Add these tags to all imported bookmarks

Examples:
  ld import bookmarks.json
  ld import bookmarks.html --add-tags "imported"
  ld import export.csv --dry-run
```

Format is auto-detected from file extension:
- `.json` → JSON
- `.html`, `.htm` → HTML/Netscape
- `.csv` → CSV

#### Backup

```bash
ld backup [flags]

Flags:
  -o, --output string    Output directory (default: current directory)
  --prefix string        Filename prefix (default: "linkding-backup")

Examples:
  ld backup
  ld backup -o ~/backups/
  ld backup --prefix my-backup

# Creates: linkding-backup-2026-01-22T103000.json
```

#### Restore

```bash
ld restore <backup-file> [flags]

Flags:
  --dry-run   Show what would be restored
  --wipe      Delete all existing bookmarks before restore (DANGEROUS)

Examples:
  ld restore backup.json
  ld restore backup.json --dry-run
  ld restore backup.json --wipe  # Requires confirmation

# Without --wipe: Updates existing bookmarks, adds new ones
# With --wipe: Deletes ALL bookmarks first, then imports
```

## Scripting

All commands support `--json` output for easy parsing:

```bash
# Get bookmark count
ld list --json | jq '.count'

# Export URLs from a specific tag
ld list --tags "homelab" --json | jq -r '.results[].url'

# Backup all bookmarks nightly
0 2 * * * ld backup -o ~/backups/ > /dev/null 2>&1

# Find bookmarks without tags
ld list --json | jq '.results[] | select(.tag_names | length == 0) | {id, title}'
```

## Configuration File

Location: `~/.config/ld/config.yaml`

```yaml
url: https://linkding.example.com
token: your-api-token-here
```

## Environment Variables

Environment variables override the config file:

- `LINKDING_URL`: Your LinkDing instance URL
- `LINKDING_TOKEN`: Your API token

## Exit Codes

- `0`: Success
- `1`: Error (API, network, etc.)
- `2`: Configuration error

## Architecture

- Single binary, no external dependencies
- Uses LinkDing's REST API
- Config via `~/.config/ld/config.yaml` or environment variables
- All state stored in LinkDing (no local database)

## Development

### Requirements

- Go 1.21+

### Building

```bash
go build -o ld ./cmd/ld
```

### Testing

```bash
go test ./...
```

### Code Structure

```
cmd/ld/           # Command implementations
internal/
  api/            # LinkDing API client
  config/         # Configuration loading
  models/         # Data structures
  export/         # Import/export logic
```

## License

[Your License Here]

## Acknowledgments

Built for use with [LinkDing](https://github.com/sissbruecker/linkding) by sissbruecker.
