# Specification: Test Coverage

## Jobs to Be Done
- All command logic is tested without requiring a live LinkDing instance
- Import/export round-trips are verified for all formats
- API client pagination logic is tested
- Edge cases (empty results, errors, special characters) are covered

## Current State

| Package | Status |
|---------|--------|
| `internal/api` | Partial — CRUD tested, pagination untested |
| `internal/config` | Good — load, save, env override tested |
| `internal/export` | Partial — JSON/CSV structure tests only, no HTML tests |
| `internal/models` | No tests needed (pure data structs) |
| `cmd/ld` | **None** — all command logic untested |

## API Client Tests Needed

File: `internal/api/client_test.go`

- [ ] `GetBookmarks` pagination (multi-page fetch)
- [ ] `GetBookmarks` with archived filter
- [ ] `GetTags` with pagination
- [ ] `doRequest` timeout behavior
- [ ] Error response body parsing (400 with details)

## Export Tests Needed

File: `internal/export/html_test.go` (new)

- [ ] `ExportHTML` produces valid Netscape bookmark format
- [ ] HTML-escapes URLs, titles, and descriptions
- [ ] Includes TAGS attribute when tags are present
- [ ] Omits DD element when description is empty

File: `internal/export/import_test.go` (new)

- [ ] `importHTML` parses standard Netscape format
- [ ] `importHTML` handles bookmarks with and without descriptions
- [ ] `importHTML` extracts TAGS attribute
- [ ] `importJSON` round-trips with `ExportJSON` output
- [ ] `importCSV` round-trips with `ExportCSV` output
- [ ] `importCSV` handles missing columns gracefully
- [ ] `DetectFormat` returns correct format for all extensions
- [ ] `DetectFormat` returns empty string for unknown extensions
- [ ] Duplicate URL detection works (skip vs update)
- [ ] `--add-tags` appends to imported bookmarks

## Command Tests Needed

File: `cmd/ld/commands_test.go` (new)

Pattern: Use `httptest.NewServer` to mock the API, set config via env vars, execute commands via cobra's `Execute()`.

```go
func executeCommand(args ...string) (string, error) {
    buf := new(bytes.Buffer)
    rootCmd.SetOut(buf)
    rootCmd.SetErr(buf)
    rootCmd.SetArgs(args)
    err := rootCmd.Execute()
    return buf.String(), err
}
```

Priority tests:
- [ ] `add` creates bookmark and prints confirmation
- [ ] `list` renders table with correct columns
- [ ] `list --json` outputs valid JSON
- [ ] `get` displays full bookmark details
- [ ] `update --add-tags` merges tags correctly
- [ ] `delete --force` skips confirmation
- [ ] `export -f csv` writes valid CSV to stdout
- [ ] `backup` creates file with timestamp in name
- [ ] `tags` lists tags with counts
- [ ] `config show` redacts token correctly

## Integration Test Pattern

For commands that require stdin interaction (delete confirmation, config init):

```go
func TestDeleteWithConfirmation(t *testing.T) {
    // Provide "y\n" via stdin
    oldStdin := os.Stdin
    r, w, _ := os.Pipe()
    os.Stdin = r
    w.WriteString("y\n")
    w.Close()
    defer func() { os.Stdin = oldStdin }()

    // Execute and verify deletion occurred
}
```

## Success Criteria
- [ ] `go test ./...` covers all packages (no `[no test files]` for cmd/ld)
- [ ] HTML export/import has dedicated tests
- [ ] Round-trip tests verify export-then-import fidelity
- [ ] Pagination logic tested with multi-page mock responses
- [ ] Command tests don't require network access
- [ ] `go test -cover ./...` reports >70% coverage for each package