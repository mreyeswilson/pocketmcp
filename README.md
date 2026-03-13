# pocketmcp

Go implementation of the PocketBase MCP CLI.

## Current state

- `main.go` boots the Cobra CLI.
- `cmd/pocketmcp/` contains `mcp`, `setup`, and `version`.
- `internal/config/` centralizes flag and environment resolution.
- GitHub Actions release builds and publishes Go binaries for Linux, macOS, and Windows.

## Commands

```bash
go run . --help
go run . mcp --url http://127.0.0.1:8090 --email admin@example.com --password secret
go run . setup --client all --url http://127.0.0.1:8090 --email admin@example.com --password secret
go test ./...
go build ./...
```

## Release assets

- `pocketmcp` for Linux
- `pocketmcp-macos` for macOS
- `pocketmcp.exe` for Windows

## Compatibility notes

- Environment fallback uses `POCKETBASE_URL`, `POCKETBASE_EMAIL`, `POCKETBASE_PASSWORD`, and `REQUEST_TIMEOUT_MS`.
- `setup --client opencode` is accepted by config parsing alongside the other supported clients.
- `mcp` and `setup` prompt interactively for PocketBase URL, user/email, and password when they are not provided by flags.
