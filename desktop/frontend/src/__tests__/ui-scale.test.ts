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
eq(effectiveUIScale(125), 125, "manual scale applies relative to Windows DPI");
eq(isUIScale(80), true, "80 is valid");
eq(isUIScale(83), false, "non-step value is invalid");
eq(isUIScale(130), false, "out-of-range value is invalid");

if (failed > 0) process.exit(1);
