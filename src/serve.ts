import PocketBase from "npm:pocketbase";
import { Server } from "npm:@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "npm:@modelcontextprotocol/sdk/server/stdio.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
  type CallToolRequest,
} from "npm:@modelcontextprotocol/sdk/types.js";
import { createSpinner } from "./spinner.ts";

type JsonObject = Record<string, unknown>;

type StartupConfig = {
  url: string;
  email: string;
  password: string;
  timeoutMs: number;
};

type ParsedCliArgs = {
  url?: string;
  email?: string;
  password?: string;
  timeoutMs?: number;
};

const DEFAULT_REQUEST_TIMEOUT_MS = 15000;

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

function parseCliArgs(args: string[]): ParsedCliArgs {
  const parsed: ParsedCliArgs = {};

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];

    if (arg === "--url" || arg.startsWith("--url=")) {
      const { value, nextIndex } = readFlagValue(args, i, "--url");
      parsed.url = value.trim();
      i = nextIndex;
      continue;
    }

    if (arg === "--email" || arg.startsWith("--email=")) {
      const { value, nextIndex } = readFlagValue(args, i, "--email");
      parsed.email = value.trim();
      i = nextIndex;
      continue;
    }

    if (arg === "--user" || arg.startsWith("--user=")) {
      const { value, nextIndex } = readFlagValue(args, i, "--user");
      parsed.email = value.trim();
      i = nextIndex;
      continue;
    }

    if (arg === "--password" || arg.startsWith("--password=")) {
      const { value, nextIndex } = readFlagValue(args, i, "--password");
      parsed.password = value;
      i = nextIndex;
      continue;
    }

    if (arg === "--timeout-ms" || arg.startsWith("--timeout-ms=")) {
      const { value, nextIndex } = readFlagValue(args, i, "--timeout-ms");
      parsed.timeoutMs = parseTimeoutMs(value, "--timeout-ms");
      i = nextIndex;
      continue;
    }

    throw new Error(`Unknown CLI flag for serve: ${arg}`);
  }

  return parsed;
}

function envTrimmed(name: string): string | undefined {
  const value = Deno.env.get(name);
  if (value === undefined) return undefined;
  const trimmed = value.trim();
  return trimmed === "" ? undefined : trimmed;
}

function resolveStartupConfig(args: string[]): StartupConfig {
  const cli = parseCliArgs(args);

  const url = cli.url ?? envTrimmed("POCKETBASE_URL");
  const email = cli.email ?? envTrimmed("POCKETBASE_EMAIL");
  const password = cli.password ?? envTrimmed("POCKETBASE_PASSWORD");

  const missing: string[] = [];
  if (!url) missing.push("--url or POCKETBASE_URL");
  if (!email) missing.push("--email/--user or POCKETBASE_EMAIL");
  if (!password) missing.push("--password or POCKETBASE_PASSWORD");

  if (missing.length > 0) {
    throw new Error(`Missing required configuration: ${missing.join(", ")}`);
  }

  const envTimeout = envTrimmed("REQUEST_TIMEOUT_MS");
  const timeoutMs = cli.timeoutMs ??
    (envTimeout ? parseTimeoutMs(envTimeout, "REQUEST_TIMEOUT_MS") : DEFAULT_REQUEST_TIMEOUT_MS);

  return {
    url: url as string,
    email: email as string,
    password: password as string,
    timeoutMs,
  };
}

function isObject(value: unknown): value is JsonObject {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function deepMerge(base: unknown, patch: unknown): unknown {
  if (!isObject(base) || !isObject(patch)) return patch;

  const merged: JsonObject = { ...base };
  for (const [key, value] of Object.entries(patch)) {
    const prev = merged[key];
    merged[key] = isObject(prev) && isObject(value) ? deepMerge(prev, value) : value;
  }
  return merged;
}

function withTimeout<T>(promise: Promise<T>, timeoutMs: number): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    const timeout = setTimeout(() => {
      reject(new Error(`Operation timed out after ${timeoutMs}ms`));
    }, timeoutMs);

    promise
      .then((result) => {
        clearTimeout(timeout);
        resolve(result);
      })
      .catch((error) => {
        clearTimeout(timeout);
        reject(error);
      });
  });
}

class ToolInputError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "ToolInputError";
  }
}

