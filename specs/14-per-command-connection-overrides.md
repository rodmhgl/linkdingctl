# Specification: Per-Command Connection Overrides

## Jobs to Be Done
- User can target a different LinkDing instance without modifying their config file
- User can use a different API token for a one-off command (e.g., testing a new token)
- User can script operations across multiple LinkDing instances in a single shell session
- User can run commands in CI/CD without a config file, using only CLI flags

## Global Flags
```
ld [command] [flags]

New Global Flags:
  --url string     LinkDing instance URL (overrides config and env)
  --token string   API token (overrides config and env)
```

These flags override both the config file and environment variables, establishing the highest-priority configuration source.

## Precedence Order (highest to lowest)

1. CLI flags (`--url`, `--token`)
2. Environment variables (`LINKDING_URL`, `LINKDING_TOKEN`)
3. Config file (`~/.config/ld/config.yaml`)

## Examples

```bash
# One-off command against a different instance
ld list --url https://other-linkding.example.com --token abc123

# Use a different token for testing
ld user profile --token new-token-to-test

# Script across multiple instances
ld export --url https://instance-a.example.com --token tokenA > a.json
ld export --url https://instance-b.example.com --token tokenB > b.json

# CI/CD usage without config file
ld backup --url "$STAGING_URL" --token "$STAGING_TOKEN" -o /tmp/
```

## Partial Override Behavior

Either flag can be used independently:
- `--url` alone: Uses provided URL with token from config/env
- `--token` alone: Uses provided token with URL from config/env
- Both together: Fully overrides connection without needing config/env

## Implementation Notes

- Add `--url` and `--token` as persistent flags on the root command in `cmd/ld/root.go`
- Modify `loadConfig()` to apply CLI flag overrides after loading config file and env vars
- The `--token` value must not be logged even when `--debug` is enabled (redact in debug output)
- If neither config file, env vars, nor CLI flags provide URL/token, the existing "Run 'ld config init'" error applies
- `ld config show` should indicate when overrides are active (e.g., "URL: https://... (--url flag)")
- `ld config test` should test the effective configuration (including any overrides)

## Success Criteria
- [ ] `--url` flag overrides config file and environment variable URL
- [ ] `--token` flag overrides config file and environment variable token
- [ ] Partial overrides work (only `--url` or only `--token`)
- [ ] Commands function with only CLI flags and no config file present
- [ ] `--token` value is never printed in debug output
- [ ] `ld config show` reflects active overrides
- [ ] `ld config test` tests effective configuration including overrides
- [ ] Existing config file and env var behavior is unchanged when flags are not provided
- [ ] All commands respect the override flags
