import { dirname, fromFileUrl, join } from "jsr:@std/path";
import { createSpinner } from "./spinner.ts";

const SERVER_NAME = "pocketbase-admin";
const CLI_RELATIVE_PATH = "src/cli.ts";
const DEFAULT_TIMEOUT_MS = 15000;

type SupportedClient = "claude-desktop" | "cursor" | "vscode" | "windsurf";

type CliOptions = {
  client: "all" | SupportedClient;
  uninstall: boolean;
  url?: string;
  email?: string;
  password?: string;
  timeoutMs?: number;
  binary?: string;
};

type InstallConfig = {
  command: string;
  args: string[];
};

type ClientReport = {
  client: SupportedClient;
  path: string;
  status: "installed" | "updated" | "removed" | "skipped" | "error";
  message: string;
};

function usageText(): string {
  return `Usage:
  pocketmcp install --client <all|claude-desktop|cursor|vscode|windsurf> [options]
  pocketmcp install --uninstall --client <all|claude-desktop|cursor|vscode|windsurf>

Install options:
  --url <url>
  --email <email> (alias: --user)
  --password <password>
  --timeout-ms <number>
  --binary <path>

Fallback env values:
  --url -> POCKETBASE_URL
  --email/--user -> POCKETBASE_EMAIL
  --password -> POCKETBASE_PASSWORD
  --timeout-ms -> REQUEST_TIMEOUT_MS

Notes:
  - In install mode, url/email/password are required (flags or env).
  - Password is never printed in clear text.
`;
}

function readFlagValue(args: string[], index: number, flag: string): { value: string; nextIndex: number } {
  const current = args[index];
  const equalsPrefix = `${flag}=`;
  if (current.startsWith(equalsPrefix)) {
    return { value: current.slice(equalsPrefix.length), nextIndex: index };
  }

  const next = args[index + 1];
  if (next === undefined || next.startsWith("--")) {
    throw new Error(`Missing value for ${flag}`);
  }

  return { value: next, nextIndex: index + 1 };
}

function parseTimeoutMs(raw: string, source: string): number {
  const value = Number(raw);
  if (!Number.isFinite(value) || value <= 0) {
    throw new Error(`${source} must be a positive number`);
  }
  return value;
}

function envTrimmed(name: string): string | undefined {
  const value = Deno.env.get(name);
  if (value === undefined) return undefined;
  const trimmed = value.trim();
  return trimmed === "" ? undefined : trimmed;
}

function parseClient(raw: string): "all" | SupportedClient {
  const client = raw.trim().toLowerCase();
  if (
    client === "all" ||
    client === "claude-desktop" ||
    client === "cursor" ||
    client === "vscode" ||
    client === "windsurf"
  ) {
    return client;
  }
  throw new Error(`Unsupported client: ${raw}`);
}

function parseCliArgs(args: string[]): CliOptions {
  const options: CliOptions = {
    client: "all",
    uninstall: false,
  };

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];

    if (arg === "--uninstall") {
      options.uninstall = true;
      continue;
    }

    if (arg === "--client" || arg.startsWith("--client=")) {
      const { value, nextIndex } = readFlagValue(args, i, "--client");
      options.client = parseClient(value);
      i = nextIndex;
      continue;
    }

    if (arg === "--url" || arg.startsWith("--url=")) {
      const { value, nextIndex } = readFlagValue(args, i, "--url");
      options.url = value.trim();
      i = nextIndex;
      continue;
    }

    if (arg === "--email" || arg.startsWith("--email=")) {
      const { value, nextIndex } = readFlagValue(args, i, "--email");
      options.email = value.trim();
      i = nextIndex;
      continue;
    }

    if (arg === "--user" || arg.startsWith("--user=")) {
      const { value, nextIndex } = readFlagValue(args, i, "--user");
      options.email = value.trim();
      i = nextIndex;
      continue;
    }

    if (arg === "--password" || arg.startsWith("--password=")) {
      const { value, nextIndex } = readFlagValue(args, i, "--password");
      options.password = value;
      i = nextIndex;
      continue;
    }

    if (arg === "--timeout-ms" || arg.startsWith("--timeout-ms=")) {
      const { value, nextIndex } = readFlagValue(args, i, "--timeout-ms");
      options.timeoutMs = parseTimeoutMs(value, "--timeout-ms");
      i = nextIndex;
      continue;
    }

    if (arg === "--binary" || arg.startsWith("--binary=")) {
      const { value, nextIndex } = readFlagValue(args, i, "--binary");
      options.binary = value.trim();
      i = nextIndex;
      continue;
    }

    if (arg === "--help" || arg === "-h") {
      console.log(usageText());
      Deno.exit(0);
    }

    throw new Error(`Unknown CLI flag for install: ${arg}`);
  }

  if (!options.uninstall) {
    const url = options.url ?? envTrimmed("POCKETBASE_URL");
    const email = options.email ?? envTrimmed("POCKETBASE_EMAIL");
    const password = options.password ?? envTrimmed("POCKETBASE_PASSWORD");
    const envTimeout = envTrimmed("REQUEST_TIMEOUT_MS");

    const missing: string[] = [];
    if (!url) missing.push("--url or POCKETBASE_URL");
    if (!email) missing.push("--email/--user or POCKETBASE_EMAIL");
    if (!password) missing.push("--password or POCKETBASE_PASSWORD");
    if (missing.length > 0) {
      throw new Error(`Missing required options for install: ${missing.join(", ")}`);
    }

    options.url = url;
    options.email = email;
    options.password = password;
    options.timeoutMs = options.timeoutMs ??
      (envTimeout ? parseTimeoutMs(envTimeout, "REQUEST_TIMEOUT_MS") : DEFAULT_TIMEOUT_MS);
  }

  return options;
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function getSupportedClients(selected: "all" | SupportedClient): SupportedClient[] {
  if (selected === "all") {
    return ["claude-desktop", "cursor", "vscode", "windsurf"];
  }
  return [selected];
}

