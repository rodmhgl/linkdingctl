# Specification: Tag CRUD Enhancements (Create & Get)

## Jobs to Be Done
- User can create tags without needing to attach them to a bookmark first
- User can retrieve a specific tag's details by ID
- User can pre-populate tags before bulk-importing bookmarks

## Create Tag
```
linkdingctl tags create <tag-name>
```

Creates a new tag in LinkDing via `POST /api/tags/`.

Examples:
```bash
linkdingctl tags create "kubernetes"
linkdingctl tags create "home-automation"
linkdingctl tags create "kubernetes"  # already exists → error
```

Output (human):
```
Tag created: kubernetes
  ID: 42
```

Output (JSON):
```json
{"id": 42, "name": "kubernetes", "date_added": "2026-01-23T10:30:00Z"}
```

Error (duplicate):
```
Error: Tag "kubernetes" already exists (ID: 42).
```

### API Details

`POST /api/tags/`

Request body:
```json
{"name": "kubernetes"}
```

Response (201 Created):
```json
{"id": 42, "name": "kubernetes", "date_added": "2026-01-23T10:30:00Z"}
```

Response (400 Bad Request / conflict): Tag already exists.

## Get Tag
```
linkdingctl tags get <tag-id>
```

Retrieves a specific tag by its numeric ID via `GET /api/tags/{id}/`.

Examples:
```bash
linkdingctl tags get 42
linkdingctl tags get 42 --json
```

Output (human):
```
ID:         42
Name:       kubernetes
Date Added: 2026-01-23
```

Output (JSON):
```json
{"id": 42, "name": "kubernetes", "date_added": "2026-01-23T10:30:00Z"}
```

Error (not found):
```
Error: Tag with ID 42 not found.
```

### API Details

`GET /api/tags/{id}/`

Response (200 OK):
```json
{"id": 42, "name": "kubernetes", "date_added": "2026-01-23T10:30:00Z"}
```

Response (404 Not Found): Tag does not exist.

## Implementation Notes

- Add `CreateTag(name string)` method to `internal/api/client.go`
- Add `GetTag(id int)` method to `internal/api/client.go`
- Add `TagCreate` struct to `internal/models/` (request body for POST)
- Add `create` and `get` subcommands to the existing `tags` command in `cmd/linkdingctl/tags.go`
- The existing `Tag` model already has the required fields (ID, Name, DateAdded)

## Success Criteria
- [ ] `linkdingctl tags create <name>` creates a tag and returns its ID
- [ ] `linkdingctl tags create` with duplicate name reports a clear error
- [ ] `linkdingctl tags create` with empty name reports a validation error
- [ ] `linkdingctl tags get <id>` displays full tag details
- [ ] `linkdingctl tags get` with non-existent ID reports "not found" error
- [ ] Both commands respect `--json` flag
- [ ] Both commands respect global config and auth handling
