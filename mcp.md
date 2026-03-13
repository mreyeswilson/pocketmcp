## App overview

- Current app: `pocketmcp`, a Go CLI that exposes a PocketBase admin MCP server over `stdio` and can install MCP client config entries for local desktop tools.
- Primary user flows:
  - `serve`: authenticate against PocketBase and expose MCP tools over `stdio`.
  - `install`: write/remove `mcpServers.pocketbase-admin` entries in supported client config files.
- Scope today is intentionally narrow: collections and global settings administration.

## Current stack

- Runtime/build: Go 1.25+
- Language: Go
- CLI parser: `cobra`
- UX: `survey` for interactive `install` prompts when credentials or URL are missing
- Packaging: `go build`
- CI/CD:
  - `.github/workflows/release.yml` builds release binaries for Linux/macOS/Windows on tag push.
  - `.github/workflows/pages.yml` publishes `docs/` to GitHub Pages.

## Repo structure

- `main.go`: minimal CLI entrypoint.
- `cmd/pocketmcp/root.go`: root Cobra command plus subcommand registration.
- `cmd/pocketmcp/serve.go`: `serve` command validation and placeholder runtime wiring message.
- `cmd/pocketmcp/install.go`: `install` command validation and interactive Survey prompts for missing URL/email/password.
- `cmd/pocketmcp/version.go`: build metadata output.
- `internal/config/config.go`: shared config parsing, env fallback, timeout validation, and client normalization.
- `internal/config/config_test.go`: unit tests for config resolution.
- `docs/index.html`: landing page.
- `install.sh`, `install.ps1`: public installers for released binaries.
- `build/`: local compiled artifact location.

## Execution flow

### `serve`

1. `main.go` calls `pocketmcp.Execute()`.
2. `cmd/pocketmcp/root.go` registers `serve`, `install`, and `version`.
3. `cmd/pocketmcp/serve.go` parses flags, then falls back to env vars through `internal/config`:
   - `POCKETBASE_URL`
   - `POCKETBASE_EMAIL`
   - `POCKETBASE_PASSWORD`
   - `REQUEST_TIMEOUT_MS`
4. The current root Go app validates configuration and prints the next runtime wiring step.

### `install`

1. `cmd/pocketmcp/install.go` parses install/uninstall flags.
2. For install mode only, it uses Survey to ask interactively for any missing `url`, `email`, or `password` values; URL defaults to `http://localhost:8090`.
3. `internal/config/config.go` applies env fallback and shared validation.
4. The current root Go app validates installation inputs and prints the next config-writing step.

## MCP tools exposed today

- `get_collections`
- `get_collection`
- `create_collection`
- `update_collection`
- `delete_collection`
- `get_settings`
- `update_settings`

Notes:

- All tool schemas use root `type: object` for validator compatibility.
- `get_collection`, `update_collection`, and `delete_collection` enforce exactly one of `id` or `name` at runtime.

## Development and validation

- Main dev commands:
  - `go run . --help`
  - `go run . serve --url http://127.0.0.1:8090 --email admin@example.com --password secret`
  - `go run . install --client all --url http://127.0.0.1:8090 --email admin@example.com --password secret`
  - `go test ./...`
  - `go build ./...`
- Build command:
  - `make compile`
  - or `go build -o build/pocketmcp .`

## Integrations and external touchpoints

- MCP clients via local JSON config edits.
- GitHub Releases for binary distribution.
- GitHub Pages for landing docs.

## Migration-relevant findings

- The Go root app currently preserves config validation and command shape, but `serve` and `install` still print placeholder next steps instead of performing the full runtime work.
- `install` prompting is intentionally limited to this command; `serve` still relies on flags/env only.
- Client support remains intentionally asymmetric for compatibility: parsing accepts `opencode`, while help text and `all` expansion stay aligned with the previous public contract.
- Release automation now needs to build Go binaries directly from the repo root.
- `git status` could not be inspected in this environment because the repo is flagged as a dubious ownership directory by Git. No git config was changed.

## Suggested migration slices

1. Port PocketBase service startup/auth logic into the root Go app.
2. Port MCP tool definitions and stdio server handling into Go.
3. Port client config write/remove behavior for supported MCP clients.
4. Add compatibility tests against the previous CLI contract.
5. Keep release automation aligned with the root Go layout.
