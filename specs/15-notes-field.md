# Specification: Notes Field (Separate from Description)

## Jobs to Be Done
- User can add Markdown-formatted notes to a bookmark (distinct from description)
- User can update notes independently from description
- User can view both description and notes when inspecting a bookmark
- User can leverage LinkDing's Markdown rendering for rich notes content

## Background

The LinkDing API distinguishes between two text fields on bookmarks:
- **`description`**: A short plain-text description/summary
- **`notes`**: A longer Markdown-formatted field for personal annotations

Currently, `linkdingctl` only exposes `--description` (`-d`) on `add` and `update` commands. The `notes` field in the data model exists but is not settable via CLI flags. This creates a gap where users cannot use the Markdown-capable notes functionality that LinkDing provides.

## Add Bookmark (Updated)
```
linkdingctl add <url> [flags]

New Flag:
  -n, --notes string   Markdown notes for the bookmark
```

Examples:
```bash
linkdingctl add https://example.com -d "Quick reference" -n "## Setup\n- Step 1\n- Step 2"
linkdingctl add https://example.com --notes "Important: check section 3 for config"
```

## Update Bookmark (Updated)
```
linkdingctl update <id> [flags]

New Flag:
  -n, --notes string   Update notes (Markdown supported)
```

Examples:
```bash
linkdingctl update 123 --notes "Updated: now requires v2.0+"
linkdingctl update 123 -d "New description" -n "New notes"
linkdingctl update 123 --notes ""  # Clear notes
```

## Get Bookmark (Updated)

`linkdingctl get <id>` already displays the notes field. No changes needed to output, but ensure notes are displayed distinctly from description.

Output (human):
```
ID:           123
URL:          https://example.com
Title:        Example Site
Description:  Quick reference
Notes:        ## Setup
              - Step 1
              - Step 2
Tags:         reference, docs
Date Added:   2026-01-23
Unread:       no
Shared:       no
Archived:     no
```

Output (JSON): The `notes` field is already included in JSON output via the model.

## Export/Import Impact

The existing export/import functionality already handles the notes field through the bookmark model. No changes needed to export formats — notes are preserved in JSON exports and should round-trip correctly.

## Implementation Notes

- Add `--notes` (`-n`) string flag to `add` command in `cmd/linkdingctl/add.go`
- Add `--notes` (`-n`) string flag to `update` command in `cmd/linkdingctl/update.go`
- Map the flag value to `Notes` field on `BookmarkCreate` (add) and `BookmarkUpdate` (update)
- The `BookmarkCreate` and `BookmarkUpdate` models already include the `Notes` field
- Update the `--description` flag help text from "Description/notes" to "Description" to avoid confusion
- An empty string for `--notes` on update should clear the field (set to `""`)

## Success Criteria
- [ ] `linkdingctl add --notes` sets the notes field on bookmark creation
- [ ] `linkdingctl update --notes` updates the notes field
- [ ] `linkdingctl update --notes ""` clears the notes field
- [ ] `linkdingctl get` displays notes separately from description
- [ ] Notes and description can be set independently
- [ ] Notes and description can be set together in one command
- [ ] `-n` shorthand works for `--notes` (no conflict with existing flags)
- [ ] `--description` help text no longer references "notes"
- [ ] All commands respect `--json` flag
- [ ] Export/import preserves notes field correctly (existing behavior, verify only)
