# Specification: Tag Management

## Jobs to Be Done
- User can list all tags with bookmark counts
- User can rename tags across all bookmarks
- User can delete unused tags
- User can see which bookmarks have a specific tag

## List Tags
```
ld tags [flags]

Flags:
  -s, --sort string    Sort by: name, count (default: name)
  --unused             Show only tags with 0 bookmarks
```

Output (human):
```
TAG                COUNT
kubernetes         42
platform           38
homelab            25
...
```

Output (JSON):
```json
[{"name": "kubernetes", "count": 42}, ...]
```

## Tag Details
```
ld tags show <tag-name>
```

Lists all bookmarks with that tag (delegates to `ld list --tags <tag>`).

## Rename Tag
```
ld tags rename <old-name> <new-name> [flags]

Flags:
  -f, --force    Skip confirmation
```

Note: LinkDing API may not support direct tag rename. Implementation should:
1. Get all bookmarks with old tag
2. Update each bookmark: remove old tag, add new tag
3. Report progress: "Updating bookmark 1/42..."

## Delete Tag
```
ld tags delete <tag-name> [flags]

Flags:
  -f, --force    Skip confirmation
```

Only works if tag has 0 bookmarks. Otherwise error:
"Tag 'kubernetes' has 42 bookmarks. Remove tag from bookmarks first or use --force to remove from all."

With `--force`: removes tag from all bookmarks, then deletes tag.

## Success Criteria
- [ ] `ld tags` lists all tags with counts
- [ ] `ld tags --sort count` sorts by bookmark count descending
- [ ] `ld tags --unused` shows only zero-count tags
- [ ] `ld tags show <tag>` lists bookmarks with that tag
- [ ] `ld tags rename` updates all affected bookmarks
- [ ] `ld tags delete` fails safely if tag is in use
- [ ] `ld tags delete --force` removes tag from all bookmarks
- [ ] All commands respect `--json` flag
