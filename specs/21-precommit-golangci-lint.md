# Specification: Pre-Commit Hook — golangci-lint

## Jobs to Be Done
- Code that fails `golangci-lint` is blocked from being committed, catching lint violations locally before CI
- The check mirrors the `golangci-lint-action@v6` step in `.github/workflows/release.yaml`
- Developers get immediate feedback on lint issues (errcheck, ineffassign, staticcheck, unused, etc.) without waiting for a CI round-trip

## Context

The CI workflow runs `golangci-lint` via `golangci/golangci-lint-action@v6` with `version: latest` as a quality gate. This action runs `golangci-lint run` against the full module. When violations are detected remotely, the feedback loop is slow. A pre-commit hook catches these issues at commit time.

The Makefile already has a `lint` target that checks for `golangci-lint` availability and runs it. This hook reuses the same tool invocation in the pre-commit context.

## Scope

Add a Lefthook pre-commit command that runs:

```bash
golangci-lint run
```

The hook must:
1. Run against the **entire module** (default `./...` scope) — this matches CI behavior where the action runs against the full checkout
2. Block the commit (exit non-zero) if any lint violations are reported
3. Print the full linter output so the developer can see exactly what failed and where
4. Skip gracefully if `golangci-lint` is not installed, with a warning message directing the developer to install it
5. Integrate into the existing `lefthook.yml` created by the `go vet` hook spec (spec 20)

## Prerequisite

This spec depends on spec `20-precommit-go-vet.md` which establishes the Lefthook infrastructure (`lefthook.yml`, Makefile `hooks` target, and `bd` hook chaining).

## Configuration

Add the `golangci-lint` command to the existing `lefthook.yml`:

```yaml
pre-commit:
  commands:
    beads-sync:
      run: bd hook pre-commit
      skip:
        - run: "! command -v bd >/dev/null 2>&1"
    go-vet:
      run: go vet ./...
    golangci-lint:
      run: golangci-lint run
```

### Handling Missing golangci-lint

Unlike `go vet` (which is always available with a Go installation), `golangci-lint` is an external tool that developers install separately. The hook should not hard-fail if the tool is missing — this would block all commits for developers who haven't installed it yet. Instead, use a wrapper that warns and exits 0:

```yaml
    golangci-lint:
      run: |
        if ! command -v golangci-lint >/dev/null 2>&1; then
          echo "Warning: golangci-lint not found, skipping lint check" >&2
          echo "  Install: https://golangci-lint.run/usage/install/" >&2
          exit 0
        fi
        golangci-lint run
```

This matches the pattern used by the Makefile's `lint` target.

## Execution Order

Lefthook runs commands in parallel by default. Since `go vet` is fast and `golangci-lint` is more comprehensive (and slower), parallel execution is acceptable — both operate read-only on the source. If sequential execution is preferred (to fail fast on cheaper checks), add:

```yaml
pre-commit:
  parallel: false
```

The recommended configuration keeps parallel execution enabled, since both checks are independent and the total wall-clock time is reduced.

## Performance Consideration

`golangci-lint run` on the full module can take several seconds on larger codebases. For this project's current size (~20 Go files), this is acceptable. If commit-time latency becomes an issue in the future, the command can be changed to `golangci-lint run --new-from-rev=HEAD` to only lint changed code — but this diverges from CI behavior and is out of scope for this spec.

## Affected Files

- `lefthook.yml` — add `golangci-lint` command to existing pre-commit section

## Success Criteria
- [ ] `lefthook.yml` defines a `pre-commit` → `golangci-lint` command
- [ ] The command runs `golangci-lint run` (full module scope, matching CI)
- [ ] If `golangci-lint` is not installed, the hook prints a warning and exits 0 (non-blocking)
- [ ] If `golangci-lint` is installed and finds violations, the commit is blocked with diagnostic output
- [ ] If `golangci-lint` is installed and finds no violations, the commit proceeds
- [ ] After `lefthook install`, running `git commit` on code with a known lint violation (e.g., unchecked error return) is blocked
- [ ] `golangci-lint run` passes on the current codebase (no pre-existing violations — confirmed by spec 19)
- [ ] The hook runs in parallel with `go vet` for faster feedback
