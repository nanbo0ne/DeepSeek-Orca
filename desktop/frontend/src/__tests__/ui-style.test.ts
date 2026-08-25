import { readFileSync } from "node:fs";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { normalizeUIStyle } from "../lib/uiStyle";

let failed = 0;
function check(condition: boolean, label: string) {
  if (condition) process.stdout.write(`  PASS  ${label}\n`);
  else {
    process.stderr.write(`  FAIL  ${label}\n`);
    failed += 1;
  }
}

const root = fileURLToPath(new URL("..", import.meta.url));
const css = readFileSync(join(root, "styles.css"), "utf8");

console.log("\nUI style");
check(normalizeUIStyle(undefined) === "modern", "missing preference defaults to modern");
check(normalizeUIStyle("modern") === "modern", "modern preference is retained");
check(normalizeUIStyle("classic") === "classic", "classic preference is retained");
check(css.includes(':root[data-theme-style] .layout'), "modern layout overlay is present");
check(css.includes(".layout {"), "classic baseline remains available");

if (failed > 0) process.exit(1);
