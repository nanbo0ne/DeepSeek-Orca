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
  timelineKinds(buildTimelineSegments(chronology, false)),
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
  timelineKinds(buildTimelineSegments(adjacent, false)),
  ["user", "process:assistant,tool:read,notice", "assistant", "process:tool:bash"],
);

const withStats: Item[] = [
  ...chronology,
  { kind: "turn_stats", id: "s1", elapsedMs: 5000, tokens: 120 },
];
equal(
  "turn stats render immediately after the user message",
  timelineKinds(buildTimelineSegments(withStats, false)),
  ["user", "stats", "assistant", "process:tool:read", "assistant", "process:tool:bash"],
);

const runningSegments = buildTimelineSegments(chronology, true);
const completedSegments = buildTimelineSegments(chronology, false);
const lastRunningProcess = [...runningSegments].reverse().find((segment) => segment.kind === "process");
const lastCompletedProcess = [...completedSegments].reverse().find((segment) => segment.kind === "process");
equal(
  "current process segment remains live before TurnDone",
  lastRunningProcess?.kind === "process" ? lastRunningProcess.completed : undefined,
  false,
);
equal(
  "current process segment completes after TurnDone",
  lastCompletedProcess?.kind === "process" ? lastCompletedProcess.completed : undefined,
  true,
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
