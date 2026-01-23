# Specification: Core CLI Structure & Configuration

## Jobs to Be Done
- User can configure LinkDing connection once and reuse across commands
- User gets helpful error messages when misconfigured
- User can override config via environment variables for scripting/CI

## Commands Structure
```
linkdingctl [command] [subcommand] [flags]

Global Flags:
  --config string   Config file path (default ~/.config/linkdingctl/config.yaml)
  --json            Output as JSON instead of human-readable
  --debug           Enable debug logging
```

## Configuration
Config file location: `~/.config/linkdingctl/config.yaml`

```yaml
url: https://linkding.example.com
token: your-api-token-here
```

Environment variable overrides (higher priority than config file):
- `LINKDING_URL` - Base URL of LinkDing instance
- `LINKDING_TOKEN` - API token

## Setup Command
```
linkdingctl config init          # Interactive setup, writes config file
linkdingctl config show          # Display current config (token redacted)
linkdingctl config test          # Test connection to LinkDing
```

## Success Criteria
- [ ] `linkdingctl --help` displays all available commands
- [ ] `linkdingctl config init` creates config file with user input
- [ ] `linkdingctl config test` validates connection and reports success/failure
- [ ] Missing config produces clear error: "Run 'linkdingctl config init' to set up"
- [ ] Environment variables override config file values
- [ ] `--json` flag works on all commands that produce output

## Error Handling
- HTTP 401: "Authentication failed. Check your API token."
- HTTP 404: "LinkDing not found at {url}. Check your URL."
- Connection refused: "Cannot connect to {url}. Is LinkDing running?"
- Missing config: "No configuration found. Run 'linkdingctl config init' to set up."
