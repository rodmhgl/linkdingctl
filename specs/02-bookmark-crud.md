# Specification: Bookmark CRUD Operations

## Jobs to Be Done
- User can quickly add a bookmark from the command line
- User can list bookmarks with filtering options
- User can view details of a specific bookmark
- User can update bookmark metadata
- User can delete bookmarks

## Add Bookmark
```
linkdingctl add <url> [flags]

Flags:
  -t, --title string        Custom title (default: auto-fetch)
  -d, --description string  Description/notes
  -T, --tags strings        Comma-separated tags
  -u, --unread              Mark as unread (default: false)
  -s, --shared              Make publicly shared (default: false)
```

Examples:
```bash
linkdingctl add https://example.com
linkdingctl add https://example.com -t "Example Site" -T "reference,docs"
linkdingctl add https://example.com --tags k8s,platform --unread
```

Output (human):
```
✓ Bookmark added: Example Site
  ID: 123
  URL: https://example.com
  Tags: reference, docs
```

Output (JSON):
```json
{"id": 123, "url": "https://example.com", "title": "Example Site", "tags": ["reference", "docs"]}
```

## List Bookmarks
```
linkdingctl list [flags]

Flags:
  -q, --query string    Search query
  -T, --tags strings    Filter by tags (AND logic)
  -u, --unread          Show only unread
  -a, --archived        Show only archived
  -l, --limit int       Max results (default: 100)
  -o, --offset int      Pagination offset
```

Examples:
```bash
linkdingctl list
linkdingctl list --tags k8s,platform
linkdingctl list -q "kubernetes" --unread
linkdingctl list --limit 10
```

Output (human): Table format with ID, Title (truncated), Tags, Date

## Get Bookmark
```
linkdingctl get <id>
```

Shows full bookmark details including description.

## Update Bookmark
```
linkdingctl update <id> [flags]

Flags:
  -t, --title string        New title
  -d, --description string  New description
  -T, --tags strings        Replace tags (use --add-tags/--remove-tags for partial)
  --add-tags strings        Add tags to existing
  --remove-tags strings     Remove specific tags
  -u, --unread              Set unread status
  -s, --shared              Set shared status
  -a, --archive             Archive the bookmark
  --unarchive               Unarchive the bookmark
```

Examples:
```bash
linkdingctl update 123 --add-tags "reviewed"
linkdingctl update 123 --title "New Title" --archive
```

## Delete Bookmark
```
linkdingctl delete <id> [flags]

Flags:
  -f, --force    Skip confirmation prompt
```

Requires confirmation unless `--force` or `--json` flag is set.

## Success Criteria
- [ ] `linkdingctl add <url>` creates bookmark and returns ID
- [ ] `linkdingctl add` with duplicate URL updates existing bookmark (LinkDing behavior)
- [ ] `linkdingctl list` shows bookmarks in table format
- [ ] `linkdingctl list --tags` filters correctly (AND logic)
- [ ] `linkdingctl list -q` searches title, description, URL
- [ ] `linkdingctl get <id>` shows full bookmark details
- [ ] `linkdingctl update <id>` modifies only specified fields
- [ ] `linkdingctl update --add-tags` appends without replacing
- [ ] `linkdingctl delete <id>` prompts for confirmation
- [ ] `linkdingctl delete --force` skips confirmation
- [ ] All commands respect `--json` flag
