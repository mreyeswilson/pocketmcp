import { runCli } from "./cli.ts";

if (import.meta.main) {
  await runCli(Deno.args);
}
