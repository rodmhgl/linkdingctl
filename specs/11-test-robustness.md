# Specification: Test Robustness

## Jobs to Be Done
- Config tests do not panic on unexpected error messages
- Command tests reset ALL global flags between runs to prevent test pollution

## Problem 1: Unsafe String Slicing in Config Test

`internal/config/config_test.go` — `TestLoad_NonYAMLFile` uses direct string slicing to compare error messages:

```go
expectedErrSubstring := "failed to read config file"
if err.Error()[:len(expectedErrSubstring)] != expectedErrSubstring {
    t.Errorf(...)
}
```

If the error message is shorter than `expectedErrSubstring`, this causes an index out-of-range panic, crashing the test suite. This is a latent bug that will trigger if the error format changes.

### Implementation

Replace the unsafe slice with `strings.HasPrefix` or `strings.Contains`:

```go
if !strings.Contains(err.Error(), expectedErrSubstring) {
    t.Errorf("expected error containing '%s', got '%v'", expectedErrSubstring, err)
}
```

Affected files:
- `internal/config/config_test.go` — `TestLoad_NonYAMLFile`

## Problem 2: Incomplete Flag Reset in Command Tests

`cmd/linkdingctl/commands_test.go` — the `executeCommand` helper resets global flags between test runs, but misses several flags:

```go
// Currently reset:
jsonOutput = false
debugMode = false
forceDelete = false
updateArchive = false
updateUnarchive = false
updateTags = nil
updateAddTags = nil
updateRemoveTags = nil
updateTitle = ""
updateDescription = ""
tagsSort = "name"
tagsUnused = false

// Missing resets:
// backupOutput
// backupPrefix
// tagsRenameForce
// tagsDeleteForce
```

This causes test pollution: a test that sets `tagsRenameForce = true` will leak that state into subsequent tests, causing incorrect pass/fail results depending on test execution order.

### Implementation

Add the missing flag resets to the `executeCommand` helper:

```go
// Reset ALL global command flags
backupOutput = ""
backupPrefix = ""
tagsRenameForce = false
tagsDeleteForce = false
```

Additionally, consider extracting the reset logic into a dedicated `resetGlobalFlags()` function to make it easier to audit completeness when new flags are added.

Affected files:
- `cmd/linkdingctl/commands_test.go` — `executeCommand` function

## Success Criteria
- [ ] `TestLoad_NonYAMLFile` uses `strings.Contains` or `strings.HasPrefix` (no direct slice)
- [ ] No test can panic due to error message length mismatch
- [ ] `executeCommand` resets `backupOutput`, `backupPrefix`, `tagsRenameForce`, `tagsDeleteForce`
- [ ] No test pollution from leaked global flag state
- [ ] All existing tests continue to pass
