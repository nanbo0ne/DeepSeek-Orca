import type { Item } from "./useController";

export type TimelineProcessItem = Exclude<Item, { kind: "user" | "turn_stats" }>;

export type TimelineSegment =
  | { kind: "user"; item: Extract<Item, { kind: "user" }> }
  | { kind: "assistant"; item: Extract<Item, { kind: "assistant" }> }
  | { kind: "steer"; item: Extract<Item, { kind: "steer" }> }
  | { kind: "process"; id: string; items: TimelineProcessItem[]; completed: boolean }
  | { kind: "stats"; item: Extract<Item, { kind: "turn_stats" }> }
  | {
      kind: "completed";
      id: string;
      stats: Extract<Item, { kind: "turn_stats" }>;
      hidden: Item[];
      final: Extract<Item, { kind: "assistant" }>;
    };

const timelineCache = new WeakMap<readonly Item[], Map<boolean, TimelineSegment[]>>();

export function requiredWarmPage(warmTurnCount: number, questionTurn: number, pageSize: number): number {
  if (warmTurnCount <= 0 || questionTurn >= warmTurnCount || pageSize <= 0) return 0;
  return Math.max(1, Math.ceil((warmTurnCount - Math.max(0, questionTurn)) / pageSize));
}

export function visibleWarmStart(warmTurnCount: number, page: number, pageSize: number): number {
  if (warmTurnCount <= 0) return 0;
  if (page <= 0 || pageSize <= 0) return warmTurnCount;
  return Math.max(0, warmTurnCount - page * pageSize);
}

function visibleProcessItem(item: Item): item is TimelineProcessItem {
  if (item.kind === "assistant") return Boolean(item.reasoning) || item.streaming;
  if (item.kind === "tool") return !item.parentId && item.name !== "todo_write" && item.name !== "exit_plan_mode";
  return item.kind === "notice" || item.kind === "phase" || item.kind === "compaction" || item.kind === "steer";
}

function pushProcess(out: TimelineSegment[], item: TimelineProcessItem, completed: boolean) {
  const last = out[out.length - 1];
  if (last?.kind === "process" && last.completed === completed) {
    last.items.push(item);
    return;
  }
  out.push({ kind: "process", id: `process-${item.id}`, items: [item], completed });
}

function completedTurnSegment(items: readonly Item[], completed: boolean, fallbackID: string): TimelineSegment | null {
  if (!completed) return null;
  const explicitStats = items.find((item): item is Extract<Item, { kind: "turn_stats" }> => item.kind === "turn_stats");
  if (explicitStats && !explicitStats.success) return null;
  const hasFailureEvidence = items.some((item) =>
    (item.kind === "user" && item.failed) ||
    (item.kind === "tool" && (item.status === "error" || item.status === "stopped")),
  );
  if (!explicitStats && hasFailureEvidence) return null;
  let finalIndex = -1;
  for (let i = items.length - 1; i >= 0; i--) {
    const item = items[i];
    if (item.kind === "assistant" && !item.streaming && item.text.trim() !== "") {
      finalIndex = i;
      break;
    }
  }
  if (finalIndex < 0) return null;
  const final = items[finalIndex] as Extract<Item, { kind: "assistant" }>;
  const stats = explicitStats ?? {
    kind: "turn_stats" as const,
    id: `legacy-stats-${fallbackID}`,
    success: true,
  };
  const hidden = items.flatMap((item, index): Item[] => {
    if (item.kind === "user" || item.kind === "turn_stats") return [];
    if (index !== finalIndex) return [item];
    return final.reasoning.trim() !== "" ? [{ ...final, text: "" }] : [];
  });
  if (hidden.length === 0) return null;
  return { kind: "completed", id: `completed-${fallbackID}`, stats, hidden, final: { ...final, reasoning: "" } };
}

