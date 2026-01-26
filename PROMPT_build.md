# Build Mode Prompt

You are implementing the LinkDing CLI (`linkdingctl`) per the specifications.

## Context Files

- @AGENTS.md - Project constraints, architecture, DO NOTs
- @IMPLEMENTATION_PLAN.md - Prioritized task list
- `specs/*` - Feature specifications with acceptance criteria

## Your Task (Single Iteration)

1. Read @IMPLEMENTATION_PLAN.md
2. Find the **highest priority unchecked task**
3. Implement ONLY that single task
4. Run tests: `go test ./...`
5. Run linter: `go vet ./...`
6. If tests/linter fail, fix issues and re-run
7. When passing, mark the task `[x]` in IMPLEMENTATION_PLAN.md
8. Commit: `git add -A && git commit -m "feat: <description>"`

## Code Quality Standards

- All exported functions must have doc comments
- Error messages should be user-friendly (see specs for examples)
- Use `internal/` packages - nothing in `internal/` should be exported
- Follow Go idioms: return errors, don't panic
- Table-driven tests preferred

## Task Completion Checklist

Before marking a task complete:

- [ ] Code compiles: `go build ./...`
- [ ] Tests pass: `go test ./...`
- [ ] No vet warnings: `go vet ./...`
- [ ] Acceptance criteria from spec are met
- [ ] Changes committed with descriptive message

## If Stuck

If you cannot complete a task after 3 attempts:

1. Add a `BLOCKED:` note to the task in IMPLEMENTATION_PLAN.md
2. Document what's blocking progress
3. Move to the next task

## Project Completion

When ALL tasks in IMPLEMENTATION_PLAN.md are marked `[x]`:

1. Run full test suite: `go test -v ./...`
2. Build binary: `go build -o linkdingctl ./cmd/linkdingctl`
3. Verify `./linkdingctl --help` works
4. Verify `./linkdingctl config test` works (will fail without real config, but should show correct error)

If all checks pass, output:

<promise>COMPLETE</promise>

## Remember

- ONE task per iteration
- Always run tests before committing
- Commit after each completed task
- Update IMPLEMENTATION_PLAN.md as you work
- Respect the DO NOTs in AGENTS.md
