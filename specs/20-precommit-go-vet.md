# Specification: Pre-Commit Hook — go vet

## Jobs to Be Done
- Code that fails `go vet` is blocked from being committed, catching issues locally before CI
- The check mirrors the `go vet ./...` step in `.github/workflows/release.yaml` exactly
- Developers get immediate feedback on suspicious constructs (printf format mismatches, unreachable code, invalid struct tags, etc.) without waiting for a CI round-trip

## Context

The CI workflow runs `go vet ./...` as a quality gate before tests and release. When this step fails remotely, the feedback loop is slow: push → wait for CI → read logs → fix → push again. A pre-commit hook eliminates this by failing fast at commit time.

## Scope

Install a git pre-commit hook that runs:

```bash
go vet ./...
```

The hook must:
1. Run against the **entire module** (`./...`), not just staged files — this matches CI behavior and avoids partial-analysis false negatives (e.g., a staged file references a symbol in an unstaged file that would cause a vet failure)
2. Block the commit (exit non-zero) if `go vet` reports any diagnostics
3. Print the full `go vet` output so the developer can see exactly what failed
4. Coexist with the existing `bd` (beads) pre-commit shim in `.git/hooks/pre-commit`

## Hook Management Mechanism

Use [Lefthook](https://github.com/evilmartians/lefthook) as the hook runner:

- Written in Go, single binary, no runtime dependencies
- Configured via a committed `lefthook.yml` at the repo root
- Handles hook chaining — can invoke the existing `bd hook pre-commit` as part of the pre-commit sequence
- Developers install once (`go install github.com/evilmartians/lefthook@latest`) and run `lefthook install` to set up `.git/hooks`

### Configuration

Create `lefthook.yml` at the repo root with the `go vet` hook:

```yaml
pre-commit:
  commands:
    go-vet:
      run: go vet ./...
```

### Preserving the Existing `bd` Hook

The existing `.git/hooks/pre-commit` delegates to `bd hook pre-commit`. Lefthook replaces `.git/hooks/pre-commit` with its own runner. To preserve `bd` functionality, add it as a command in the pre-commit sequence:

```yaml
pre-commit:
  commands:
    beads-sync:
      run: bd hook pre-commit
      skip:
        - run: "! command -v bd >/dev/null 2>&1"
    go-vet:
      run: go vet ./...
```

The `skip` condition ensures the hook doesn't fail if `bd` is not installed (mirrors the existing shim's behavior).

## Developer Setup

After cloning or pulling this change, developers run:

```bash
lefthook install
```

This replaces `.git/hooks/pre-commit` with Lefthook's dispatcher. Subsequent `git commit` invocations trigger the configured hooks automatically.

Add a note to the project README or a `Makefile` target:

```makefile
hooks: ## Install git hooks via lefthook
	@which lefthook > /dev/null || (echo "lefthook not found. Install: go install github.com/evilmartians/lefthook@latest" && exit 1)
	lefthook install
```

## Affected Files

- `lefthook.yml` (new) — hook configuration
- `Makefile` — add `hooks` target for developer convenience
- `.git/hooks/pre-commit` — replaced by Lefthook's dispatcher on `lefthook install`

## Success Criteria
- [ ] `lefthook.yml` exists at the repo root and is valid YAML
- [ ] `lefthook.yml` defines a `pre-commit` → `go-vet` command that runs `go vet ./...`
- [ ] `lefthook.yml` chains the existing `bd hook pre-commit` with a skip condition for missing `bd`
- [ ] After `lefthook install`, running `git commit` on a clean working tree succeeds
- [ ] After `lefthook install`, running `git commit` on code with a `go vet` violation (e.g., `fmt.Printf("%d", "not-an-int")`) is blocked with diagnostic output
- [ ] Makefile has a `hooks` target that installs Lefthook hooks
- [ ] `go vet ./...` passes on the current codebase (no pre-existing violations)
