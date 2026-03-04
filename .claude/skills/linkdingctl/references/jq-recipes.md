# Advanced jq Recipes for linkdingctl

These recipes extend linkdingctl's native filtering with jq post-processing. All examples pipe from `linkdingctl list --json` (or `--limit 0` to fetch all bookmarks).

---

## JSON Schema Reference

The `list --json` output wraps bookmarks in a paginated envelope:

```json
{
  "count": 42,
  "next": null,
  "previous": null,
  "results": [
    {
      "id": 1,
      "url": "https://example.com",
      "title": "Example",
      "description": "",
      "notes": "",
      "website_title": "Example Domain",
      "website_description": "",
      "is_archived": false,
      "unread": true,
      "shared": false,
      "tag_names": ["reading", "tech"],
      "date_added": "2026-01-15T10:30:00Z",
      "date_modified": "2026-01-20T14:00:00Z"
    }
  ]
}
```

**Field types:**
- `id`: integer
- `url`, `title`, `description`, `notes`, `website_title`, `website_description`: string
- `is_archived`, `unread`, `shared`: boolean
- `tag_names`: array of strings
- `date_added`, `date_modified`: ISO 8601 timestamp string

Use `--limit 0` to bypass the default 100-result limit and fetch all bookmarks.

---

## 1. Negative Tag Filtering

The `--tags` flag only supports AND inclusion. Use jq to exclude bookmarks that have specific tags.

### Exclude a single tag

```bash
# All bookmarks EXCEPT those tagged "archive-later"
linkdingctl list --json --limit 0 | jq '[.results[] | select((.tag_names | index("archive-later")) | not)]'
```

### Exclude multiple tags

```bash
# Exclude bookmarks tagged "spam" OR "deprecated"
linkdingctl list --json --limit 0 | jq '[.results[] | select(.tag_names as $t | ["spam", "deprecated"] | all(. as $ex | $t | index($ex) | not))]'
```

### Include one tag, exclude another

```bash
# Tagged "kubernetes" but NOT "outdated"
linkdingctl list --tags kubernetes --json --limit 0 | jq '[.results[] | select((.tag_names | index("outdated")) | not)]'
```

---

## 2. Tag Set Operations

### OR logic (any of these tags)

```bash
# Bookmarks with "k8s" OR "docker" OR "containers"
linkdingctl list --json --limit 0 | jq '[.results[] | select(.tag_names as $t | ["k8s", "docker", "containers"] | any(. as $tag | $t | index($tag)))]'
```

### Find untagged bookmarks

```bash
# Bookmarks with no tags at all
linkdingctl list --json --limit 0 | jq '[.results[] | select(.tag_names | length == 0)]'
```

### Bookmarks with exactly N tags

```bash
# Bookmarks with exactly one tag
linkdingctl list --json --limit 0 | jq '[.results[] | select(.tag_names | length == 1)]'
```

### Tag difference (has tag A but not tag B)

```bash
# Tagged "homelab" but missing "documented"
linkdingctl list --json --limit 0 | jq '[.results[] | select((.tag_names | index("homelab")) and ((.tag_names | index("documented")) | not))]'
```

---

## 3. Date Filtering

Dates are ISO 8601 strings, so lexicographic comparison works directly.

### Bookmarks added after a date

```bash
# Added after 2026-01-01
linkdingctl list --json --limit 0 | jq '[.results[] | select(.date_added > "2026-01-01")]'
```

### Bookmarks added in a date range

```bash
# Added in January 2026
linkdingctl list --json --limit 0 | jq '[.results[] | select(.date_added >= "2026-01-01" and .date_added < "2026-02-01")]'
```

### Recently modified (last N days)

```bash
# Modified in the last 7 days (compute cutoff dynamically)
linkdingctl list --json --limit 0 | jq --arg cutoff "$(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%SZ)" '[.results[] | select(.date_modified >= $cutoff)]'
```

### Stale bookmarks (not modified since a date)

```bash
# Not modified since 2025-06-01
linkdingctl list --json --limit 0 | jq '[.results[] | select(.date_modified < "2025-06-01")]'
```

### Sort by date added (newest first)

```bash
linkdingctl list --json --limit 0 | jq '[.results[] | select(.unread)] | sort_by(.date_added) | reverse'
```

---

## 4. Multi-Field Queries

Combine boolean, tag, and text field checks that the CLI can't express natively.

### Unread AND tagged with any of several tags

```bash
# Unread bookmarks tagged "reading" or "toread"
linkdingctl list --unread --json --limit 0 | jq '[.results[] | select(.tag_names as $t | ["reading", "toread"] | any(. as $tag | $t | index($tag)))]'
```

### Search in notes field

```bash
# Bookmarks where notes contain "TODO"
linkdingctl list --json --limit 0 | jq '[.results[] | select(.notes | test("TODO"; "i"))]'
```

