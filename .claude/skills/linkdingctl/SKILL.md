---
name: linkdingctl
description: Use when the user mentions linkdingctl, linkding, bookmarks, bookmark manager, saving a URL or link, tagging bookmarks, exporting or importing bookmarks, backup bookmarks, restore bookmarks, bundles, or managing saved links. This skill provides accurate CLI syntax for linkdingctl.
---

# linkdingctl — LinkDing CLI Reference

`linkdingctl` is a Go CLI for managing bookmarks in a [LinkDing](https://github.com/sissbruecker/linkding) instance. Single binary, no local state — LinkDing is the source of truth.

**Scope:** This skill covers the `linkdingctl` CLI only — not the LinkDing REST API directly.

Every command supports `--json` for machine-readable output.

---

## 1. Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--json` | | `false` | Output as JSON instead of human-readable |
| `--debug` | | `false` | Enable debug logging (to stderr) |
| `--config` | | `~/.config/linkdingctl/config.yaml` | Config file path |
| `--url` | | | LinkDing instance URL (overrides config + env) |
| `--token` | | | API token (overrides config + env) |

**Config precedence:** config file < `LINKDING_URL`/`LINKDING_TOKEN` env vars < `--url`/`--token` CLI flags.

If both `--url` and `--token` are provided, no config file is needed at all.

---

## 2. Configuration

### `config init`
Interactive setup. Prompts for URL and token (token input is masked on TTY).
Saves to `~/.config/linkdingctl/config.yaml` with `0600` permissions (dir `0700`).

```bash
linkdingctl config init
```

### `config show`
Displays current config with token redacted. Shows source of each value (config file, env var, or CLI flag).

```bash
linkdingctl config show
```

### `config test`
Tests connectivity to the LinkDing instance.

```bash
linkdingctl config test
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `LINKDING_URL` | LinkDing instance URL |
| `LINKDING_TOKEN` | API authentication token |

---

## 3. Bookmark Commands

### `add <url>`

Add a new bookmark. URL is a positional argument (required).

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--title` | `-t` | auto-fetch | Custom title |
| `--description` | `-d` | `""` | Description |
| `--notes` | `-n` | `""` | Notes |
| `--tags` | `-T` | `nil` | Comma-separated tags |
| `--unread` | `-u` | `false` | Mark as unread |
| `--shared` | `-s` | `false` | Make publicly shared |

```bash
linkdingctl add "https://example.com" -T "reading,tech" -u
```

### `list`

List bookmarks with optional filtering.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--query` | `-q` | `""` | Search query |
| `--tags` | `-T` | `[]` | Filter by tags (AND logic) |
| `--unread` | `-u` | `false` | Show only unread |
| `--archived` | `-a` | `false` | Show only archived |
| `--limit` | `-l` | `100` | Max results |
| `--offset` | `-o` | `0` | Pagination offset |

```bash
linkdingctl list --tags k8s,platform -q "kubernetes" --limit 10
```

### `get <id>`

Get full details of a single bookmark by ID. No additional flags beyond globals.

```bash
linkdingctl get 123
```

### `update <id>`

Update a bookmark's metadata. Only specified fields are modified (PATCH semantics).

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--title` | `-t` | `""` | New title |
| `--description` | `-d` | `""` | New description |
| `--notes` | `-n` | `""` | New notes |
| `--tags` | `-T` | `nil` | Replace ALL tags (comma-separated) |
| `--add-tags` | | `nil` | Add tags to existing (comma-separated) |
| `--remove-tags` | | `nil` | Remove specific tags (comma-separated) |
| `--archive` | `-a` | `false` | Archive the bookmark |
| `--unarchive` | | `false` | Unarchive the bookmark |
| `--unread` | `-u` | `false` | Set unread status |
| `--shared` | `-s` | `false` | Set shared status |

**Mutual exclusivity:**
- `--tags` cannot be used with `--add-tags` or `--remove-tags` (error if combined)
- `--archive` and `--unarchive` cannot be used together

When using `--add-tags`/`--remove-tags`, the CLI fetches the current bookmark first, merges the tag changes, then sends the update.

```bash
linkdingctl update 123 --title "New Title" --add-tags "reviewed" --remove-tags "draft"
```

### `delete <id>`

Delete a bookmark by ID. Shows confirmation prompt unless `--force` or `--json` is set.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | `-f` | `false` | Skip confirmation prompt |

**Note:** Both `--force` and `--json` skip the confirmation prompt.

```bash
linkdingctl delete 123 --force
```

---

## 4. Tag Commands

### `tags`

List all tags with bookmark counts. Counts are computed client-side by fetching all bookmarks (including archived).

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--sort` | `-s` | `name` | Sort by: `name`, `count` (count = descending) |
| `--unused` | | `false` | Show only tags with 0 bookmarks |

```bash
linkdingctl tags --sort count
linkdingctl tags --unused
```

### `tags create <name>`

Create a new tag.

```bash
linkdingctl tags create "kubernetes"
```

### `tags get <id>`

Get a tag by its numeric ID.

```bash
linkdingctl tags get 42
```

### `tags rename <old-name> <new-name>`

Rename a tag across all bookmarks. Fetches all bookmarks with the old tag, replaces it with the new tag on each one. Shows progress.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | `-f` | `false` | Skip confirmation |

```bash
linkdingctl tags rename "oldtag" "newtag" --force
```

### `tags delete <tag-name>`

Delete a tag. By default only works if the tag has 0 bookmarks. With `--force`, removes the tag from all bookmarks first (with confirmation prompt).

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--force` | `-f` | `false` | Skip safety check and remove tag from all bookmarks |

```bash
linkdingctl tags delete "unused-tag"
linkdingctl tags delete "old-tag" --force
```

### `tags show <tag-name>`

Show all bookmarks with a specific tag (including archived). Equivalent to `list --tags <tag-name>` but fetches all pages.

```bash
linkdingctl tags show "kubernetes"
```

---

## 5. Bundle Commands

Bundles are saved search configurations. Tag fields (`--any-tags`, `--all-tags`, `--excluded-tags`) are **comma-separated strings**, not slices.

### `bundles list`

List all bundles.

```bash
linkdingctl bundles list
```

### `bundles get <id>`

Get full details of a bundle by ID.

```bash
linkdingctl bundles get 1
```

### `bundles create <name>`

Create a new bundle.

| Flag | Default | Description |
|------|---------|-------------|
| `--search` | `""` | Search query for the bundle |
| `--any-tags` | `""` | Comma-separated tags (any match) |
| `--all-tags` | `""` | Comma-separated tags (all required) |
| `--excluded-tags` | `""` | Comma-separated tags to exclude |
| `--order` | `0` | Display order |

```bash
linkdingctl bundles create "Tech" --search "kubernetes" --any-tags "k8s,docker"
```

### `bundles update <id>`

Update an existing bundle. **PATCH semantics** — only sends fields that were explicitly set via flags.

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | `""` | New name |
| `--search` | `""` | Search query |
| `--any-tags` | `""` | Comma-separated tags (any match) |
| `--all-tags` | `""` | Comma-separated tags (all required) |
| `--excluded-tags` | `""` | Comma-separated tags to exclude |
| `--order` | `-1` | Display order |

Errors if no flags are provided.

```bash
linkdingctl bundles update 1 --name "Renamed" --any-tags "tag1,tag2"
```

### `bundles delete <id>`

Delete a bundle by ID. No confirmation prompt.

```bash
linkdingctl bundles delete 1
```

---

## 6. Export & Import

### `export`

Export bookmarks to various formats.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--format` | `-f` | `json` | Output format: `json`, `html`, `csv` |
| `--output` | `-o` | stdout | Output file path |
| `--tags` | `-T` | `[]` | Export only bookmarks with these tags |
| `--archived` | | `true` | Include archived bookmarks |

**Note:** `--archived` defaults to `true` (includes archived by default).

```bash
linkdingctl export > bookmarks.json
linkdingctl export -f html -o bookmarks.html
linkdingctl export --tags homelab -f csv -o homelab.csv
linkdingctl export --archived=false -o active-only.json
```

### `import <file>`

Import bookmarks from a file. Format is auto-detected from extension (`.json`, `.html`/`.htm`, `.csv`).

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--format` | `-f` | `auto` | Input format: `json`, `html`, `csv`, `auto` |
| `--dry-run` | | `false` | Preview without making changes |
| `--skip-duplicates` | | `false` | Skip existing URLs (default: update them) |
| `--add-tags` | `-T` | `[]` | Add these tags to all imported bookmarks |

```bash
linkdingctl import bookmarks.json
linkdingctl import bookmarks.html --add-tags "imported" --dry-run
linkdingctl import export.csv --skip-duplicates
```

---

## 7. Backup & Restore

### `backup`

Create a timestamped JSON backup of all bookmarks (including archived).
Filename format: `{prefix}-{timestamp}.json` (e.g. `linkding-backup-2026-01-22T103000.json`).

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `.` (current dir) | Output directory |
| `--prefix` | | `linkding-backup` | Filename prefix |

```bash
linkdingctl backup
linkdingctl backup -o ~/backups/ --prefix my-instance
```

### `restore <backup-file>`

Restore bookmarks from a backup file.

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `false` | Preview without making changes |
| `--wipe` | `false` | Delete ALL existing bookmarks before restore (DANGEROUS) |

**Without `--wipe`:** Equivalent to `import` — existing bookmarks are updated, new ones are added.

**With `--wipe`:** Deletes all existing bookmarks first. Requires typing `yes` at the confirmation prompt. **Cannot be combined with `--json`** (errors out because interactive confirmation is required).

```bash
linkdingctl restore backup.json --dry-run
linkdingctl restore backup.json --wipe
```

---

## 8. User & Utility Commands

### `user profile`

Display the authenticated user's profile preferences (theme, sharing settings, search preferences, etc.).

```bash
linkdingctl user profile
linkdingctl user profile --json
```

### `version`

Print version, commit, build date, Go version, and OS/arch.

```bash
linkdingctl version
```

### `completion`

Generate shell completion scripts (built-in Cobra subcommand).

```bash
linkdingctl completion bash > /etc/bash_completion.d/linkdingctl
linkdingctl completion zsh > "${fpath[1]}/_linkdingctl"
linkdingctl completion fish > ~/.config/fish/completions/linkdingctl.fish
```

---

## 9. Common Workflows

### First-time setup
```bash
linkdingctl config init        # Interactive: URL + token
linkdingctl config test        # Verify connectivity
linkdingctl config show        # Confirm settings
```

### Save and tag a URL
```bash
linkdingctl add "https://example.com/article" -t "Great Article" -T "reading,tech" -u
```

### Search bookmarks
```bash
linkdingctl list -q "kubernetes" --tags k8s --limit 20
```

### Tag cleanup
```bash
linkdingctl tags --unused                          # Find unused tags
linkdingctl tags rename "k8s" "kubernetes" --force # Standardize naming
linkdingctl tags delete "old-tag" --force          # Remove from all bookmarks
```

### Backup cycle
```bash
linkdingctl backup -o ~/backups/                   # Create timestamped backup
linkdingctl restore ~/backups/linkding-backup-2026-01-22T103000.json --dry-run  # Preview
linkdingctl restore ~/backups/linkding-backup-2026-01-22T103000.json            # Restore
```

### Filtered export
```bash
linkdingctl export --tags homelab -f csv -o homelab.csv
linkdingctl export --archived=false -f html -o active.html
```

### Safe import with preview
```bash
linkdingctl import bookmarks.html --dry-run             # Preview first
linkdingctl import bookmarks.html --add-tags "imported"  # Then import
```

### Ad-hoc connection (no config file)
```bash
linkdingctl list --url "https://linkding.example.com" --token "abc123"
```

### Scripting with JSON + jq

**Note:** `list` returns at most 100 results by default. Use `--limit 0` to fetch all bookmarks when piping to jq.

```bash
# Get all bookmark URLs
linkdingctl list --json --limit 0 | jq -r '.results[].url'

# Count bookmarks per tag
linkdingctl tags --json | jq '.[] | "\(.name): \(.count)"'

# Get IDs of unread bookmarks
linkdingctl list --unread --json --limit 0 | jq '.results[].id'

# Export bookmark data as TSV
linkdingctl list --json --limit 0 | jq -r '.results[] | [.id, .url, .title] | @tsv'
```

**Important:** The native CLI cannot do OR-tag logic, negative tag filtering, date range filtering, or cross-field boolean queries. Before improvising jq pipelines for these, **read `~/.claude/skills/linkdingctl/references/jq-recipes.md`** — it has tested, efficient single-query recipes and the complete JSON field reference.

---

## 10. Common Mistakes

| Mistake | What happens | Fix |
|---------|-------------|-----|
| Using `--tags` with `--add-tags` or `--remove-tags` on `update` | CLI errors — these are mutually exclusive | Use `--tags` to replace all tags, OR `--add-tags`/`--remove-tags` to modify incrementally |
| Piping `list --json` to jq without `--limit 0` | Only processes the first 100 bookmarks (default limit) | Always add `--limit 0` when you need all results |
| Using `restore --wipe` with `--json` | CLI errors — `--wipe` requires interactive `yes` confirmation | Run `--wipe` restores interactively without `--json` |
| Using `--archive` and `--unarchive` together on `update` | CLI errors — mutually exclusive | Use one or the other |
| Assuming `export` excludes archived bookmarks | `--archived` defaults to `true` — archived bookmarks are included | Pass `--archived=false` to exclude them |

---

## 11. Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `1` | Error (API failure, invalid input, etc.) |
| `2` | Configuration error |
