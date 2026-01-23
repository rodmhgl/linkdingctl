# Specification: Makefile Coverage Portability

## Jobs to Be Done
- `make cover` validates coverage for ALL packages with test files, including `cmd/ld`
- Coverage validation runs reliably on minimal environments (CI, Docker) without `bc`

## Problem 1: cmd/ Packages Skipped

The `cover` target skips all packages matching `cmd/` from coverage threshold checks. This made sense before command tests existed, but `cmd/ld/commands_test.go` now provides coverage for the `cmd/ld` package. The skip condition silently exempts `cmd/ld` from the 70% threshold, defeating the purpose of adding those tests.

Relevant Makefile logic:

```makefile
if echo "$$pkg" | grep -q "cmd/"; then \
    echo "SKIP: $$pkg (no test files)"; \
```

### Implementation

Remove the `cmd/` skip condition entirely. All packages reported by `go test -cover` already have test files (packages without test files report `[no test files]` and do not emit a coverage line).

## Problem 2: `bc` Dependency

The coverage comparison uses `bc -l` for floating-point comparison:

```makefile
result=$$(echo "$$cov < 70" | bc -l); \
```

`bc` is not available on all systems (minimal Docker images, Alpine, some CI runners). `awk` is universally available and handles float comparison natively.

### Implementation

Replace the `bc` comparison with `awk`:

```makefile
result=$$(echo "$$cov 70" | awk '{print ($$1 < $$2)}'); \
```

Affected files:
- `Makefile` — `cover` target

## Success Criteria
- [ ] `make cover` enforces the 70% threshold on `cmd/ld` package
- [ ] `make cover` does not skip any package that has test files
- [ ] `make cover` does not require `bc` to be installed
- [ ] Coverage comparison uses `awk` for portable float comparison
- [ ] No functional change to pass/fail behavior for non-cmd packages