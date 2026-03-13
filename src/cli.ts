import { Command } from "npm:commander";
import { runInstall } from "./install.ts";
import { runServe } from "./serve.ts";

function buildProgram(rawArgs: string[]): Command {
  const program = new Command();

  program
    .name("mcp-pocketbase-admin")
    .description("PocketBase MCP CLI")
    .showHelpAfterError("\nUse --help for usage details.")
    .helpCommand(true)
    .configureOutput({
      outputError: (text) => {
        Deno.stderr.writeSync(new TextEncoder().encode(text));
      },
    });

  program
    .command("serve")
    .description("Start the MCP stdio server")
    .option("--url <url>", "PocketBase URL")
    .option("--email <email>", "PocketBase user email")
    .option("--user <email>", "Alias for --email")
    .option("--password <password>", "PocketBase password")
    .option("--timeout-ms <number>", "Request timeout in milliseconds")
    .action(async () => {
      await runServe(rawArgs.slice(1));
    });

  program
    .command("install")
    .description("Install or uninstall MCP client entries")
    .option("--client <client>", "Target client: all|claude-desktop|cursor|vscode|windsurf")
    .option("--uninstall", "Remove existing entry instead of installing")
    .option("--binary <path>", "Path to compiled CLI binary")
    .option("--url <url>", "PocketBase URL")
    .option("--email <email>", "PocketBase user email")
    .option("--user <email>", "Alias for --email")
    .option("--password <password>", "PocketBase password")
    .option("--timeout-ms <number>", "Request timeout in milliseconds")
    .action(async () => {
      await runInstall(rawArgs.slice(1));
    });

  return program;
}

export async function runCli(args: string[]): Promise<void> {
  const program = buildProgram(args);
  if (args.length === 0) {
    program.outputHelp();
    return;
  }
  await program.parseAsync(args, { from: "user" });
}

if (import.meta.main) {
  try {
    await runCli(Deno.args);
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unknown error";
    console.error(`Error: ${message}`);
    Deno.exit(1);
  }
}
