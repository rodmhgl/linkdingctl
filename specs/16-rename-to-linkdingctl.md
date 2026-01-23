# Specification: Rename Binary from `linkdingctl` to `linkdingctl`

**Priority: HIGHEST** — This rename must be completed before all other pending work. The current binary name `linkdingctl` conflicts with the system linker (`/usr/bin/linkdingctl`) on Linux, causing PATH ambiguity and potential user confusion.

## Jobs to Be Done
- User can install and invoke the CLI without conflicting with the GNU linker (`linkdingctl`)
- User's existing config at `~/.config/linkdingctl/` is migrated transparently on first run
- All documentation, specs, and build tooling reflect the new name
- The binary follows the `<service>ctl` naming convention common to CLI management tools

## Rename Scope

### Binary & Build
| Item | Old | New |
|------|-----|-----|
| Binary name | `linkdingctl` | `linkdingctl` |
| Build output | `./linkdingctl` | `./linkdingctl` |
| Makefile `BINARY_NAME` | `linkdingctl` | `linkdingctl` |
| Makefile build path | `./cmd/linkdingctl` | `./cmd/linkdingctl` |
| .gitignore entry | `/linkdingctl` | `/linkdingctl` |

### Source Directory
| Item | Old | New |
|------|-----|-----|
| Command package | `cmd/linkdingctl/` | `cmd/linkdingctl/` |
| All Go files within | `cmd/linkdingctl/*.go` | `cmd/linkdingctl/*.go` |

### Config Paths
| Item | Old | New |
|------|-----|-----|
| Config directory | `~/.config/linkdingctl/` | `~/.config/linkdingctl/` |
| Config file | `~/.config/linkdingctl/config.yaml` | `~/.config/linkdingctl/config.yaml` |

### Cobra Root Command
| Item | Old | New |
|------|-----|-----|
| `Use` field | `"linkdingctl"` | `"linkdingctl"` |
| Help text / Long description | References to `linkdingctl` | References to `linkdingctl` |
| Error messages | `"Run 'linkdingctl config init'"` | `"Run 'linkdingctl config init'"` |

## Config Migration

On startup, if no config exists at the new path (`~/.config/linkdingctl/config.yaml`) but one exists at the old path (`~/.config/linkdingctl/config.yaml`):

1. Create `~/.config/linkdingctl/` with `0700` permissions
2. Copy `~/.config/linkdingctl/config.yaml` to `~/.config/linkdingctl/config.yaml` with `0600` permissions
3. Print notice to stderr: `Migrated config from ~/.config/linkdingctl/ to ~/.config/linkdingctl/`
4. Do **not** delete the old config (user may have other tools or scripts referencing it)

This migration runs once — subsequent invocations find the new config and skip migration.

## Files Requiring Changes

### Source Code
- `cmd/linkdingctl/root.go` → `cmd/linkdingctl/root.go`: `Use` field, Long description, help text
- `cmd/linkdingctl/config.go` → `cmd/linkdingctl/config.go`: any hardcoded references
- `internal/config/config.go`: `DefaultConfigPath()` returns `~/.config/linkdingctl/config.yaml`, error messages reference `linkdingctl`
- `internal/config/config_test.go`: updated expected error strings

### Build & Project
- `Makefile`: `BINARY_NAME=linkdingctl`, build path `./cmd/linkdingctl`
- `.gitignore`: `/linkdingctl`

### Documentation
- `README.md`: all command examples and paths
- `CLAUDE.md`: architecture paths, build commands, test commands
- `AGENTS.md`: binary name and config path references

### Specifications (all files in `specs/`)
- `01-core-cli.md` through `15-notes-field.md`: command examples and file paths
- `specs/archive/*.md`: historical references

## Examples (Post-Rename)

```bash
# Build
make build        # produces ./linkdingctl

# Install
go build -trimpath -ldflags "-s -w" -o linkdingctl ./cmd/linkdingctl
sudo mv linkdingctl /usr/local/bin/

# Usage
linkdingctl config init
linkdingctl add https://example.com --tags "reference"
linkdingctl list --query "kubernetes"
linkdingctl tags --sort count
linkdingctl export -f json -o backup.json
```

## Success Criteria
- [ ] `make build` produces a binary named `linkdingctl`
- [ ] `linkdingctl --help` displays correct binary name throughout
- [ ] `linkdingctl config init` creates config at `~/.config/linkdingctl/config.yaml`
- [ ] Existing config at `~/.config/linkdingctl/config.yaml` is auto-migrated on first run
- [ ] Migration prints a notice to stderr (not stdout, to avoid breaking piped output)
- [ ] Migration does not delete the old config directory
- [ ] `go test ./...` passes with all references updated
- [ ] `make cover` passes the 70% threshold
- [ ] No remaining references to `cmd/linkdingctl` or binary name `linkdingctl` in source code
- [ ] All spec files updated to reference `linkdingctl`
- [ ] All documentation updated to reference `linkdingctl`
- [ ] `/usr/bin/linkdingctl` (system linker) is no longer shadowed when `linkdingctl` is on PATH
