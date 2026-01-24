# Git Hook Validation Guide

This document describes how to validate that the lefthook-based pre-commit hooks are working correctly.

## Prerequisites

1. Install lefthook:
   ```bash
   # macOS
   brew install lefthook

   # Linux/Windows
   # Download from https://github.com/evilmartians/lefthook/releases
   ```

2. Install the hooks:
   ```bash
   make hooks
   # or directly: lefthook install
   ```

## Validation Tests

### Test 1: Verify go vet catches violations

1. Introduce a go vet violation:
   ```bash
   # Add invalid code to a test file
   echo 'package main; func test() { x := 1 }' > /tmp/test_violation.go
   cp /tmp/test_violation.go cmd/linkdingctl/test_violation.go
   git add cmd/linkdingctl/test_violation.go
   ```

2. Attempt to commit:
   ```bash
   git commit -m "test: verify go vet hook"
   ```

3. Expected result:
   - Commit should be **blocked**
   - Error message should indicate go vet failure
   - Output should show the unused variable violation

4. Clean up:
   ```bash
   git reset HEAD cmd/linkdingctl/test_violation.go
   rm cmd/linkdingctl/test_violation.go
   ```

### Test 2: Verify golangci-lint runs (if installed)

1. Check if golangci-lint is installed:
   ```bash
   command -v golangci-lint
   ```

2. If installed, the hook will run `golangci-lint run`
3. If not installed, the hook will print a warning and exit 0 (non-blocking)

### Test 3: Verify beads hook runs (if bd installed)

1. Check if bd is installed:
   ```bash
   command -v bd
   ```

2. If installed, `bd hook pre-commit` will be called
3. If not installed, the hook will be skipped silently

### Test 4: Verify clean code passes all hooks

1. Make a legitimate change:
   ```bash
   echo "# Test" >> README.md
   git add README.md
   ```

2. Commit:
   ```bash
   git commit -m "docs: test hooks"
   ```

3. Expected result:
   - Commit should **succeed**
   - All hooks should run without errors
   - Output should show successful execution of each hook

4. Clean up:
   ```bash
   git reset HEAD~1
   git restore README.md
   ```

### Test 5: Verify parallel execution

1. Make a change and commit:
   ```bash
   echo "# Test" >> README.md
   git add README.md
   git commit -m "test: parallel hooks"
   ```

2. Observe the output - all three hooks should execute concurrently
3. Total execution time should be close to the slowest hook, not the sum of all hooks

## Expected Hook Behavior

| Hook | Condition | Behavior |
|------|-----------|----------|
| `beads-sync` | bd not installed | Skipped (no output) |
| `beads-sync` | bd installed | Runs `bd hook pre-commit` |
| `go-vet` | Always | Runs `go vet ./...`, blocks on errors |
| `golangci-lint` | Tool not installed | Prints warning, exits 0 (non-blocking) |
| `golangci-lint` | Tool installed | Runs `golangci-lint run`, blocks on errors |

## Troubleshooting

### Hooks not running

```bash
# Check lefthook installation
lefthook version

# Reinstall hooks
lefthook uninstall
lefthook install
```

### Hooks running but failing

```bash
# Run hooks manually to see detailed output
lefthook run pre-commit
```

### Check hook configuration

```bash
# View installed hook
cat .git/hooks/pre-commit

# View lefthook configuration
cat lefthook.yml
```
