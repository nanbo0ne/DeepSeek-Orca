import { initialState, reducer, type Item } from "../lib/useController";
import { buildTimelineSegments, timelineKinds } from "../lib/transcriptTimeline";

let failed = 0;

function equal<T>(label: string, actual: T, expected: T) {
  if (JSON.stringify(actual) === JSON.stringify(expected)) {
    process.stdout.write(`  PASS  ${label}\n`);
    return;
  }
  failed += 1;
  process.stdout.write(`  FAIL  ${label}: expected ${JSON.stringify(expected)}, got ${JSON.stringify(actual)}\n`);
}

const chronology: Item[] = [
  { kind: "user", id: "u1", text: "start" },
  { kind: "assistant", id: "a1", text: "first", reasoning: "", streaming: false },
  { kind: "tool", id: "t1", name: "read", args: "{}", readOnly: true, status: "done" },
  { kind: "assistant", id: "a2", text: "second", reasoning: "", streaming: false },
  { kind: "tool", id: "t2", name: "bash", args: "{}", readOnly: false, status: "done" },
];

equal(
  "assistant text and tools preserve insertion order",
  timelineKinds(buildTimelineSegments(chronology, true)),
  ["user", "assistant", "process:tool:read", "assistant", "process:tool:bash"],
);

const adjacent: Item[] = [
  { kind: "user", id: "u1", text: "start" },
  { kind: "assistant", id: "r1", text: "", reasoning: "thinking", streaming: false },
  { kind: "tool", id: "t1", name: "read", args: "{}", readOnly: true, status: "done" },
  { kind: "notice", id: "n1", level: "info", text: "done" },
  { kind: "assistant", id: "a1", text: "answer", reasoning: "", streaming: false },
  { kind: "tool", id: "t2", name: "bash", args: "{}", readOnly: false, status: "done" },
];

equal(
  "adjacent process items merge but never cross assistant text",
  timelineKinds(buildTimelineSegments(adjacent, true)),
  ["user", "process:assistant,tool:read,notice", "assistant", "process:tool:bash"],
);

const withStats: Item[] = [
  ...chronology,
  { kind: "turn_stats", id: "s1", elapsedMs: 5000, tokens: 120, success: true },
];
equal(
  "successful completed turn collapses intermediate activity",
  timelineKinds(buildTimelineSegments(withStats, false)),
  ["user", "completed:assistant,tool:read,tool:bash"],
);

const recoveredStats: Item[] = [
  { kind: "user", id: "ru1", text: "fix it" },
  { kind: "tool", id: "rt1", name: "bash", args: "{}", readOnly: false, status: "error", error: "first attempt failed" },
  { kind: "tool", id: "rt2", name: "bash", args: "{}", readOnly: false, status: "done" },
  { kind: "assistant", id: "ra1", text: "Fixed and verified.", reasoning: "", streaming: false },
  { kind: "turn_stats", id: "rs1", elapsedMs: 7000, success: true },
];
equal(
  "explicit successful completion collapses a recovered tool failure",
  timelineKinds(buildTimelineSegments(recoveredStats, false)),
  ["user", "completed:tool:bash,tool:bash"],
);

const failedStats: Item[] = [
  ...chronology,
  { kind: "notice", id: "error", level: "warn", text: "failed" },
  { kind: "turn_stats", id: "s2", elapsedMs: 5000, success: false },
];
equal(
  "failed turn keeps its diagnostic timeline visible",
  timelineKinds(buildTimelineSegments(failedStats, false)),
  ["user", "stats", "assistant", "process:tool:read", "assistant", "process:tool:bash,notice"],
);

equal(
  "final legacy turn without completion evidence remains expanded",
  timelineKinds(buildTimelineSegments(chronology, false)),
  ["user", "assistant", "process:tool:read", "assistant", "process:tool:bash"],
);

const legacyHistory: Item[] = [
  ...chronology,
  { kind: "user", id: "u2", text: "next question" },
  { kind: "assistant", id: "a3", text: "working", reasoning: "", streaming: false },
];
equal(
  "next real user turn confirms the preceding legacy turn completed",
  timelineKinds(buildTimelineSegments(legacyHistory, true)),
  ["user", "completed:assistant,tool:read,tool:bash", "user", "assistant"],
);

const failedLegacyHistory: Item[] = [
  { kind: "user", id: "fu1", text: "run" },
  { kind: "tool", id: "ft1", name: "bash", args: "{}", readOnly: false, status: "error", error: "failed" },
  { kind: "assistant", id: "fa1", text: "The command failed.", reasoning: "", streaming: false },
  { kind: "user", id: "fu2", text: "try something else" },
];
equal(
  "legacy failure evidence prevents automatic collapse",
  timelineKinds(buildTimelineSegments(failedLegacyHistory, false)),
  ["user", "process:tool:bash", "assistant", "user"],
);

const runningSegments = buildTimelineSegments(chronology, true);
const completedSegments = buildTimelineSegments(failedStats, false);
const lastRunningProcess = [...runningSegments].reverse().find((segment) => segment.kind === "process");
const lastCompletedProcess = [...completedSegments].reverse().find((segment) => segment.kind === "process");
equal(
  "current process segment remains live before TurnDone",
  lastRunningProcess?.kind === "process" ? lastRunningProcess.completed : undefined,
  false,
);
equal(
  "failed process segment stays expanded after TurnDone",
  lastCompletedProcess?.kind === "process" ? lastCompletedProcess.completed : undefined,
  false,
);

const active = reducer(initialState, { type: "event", e: { kind: "turn_started" } });
const done = reducer(active, { type: "event", e: { kind: "turn_done" } });
const backgroundNotice = reducer(done, {
  type: "event",
  e: { kind: "notice", level: "info", text: "background job finished" },
});
equal("background notice does not reactivate running state", backgroundNotice.running, false);
equal("background notice does not reactivate the turn", backgroundNotice.turnActive, false);

if (failed > 0) process.exit(1);
