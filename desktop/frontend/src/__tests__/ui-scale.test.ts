import { automaticUIScale, effectiveUIScale, isUIScale } from "../lib/uiScale";

let failed = 0;
function eq(actual: unknown, expected: unknown, label: string) {
  if (actual === expected) process.stdout.write(`  PASS  ${label}\n`);
  else {
    process.stderr.write(`  FAIL  ${label}: got ${String(actual)}, want ${String(expected)}\n`);
    failed += 1;
  }
}

console.log("\nui scale");
eq(automaticUIScale(1366, 768), 90, "small logical screen clamps to 90 percent");
eq(automaticUIScale(1920, 1080), 100, "1080p logical screen uses 100 percent");
eq(automaticUIScale(2560, 1440), 110, "1440p logical screen clamps to 110 percent");
eq(automaticUIScale(3840, 2160), 110, "4K logical screen remains capped");
eq(effectiveUIScale(0, { width: 1920, height: 1080 }), 100, "zero selects automatic scale");
eq(effectiveUIScale(125, { width: 1366, height: 768 }), 125, "manual scale overrides automatic scale");
eq(isUIScale(80), true, "80 is valid");
eq(isUIScale(83), false, "non-step value is invalid");
eq(isUIScale(130), false, "out-of-range value is invalid");

if (failed > 0) process.exit(1);
