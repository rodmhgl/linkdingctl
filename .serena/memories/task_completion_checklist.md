# Task Completion Workflow

## Before Marking Task Complete

1. **Code compiles**: `go build ./...`
2. **Tests pass**: `go test ./...`
3. **No vet warnings**: `go vet ./...`
4. **Acceptance criteria met**: Check spec file for the feature
5. **Changes committed**: `git add -A && git commit -m "feat: description"`

## Quality Gates

### Must Pass
- `go build ./...` - No compilation errors
- `go test ./...` - All tests pass
- `go vet ./...` - No static analysis warnings

### Code Quality
- All exported functions have doc comments
- Error messages are user-friendly
- No panics (return errors instead)
- Follow Go idioms

## Integration Tests
- Mark with `//go:build integration` tag
- Require real LinkDing instance
- Not run in regular test suite

## Session Completion (Landing the Plane)

When ending a work session, MUST complete ALL steps:

1. File issues for remaining work
2. Run quality gates (if code changed)
3. Update issue status
4. **PUSH TO REMOTE** (MANDATORY):
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. Clean up stashes/branches
6. Verify all changes committed AND pushed
7. Provide context for next session

**CRITICAL**: Work is NOT complete until `git push` succeeds.
