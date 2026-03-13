# PocketMCP CLI (Deno)

CLI para ejecutar un servidor MCP sobre `stdio` para PocketBase y para instalar/desinstalar su configuracion en clientes MCP.

La UX del CLI usa `commander` (comandos/flags/help) y `ora` (spinners de progreso en pasos clave, con comportamiento seguro en modo no interactivo).

Comandos principales del binario:

- `serve` -> inicia el servidor MCP.
- `install` -> instala/desinstala la entrada MCP en clientes (`claude-desktop`, `cursor`, `vscode`, `windsurf`).

## Instalacion rapida

Reemplaza `mreyeswilson/pocketmcp` si usas un fork propio:

```bash
curl -fsSL https://raw.githubusercontent.com/mreyeswilson/pocketmcp/main/install.sh | bash
```

```powershell
powershell -c "irm https://raw.githubusercontent.com/mreyeswilson/pocketmcp/main/install.ps1 | iex"
```

## Requisitos (desarrollo)

- Deno 2.x
- Instancia de PocketBase accesible por URL
- Credenciales de admin/superuser

## Uso CLI

```bash
deno run -A src/cli.ts <command> [flags]
```

Tambien puedes ver ayuda estructurada por comando:

```bash
deno run -A src/cli.ts --help
deno run -A src/cli.ts serve --help
deno run -A src/cli.ts install --help
```

### `serve`

Flags:

- `--url`
- `--email` (alias: `--user`)
- `--password`
- `--timeout-ms` (opcional, default `15000`)

Fallback por variables de entorno:

- `POCKETBASE_URL`
- `POCKETBASE_EMAIL`
- `POCKETBASE_PASSWORD`
- `REQUEST_TIMEOUT_MS`

Ejemplo:

```bash
deno task start -- --url http://127.0.0.1:8090 --email admin@example.com --password 'tu_password'
```

### `install`

Flags:

- `--client <all|claude-desktop|cursor|vscode|windsurf>`
- `--uninstall`
- `--binary <ruta>` (opcional, fuerza binario especifico)
- `--url`, `--email`/`--user`, `--password`, `--timeout-ms`

Notas:

- En modo instalacion, `url/email/password` son obligatorios (flags o env).
- En modo uninstall no hacen falta credenciales.
- El password se enmascara en logs.

Instalar:

```bash
deno task install -- --client all --url http://127.0.0.1:8090 --email admin@example.com --password 'tu_password'
```

Desinstalar:

```bash
deno task install -- --uninstall --client all
```

## Tareas Deno

- `deno task dev` -> watch mode de `serve`
- `deno task start` -> ejecuta `serve`
- `deno task install` -> ejecuta subcomando `install`
- `deno task cli --help` -> entrypoint unico del CLI (commander)
- `deno task check` -> type check de CLI y modulos

## Compilar a binario

```bash
deno compile --allow-env --allow-net --allow-read --allow-write --output ./dist/pocketmcp src/cli.ts
```

## Releases en GitHub Actions

Al pushear un tag `v*` (por ejemplo `v0.2.0`) se ejecuta `.github/workflows/release.yml` para:

- Compilar binarios con `deno compile` para:
  - `x86_64-unknown-linux-gnu`
  - `x86_64-apple-darwin`
  - `x86_64-pc-windows-msvc`
- Publicar release con assets versionados por tag.

## Landing docs

Landing simple disponible en:

- `docs/index.html`

Incluye instalacion one-liner y quick start de `serve`/`install`.

La publicacion en GitHub Pages se hace con `.github/workflows/pages.yml` en cada push a `main` (y tambien por ejecucion manual de workflow).

## Verificacion

```bash
deno task check
```
