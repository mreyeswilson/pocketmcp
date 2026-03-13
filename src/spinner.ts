import ora, { type Ora } from "npm:ora";

function isInteractiveCli(): boolean {
  return Deno.stderr.isTerminal() && Deno.env.get("CI") !== "true" && Deno.env.get("TERM") !== "dumb";
}

export function createSpinner(text: string): Ora {
  return ora({ text, isEnabled: isInteractiveCli() });
}