function asObject(value: unknown, label = "input"): JsonObject {
  if (!isObject(value)) {
    throw new ToolInputError(`${label} must be an object`);
  }
  return value;
}

function asNonEmptyString(value: unknown, label: string): string {
  if (typeof value !== "string" || value.trim() === "") {
    throw new ToolInputError(`${label} must be a non-empty string`);
  }
  return value.trim();
}

function asCollectionRef(args: JsonObject): string {
  const id = args.id;
  const name = args.name;

  if (id !== undefined && name !== undefined) {
    throw new ToolInputError("Provide either 'id' or 'name', not both");
  }

  if (id !== undefined) return asNonEmptyString(id, "id");
  if (name !== undefined) return asNonEmptyString(name, "name");

  throw new ToolInputError("Missing collection reference: provide 'id' or 'name'");
}

function toErrorDetails(error: unknown): JsonObject {
  if (error instanceof ToolInputError) {
    return {
      type: error.name,
      message: error.message,
      status: 400,
    };
  }

  if (error instanceof Error) {
    const anyErr = error as Error & { status?: number; response?: unknown; data?: unknown };
    return {
      type: anyErr.name || "Error",
      message: anyErr.message,
      status: anyErr.status ?? 500,
      response: anyErr.response,
      data: anyErr.data,
    };
  }

  return {
    type: "UnknownError",
    message: "Unexpected error",
    status: 500,
    raw: error,
  };
}

function success(tool: string, data: unknown) {
  return {
    content: [{
      type: "text" as const,
      text: JSON.stringify({ ok: true, tool, data }, null, 2),
    }],
  };
}

function failure(tool: string, error: unknown) {
  return {
    isError: true,
    content: [{
      type: "text" as const,
      text: JSON.stringify({ ok: false, tool, error: toErrorDetails(error) }, null, 2),
    }],
  };
}

