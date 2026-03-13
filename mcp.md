## App overview

- Current app: `pocketmcp`, a Go CLI that exposes a PocketBase admin MCP server over `stdio` and can install MCP client config entries for local desktop tools.
- Primary user flows:
  - `mcp`: authenticate against PocketBase and expose MCP tools over `stdio`.
  - `setup`: write/remove `pocketbase-admin` MCP entries in supported client config files.
- Scope today is intentionally narrow: collections and global settings administration.

## Current stack

- Runtime/build: Go 1.25+
- Language: Go
- CLI parser: `cobra`
- UX: `survey` for interactive `setup` and `mcp` prompts when credentials or URL are missing
- Packaging: `go build`
- CI/CD:
  - `.github/workflows/release.yml` builds release binaries for Linux/macOS/Windows on tag push.
  - `.github/workflows/pages.yml` publishes `docs/` to GitHub Pages.

## Repo structure

- `main.go`: minimal CLI entrypoint.
- `cmd/pocketmcp/root.go`: root Cobra command plus subcommand registration.
- `cmd/pocketmcp/serve.go`: `mcp` command validation and placeholder runtime wiring message.
- `cmd/pocketmcp/serve.go`: `mcp` command bootstrapping for the stdio MCP server.
- `cmd/pocketmcp/install.go`: `setup` command plus MCP client config writing/removal.
- `cmd/pocketmcp/version.go`: build metadata output.
- `internal/config/config.go`: shared config parsing, env fallback, timeout validation, and client normalization.
- `internal/config/config_test.go`: unit tests for config resolution.
- `internal/install/install.go`: client-specific JSON/TOML installation logic.
- `internal/install/install_test.go`: unit tests for config mutation behavior.
- `internal/pocketbase/client.go`: authenticated PocketBase superuser REST client.
- `internal/pocketbase/client_test.go`: HTTP-level tests for auth and CRUD requests.
- `internal/mcpserver/server.go`: MCP tool registration for collections, records, and settings.
- `internal/mcpserver/server_test.go`: server tool registration tests.
- `docs/index.html`: landing page.
- `install.sh`, `install.ps1`: public installers for released binaries.
- `build/`: local compiled artifact location.

## Execution flow

### `mcp`

1. `main.go` calls `pocketmcp.Execute()`.
2. `cmd/pocketmcp/root.go` registers `mcp`, `setup`, and `version`.
3. `cmd/pocketmcp/serve.go` resolves credentials from flags/env, and only prompts when running interactively in a terminal:
   - `POCKETBASE_URL`
   - `POCKETBASE_EMAIL`
   - `POCKETBASE_PASSWORD`
   - `REQUEST_TIMEOUT_MS`
4. `internal/pocketbase/client.go` authenticates against `/api/collections/_superusers/auth-with-password`.
5. `internal/mcpserver/server.go` serves PocketBase administration tools over stdio using the official Go MCP SDK.

### `setup`

1. `cmd/pocketmcp/install.go` parses setup/uninstall flags.
2. For install mode only, it uses Survey to ask interactively for any missing `url`, `email`, or `password` values; URL defaults to `http://localhost:8090`.
3. `internal/config/config.go` applies env fallback and shared validation.
4. `internal/install/install.go` writes or removes client-specific MCP config entries.
5. Installed entries launch the global `mcp` binary with args from `setup`: subcommand `mcp`, PocketBase URL, email, password, and timeout.

## MCP tools exposed today

- `list_collections`
- `get_collection`
- `create_collection`
- `update_collection`
- `delete_collection`
- `get_settings`
- `update_settings`
- `list_records`
- `get_record`
- `create_record`
- `update_record`
- `delete_record`

Notes:

- All tool schemas use root `type: object` for validator compatibility.
- Record CRUD works with any collection, including auth collections used for users.

## Development and validation

- Main dev commands:
  - `go run . --help`
  - `go run . mcp --url http://127.0.0.1:8090 --email admin@example.com --password secret`
  - `go run . setup --client all --url http://127.0.0.1:8090 --email admin@example.com --password secret`
  - `go test ./...`
  - `go build ./...`
- Build command:
  - `make compile`
  - or `go build -o build/pocketmcp .`

## Integrations and external touchpoints

- MCP clients via local JSON config edits.
- MCP clients via local JSON/TOML config edits:
  - `claude-code`
  - `claude-desktop`
  - `codex`
  - `cursor`
  - `gemini`
  - `opencode`
  - `vscode`
  - `windsurf`
- GitHub Releases for binary distribution.
- GitHub Pages for landing docs.

## Migration-relevant findings

- The Go root app now performs real client configuration writes for supported MCP clients and runs a real PocketBase admin MCP server over stdio.
- `setup` and `mcp` both support interactive credential prompting through Survey.
- `setup --client all` now targets `claude-code`, `codex`, `cursor`, `gemini`, `opencode`, `vscode`, and `windsurf`, with `claude-desktop` available explicitly.
- Release automation now needs to build Go binaries directly from the repo root.
- `git status` could not be inspected in this environment because the repo is flagged as a dubious ownership directory by Git. No git config was changed.

## Suggested migration slices

1. Port PocketBase service startup/auth logic into the root Go app.
2. Port MCP tool definitions and stdio server handling into Go.
3. Complete any remaining client-specific compatibility checks for MCP config formats.
4. Add compatibility tests against the previous CLI contract.
5. Keep release automation aligned with the root Go layout.
