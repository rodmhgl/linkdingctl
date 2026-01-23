# Implementation Plan - LinkDing CLI

> Gap analysis based on specs 01-18 vs current codebase (2026-01-23).
> Specs 01-17 are fully implemented and passing. Spec 18 (CI/Release Workflow) is the only remaining gap.

## Current State Summary

| Spec | Status | Notes |
|------|--------|-------|
| 01 - Core CLI | Complete | Config init/show/test, env overrides, --json, --debug |
| 02 - Bookmark CRUD | Complete | add/list/get/update/delete with --json support |
| 03 - Tags | Complete | tags list/show/rename/delete |
| 04 - Import/Export | Complete | JSON/HTML/CSV export+import, backup, restore --wipe |
| 05 - Security | Complete | 0700/0600 perms, token masking, safe JSON output |
| 06 - Tags Performance | Complete | FetchAllBookmarks/FetchAllTags, client-side counting |
| 07 - Test Coverage | Complete | All packages >70% threshold |
| 08 - Tags Show Pagination | Complete | Uses FetchAllBookmarks |
| 09 - Makefile Portability | Complete | awk-based coverage comparison |
| 10 - Config Token Trim | Complete | Clean per-branch trim |
| 11 - Test Robustness | Complete | strings.Contains, full flag reset |
| 12 - User Profile | Complete | Superseded by spec 17 |
| 13 - Tags CRUD | Complete | tags create/get subcommands |
| 14 - Per-Command Overrides | Complete | --url and --token global flags |
| 15 - Notes Field | Complete | --notes on add/update commands |
| 16 - Rename | Complete | All references updated to linkdingctl |
| 17 - Fix User Profile | Complete | Correct API model, all tests passing |
| 18 - CI/Release Workflow | **NOT STARTED** | No .github/workflows/ directory, no version variable |

## Coverage Status

```
Package                    Current   Target   Status
cmd/linkdingctl            81.1%     70%      PASS
internal/api               79.5%     70%      PASS
internal/config            83.1%     70%      PASS
internal/export            78.4%     70%      PASS
```

---

## Remaining Tasks

### Phase 1: Source Code Prerequisite (spec 18)

- [x] **P0** | Add `version` variable to main.go | ~small
  - Acceptance: `cmd/linkdingctl/main.go` contains `var version = "dev"` at package level; build succeeds with and without `-X main.version=...` linker flag
  - Files: `cmd/linkdingctl/main.go`
  - Details: The CI workflow embeds the release version via `-X main.version=${VERSION}`. Without this variable, the linker flag would fail silently. The default value `"dev"` is used for local builds.

### Phase 2: GitHub Actions Workflow (spec 18)

- [x] **P1** | Create lint-test job | ~medium
  - Acceptance: Job runs on `push` to `main`; uses `actions/checkout@v4`; sets up Go via `actions/setup-go@v5` with `go-version-file: go.mod`; runs `go mod download`, `go vet ./...`, `golangci-lint` (via `golangci/golangci-lint-action@v6`), `go test -v -race ./...`, and `go build -trimpath -ldflags "-s -w" -o linkdingctl ./cmd/linkdingctl`
  - Files: `.github/workflows/release.yaml`

- [x] **P1** | Create release job | ~medium
  - Acceptance: Job depends on `lint-test`; uses `actions/checkout@v4` with `fetch-depth: 0`; uses `go-semantic-release/action@v1` with `github-token: ${{ secrets.GITHUB_TOKEN }}`; has `permissions: contents: write`; outputs `version` from the semantic release step; version output is empty when no release-worthy commits exist
  - Files: `.github/workflows/release.yaml`

- [ ] **P1** | Create build-binaries job | ~medium
  - Acceptance: Job depends on `release`; conditional on `needs.release.outputs.version != ''`; checks out tagged commit via `ref: v${{ needs.release.outputs.version }}`; uses matrix strategy for 5 targets (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64); builds with `-X main.version=${VERSION}` linker flag; binary names follow `linkdingctl-<os>-<arch>[.exe]` pattern; uploads via `gh release upload` with `--clobber`; has `permissions: contents: write`
  - Files: `.github/workflows/release.yaml`

### Phase 3: Verification

- [ ] **P2** | Validate workflow YAML syntax | ~small
  - Acceptance: `python -c "import yaml; yaml.safe_load(open('.github/workflows/release.yaml'))"` or equivalent linting passes; no YAML syntax errors; all action version references are valid (checkout@v4, setup-go@v5, golangci-lint-action@v6, go-semantic-release/action@v1)
  - Files: `.github/workflows/release.yaml`

- [ ] **P2** | Verify local build with version flag | ~small
  - Acceptance: `go build -ldflags "-X main.version=1.2.3" -o /dev/null ./cmd/linkdingctl` compiles without error; confirms the `version` variable is accessible to the linker
  - Files: (verification only)

- [ ] **P2** | Verify `make check` passes | ~small
  - Acceptance: `make check` (fmt + vet + test) passes with no failures; no regressions from the `main.go` change
  - Files: (verification only)

---

## Dependency Graph

```
Phase 1 (version variable) ─── must be first (binary build embeds version)
        │
        ▼
Phase 2 (workflow YAML) ─────── depends on Phase 1 (references main.version)
        │
        ▼
Phase 3 (verification) ──────── depends on Phase 2 (validates everything)
```

## Implementation Notes

- The entire remaining work is Spec 18. No other specs have outstanding gaps.
- The `version` variable change to `main.go` is trivial (1 line) but blocking for the CI workflow.
- The workflow file is self-contained — no changes to existing source code beyond the version variable.
- The workflow uses only GitHub's built-in `GITHUB_TOKEN` — no additional secrets required.
- Conventional Commits are already used in this repo (visible in git log: `feat:`, `fix:`, `chore:` prefixes).
- The `go-semantic-release/action@v1` action handles tagging and release creation automatically.
- Cross-compilation is trivial in Go — `GOOS`/`GOARCH` environment variables are all that's needed.
- The `.exe` extension for Windows is handled in the build script via conditional logic.
- No tests need to be written for the CI workflow itself — it's infrastructure, validated by GitHub Actions.