### Search across multiple text fields

```bash
# "kubernetes" appears in title, description, or notes
linkdingctl list --json --limit 0 | jq '[.results[] | select((.title + " " + .description + " " + .notes) | test("kubernetes"; "i"))]'
```

### Shared but unread

```bash
linkdingctl list --json --limit 0 | jq '[.results[] | select(.shared and .unread)]'
```

### Complex boolean: (tagged A OR tagged B) AND unread AND not archived

```bash
linkdingctl list --json --limit 0 | jq '[.results[] | select(
  (.tag_names as $t | ["devops", "sre"] | any(. as $tag | $t | index($tag)))
  and .unread
  and (.is_archived | not)
)]'
```

---

## 5. Aggregation & Reporting

### Count bookmarks per domain

```bash
linkdingctl list --json --limit 0 | jq '[.results[].url | capture("https?://(?<domain>[^/]+)").domain] | group_by(.) | map({domain: .[0], count: length}) | sort_by(.count) | reverse'
```

### Tag co-occurrence (which tags appear together most often)

```bash
linkdingctl list --json --limit 0 | jq '
  [.results[] | .tag_names | select(length >= 2) | combinations(2)] |
  map(sort | join(" + ")) |
  group_by(.) | map({pair: .[0], count: length}) |
  sort_by(.count) | reverse | .[0:10]
'
```

### Summary statistics

```bash
linkdingctl list --json --limit 0 | jq '{
  total: (.results | length),
  unread: [.results[] | select(.unread)] | length,
  archived: [.results[] | select(.is_archived)] | length,
  shared: [.results[] | select(.shared)] | length,
  untagged: [.results[] | select(.tag_names | length == 0)] | length
}'
```

### Top 10 most-used tags (from bookmark data)

```bash
linkdingctl list --json --limit 0 | jq '[.results[].tag_names[]] | group_by(.) | map({tag: .[0], count: length}) | sort_by(.count) | reverse | .[0:10]'
```

### Bookmarks per month

```bash
linkdingctl list --json --limit 0 | jq '[.results[].date_added[0:7]] | group_by(.) | map({month: .[0], count: length}) | sort_by(.month)'
```

---

## 6. Batch Operations

linkdingctl operates on one bookmark at a time. Combine jq ID extraction with xargs for bulk mutations.

### Archive all bookmarks with a specific tag

```bash
linkdingctl list --tags "to-archive" --json --limit 0 | jq -r '.results[].id' | xargs -I{} linkdingctl update {} --archive
```

### Add a tag to all unread bookmarks

```bash
linkdingctl list --unread --json --limit 0 | jq -r '.results[].id' | xargs -I{} linkdingctl update {} --add-tags "needs-review"
```

### Delete bookmarks matching a domain

```bash
# Preview first
linkdingctl list --json --limit 0 | jq -r '.results[] | select(.url | test("spam-domain\\.com")) | "\(.id) \(.url)"'

# Then delete (with force to skip prompts)
linkdingctl list --json --limit 0 | jq -r '.results[] | select(.url | test("spam-domain\\.com")) | .id' | xargs -I{} linkdingctl delete {} --force
```

### Remove a tag from all bookmarks that have it

```bash
linkdingctl list --tags "old-tag" --json --limit 0 | jq -r '.results[].id' | xargs -I{} linkdingctl update {} --remove-tags "old-tag"
```

### Mark all archived bookmarks as read

```bash
linkdingctl list --archived --json --limit 0 | jq -r '.results[] | select(.unread) | .id' | xargs -I{} linkdingctl update {} --unread=false
```

### Bulk operations with rate limiting

```bash
# Add 0.5s delay between API calls to avoid rate limiting
linkdingctl list --tags "process-me" --json --limit 0 | jq -r '.results[].id' | while read -r id; do
  linkdingctl update "$id" --add-tags "processed" --remove-tags "process-me"
  sleep 0.5
done
```

---

## Output Formatting Tips

### Compact table output

```bash
linkdingctl list --json --limit 0 | jq -r '.results[] | [.id, (.tag_names | join(",")), .title[0:60]] | @tsv'
```

### JSON output for further piping

```bash
# Extract just the fields you need
linkdingctl list --json --limit 0 | jq '[.results[] | {id, url, tags: .tag_names, added: .date_added[0:10]}]'
```

### CSV output

```bash
linkdingctl list --json --limit 0 | jq -r '.results[] | [.id, .url, .title, (.tag_names | join(";")), .date_added[0:10]] | @csv'
```

### Count results from any filter

Append `| length` to any array-producing recipe:

```bash
linkdingctl list --json --limit 0 | jq '[.results[] | select(.unread and (.tag_names | length == 0))] | length'
```
