// Run: tsx src/__tests__/context-panel-usage.test.ts

import { computeContextPanelUsage } from "../lib/contextPanelUsage";

let passed = 0;
let failed = 0;

function eq(a: unknown, b: unknown, label: string) {
  if (JSON.stringify(a) === JSON.stringify(b)) {
    process.stdout.write(`  PASS  ${label}\n`);
    passed += 1;
  } else {
    process.stdout.write(`  FAIL  ${label}: expected ${JSON.stringify(b)}, got ${JSON.stringify(a)}\n`);
    failed += 1;
  }
}

console.log("\ncontext panel usage");

eq(
  computeContextPanelUsage({
    context: { used: 0, window: 1000000, sessionTokens: 19337 },
    info: {
      usedTokens: 0,
      windowTokens: 1000000,
      promptTokens: 0,
      completionTokens: 0,
      totalTokens: 19337,
      reasoningTokens: 0,
      cacheHitTokens: 0,
      cacheMissTokens: 0,
      requestCount: 2,
      readFiles: [],
      changedFiles: [],
    },
    usage: {
      promptTokens: 12000,
      completionTokens: 7000,
      totalTokens: 19000,
      reasoningTokens: 300,
      cacheHitTokens: 0,
      cacheMissTokens: 0,
      sessionCacheHitTokens: 0,
      sessionCacheMissTokens: 0,
    },
    sessionTokens: 19337,
  }).promptTokens,
  12000,
  "falls back to latest usage breakdown when panel only has totals",
);

eq(
  computeContextPanelUsage({
    context: { used: 0, window: 1000000, sessionTokens: 19337 },
    info: {
      usedTokens: 0,
      windowTokens: 1000000,
      promptTokens: 0,
      completionTokens: 0,
      totalTokens: 19337,
      reasoningTokens: 0,
      cacheHitTokens: 0,
      cacheMissTokens: 0,
      requestCount: 2,
      readFiles: [],
      changedFiles: [],
    },
    sessionTokens: 19337,
  }).usedTokens,
  0,
  "does not treat cumulative session totals as current context usage",
);

eq(
  computeContextPanelUsage({
    context: { used: 0, window: 1000000, sessionTokens: 19337 },
    info: {
      usedTokens: 0,
      windowTokens: 1000000,
      promptTokens: 0,
      completionTokens: 0,
      totalTokens: 19337,
      reasoningTokens: 0,
      cacheHitTokens: 0,
      cacheMissTokens: 0,
      sessionPromptTokens: 12000,
      sessionCompletionTokens: 7000,
      sessionReasoningTokens: 300,
      requestCount: 2,
      readFiles: [],
      changedFiles: [],
    },
    sessionTokens: 19337,
  }).reasoningTokens,
  300,
  "uses persisted session breakdown after restart",
);

eq(
  computeContextPanelUsage({
    context: { used: 5000, window: 1000000, sessionTokens: 19337 },
    info: {
      usedTokens: 0,
      windowTokens: 1000000,
      promptTokens: 3200,
      completionTokens: 900,
      totalTokens: 19337,
      reasoningTokens: 100,
      cacheHitTokens: 0,
      cacheMissTokens: 0,
      readFiles: [],
      changedFiles: [],
    },
  }).usedTokens,
  5000,
  "keeps the real context snapshot when it is available",
);

console.log(`\n${passed} passed, ${failed} failed, ${passed + failed} total`);
if (failed > 0) process.exit(1);
