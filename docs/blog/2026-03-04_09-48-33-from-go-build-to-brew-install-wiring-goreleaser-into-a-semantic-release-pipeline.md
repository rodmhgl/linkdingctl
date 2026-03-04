# From `go build` to `brew install`: Wiring GoReleaser into a Semantic Release Pipeline

**Date:** March 4, 2026
**Author:** Rod Stewart (with Claude Opus 4.6)
**Project:** linkdingctl — CLI for managing LinkDing bookmarks
**Task:** Replace manual cross-compilation with GoReleaser, add Homebrew tap, shell completions, and version command

## The Problem

linkdingctl was feature-complete — 23 specs implemented, 70-86% test coverage, CI/CD running. But the distribution story was primitive: a GitHub Actions matrix job that cross-compiled 5 binaries and uploaded them individually with `gh release upload`. No Homebrew tap, no shell completions, no version command, and the README had placeholder URLs from day one.

The goal: replace all of that with a proper release pipeline that takes a conventional commit and turns it into a `brew install`.

## The Architecture: Two Tools, One Pipeline

The key insight is that **semantic release and goreleaser solve different problems**, and they compose beautifully when you understand the handoff point.

### Semantic Release: The Decision Maker

`go-semantic-release` answers exactly one question: *"Does this push deserve a new version?"*

It reads commit history since the last tag and applies Conventional Commits rules:

| Prefix | Bump | Example |
| --- | --- | --- |
| `fix:` | Patch (1.3.0 → 1.3.1) | `fix: handle nil tags in export` |
| `feat:` | Minor (1.3.0 → 1.4.0) | `feat: add bundle support` |
| `feat!:` | Major (1.3.0 → 2.0.0) | `feat!: change config format` |
| `chore:` | No release | `chore: update README` |

When it decides yes, it creates the git tag, creates a GitHub Release with auto-generated changelog, and outputs the version string as a GitHub Actions variable. That output is the contract between the two tools.

### GoReleaser: The Builder

GoReleaser picks up where semantic release left off. It doesn't decide *if* there's a release — it builds *what* gets released.

In a single job, it:

1. Runs before-hooks (dependency tidying, completion generation)
2. Cross-compiles 5 binaries (linux/darwin amd64+arm64, windows/amd64)
3. Packages them into archives with LICENSE, README, and shell completions
4. Generates SHA256 checksums
5. Uploads everything to the existing GitHub Release
6. Pushes an auto-generated Homebrew formula to a separate tap repository

This replaced a 5-runner matrix job with a single job that does more work.

## The Critical Integration Detail: `release.mode: append`

The most important line in the entire `.goreleaser.yaml`:

```yaml
release:
  mode: append
```

Without this, goreleaser tries to *create* a GitHub Release — but semantic release already created one. The job would fail with a "release already exists" error.

With `mode: append`, goreleaser adds its artifacts (binaries, archives, checksums) to the existing release. The two tools share the same GitHub Release without stepping on each other.

Similarly, `--skip=validate` in the goreleaser args is necessary because goreleaser's validation checks that the current git tag matches expectations. Since semantic release created the tag in a previous job, goreleaser's tag detection can get confused. Passing the tag explicitly via `GORELEASER_CURRENT_TAG` and skipping validation solves this cleanly.

## Shell Completions: The Before-Hook Chicken-and-Egg

Cobra auto-generates shell completions, but it needs a compiled binary to do so. GoReleaser's `before.hooks` run *before* any builds happen. So how do you generate completions if the binary doesn't exist yet?

The solution is a small script that builds a throwaway binary:

```bash
#!/bin/sh
set -e
rm -rf completions
mkdir -p completions
go build -o /tmp/linkdingctl-completions ./cmd/linkdingctl
/tmp/linkdingctl-completions completion bash > completions/linkdingctl.bash
/tmp/linkdingctl-completions completion zsh  > completions/linkdingctl.zsh
/tmp/linkdingctl-completions completion fish > completions/linkdingctl.fish
rm -f /tmp/linkdingctl-completions
```