function getHomeDir(): string {
  const home = Deno.env.get("HOME") || Deno.env.get("USERPROFILE");
  if (!home) {
    throw new Error("Could not determine user home directory");
  }
  return home;
}

function getAppDataDir(home: string): string {
  return Deno.env.get("APPDATA") || join(home, "AppData", "Roaming");
}

function getClientConfigPath(client: SupportedClient): string {
  const home = getHomeDir();
  const os = Deno.build.os;

  if (client === "claude-desktop") {
    if (os === "windows") return join(getAppDataDir(home), "Claude", "claude_desktop_config.json");
    if (os === "darwin") return join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json");
    return join(home, ".config", "Claude", "claude_desktop_config.json");
  }

  const appName = client === "cursor" ? "Cursor" : client === "vscode" ? "Code" : "Windsurf";
  if (os === "windows") return join(getAppDataDir(home), appName, "User", "settings.json");
  if (os === "darwin") return join(home, "Library", "Application Support", appName, "User", "settings.json");
  return join(home, ".config", appName, "User", "settings.json");
}

async function pathExists(path: string): Promise<boolean> {
  try {
    await Deno.stat(path);
    return true;
  } catch {
    return false;
  }
}

function getRepoRoot(): string {
  const scriptPath = fromFileUrl(import.meta.url);
  return join(dirname(scriptPath), "..");
}

async function resolveDefaultBinaryPath(repoRoot: string): Promise<string | undefined> {
  const os = Deno.build.os;
  const candidates = [
    join(repoRoot, "build", os === "windows" ? "pocketmcp.exe" : "pocketmcp"),
    join(repoRoot, "dist", os === "windows" ? "pocketmcp.exe" : "pocketmcp"),
  ];

  for (const candidate of candidates) {
    if (await pathExists(candidate)) {
      return candidate;
    }
  }

  return undefined;
}

async function buildInstallConfig(options: CliOptions): Promise<InstallConfig> {
  if (!options.url || !options.email || !options.password) {
    throw new Error("Invalid install options");
  }

  const repoRoot = getRepoRoot();
  const timeoutMs = options.timeoutMs ?? DEFAULT_TIMEOUT_MS;

  const baseArgs = [
    "serve",
    "--url",
    options.url,
    "--email",
    options.email,
    "--password",
    options.password,
    "--timeout-ms",
    String(timeoutMs),
  ];

  const explicitBinary = options.binary && options.binary.trim() !== "" ? options.binary : undefined;
  const binaryPath = explicitBinary ?? await resolveDefaultBinaryPath(repoRoot);

  if (binaryPath) {
    return {
      command: binaryPath,
      args: baseArgs,
    };
  }

  return {
    command: "deno",
    args: ["run", "-A", join(repoRoot, CLI_RELATIVE_PATH), ...baseArgs],
  };
}

async function loadConfig(path: string): Promise<Record<string, unknown>> {
  const exists = await pathExists(path);
  if (!exists) return {};

  const raw = await Deno.readTextFile(path);
  if (raw.trim() === "") return {};

  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch (error) {
    throw new Error(`Invalid JSON in ${path}: ${error instanceof Error ? error.message : "parse error"}`);
  }

  if (!isObject(parsed)) {
    throw new Error(`Config root in ${path} must be a JSON object`);
  }

  return parsed;
}

function configsEqual(a: unknown, b: unknown): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

function buildServerEntry(install: InstallConfig): Record<string, unknown> {
  return {
    command: install.command,
    args: install.args,
  };
}

async function saveConfig(path: string, data: Record<string, unknown>): Promise<void> {
  await Deno.mkdir(dirname(path), { recursive: true });
  await Deno.writeTextFile(path, `${JSON.stringify(data, null, 2)}\n`);
}

