# Specification: Config Token Trim Cleanup

## Jobs to Be Done
- Token input in `ld config init` is trimmed exactly once, with clear intent
- No redundant string operations that obscure control flow

## Problem

In `cmd/ld/config.go`, the token input path has a redundant `strings.TrimSpace` call:

```go
var token string
if term.IsTerminal(int(os.Stdin.Fd())) {
    tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
    // ...
    token = string(tokenBytes)        // ReadPassword does NOT include newline
    fmt.Println()
} else {
    tokenInput, err := reader.ReadString('\n')
    // ...
    token = strings.TrimSpace(tokenInput)  // Already trimmed here
}
token = strings.TrimSpace(token)           // Redundant for both branches
```

- **TTY branch**: `term.ReadPassword` returns raw bytes without trailing newline — `TrimSpace` is unnecessary but harmless.
- **Non-TTY branch**: `tokenInput` is already trimmed before assignment — the second `TrimSpace` is fully redundant.

The double-trim makes the code harder to reason about and suggests the author was unsure whether trimming happened upstream.

## Implementation

Remove the final `strings.TrimSpace(token)` line and ensure each branch produces the correct value:

```go
var token string
if term.IsTerminal(int(os.Stdin.Fd())) {
    tokenBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
    if err != nil {
        return fmt.Errorf("failed to read token: %w", err)
    }
    token = strings.TrimSpace(string(tokenBytes))
    fmt.Println()
} else {
    tokenInput, err := reader.ReadString('\n')
    if err != nil {
        return fmt.Errorf("failed to read token: %w", err)
    }
    token = strings.TrimSpace(tokenInput)
}
// token is already trimmed in both branches — no additional trim needed
```

Affected files:
- `cmd/ld/config.go` — `configInitCmd` RunE function

## Success Criteria
- [ ] Token is trimmed exactly once per code path
- [ ] No redundant `strings.TrimSpace` calls on `token`
- [ ] TTY input still works correctly (masked, no echo)
- [ ] Non-TTY (piped) input still works correctly
- [ ] Existing tests continue to pass
