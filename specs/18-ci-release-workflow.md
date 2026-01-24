# Specification: GitHub Actions CI/Release Workflow

## Jobs to Be Done
- On every push to `main`, the project is linted, tested, and built to catch regressions immediately
- Semantic versioning is applied automatically based on commit messages (Conventional Commits)
- A GitHub Release is created with the new version tag when commits warrant a version bump
- Pre-built binaries for all supported platforms are attached to the release for easy installation
- Users can download the correct binary for their OS/architecture without needing a Go toolchain

## Trigger

```yaml
on:
  push:
    branches:
      - main
```

The workflow runs on every merge/push to `main`. Feature branches are not included — PRs should use a separate CI workflow if desired in the future.

## Workflow Jobs

### Job 1: `lint-test`

Runs linting, vetting, and tests to gate the release.

```yaml
lint-test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version-file: go.mod

    - name: Download dependencies
      run: go mod download

    - name: Run go vet
      run: go vet ./...

    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v6
      with:
        version: latest

    - name: Run tests
      run: go test -v -race ./...

    - name: Build
      run: go build -trimpath -ldflags "-s -w" -o linkdingctl ./cmd/linkdingctl
```

Key points:
- `go-version-file: go.mod` ensures the workflow uses the same Go version as the project
- `-race` flag enables the race detector during tests
- The build step validates that the binary compiles successfully

### Job 2: `release`

Creates a semantic version release using Conventional Commits analysis.

```yaml
release:
  needs: lint-test
  runs-on: ubuntu-latest
  outputs:
    version: ${{ steps.semantic.outputs.version }}
  permissions:
    contents: write
  steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - name: Semantic Release
      id: semantic
      uses: go-semantic-release/action@v1
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
```

Key points:
- `fetch-depth: 0` is required for semantic-release to analyze the full commit history
- `go-semantic-release/action` determines the version bump from Conventional Commit prefixes:
  - `feat:` → minor bump (e.g., 1.0.0 → 1.1.0)
  - `fix:` → patch bump (e.g., 1.0.0 → 1.0.1)
  - `feat!:` or `BREAKING CHANGE:` → major bump (e.g., 1.0.0 → 2.0.0)
- The `version` output is empty if no release-worthy commits are found (no bump needed)
- `permissions: contents: write` allows the action to create tags and releases

### Job 3: `build-binaries`

Cross-compiles the CLI for all target platforms and uploads binaries to the GitHub Release.

```yaml
build-binaries:
  if: needs.release.outputs.version != ''
  needs: release
  runs-on: ubuntu-latest
  permissions:
    contents: write
  strategy:
    matrix:
      include:
        - goos: linux
          goarch: amd64
        - goos: linux
          goarch: arm64
        - goos: darwin
          goarch: amd64
        - goos: darwin
          goarch: arm64
        - goos: windows
          goarch: amd64
  steps:
    - uses: actions/checkout@v4
      with:
        ref: v${{ needs.release.outputs.version }}

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version-file: go.mod

    - name: Build binary
      env:
        GOOS: ${{ matrix.goos }}
        GOARCH: ${{ matrix.goarch }}
        VERSION: ${{ needs.release.outputs.version }}
      run: |
        EXT=""
        if [ "$GOOS" = "windows" ]; then EXT=".exe"; fi
        OUTNAME="linkdingctl-${GOOS}-${GOARCH}${EXT}"
        go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" \
          -o "${OUTNAME}" ./cmd/linkdingctl

    - name: Upload release asset
      env:
        GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        VERSION: ${{ needs.release.outputs.version }}
        GOOS: ${{ matrix.goos }}
        GOARCH: ${{ matrix.goarch }}
      run: |
        EXT=""
        if [ "$GOOS" = "windows" ]; then EXT=".exe"; fi
        OUTNAME="linkdingctl-${GOOS}-${GOARCH}${EXT}"
        gh release upload "v${VERSION}" "${OUTNAME}" --clobber
```

Key points:
- Only runs when `release` produces a version (i.e., a new release was created)
- Checks out the tagged commit (`ref: v<version>`) to ensure binary matches the release
- Builds for 5 platform/architecture combinations covering the most common targets
- Uses `-X main.version=${VERSION}` to embed the version string in the binary (requires a `version` variable in `main.go`)
- Uses `gh release upload` to attach each binary to the existing release
- `--clobber` allows re-running the workflow without failing on existing assets

## Platform Matrix

| OS      | Architecture | Binary Name                     |
|---------|-------------|---------------------------------|
| Linux   | amd64       | `linkdingctl-linux-amd64`       |
| Linux   | arm64       | `linkdingctl-linux-arm64`       |
| macOS   | amd64       | `linkdingctl-darwin-amd64`      |
| macOS   | arm64       | `linkdingctl-darwin-arm64`      |
| Windows | amd64       | `linkdingctl-windows-amd64.exe` |

## Semantic Versioning Convention

This workflow uses [Conventional Commits](https://www.conventionalcommits.org/) to determine version bumps automatically:

| Commit prefix        | Version bump | Example              |
|---------------------|-------------|----------------------|
| `fix:`              | Patch       | 1.2.3 → 1.2.4       |
| `perf:`             | Patch       | 1.2.3 → 1.2.4       |
| `feat:`             | Minor       | 1.2.3 → 1.3.0       |
| `feat!:`            | Major       | 1.2.3 → 2.0.0       |
| `BREAKING CHANGE:`  | Major       | 1.2.3 → 2.0.0       |
| `chore:`, `docs:`, `ci:`, `test:` | No release | — |

## Source Code Change: Version Variable

To support the `-X main.version` linker flag, add a version variable to `cmd/linkdingctl/main.go`:

```go
var version = "dev"
```

This variable can then be used in a `--version` flag or `version` subcommand (out of scope for this spec, but the infrastructure is in place).

## File to Create

```
.github/workflows/release.yaml
```

## Dependencies

- Repository must have `GITHUB_TOKEN` available (automatic for GitHub Actions)
- Commit messages should follow Conventional Commits for semantic-release to detect version bumps
- No additional secrets or external services required

## Success Criteria
- [ ] `.github/workflows/release.yaml` exists and is valid YAML
- [ ] Workflow triggers on push to `main` only
- [ ] `lint-test` job runs `go vet`, `golangci-lint`, tests with `-race`, and a build
- [ ] `lint-test` uses `go-version-file: go.mod` to match project Go version
- [ ] `release` job only runs after `lint-test` passes
- [ ] `release` job uses `go-semantic-release/action@v1` with full commit history (`fetch-depth: 0`)
- [ ] `release` job outputs the new version (or empty if no bump)
- [ ] `build-binaries` job only runs when a new version is released
- [ ] `build-binaries` checks out the tagged commit, not HEAD
- [ ] Binaries are built for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- [ ] Binary names follow the pattern `linkdingctl-<os>-<arch>[.exe]`
- [ ] Version is embedded in binaries via `-X main.version`
- [ ] Binaries are attached to the GitHub Release via `gh release upload`
- [ ] Windows binary includes `.exe` extension
- [ ] `cmd/linkdingctl/main.go` has a `version` variable defaulting to `"dev"`
- [ ] `make check` passes locally before merging
