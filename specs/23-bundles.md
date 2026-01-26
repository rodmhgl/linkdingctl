# PRD: Bundles Support for linkdingctl

## Executive Summary

Add full CRUD support for LinkDing bundles to linkdingctl, enabling users to manage saved search/filter groups through the CLI. Bundles allow organizing bookmarks via named collections with tag-based filtering, providing power users with scriptable workflow automation.

## Problem Statement

### Current Situation

linkdingctl supports bookmarks and tags but lacks support for LinkDing's bundles feature. Bundles are saved search configurations that combine search terms with tag filters (any_tags, all_tags, excluded_tags) and ordering.

### User Impact

- Users cannot create or manage bundles via CLI
- Scripting workflows cannot automate bundle operations
- No way to programmatically organize bookmark collections
- CLI feature parity gap with LinkDing web interface

### Business Impact

- Incomplete API coverage reduces CLI utility
- Users must switch to web UI for bundle management
- Limits automation and integration possibilities

### Why Solve This Now

- Bundles API is stable and documented
- Completes the core LinkDing resource coverage (bookmarks, tags, bundles)
- Follows established patterns from existing commands

## Goals & Success Metrics

| Goal | Metric | Baseline | Target | Timeframe |
|------|--------|----------|--------|-----------|
| Feature completeness | Bundle API operations covered | 0/5 | 5/5 | Implementation |
| Code quality | Test coverage for bundle code | 0% | 70%+ | Implementation |
| Consistency | Follows existing CLI patterns | N/A | 100% | Implementation |
| User adoption | Bundle commands used without errors | N/A | Works first try | Post-release |

## User Stories

### US-1: List Bundles

**As a** linkdingctl user
**I want to** list all my bundles
**So that** I can see my saved search configurations

**Acceptance Criteria:**

- `linkdingctl bundles list` displays all bundles in table format
- `linkdingctl bundles list --json` outputs valid JSON array
- Shows ID, Name, Search, and Order columns in table view
- Handles pagination transparently (fetches all pages)

### US-2: Get Bundle Details

**As a** linkdingctl user
**I want to** view a specific bundle's full details
**So that** I can see all filter settings

**Acceptance Criteria:**

- `linkdingctl bundles get <id>` displays full bundle info
- Shows all fields: name, search, any_tags, all_tags, excluded_tags, order, dates
- Returns clear error if bundle ID not found
- Supports `--json` output format

### US-3: Create Bundle

**As a** linkdingctl user
**I want to** create a new bundle from the command line
**So that** I can define saved searches programmatically

**Acceptance Criteria:**

- `linkdingctl bundles create <name>` creates bundle with just name
- Supports `--search`, `--any-tags`, `--all-tags`, `--excluded-tags`, `--order` flags
- Auto-assigns order if not specified
- Returns created bundle ID on success
- Validates name is not empty

### US-4: Update Bundle

**As a** linkdingctl user
**I want to** modify an existing bundle
**So that** I can adjust my saved searches

**Acceptance Criteria:**

- `linkdingctl bundles update <id>` with flags updates only specified fields
- Supports all create flags plus `--name` for renaming
- Uses PATCH semantics (only sends provided fields)
- Returns clear error if bundle ID not found

### US-5: Delete Bundle

**As a** linkdingctl user
**I want to** delete a bundle
**So that** I can remove unwanted saved searches

**Acceptance Criteria:**

- `linkdingctl bundles delete <id>` removes the bundle
- Returns success message with deleted ID
- Returns clear error if bundle ID not found

## Functional Requirements

### REQ-001: Bundle Model (Must Have)

Create Bundle data structures in `internal/models/`:

- `Bundle` struct with all API fields (id, name, search, any_tags, all_tags, excluded_tags, order, date_created, date_modified)
- `BundleCreate` struct for POST requests
- `BundleUpdate` struct with pointer fields for PATCH semantics
- `BundleList` struct for paginated responses

**Implementation hint:** Follow exact pattern from `BookmarkUpdate` for pointer fields.

### REQ-002: Bundle API Client Methods (Must Have)

Add to `internal/api/client.go`:

- `GetBundles(limit, offset int) (*models.BundleList, error)`
- `FetchAllBundles() ([]models.Bundle, error)` - handles pagination
- `GetBundle(id int) (*models.Bundle, error)`
- `CreateBundle(bundle *models.BundleCreate) (*models.Bundle, error)`
- `UpdateBundle(id int, update *models.BundleUpdate) (*models.Bundle, error)`
- `DeleteBundle(id int) error`

**Implementation hint:** Follow existing `GetTags`/`GetBookmarks` pattern for pagination.

### REQ-003: Bundles List Command (Must Have)

`linkdingctl bundles list`:

- Fetches all bundles via `FetchAllBundles()`
- Displays table: ID, Name, Search, Order
- Supports `--json` flag for JSON array output

### REQ-004: Bundles Get Command (Must Have)

`linkdingctl bundles get <id>`:

- Retrieves single bundle via `GetBundle(id)`
- Displays all fields in human-readable format
- Supports `--json` flag

