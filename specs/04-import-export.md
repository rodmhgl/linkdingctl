# Specification: Import, Export & Backup

## Jobs to Be Done
- User can export all bookmarks for backup
- User can import bookmarks from various formats
- User can create timestamped backups for disaster recovery
- User can restore from a backup file

## Export Bookmarks
```
linkdingctl export [flags]

Flags:
  -f, --format string    Output format: json, html, csv (default: json)
  -o, --output string    Output file (default: stdout)
  -T, --tags strings     Export only bookmarks with these tags
  --archived             Include archived bookmarks (default: true)
```

Examples:
```bash
linkdingctl export > bookmarks.json
linkdingctl export -f html -o bookmarks.html
linkdingctl export --tags homelab -f csv -o homelab.csv
```

### Export Formats

**JSON** (default): Full fidelity, includes all metadata
```json
{
  "version": "1",
  "exported_at": "2026-01-22T10:30:00Z",
  "source": "linkding",
  "bookmarks": [
    {
      "id": 123,
      "url": "https://example.com",
      "title": "Example",
      "description": "Notes here",
      "tags": ["tag1", "tag2"],
      "date_added": "2025-06-15T08:00:00Z",
      "date_modified": "2025-06-20T12:00:00Z",
      "unread": false,
      "shared": false,
      "archived": false
    }
  ]
}
```

**HTML**: Netscape bookmark format (browser-compatible)
```html
<!DOCTYPE NETSCAPE-Bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
    <DT><A HREF="https://example.com" ADD_DATE="1718438400" TAGS="tag1,tag2">Example</A>
    <DD>Notes here
</DL><p>
```

**CSV**: Simple tabular format
```csv
url,title,description,tags,date_added,unread,shared,archived
https://example.com,Example,Notes here,"tag1,tag2",2025-06-15T08:00:00Z,false,false,false
```

## Import Bookmarks
```
linkdingctl import <file> [flags]

Flags:
  -f, --format string    Input format: json, html, csv (default: auto-detect)
  --dry-run              Show what would be imported without making changes
  --skip-duplicates      Skip URLs that already exist (default: update them)
  -T, --add-tags strings Add these tags to all imported bookmarks
```

Examples:
```bash
linkdingctl import bookmarks.json
linkdingctl import bookmarks.html --add-tags "imported"
linkdingctl import export.csv --dry-run
```

Auto-detection:
- `.json` → JSON format
- `.html`, `.htm` → HTML/Netscape format
- `.csv` → CSV format

Output:
```
Importing bookmarks...
  ✓ 150 new bookmarks added
  ✓ 23 existing bookmarks updated
  ⊘ 5 skipped (--skip-duplicates)
  ✗ 2 failed (see errors below)

Errors:
  Line 45: Invalid URL "not-a-url"
  Line 89: Missing required field "url"
```

## Backup Command
```
linkdingctl backup [flags]

Flags:
  -o, --output string    Output directory (default: current directory)
  --prefix string        Filename prefix (default: "linkding-backup")
```

Creates timestamped backup file: `linkding-backup-2026-01-22T103000.json`

Equivalent to: `linkdingctl export -f json -o <timestamped-file>`

Example:
```bash
linkdingctl backup -o ~/backups/
# Creates: ~/backups/linkding-backup-2026-01-22T103000.json
```

## Restore Command
```
linkdingctl restore <backup-file> [flags]

Flags:
  --dry-run              Show what would be restored
  --wipe                 Delete all existing bookmarks before restore (DANGEROUS)
```

Without `--wipe`: Equivalent to `linkdingctl import <file>`
With `--wipe`: Clears all bookmarks first, then imports. Requires interactive confirmation:

```
WARNING: This will delete ALL 500 existing bookmarks before restoring.
Type 'yes' to confirm: 
```

## Success Criteria
- [ ] `linkdingctl export` outputs valid JSON to stdout
- [ ] `linkdingctl export -f html` produces browser-importable HTML
- [ ] `linkdingctl export -f csv` produces valid CSV with headers
- [ ] `linkdingctl export -o <file>` writes to file instead of stdout
- [ ] `linkdingctl export --tags` filters exported bookmarks
- [ ] `linkdingctl import` auto-detects format from extension
- [ ] `linkdingctl import` handles all three formats correctly
- [ ] `linkdingctl import --dry-run` shows preview without changes
- [ ] `linkdingctl import` reports success/update/skip/error counts
- [ ] `linkdingctl backup` creates timestamped JSON file
- [ ] `linkdingctl restore` imports from backup file
- [ ] `linkdingctl restore --wipe` requires confirmation
- [ ] All commands respect `--json` flag for status output
