# Specification: Security Hardening

## Jobs to Be Done

- User's API token is protected from other local users
- User's token is not visible on-screen during entry
- JSON output is safe from injection regardless of file paths

## Config File Permissions

The config file contains a plaintext API token. File and directory permissions must restrict access.

**Directory**: `~/.config/ld/` created with `0700` (owner-only)
**File**: `config.yaml` created with `0600` (owner read/write only)

Current behavior: directory uses `0755`, file permissions depend on umask.

### Permissions Implementation

`internal/config/config.go` — `Save()`:

- `os.MkdirAll(dir, 0700)` instead of `0755`
- After `v.WriteConfig()`, call `os.Chmod(configPath, 0600)`

## Token Input Masking

`ld config init` currently echoes the API token in plaintext during entry.

### Masking Implementation

Use `golang.org/x/term.ReadPassword(int(os.Stdin.Fd()))` for the token prompt:

```text
LinkDing URL: https://linkding.example.com
API Token: ********
```

Fallback: if stdin is not a terminal (piped input), read normally.

## Safe JSON Output in Backup Command

`cmd/ld/backup.go` constructs JSON via `fmt.Printf` with unescaped path:

```go
fmt.Printf("{\"file\": \"%s\"}\n", fullPath)  // BROKEN if path contains quotes
```

### JSON Output Implementation

Replace with `json.Marshal` or `json.NewEncoder`:

```go
output := map[string]string{"file": fullPath}
json.NewEncoder(os.Stdout).Encode(output)
```

## Success Criteria

- [ ] Config directory created with `0700`
- [ ] Config file created with `0600`
- [ ] `ld config init` does not echo token to terminal
- [ ] Token input works when stdin is piped (non-TTY fallback)
- [ ] `ld backup --json` produces valid JSON regardless of output path characters
- [ ] Existing configs are not re-permissioned (only new writes)