export async function runServe(args: string[]): Promise<void> {
  const validationSpinner = createSpinner("Validating startup configuration").start();
  let startupConfig: StartupConfig;
  try {
    startupConfig = resolveStartupConfig(args);
    validationSpinner.succeed("Startup configuration validated");
  } catch (error) {
    validationSpinner.fail("Startup configuration validation failed");
    throw error;
  }

  const requestTimeoutMs = startupConfig.timeoutMs;

  async function initPocketBase(): Promise<PocketBase> {
    const authSpinner = createSpinner("Authenticating with PocketBase").start();
    const pb = new PocketBase(startupConfig.url);

    pb.beforeSend = (reqUrl: string, options: RequestInit) => {
      return {
        url: reqUrl,
        options: {
          ...options,
          signal: AbortSignal.timeout(requestTimeoutMs),
        },
      };
    };

    const candidate = pb as unknown as {
      admins?: { authWithPassword?: (email: string, password: string) => Promise<unknown> };
      collection: (name: string) => { authWithPassword: (email: string, password: string) => Promise<unknown> };
    };

    try {
      if (candidate.admins?.authWithPassword) {
        await withTimeout(candidate.admins.authWithPassword(startupConfig.email, startupConfig.password), requestTimeoutMs);
      } else {
        await withTimeout(
          candidate.collection("_superusers").authWithPassword(startupConfig.email, startupConfig.password),
          requestTimeoutMs,
        );
      }
      authSpinner.succeed("PocketBase authentication successful");
      return pb;
    } catch (error) {
      authSpinner.fail("PocketBase authentication failed");
      throw error;
    }
  }

  let pbClientPromise: Promise<PocketBase> | null = null;

  async function getPocketBase(): Promise<PocketBase> {
    if (!pbClientPromise) {
      pbClientPromise = initPocketBase();
      return pbClientPromise;
    }

    try {
      const pb = await pbClientPromise;
      if (!pb.authStore.isValid) {
        pbClientPromise = initPocketBase();
      }
      return await pbClientPromise;
    } catch {
      pbClientPromise = initPocketBase();
      return await pbClientPromise;
    }
  }

  const tools = [
    {
      name: "get_collections",
      description: "Lista todas las colecciones de PocketBase",
      inputSchema: {
        type: "object",
        additionalProperties: false,
        properties: {},
      },
    },
    {
      name: "get_collection",
      description: "Obtiene una coleccion por id o nombre",
      inputSchema: {
        type: "object",
        additionalProperties: false,
        properties: {
          id: { type: "string", minLength: 1 },
          name: { type: "string", minLength: 1 },
        },
        anyOf: [{ required: ["id"] }, { required: ["name"] }],
      },
    },
    {
      name: "create_collection",
      description: "Crea una coleccion con payload JSON generico",
      inputSchema: {
        type: "object",
        additionalProperties: false,
        properties: {
          collection: { type: "object", additionalProperties: true },
        },
        required: ["collection"],
      },
    },
    {
      name: "update_collection",
      description: "Actualiza una coleccion por id o nombre",
      inputSchema: {
        type: "object",
        additionalProperties: false,
        properties: {
          id: { type: "string", minLength: 1 },
          name: { type: "string", minLength: 1 },
          collection: { type: "object", additionalProperties: true },
        },
        required: ["collection"],
        anyOf: [{ required: ["id"] }, { required: ["name"] }],
      },
    },
    {
      name: "delete_collection",
      description: "Elimina una coleccion por id o nombre",
      inputSchema: {
        type: "object",
        additionalProperties: false,
        properties: {
          id: { type: "string", minLength: 1 },
          name: { type: "string", minLength: 1 },
        },
        anyOf: [{ required: ["id"] }, { required: ["name"] }],
      },
    },
    {
      name: "get_settings",
      description: "Obtiene configuraciones globales de la instancia",
      inputSchema: {
        type: "object",
        additionalProperties: false,
        properties: {},
      },
    },
    {
      name: "update_settings",
      description: "Actualiza configuraciones globales con patch parcial",
      inputSchema: {
        type: "object",
        additionalProperties: false,
        properties: {
          patch: { type: "object", additionalProperties: true },
        },
        required: ["patch"],
      },
    },
  ];

  const server = new Server(
    {
      name: "mcp-pocketbase-admin",
      version: "0.1.0",
    },
    {
      capabilities: {
        tools: {},
      },
    },
  );

  server.setRequestHandler(ListToolsRequestSchema, async () => {
    return { tools };
  });

  server.setRequestHandler(CallToolRequestSchema, async (request: CallToolRequest) => {
    const toolName = request.params.name;
    const args = asObject(request.params.arguments ?? {});

    try {
      const pb = await getPocketBase();

      switch (toolName) {
        case "get_collections": {
          const list = await withTimeout(pb.collections.getFullList(), requestTimeoutMs) as unknown[];
          return success(toolName, { collections: list, count: list.length });
        }

        case "get_collection": {
          const ref = asCollectionRef(args);
          const collection = await withTimeout(pb.collections.getOne(ref), requestTimeoutMs);
          return success(toolName, { collection });
        }

        case "create_collection": {
          const payload = asObject(args.collection, "collection");
          const collection = await withTimeout(pb.collections.create(payload), requestTimeoutMs);
          return success(toolName, { collection });
        }

        case "update_collection": {
          const ref = asCollectionRef(args);
          const payload = asObject(args.collection, "collection");
          const collection = await withTimeout(pb.collections.update(ref, payload), requestTimeoutMs);
          return success(toolName, { collection });
        }

        case "delete_collection": {
          const ref = asCollectionRef(args);
          await withTimeout(pb.collections.delete(ref), requestTimeoutMs);
          return success(toolName, { deleted: true, idOrName: ref });
        }

        case "get_settings": {
          const settings = await withTimeout(pb.settings.getAll(), requestTimeoutMs);
          return success(toolName, { settings });
        }

        case "update_settings": {
          const patch = asObject(args.patch, "patch");
          const current = await withTimeout(pb.settings.getAll(), requestTimeoutMs);
          const merged = deepMerge(current, patch);
          const updated = await withTimeout(pb.settings.update(merged as JsonObject), requestTimeoutMs);
          return success(toolName, { settings: updated });
        }

        default:
          throw new ToolInputError(`Unknown tool: ${toolName}`);
      }
    } catch (error) {
      return failure(toolName, error);
    }
  });

  const transport = new StdioServerTransport();
  const startupSpinner = createSpinner("Starting MCP stdio server").start();
  await server.connect(transport);
  startupSpinner.succeed("MCP stdio server is running");
}
