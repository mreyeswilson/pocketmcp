# pocketmcp

Go implementation of the PocketBase MCP CLI.

## Current state

- `main.go` boots the Cobra CLI.
- `cmd/pocketmcp/` contains `mcp`, `setup`, `uninstall`, and `version`.
- `internal/config/` centralizes flag and environment resolution.
- `internal/install/` writes or removes MCP client configuration for supported AI tools.
- `internal/pocketbase/` wraps the PocketBase superuser REST API.
- `internal/mcpserver/` exposes PocketBase admin tools over MCP stdio.
- GitHub Actions release builds and publishes Go binaries for Linux, macOS, and Windows.

## Commands

```bash
go run . --help
go run . mcp --url http://127.0.0.1:8090 --email admin@example.com --password secret
go run . setup --client all --url http://127.0.0.1:8090 --email admin@example.com --password secret
go run . uninstall --client all
go test ./...
go build ./...
```

## Release assets

- `pocketmcp` for Linux
- `pocketmcp` for macOS
- `pocketmcp.exe` for Windows

## Compatibility notes

- Environment fallback uses `POCKETBASE_URL`, `POCKETBASE_EMAIL`, `POCKETBASE_PASSWORD`, and `REQUEST_TIMEOUT_MS`.
- `setup` supports `claude-code`, `claude-desktop`, `codex`, `cursor`, `gemini`, `opencode`, `vscode`, and `windsurf`.
- `setup --client all` expands to `claude-code`, `codex`, `cursor`, `gemini`, `opencode`, `vscode`, and `windsurf`.
- `setup` writes MCP entries that launch the global `pocketmcp` binary with the credentials captured during setup.
- `setup` prompts interactively for PocketBase URL, user/email, and password when they are not provided by flags.
- `mcp` runs a real stdio MCP server with tools for collections, records/users, and settings administration.
