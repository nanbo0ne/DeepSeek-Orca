import type { Item } from "./useController";

export type TimelineProcessItem = Exclude<Item, { kind: "user" | "steer" | "turn_stats" }>;

export type TimelineSegment =
  | { kind: "user"; item: Extract<Item, { kind: "user" }> }
  | { kind: "assistant"; item: Extract<Item, { kind: "assistant" }> }
  | { kind: "steer"; item: Extract<Item, { kind: "steer" }> }
  | { kind: "process"; id: string; items: TimelineProcessItem[]; completed: boolean }
  | { kind: "stats"; item: Extract<Item, { kind: "turn_stats" }> };

function visibleProcessItem(item: Item): item is TimelineProcessItem {
  if (item.kind === "assistant") return Boolean(item.reasoning) || item.streaming;
  if (item.kind === "tool") return !item.parentId && item.name !== "todo_write" && item.name !== "exit_plan_mode";
  return item.kind === "notice" || item.kind === "phase" || item.kind === "compaction";
}

function pushProcess(out: TimelineSegment[], item: TimelineProcessItem, completed: boolean) {
  const last = out[out.length - 1];
  if (last?.kind === "process" && last.completed === completed) {
    last.items.push(item);
    return;
  }
  out.push({ kind: "process", id: `process-${item.id}`, items: [item], completed });
}

function buildTurn(items: readonly Item[], completed: boolean): TimelineSegment[] {
  const out: TimelineSegment[] = [];
  const stats = items.find((item): item is Extract<Item, { kind: "turn_stats" }> => item.kind === "turn_stats");
  const user = items.find((item): item is Extract<Item, { kind: "user" }> => item.kind === "user");
  if (user) out.push({ kind: "user", item: user });
  if (stats) out.push({ kind: "stats", item: stats });

  for (const item of items) {
    if (item.kind === "user" || item.kind === "turn_stats") continue;
    if (item.kind === "assistant") {
      if (visibleProcessItem(item)) {
        pushProcess(out, { ...item, text: "" }, completed);
      }
      if (item.text.trim() !== "" || item.streaming) {
        out.push({ kind: "assistant", item: { ...item, reasoning: "" } });
      }
      continue;
    }
    if (item.kind === "steer") {
      out.push({ kind: "steer", item });
      continue;
    }
    if (visibleProcessItem(item)) pushProcess(out, item, completed);
  }
  return out;
}

export function buildTimelineSegments(items: readonly Item[], running: boolean): TimelineSegment[] {
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
    out.push(...buildTurn(turn, index < turns.length - 1 || !running));
  });
  return out;
}

export function timelineKinds(segments: readonly TimelineSegment[]): string[] {
  return segments.map((segment) => {
    if (segment.kind !== "process") return segment.kind;
    return `process:${segment.items.map((item) => item.kind === "tool" ? `tool:${item.name}` : item.kind).join(",")}`;
  });
}