function buildTurn(items: readonly Item[], completed: boolean): TimelineSegment[] {
  const out: TimelineSegment[] = [];
  const stats = items.find((item): item is Extract<Item, { kind: "turn_stats" }> => item.kind === "turn_stats");
  const user = items.find((item): item is Extract<Item, { kind: "user" }> => item.kind === "user");
  if (user) out.push({ kind: "user", item: user });
  const collapsed = completedTurnSegment(items, completed, user?.id ?? stats?.id ?? "turn");
  if (collapsed) {
    out.push(collapsed);
    return out;
  }
  if (stats) out.push({ kind: "stats", item: stats });

  let lastNonAssistantProcessIndex = -1;
  items.forEach((item, index) => {
    if (item.kind !== "assistant" && visibleProcessItem(item)) lastNonAssistantProcessIndex = index;
  });
  let finalAssistantIndex = -1;
  for (let index = items.length - 1; index >= 0; index -= 1) {
    const item = items[index];
    if (index > lastNonAssistantProcessIndex && item.kind === "assistant" && (item.text.trim() !== "" || item.streaming)) {
      finalAssistantIndex = index;
      break;
    }
  }

  const processItems: TimelineProcessItem[] = [];
  let finalAssistant: Extract<Item, { kind: "assistant" }> | undefined;
  items.forEach((item, index) => {
    if (item.kind === "user" || item.kind === "turn_stats") return;
    if (item.kind === "assistant") {
      const isFinal = index === finalAssistantIndex;
      if (item.reasoning || item.streaming) processItems.push({ ...item, text: "" });
      if (isFinal) {
        finalAssistant = { ...item, reasoning: "" };
      } else if (item.text.trim() !== "") {
        processItems.push({ ...item, reasoning: "" });
      }
      return;
    }
    if (visibleProcessItem(item)) processItems.push(item);
  });

  if (processItems.length > 0) {
    out.push({
      kind: "process",
      id: `process-${processItems[0].id}`,
      items: processItems,
      completed,
    });
  }
  if (finalAssistant) out.push({ kind: "assistant", item: finalAssistant });
  return out;
}

export function buildTimelineSegments(items: readonly Item[], running: boolean): TimelineSegment[] {
  const cached = timelineCache.get(items)?.get(running);
  if (cached) return cached;

  const turns: Item[][] = [];
  const prelude: Item[] = [];
  for (const item of items) {
    if (item.kind === "user") {
      turns.push([item]);
      continue;
    }
    if (turns.length === 0) prelude.push(item);
    else turns[turns.length - 1].push(item);
  }

  const out: TimelineSegment[] = [];
  for (const item of prelude) {
    if (item.kind === "steer") out.push({ kind: "steer", item });
    else if (item.kind === "assistant") {
      if (visibleProcessItem(item)) pushProcess(out, { ...item, text: "" }, true);
      if (item.text.trim() !== "") out.push({ kind: "assistant", item: { ...item, reasoning: "" } });
    } else if (visibleProcessItem(item)) pushProcess(out, item, true);
  }

  turns.forEach((turn, index) => {
    const hasSuccessfulStats = turn.some((item) => item.kind === "turn_stats" && item.success);
    const isHistoricalTurn = index < turns.length - 1;
    const explicitlyCompleted = !running && hasSuccessfulStats;
    out.push(...buildTurn(turn, isHistoricalTurn || explicitlyCompleted));
  });
  const variants = timelineCache.get(items) ?? new Map<boolean, TimelineSegment[]>();
  variants.set(running, out);
  timelineCache.set(items, variants);
  return out;
}

export type ActivityIndicatorPhase = "model" | "tool";

export function activityIndicatorPhase(
  items: readonly Item[],
  enabled: boolean,
  running: boolean,
  paused: boolean,
): ActivityIndicatorPhase | undefined {
  if (!enabled || !running || paused) return undefined;

  let turnStart = 0;
  for (let index = items.length - 1; index >= 0; index -= 1) {
    if (items[index].kind === "user") {
      turnStart = index;
      break;
    }
  }
  for (let index = items.length - 1; index >= turnStart; index -= 1) {
    const item = items[index];
    if (item.kind === "tool" && item.status === "running") return "tool";
    if (item.kind === "compaction" && item.pending) return "tool";
  }
  return "model";
}

export function timelineKinds(segments: readonly TimelineSegment[]): string[] {
  return segments.map((segment) => {
    if (segment.kind === "completed") return `completed:${segment.hidden.map((item) => item.kind === "tool" ? `tool:${item.name}` : item.kind).join(",")}`;
    if (segment.kind !== "process") return segment.kind;
    return `process:${segment.items.map((item) => item.kind === "tool" ? `tool:${item.name}` : item.kind).join(",")}`;
  });
}