### REQ-005: Bundles Create Command (Must Have)

`linkdingctl bundles create <name> [flags]`:

- Name is required positional argument
- Flags: `--search`, `--any-tags`, `--all-tags`, `--excluded-tags`, `--order`
- Creates via `CreateBundle()`
- Shows created bundle ID

### REQ-006: Bundles Update Command (Must Have)

`linkdingctl bundles update <id> [flags]`:

- ID is required positional argument
- Flags: `--name`, `--search`, `--any-tags`, `--all-tags`, `--excluded-tags`, `--order`
- Updates via `UpdateBundle()` with only provided fields
- Shows updated bundle info

### REQ-007: Bundles Delete Command (Must Have)

`linkdingctl bundles delete <id>`:

- ID is required positional argument
- Deletes via `DeleteBundle()`
- Shows success message

### REQ-008: Global Flag Support (Must Have)

All bundle commands must support:

- `--json` global flag for JSON output
- Per-command connection overrides (`--url`, `--token`)
- Config file and environment variable precedence

### REQ-009: Error Handling (Must Have)

- "Bundle not found" error for invalid IDs (HTTP 404)
- "Validation error" for invalid input (HTTP 400)
- Authentication errors handled consistently

## Non-Functional Requirements

### Performance

- Pagination handled transparently; list command fetches all pages
- API timeout: 30 seconds (existing client default)
- No local caching (LinkDing is source of truth)

### Security

- Token passed via `Authorization: Token <token>` header
- Config file permissions: 0600
- No secrets logged or displayed

### Reliability

- Exit code 0 on success, 1 on error, 2 on config error
- All API errors converted to user-friendly messages

### Compatibility

- Go 1.21+ (existing requirement)
- Works with LinkDing API that supports bundles endpoint

### Testability

- Unit tests with `httptest.NewServer` mocks
- 70% code coverage minimum (enforced by Makefile)

## Technical Considerations

### Architecture

```
internal/models/bundle.go     # New file: Bundle, BundleCreate, BundleUpdate, BundleList
internal/api/client.go        # Add: Bundle CRUD methods
cmd/linkdingctl/bundles.go    # New file: bundles command + subcommands
cmd/linkdingctl/commands_test.go  # Add: Bundle command tests
```

### API Specification

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/bundles/` | List (paginated: limit, offset) |
| GET | `/api/bundles/<id>/` | Retrieve single bundle |
| POST | `/api/bundles/` | Create bundle |
| PATCH | `/api/bundles/<id>/` | Partial update |
| DELETE | `/api/bundles/<id>/` | Delete bundle |

### Request/Response Examples

**Create Bundle Request:**

```json
{
  "name": "Dev Resources",
  "search": "golang tutorials",
  "any_tags": "golang rust",
  "all_tags": "",
  "excluded_tags": "outdated",
  "order": 5
}
```

**Bundle Response:**

```json
{
  "id": 3,
  "name": "Dev Resources",
  "search": "golang tutorials",
  "any_tags": "golang rust",
  "all_tags": "",
  "excluded_tags": "outdated",
  "order": 5,
  "date_created": "2025-01-24T10:00:00.000000Z",
  "date_modified": "2025-01-24T10:00:00.000000Z"
}
```

### Dependencies

- No new dependencies required
- Uses existing: Cobra, Viper, stdlib net/http

### Testing Strategy

- Unit tests for all API client methods
- Integration-style tests for CLI commands using mock HTTP servers
- Test error cases: not found, bad request, unauthorized

## Implementation Roadmap

### Phase 1: Models & API Client

1. Create `internal/models/bundle.go` with all structs
2. Add bundle methods to `internal/api/client.go`
3. Add unit tests for API client methods

### Phase 2: CLI Commands

4. Create `cmd/linkdingctl/bundles.go` with parent command
5. Implement list subcommand
6. Implement get subcommand
7. Implement create subcommand
8. Implement update subcommand
9. Implement delete subcommand
10. Add CLI command tests

### Phase 3: Integration & Polish

11. Verify all commands work with real LinkDing instance
12. Ensure coverage meets 70% threshold
13. Update any documentation if needed

## Out of Scope

- Batch operations on bundles (create/update multiple at once)
- Bundle import/export functionality
- Bundle reordering command (update --order handles this)
- Interactive bundle creation wizard
- Bundle-based bookmark listing (use `list --tags` instead)

## Open Questions & Risks

### Open Questions

1. Should we add a `bundles bookmarks <id>` command to list bookmarks matching a bundle's filters? (Deferred to future)

### Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| LinkDing version without bundles API | Commands fail | Document minimum LinkDing version |
| API changes | Breaking changes | Follow existing error handling patterns |

## Validation Checkpoints

- [ ] All bundle CRUD operations work against mock server
- [ ] All bundle CRUD operations work against real LinkDing
- [ ] JSON output is valid and matches API response format
- [ ] Table output is readable and consistent with other commands
- [ ] Error messages are user-friendly
- [ ] Code coverage >= 70%
- [ ] `make check` passes