This temporary binary has no version metadata injected (it'll say "dev"), but that doesn't matter — completions are about command structure, not version strings. The real release binaries get built afterward with proper ldflags.

## Version Injection: Three Variables, One Pattern

Go's standard pattern for build-time metadata injection uses `ldflags -X`:

```go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

GoReleaser populates these with template variables:

```yaml
ldflags:
  - -s -w
  - -X main.version={{.Version}}
  - -X main.commit={{.Commit}}
  - -X main.date={{.CommitDate}}
```

Local `go build` produces `linkdingctl dev` with commit `none`. Release builds produce `linkdingctl 1.3.0` with the real commit SHA and ISO 8601 timestamp. No code generation, no build scripts — just compiler flags.

The `mod_timestamp: "{{ .CommitTimestamp }}"` setting is a nice bonus: it sets the binary's modification time to the commit time rather than build time, making builds reproducible.

## The Homebrew Formula: Two Repos, Two Tokens

GoReleaser's `brews` section auto-generates a Ruby formula and pushes it to a separate repository (`rodmhgl/homebrew-tap`). This requires a dedicated Personal Access Token because `GITHUB_TOKEN` (the automatic Actions token) only has permissions on the current repository.

```yaml
brews:
  - repository:
      owner: rodmhgl
      name: homebrew-tap
      token: "{{ .Env.TAP_GITHUB_TOKEN }}"
```

The formula includes shell completion installation:

```ruby
install: |
  bin.install "linkdingctl"
  bash_completion.install "completions/linkdingctl.bash" => "linkdingctl"
  zsh_completion.install "completions/linkdingctl.zsh" => "_linkdingctl"
  fish_completion.install "completions/linkdingctl.fish"
```

After the first release, users get the full experience:

```bash
brew tap rodmhgl/tap
brew install linkdingctl
linkdingctl <tab>  # completions work immediately
```

## The Full Pipeline Timeline

```text
git push (feat: add distribution pipeline)
  |
  +-- lint-test: vet, lint, test, build
  |
  +-- semantic-release: reads commits, decides "1.3.0"
  |     +-- creates tag v1.3.0
  |     +-- creates GitHub Release with changelog
  |     +-- outputs version="1.3.0"
  |
  +-- goreleaser: (only runs if version was output)
        +-- go mod tidy
        +-- build throwaway binary, generate completions
        +-- cross-compile 5 release binaries with ldflags
        +-- package into tar.gz (unix) and zip (windows)
        +-- generate SHA256 checksums
        +-- upload all artifacts to existing GitHub Release
        +-- push Homebrew formula to homebrew-tap repo
```

Three jobs, fully sequential, each with a clear responsibility. The commit message is the only input.

## What Was Replaced

The old `build-binaries` job:

- 5 parallel matrix runners
- Each one: checkout, setup Go, build one binary, upload one asset
- No checksums, no archives, no completions, no Homebrew
- 35 lines of shell script with `EXT=""` conditionals

The new `goreleaser` job:

- 1 runner
- Does everything the 5 runners did, plus archives, checksums, completions, and Homebrew
- Declarative YAML config instead of imperative shell scripts

## Lessons Learned

### 1. Composition Over All-in-One

We could have used goreleaser for everything — it can create tags and releases too. But keeping semantic release as the decision-maker and goreleaser as the builder gives cleaner separation of concerns. Each tool does what it's best at.

### 2. `mode: append` is the Glue

When combining tools that both touch GitHub Releases, `mode: append` is non-negotiable. Without it, the second tool fails because the release already exists. This isn't well-documented in either tool's "getting started" guides — you discover it when the pipeline breaks.

### 3. Before-Hooks Need a Throwaway Build

If your release artifacts include generated files that depend on the binary itself (like shell completions), you need a two-phase approach: build a throwaway copy in before-hooks, generate the artifacts, then let goreleaser build the real release binaries. The generated files don't contain version-specific data, so this works cleanly.

### 4. Test the Pipeline, Not Just the Config

`goreleaser check` validates YAML syntax but doesn't catch integration issues like the `mode: append` problem. The real test is a full release cycle — push a `feat:` commit and watch the entire pipeline from semantic release through Homebrew formula push.

### 5. Conventional Commits Pay Off at Release Time

Conventional Commits felt like overhead when we adopted them. But now they're the *entire release trigger*. Write `feat:`, get a minor version bump, binaries built, Homebrew updated. Write `fix:`, get a patch. Write `chore:`, get nothing — exactly right for housekeeping commits. The discipline at commit time eliminates manual release management entirely.

## Metrics

- **Files changed:** 9 (3 created, 6 modified)
- **Lines of Go code added:** 54 (version command)
- **Test coverage maintained:** 70.5% - 86.1% across all packages
- **Pipeline runners reduced:** 5 matrix runners → 1 goreleaser runner
- **New capabilities:** version command, shell completions, Homebrew tap, checksums, proper archives
- **Time from push to `brew install`:** ~4 minutes (lint + semantic release + goreleaser)

## References

- [GoReleaser v2 Documentation](https://goreleaser.com/customization/)
- [go-semantic-release](https://github.com/go-semantic-release/semantic-release)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Cobra Shell Completions](https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md)
