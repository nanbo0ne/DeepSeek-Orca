// Run: tsx src/__tests__/send-failed.test.ts

import { initialState, nextContextRefreshDelay, reducer } from "../lib/useController";
import type { WireEvent } from "../lib/types";

let passed = 0;
let failed = 0;

function eq(a: unknown, b: unknown, label: string) {
  if (a === b) {
    process.stdout.write(`  PASS  ${label}\n`);
    passed += 1;
  } else {
    process.stdout.write(`  FAIL  ${label}: expected ${JSON.stringify(b)}, got ${JSON.stringify(a)}\n`);
    failed += 1;
  }
}

console.log("\nsend failure feedback");

const sent = reducer({ ...initialState }, { type: "user", text: "hello", seq: 0 });
eq(sent.items.length, 1, "submit appends the user bubble immediately");
eq(sent.items[0].kind === "user" && sent.items[0].text, "hello", "bubble carries the submitted text");
eq(sent.running, true, "submit marks the turn running");
eq(sent.pendingUser, "hello", "submit tracks the optimistic bubble");

const confirmed = reducer(sent, { type: "event", e: { kind: "text", text: "hi" } as WireEvent });
eq(confirmed.items.filter((it) => it.kind === "user").length, 1, "first backend event confirms without duplicating");
eq(confirmed.pendingUser, undefined, "confirmation clears the pending marker");

const failedState = reducer(sent, { type: "send_failed", error: "Send failed: bridge unavailable" });
const failedBubble = failedState.items.find((it) => it.kind === "user");
eq(failedBubble?.kind === "user" && failedBubble.failed, true, "send_failed marks the bubble failed");
const notice = failedState.items[failedState.items.length - 1];
eq(notice.kind, "notice", "send_failed appends a notice");
eq(notice.kind === "notice" && notice.level, "warn", "the notice is a warning");
eq(failedState.running, false, "send_failed stops the running indicator");
eq(failedState.pendingUser, undefined, "send_failed clears the pending marker");

const lateFailure = reducer(confirmed, { type: "send_failed", error: "Send failed: late" });
eq(lateFailure, confirmed, "send_failed after backend confirmation is a no-op");

const unsent = reducer(sent, { type: "unsend" });
eq(unsent.pendingUser, undefined, "unsend clears the pending marker");
eq(unsent.discardTurn, true, "unsend discards the in-flight turn");

const guided = reducer(confirmed, { type: "steer_sent", text: "please keep going" });
const guidedItem = guided.items[guided.items.length - 1];
eq(guidedItem.kind, "steer", "steer_sent appends a visible guidance item");
eq(guidedItem.kind === "steer" && guidedItem.text, "please keep going", "guidance item keeps the submitted text");
eq(guided.items.filter((it) => it.kind === "user").length, 1, "guidance is not counted as a new user turn");

const duplicateSteer = reducer(guided, { type: "event", e: { kind: "steer", text: "please keep going" } as WireEvent });
eq(duplicateSteer.items.filter((it) => it.kind === "steer").length, 1, "backend steer confirmation does not duplicate guidance");

const persistedTotals = reducer(
  { ...initialState, context: { used: 8000, window: 128000, sessionTokens: 10000 }, sessionTokens: 10000, sessionCost: 0.02 },
  {
    type: "event",
    e: {
      kind: "usage",
      usage: {
        promptTokens: 50000,
        completionTokens: 2000,
        totalTokens: 52000,
        reasoningTokens: 1000,
        cacheHitTokens: 45000,
        cacheMissTokens: 5000,
        sessionCacheHitTokens: 0,
        sessionCacheMissTokens: 0,
        cost: 0.1,
        currency: "¥",
      },
    } as WireEvent,
  },
);
eq(persistedTotals.context.used, 8000, "usage does not overwrite the backend context-window snapshot");
eq(persistedTotals.sessionTokens, 10000, "usage does not temporarily duplicate persisted session tokens");
eq(persistedTotals.context.sessionTokens, 10000, "context session total stays on persisted telemetry");
eq(persistedTotals.sessionCost, 0.02, "usage does not temporarily duplicate persisted session cost");

eq(nextContextRefreshDelay(10_000, undefined), 0, "context refresh runs immediately when never refreshed");
eq(nextContextRefreshDelay(10_000, 9_500), 500, "context refresh throttles recent backend snapshots");
eq(nextContextRefreshDelay(10_000, 8_000), 0, "context refresh runs after the refresh interval");

console.log(`\n${passed} passed, ${failed} failed`);
if (failed > 0) process.exit(1);
