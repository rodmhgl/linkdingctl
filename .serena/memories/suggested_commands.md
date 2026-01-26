# Suggested Commands for Development

## Building

```bash
go build ./...              # Build all packages
go build -o linkdingctl ./cmd/linkdingctl     # Build binary
```

## Testing

```bash
go test ./...               # Run all unit tests
go test -v ./...            # Verbose test output
go test ./internal/api      # Test specific package
```

## Linting and Quality

```bash
go vet ./...                # Run go vet (static analysis)
go fmt ./...                # Format code
```

## Running the Binary

```bash
./linkdingctl --help                 # Show help
./linkdingctl config init            # Initialize config
./linkdingctl config test            # Test connection to LinkDing
./linkdingctl add <url>              # Add bookmark
./linkdingctl list                   # List bookmarks
```

## Git Operations

```bash
git status                  # Check status
git add -A                  # Stage all changes
git commit -m "feat: ..."   # Commit with message
git push                    # Push to remote
```

## Module Management

```bash
go mod tidy                 # Clean up dependencies
go mod download             # Download dependencies
```
