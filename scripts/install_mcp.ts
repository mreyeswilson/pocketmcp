import { runInstall } from "../src/install.ts";

if (import.meta.main) {
  await runInstall(Deno.args);
}