async function installForClient(client: SupportedClient, install: InstallConfig): Promise<ClientReport> {
  const configPath = getClientConfigPath(client);

  try {
    const config = await loadConfig(configPath);
    const current = config.mcpServers;
    const mcpServers: Record<string, unknown> = current === undefined
      ? {}
      : isObject(current)
      ? { ...current }
      : (() => {
        throw new Error(`Field "mcpServers" in ${configPath} must be an object`);
      })();

    const nextEntry = buildServerEntry(install);
    const previousEntry = mcpServers[SERVER_NAME];

    if (configsEqual(previousEntry, nextEntry)) {
      return {
        client,
        path: configPath,
        status: "skipped",
        message: "Entry already up to date",
      };
    }

    const hadEntry = previousEntry !== undefined;
    mcpServers[SERVER_NAME] = nextEntry;
    const nextConfig = { ...config, mcpServers };
    await saveConfig(configPath, nextConfig);

    return {
      client,
      path: configPath,
      status: hadEntry ? "updated" : "installed",
      message: hadEntry ? "Entry updated" : "Entry installed",
    };
  } catch (error) {
    return {
      client,
      path: configPath,
      status: "error",
      message: error instanceof Error ? error.message : "Unknown error",
    };
  }
}

async function uninstallForClient(client: SupportedClient): Promise<ClientReport> {
  const configPath = getClientConfigPath(client);

  try {
    const exists = await pathExists(configPath);
    if (!exists) {
      return {
        client,
        path: configPath,
        status: "skipped",
        message: "Config file does not exist",
      };
    }

    const config = await loadConfig(configPath);
    const current = config.mcpServers;
    if (!isObject(current)) {
      return {
        client,
        path: configPath,
        status: "skipped",
        message: "No mcpServers object found",
      };
    }

    if (!(SERVER_NAME in current)) {
      return {
        client,
        path: configPath,
        status: "skipped",
        message: "Entry not present",
      };
    }

    const nextServers = { ...current };
    delete nextServers[SERVER_NAME];

    await saveConfig(configPath, { ...config, mcpServers: nextServers });
    return {
      client,
      path: configPath,
      status: "removed",
      message: "Entry removed",
    };
  } catch (error) {
    return {
      client,
      path: configPath,
      status: "error",
      message: error instanceof Error ? error.message : "Unknown error",
    };
  }
}

function redactArgs(args: string[]): string[] {
  const redacted = [...args];
  for (let i = 0; i < redacted.length; i++) {
    if (redacted[i] === "--password" && redacted[i + 1] !== undefined) {
      redacted[i + 1] = "***";
    }
    if (redacted[i].startsWith("--password=")) {
      redacted[i] = "--password=***";
    }
  }
  return redacted;
}

function printReport(options: CliOptions, reports: ClientReport[], installConfig?: InstallConfig): void {
  console.log(options.uninstall ? "MCP uninstall completed." : "MCP install completed.");
  console.log(`MCP server: ${SERVER_NAME}`);

  if (installConfig) {
    console.log(`Configured command: ${installConfig.command}`);
    console.log(`Configured args: ${JSON.stringify(redactArgs(installConfig.args))}`);
  }

  for (const report of reports) {
    console.log(`- ${report.client}: ${report.status} (${report.message}) -> ${report.path}`);
  }

  const hasErrors = reports.some((item) => item.status === "error");
  if (hasErrors) {
    Deno.exit(2);
  }
}

export async function runInstall(args: string[]): Promise<void> {
  const parseSpinner = createSpinner("Validating install options").start();
  let options: CliOptions;
  try {
    options = parseCliArgs(args);
    parseSpinner.succeed("Install options validated");
  } catch (error) {
    parseSpinner.fail("Install options validation failed");
    console.error(`Error: ${error instanceof Error ? error.message : "Failed to parse CLI args"}`);
    console.error("");
    console.log(usageText());
    Deno.exit(1);
  }

  const clients = getSupportedClients(options.client);
  const reports: ClientReport[] = [];

  if (options.uninstall) {
    for (const client of clients) {
      const spinner = createSpinner(`Uninstalling for ${client}`).start();
      const report = await uninstallForClient(client);
      reports.push(report);

      if (report.status === "removed") {
        spinner.succeed(`${client}: ${report.message}`);
      } else if (report.status === "error") {
        spinner.fail(`${client}: ${report.message}`);
      } else {
        spinner.warn(`${client}: ${report.message}`);
      }
    }
    printReport(options, reports);
    return;
  }

  const configSpinner = createSpinner("Preparing install command").start();
  let installConfig: InstallConfig;
  try {
    installConfig = await buildInstallConfig(options);
    configSpinner.succeed("Install command ready");
  } catch (error) {
    configSpinner.fail("Failed to prepare install command");
    throw error;
  }

  for (const client of clients) {
    const spinner = createSpinner(`Installing for ${client}`).start();
    const report = await installForClient(client, installConfig);
    reports.push(report);

    if (report.status === "installed" || report.status === "updated") {
      spinner.succeed(`${client}: ${report.message}`);
    } else if (report.status === "error") {
      spinner.fail(`${client}: ${report.message}`);
    } else {
      spinner.warn(`${client}: ${report.message}`);
    }
  }
  printReport(options, reports, installConfig);
}
