import { effectiveUIScale, isUIScale } from "../lib/uiScale";

let failed = 0;
function eq(actual: unknown, expected: unknown, label: string) {
  if (actual === expected) process.stdout.write(`  PASS  ${label}\n`);
  else {
    process.stderr.write(`  FAIL  ${label}: got ${String(actual)}, want ${String(expected)}\n`);
    failed += 1;
  }
}

console.log("\nui scale");
eq(effectiveUIScale(0), 100, "automatic mode follows Windows without extra zoom");
eq(effectiveUIScale(125), 100, "legacy manual scale is ignored");
eq(isUIScale(0), true, "system DPI mode is the only supported preference");
eq(isUIScale(80), false, "legacy manual scale is no longer valid");

if (failed > 0) process.exit(1);
