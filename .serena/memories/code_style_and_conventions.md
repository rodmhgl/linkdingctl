# Code Style and Conventions

## Go Idioms

- Follow standard Go idioms: return errors, don't panic
- Use stdlib where possible (avoid third-party HTTP clients)
- Nothing in `internal/` packages should be exported outside the module

## Documentation

- All exported functions must have doc comments
- Error messages should be user-friendly (not technical stack traces)
- Follow Go documentation standards

## Testing

- Unit tests for API client (mock HTTP responses)
- Integration tests marked with build tag `//go:build integration`
- Table-driven tests preferred for multiple test cases
- Test files named `*_test.go` in same package

## File Organization

- One cobra command per file in `cmd/linkdingctl/`
- Internal packages properly scoped
- Models in `internal/models/`
- API client in `internal/api/`

## Error Handling

- Return descriptive errors that guide users
- Handle common cases: connection errors, auth failures, not found
- Don't expose internal error details to end users
