# Specification: User Profile Command

## Jobs to Be Done
- User can verify which account they are authenticated as
- User can confirm their LinkDing instance version and permissions
- User can troubleshoot authentication issues by inspecting profile details

## Commands Structure
```
ld user profile [flags]
```

The `user` command group provides account-related subcommands. The initial subcommand is `profile`.

## User Profile
```
ld user profile
```

Retrieves the authenticated user's profile from the LinkDing API (`GET /api/user/profile/`).

Examples:
```bash
ld user profile
ld user profile --json
```

Output (human):
```
Username:     rodmhgl
Display Name: Rod Stewart
Theme:        auto
Bookmarks:    523
```

Output (JSON):
```json
{
  "theme": "auto",
  "bookmark_count": 523,
  "display_name": "Rod Stewart",
  "username": "rodmhgl"
}
```

## API Endpoint

`GET /api/user/profile/`

Response fields (per LinkDing API):
- `theme` (string): UI theme preference (auto, light, dark)
- `bookmark_count` (int): Total bookmarks owned by user
- `display_name` (string): User's display name
- `username` (string): Login username

## Implementation Notes

- Add `UserProfile` struct to `internal/models/`
- Add `GetUserProfile()` method to `internal/api/client.go`
- Create `cmd/ld/user.go` with `user` parent command and `profile` subcommand
- The `user` command itself should display help (no default action)

## Success Criteria
- [ ] `ld user profile` displays authenticated user information
- [ ] `ld user profile --json` outputs valid JSON with all profile fields
- [ ] HTTP 401 produces clear error: "Authentication failed. Check your API token."
- [ ] HTTP 403 produces clear error: "Insufficient permissions for this operation."
- [ ] Missing config produces standard config error
- [ ] `ld user --help` lists available subcommands
