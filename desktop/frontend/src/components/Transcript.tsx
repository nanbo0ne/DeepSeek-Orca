import { createContext, memo, type CSSProperties, type MouseEvent as ReactMouseEvent, type ReactNode, useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import type { Item, LiveStream } from "../lib/useController";
import type { CheckpointMeta, JobView } from "../lib/types";
import { useT } from "../lib/i18n";
import { replaceAttachmentRefsForDisplay } from "../lib/attachmentDisplay";
import { AssistantMessage, TurnActions, UserMessage } from "./Message";
import { ProcessBrainIcon, ProcessCard, ProcessCompactIcon, ProcessInfoIcon, ProcessPhaseIcon, ProcessStatusIcon } from "./ProcessCard";
import { ToolCard } from "./ToolCard";
import { ArrowDown, ChevronRight } from "lucide-react";
import { Welcome } from "./Welcome";

type ToolItem = Extract<Item, { kind: "tool" }>;
type AssistantItem = Extract<Item, { kind: "assistant" }>;
type TurnStatsItem = Extract<Item, { kind: "turn_stats" }>;
type OpenTurnAction = { turn: number; menu: "summary" | "rewind" };
type QuestionAnchor = { id: string; text: string; turn: number };

const QUESTION_NAV_MIN_COUNT = 2;
const LiveStreamContext = createContext<LiveStream | undefined>(undefined);

const LiveAssistantMessage = memo(function LiveAssistantMessage({ item, defaultExpanded = false }: { item: AssistantItem; defaultExpanded?: boolean }) {
  const live = useContext(LiveStreamContext);
  const shown = live && live.id === item.id ? { ...item, text: live.text, reasoning: live.reasoning, streaming: true } : item;
  return <AssistantMessage item={shown} defaultExpanded={defaultExpanded} />;
});

function renderProcessItem(
  item: Item,
  subcalls: ReadonlyMap<string, ToolItem[]>,
  liveToolID: string,
  defaultExpandThinking: boolean,
): ReactNode {
  switch (item.kind) {
    case "assistant":
      if (!item.reasoning) return null;
      return <AssistantMessage key={item.id} item={{ ...item, text: "" }} defaultExpanded={defaultExpandThinking} />;
    case "tool":
      if (item.parentId || item.name === "todo_write" || item.name === "exit_plan_mode") return null;
      return <ToolCard key={item.id} item={item} subcalls={subcalls.get(item.id)} livePulse={item.id === liveToolID} />;
    case "phase":
      return <PhaseCard key={item.id} text={item.text} />;
    case "notice":
      return <NoticeCard key={item.id} level={item.level} text={item.text} />;
    case "compaction":
      return <CompactionCard key={item.id} item={item} />;
    default:
      return null;
  }
}

function TurnProcessSummary({
  stats,
  items,
  subcalls,
  liveToolID,
  defaultExpandThinking,
}: {
  stats: TurnStatsItem;
  items: Item[];
  subcalls: ReadonlyMap<string, ToolItem[]>;
  liveToolID: string;
  defaultExpandThinking: boolean;
}) {
  const [open, setOpen] = useState(false);
  const visibleItems = items.filter(isProcessItem);
  if (visibleItems.length === 0) return null;
  return (
    <ProcessCard
      tone="violet"
      icon={<ProcessBrainIcon size={12} />}
      kind={turnStatsLabel(stats)}
      open={open}
      onOpenChange={setOpen}
      className="turn-process-summary"
    >
      <div className="turn-process-summary__body">
        {visibleItems.map((item) => renderProcessItem(item, subcalls, liveToolID, defaultExpandThinking))}
      </div>
    </ProcessCard>
  );
}

function BackgroundJobsCard({ jobs }: { jobs: JobView[] }) {
  const label = jobs.length === 1 ? jobs[0].label || jobs[0].id : `${jobs.length} tasks`;
  return (
    <ProcessCard
      tone="accent"
      icon={<ProcessPhaseIcon size={12} />}
      kind="research"
      name={`后台研究仍在运行：${label}`}
      meta={<ProcessStatusIcon state="running" label="后台研究仍在运行" />}
      className="background-jobs-card"
    />
  );
}

// ── Layer budgets ─────────────────────────────────────────────────────────────
// Hot zone: the most recent N user turns are always fully rendered. All data
// stays in memory (items[]), so expanding a warm turn is instant — no API call.
// Cold zone: a "load more" button paginates the warm zone in batches.
//
//   items[0]  ─┐
//   ...        │ Cold zone  ───  paginated, shown on "load more"
//              ├────────────  warmTurnStart
//   ...        │ Warm zone  ───  collapsible summary cards (individual expand)
//              ├────────────  hotStartIdx
//   items[N]  ─┤ Hot zone   ───  fully rendered
//   ...        │
//   items[end] ┘

const HOT_TURNS = 30;
const WARM_PAGE_SIZE = 20; // cold-zone pagination batch

// ── Helpers ───────────────────────────────────────────────────────────────────

function questionAnchorId(id: string): string {
  return `question-anchor-${id}`;
}

function compactQuestionText(text: string): string {
  const cleaned = replaceAttachmentRefsForDisplay(text).replace(/\s+/g, " ").trim();
  if (cleaned.length <= 80) return cleaned;
  return cleaned.slice(0, 80);
}

function scrollVersion(items: Item[]): string {
  return items
    .map((it) => {
      switch (it.kind) {
        case "assistant":
          return `${it.id}:a:${it.text?.length ?? 0}:${it.reasoning?.length ?? 0}:${it.streaming ? 1 : 0}`;
        case "tool":
          return `${it.id}:t:${it.name}:${it.status}:${it.args?.length ?? 0}:${it.output?.length ?? 0}:${it.error?.length ?? 0}:${it.truncated ? 1 : 0}`;
        default:
          return `${it.id}:${it.kind}`;
      }
    })
    .join("|");
}

function repinIfWasPinned(
  el: HTMLDivElement,
  stick: { current: boolean },
  frame: { current: number | null },
) {
  if (!stick.current) return;
  if (frame.current !== null) cancelAnimationFrame(frame.current);
  frame.current = requestAnimationFrame(() => {
    if (stick.current) el.scrollTop = el.scrollHeight;
    frame.current = null;
  });
}

function isAtBottom(el: HTMLDivElement): boolean {
  return el.scrollHeight - el.scrollTop - el.clientHeight <= 4;
}

function scrollElementToBottom(el: HTMLDivElement) {
  el.scrollTop = el.scrollHeight;
}

function lastRunningToolID(items: Item[]): string {
  for (let i = items.length - 1; i >= 0; i--) {
    const item = items[i];
    if (item.kind === "tool" && item.status === "running") return item.id;
  }
  return "";
}

function isProcessItem(item: Item): boolean {
  if (item.kind === "assistant") return Boolean(item.reasoning);
  if (item.kind === "tool") return !item.parentId && item.name !== "todo_write" && item.name !== "exit_plan_mode";
  return item.kind === "notice" || item.kind === "phase" || item.kind === "compaction";
}

function isProcessItemRunning(item: Item): boolean {
  if (item.kind === "assistant") return item.streaming;
  if (item.kind === "tool") return item.status === "running";
  if (item.kind === "compaction") return item.pending;
  return false;
}

function formatTurnElapsed(ms: number): string {
  const totalSeconds = Math.max(0, Math.round(ms / 1000));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  const pad = (n: number) => String(n).padStart(2, "0");
  if (hours > 0) return `${hours}hr${pad(minutes)}min${pad(seconds)}s`;
  if (minutes > 0) return `${minutes}min${pad(seconds)}s`;
  return `${seconds}s`;
}

function turnStatsLabel(item: TurnStatsItem): string {
  const tokens = typeof item.tokens === "number" && item.tokens > 0 ? `${item.tokens} tokens` : "token 消耗待统计";
  return `已思考 ${formatTurnElapsed(item.elapsedMs)}，${tokens}`;
}

// Summarise a warm turn for its compact card.
function warmUserPreview(text: string): string {
  const cleaned = replaceAttachmentRefsForDisplay(text).replace(/\s+/g, " ").trim();
  return cleaned.length <= 80 ? cleaned : cleaned.slice(0, 77) + "...";
}

// ── Turn grouping ─────────────────────────────────────────────────────────────
// A turn is everything from one UserMessage up to (but not including) the next
// UserMessage. This grouping is used only for warm-zone rendering; the hot zone
// still uses the flat items array to preserve the existing rendering logic.

interface TurnGroup {
  userItem: Item;
  assistantPreview: string;
  toolCount: number;
  startIdx: number; // first index in items[] (the user message)
  endIdx: number;   // exclusive end
}

function buildTurnGroups(items: Item[], questions: QuestionAnchor[]): TurnGroup[] {
  const groups: TurnGroup[] = [];
  let turnIdx = 0;
  let start = -1;
  for (let i = 0; i < items.length; i++) {
    if (items[i].kind === "user") {
      if (start >= 0) {
        // finalise previous turn
        groups[groups.length - 1].endIdx = i;
      }
      start = i;
      turnIdx = questions.findIndex((q) => q.id === items[i].id);
      if (turnIdx < 0) turnIdx = groups.length;
      groups.push({
        userItem: items[i],
        assistantPreview: "",
        toolCount: 0,
        startIdx: i,
        endIdx: items.length,
      });
    } else if (start >= 0 && groups.length > 0) {
      const g = groups[groups.length - 1];
      const it = items[i];
      if (it.kind === "assistant" && !it.streaming) {
        const previewText = it.text?.trim() || "";
        if (previewText) {
          g.assistantPreview = warmUserPreview(previewText);
        }
      }
      if (it.kind === "tool" && !it.parentId) {
        g.toolCount++;
      }
    }
  }
  return groups;
}

// ── Transcript component ──────────────────────────────────────────────────────

export function Transcript({
  items,
  live,
  footerHeight = 0,
  onPrompt,
  onEditUserMessage,
  onRewind,
  checkpoints = [],
  actionPending = false,
  rewindDisabled = false,
  questionNavigator = true,
  defaultExpandThinking = false,
  followButton = true,
  jobs = [],
}: {
  items: Item[];
  live?: LiveStream;
  footerHeight?: number;
  onPrompt: (text: string) => void;
  onEditUserMessage?: (text: string) => void;
  onRewind?: (turn: number, scope: string) => void;
  checkpoints?: CheckpointMeta[];
  actionPending?: boolean;
  rewindDisabled?: boolean;
  questionNavigator?: boolean;
  defaultExpandThinking?: boolean;
  followButton?: boolean;
  jobs?: JobView[];
}) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const stick = useRef(false);
  const resizeFrame = useRef<number | null>(null);
  const [showFollowButton, setShowFollowButton] = useState(false);
  const [activeJumpTurn, setActiveJumpTurn] = useState<number | null>(null);
  const t = useT();

  const questions = useMemo<QuestionAnchor[]>(() => {
    const anchors: QuestionAnchor[] = [];
    let turn = 0;
    for (const it of items) {
      if (it.kind !== "user") continue;
      anchors.push({ id: it.id, text: compactQuestionText(it.text), turn });
      turn += 1;
    }
    return anchors;
  }, [items]);
  const showQuestionNav = questionNavigator && questions.length >= QUESTION_NAV_MIN_COUNT;

  const updateActiveJumpTurn = useCallback((el: HTMLDivElement | null) => {
    if (!el || questions.length === 0) {
      setActiveJumpTurn(null);
      return;
    }
    const scrollerRect = el.getBoundingClientRect();
    const targetY = scrollerRect.top + Math.min(scrollerRect.height * 0.38, 220);
    let active = questions[0]?.turn ?? null;
    for (const question of questions) {
      const node = document.getElementById(questionAnchorId(question.id));
      if (!node) continue;
      const rect = node.getBoundingClientRect();
      if (rect.top <= targetY) active = question.turn;
      else break;
    }
    setActiveJumpTurn((current) => (current === active ? current : active));
  }, [questions]);

  const updateFollowButton = useCallback((el: HTMLDivElement | null) => {
    if (!el) return;
    setShowFollowButton(followButton && !isAtBottom(el));
  }, [followButton]);

  const scrollToBottom = useCallback((follow: boolean) => {
    const el = scrollRef.current;
    if (!el) return;
    stick.current = follow;
    if (resizeFrame.current !== null) {
      cancelAnimationFrame(resizeFrame.current);
      resizeFrame.current = null;
    }
    requestAnimationFrame(() => {
      scrollElementToBottom(el);
      setShowFollowButton(false);
    });
  }, []);

  const onScroll = () => {
    const el = scrollRef.current;
    if (!el) return;
    if (stick.current && !isAtBottom(el)) stick.current = false;
    updateFollowButton(el);
    updateActiveJumpTurn(el);
  };

  // Track question count so we can detect when the user sends a new message.
  const prevQuestionsLen = useRef(0);

  // When the user submits a new message, reveal it once. Do not keep the
  // viewport locked while the model streams; the user can freely scroll until
  // they explicitly press the follow button.
  useEffect(() => {
    if (questions.length > prevQuestionsLen.current) {
      const el = scrollRef.current;
      if (el) {
        requestAnimationFrame(() => {
          scrollElementToBottom(el);
          stick.current = false;
          updateFollowButton(el);
          updateActiveJumpTurn(el);
        });
      }
    }
    prevQuestionsLen.current = questions.length;
  }, [questions, updateActiveJumpTurn, updateFollowButton]);

  const contentVersion = useMemo(() => scrollVersion(items), [items]);
  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    if (!stick.current) {
      const id = requestAnimationFrame(() => {
        updateFollowButton(el);
        updateActiveJumpTurn(el);
      });
      return () => cancelAnimationFrame(id);
    }
    const id = requestAnimationFrame(() => {
      scrollElementToBottom(el);
      setShowFollowButton(false);
      updateActiveJumpTurn(el);
    });
    return () => cancelAnimationFrame(id);
  }, [contentVersion, live?.text?.length ?? 0, live?.reasoning?.length ?? 0, updateActiveJumpTurn, updateFollowButton]);

  useEffect(() => {
    const el = scrollRef.current;
    if (!el || typeof ResizeObserver === "undefined") return;
    const observer = new ResizeObserver(() => {
      repinIfWasPinned(el, stick, resizeFrame);
      updateFollowButton(el);
      updateActiveJumpTurn(el);
    });
    observer.observe(el);
    return () => {
      observer.disconnect();
      if (resizeFrame.current !== null) {
        cancelAnimationFrame(resizeFrame.current);
        resizeFrame.current = null;
      }
    };
  }, [updateActiveJumpTurn, updateFollowButton]);

  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    repinIfWasPinned(el, stick, resizeFrame);
    updateFollowButton(el);
    updateActiveJumpTurn(el);
    return () => {
      if (resizeFrame.current !== null) {
        cancelAnimationFrame(resizeFrame.current);
        resizeFrame.current = null;
      }
    };
  }, [footerHeight, updateActiveJumpTurn, updateFollowButton]);

  // Sub-agent calls carry a parentId; collect them under their parent `task`
  // call so the parent card can render them nested, and skip them at top level.
  const subcallsByParent = useMemo(() => {
    const m = new Map<string, ToolItem[]>();
    for (const it of items) {
      if (it.kind === "tool" && it.parentId) {
        const arr = m.get(it.parentId) ?? [];
        arr.push(it);
        m.set(it.parentId, arr);
      }
    }
    return m;
  }, [items]);
  const liveToolID = useMemo(() => lastRunningToolID(items), [items]);
  const backgroundTaskJobs = useMemo(() => jobs.filter((job) => job.kind === "task" && job.status === "running"), [jobs]);

  // ── Layer state ────────────────────────────────────────────────────────────
  const [expandedWarmTurns, setExpandedWarmTurns] = useState<Set<number>>(new Set());
  const [coldPage, setColdPage] = useState(0);

  // Compute turn groups (memoised — only rebuilds when user turns change,
  // not on every streaming token). The warm previews are static once built.
  const turnGroupKey = questions.length;
  const turnGroups = useMemo(() => buildTurnGroups(items, questions), [turnGroupKey, questions]);

  // hotStartIdx: first index of the hot zone in items[].
  const hotStartIdx = useMemo(() => {
    let needed = HOT_TURNS;
    for (let i = items.length - 1; i >= 0; i--) {
      if (items[i].kind === "user") {
        needed--;
        if (needed <= 0) return i;
      }
    }
    return 0;
  }, [items]);

  // How many turns are in the cold zone (not yet shown).
  const warmTurnCount = turnGroups.length - Math.min(turnGroups.length, HOT_TURNS);
  const shownWarmStart = Math.max(0, warmTurnCount - coldPage * WARM_PAGE_SIZE);
  const coldTurnCount = shownWarmStart;

  // ── The turn action menu ──────────────────────────────────────────────────
  const [openAction, setOpenAction] = useState<OpenTurnAction | null>(null);
  useEffect(() => {
    if (openAction === null) return;
    const onDown = (e: MouseEvent) => {
      const el = e.target as Element | null;
      if (!el || !el.closest(".turn-actions")) setOpenAction(null);
    };
    document.addEventListener("mousedown", onDown);
    return () => document.removeEventListener("mousedown", onDown);
  }, [openAction]);

  const userTurn = useMemo(() => new Map(questions.map((question) => [question.id, question.turn])), [questions]);
  const checkpointsByTurn = useMemo(() => new Map(checkpoints.map((checkpoint) => [checkpoint.turn, checkpoint])), [checkpoints]);

  // ── JumpBar integration ───────────────────────────────────────────────────
  const jumpToQuestion = (question: QuestionAnchor) => {
    const el = scrollRef.current;
    const node = document.getElementById(questionAnchorId(question.id));
    if (!el || !node) return;
    stick.current = false;
    if (resizeFrame.current !== null) {
      cancelAnimationFrame(resizeFrame.current);
      resizeFrame.current = null;
    }
    const scrollerRect = el.getBoundingClientRect();
    const nodeRect = node.getBoundingClientRect();
    const top = el.scrollTop + nodeRect.top - scrollerRect.top - 12;
    el.scrollTo({ top: Math.max(0, top), behavior: "smooth" });
  };

  const handleJumpToQuestion = useCallback((question: QuestionAnchor) => {
    // Auto-expand the warm turn when jumping to an old question.
    const warmTurnStart = turnGroups.length - HOT_TURNS;
    if (question.turn < warmTurnStart) {
      setExpandedWarmTurns((prev) => {
        if (prev.has(question.turn)) return prev;
        return new Set([...prev, question.turn]);
      });
    }
    jumpToQuestion(question);
  }, [turnGroups.length]);

  // ── Hot zone: fully rendered from hotStartIdx to end ─────────────────────
  // Memoized separately from the assembly so streaming tokens don't rebuild
  // the warm/cold zone JSX trees. Uses LiveStreamContext for streaming data
  // (added by upstream PR #3423) instead of per-call renderSegments.
  const empty = items.length === 0;
  const hotZoneNodes = useMemo<ReactNode[]>(() => {
    const out: ReactNode[] = [];
    let actionText = "";
    let actionReady = false;
    let activeTurn: number | undefined;
    let pendingProcessItems: Item[] = [];
    let pendingAnswerNodes: ReactNode[] = [];
    const flushProcessItems = () => {
      for (const item of pendingProcessItems) {
        const rendered = renderProcessItem(item, subcallsByParent, liveToolID, defaultExpandThinking);
        if (rendered) out.push(rendered);
      }
      pendingProcessItems = [];
    };
    const flushPendingAnswers = () => {
      out.push(...pendingAnswerNodes);
      pendingAnswerNodes = [];
    };
    const pushTurnActions = () => {
      flushProcessItems();
      flushPendingAnswers();
      if (activeTurn == null || !actionReady || actionText.trim() === "") return;
      const turn = activeTurn;
      const openMenu = openAction && openAction.turn === turn ? openAction.menu : null;
      out.push(
        <TurnActions
          key={`ta-${turn}`}
          text={actionText}
          turn={turn}
          openMenu={openMenu}
          onOpenMenu={(menu) => setOpenAction(menu ? { turn, menu } : null)}
          checkpoint={checkpointsByTurn.get(turn)}
          actionPending={actionPending}
          rewindDisabled={rewindDisabled}
          onRewind={(targetTurn, scope) => {
            onRewind?.(targetTurn, scope);
            setOpenAction(null);
          }}
        />,
      );
      actionText = "";
      actionReady = false;
    };

    for (let i = hotStartIdx; i < items.length; i++) {
      const it = items[i];
      switch (it.kind) {
        case "user": {
          pushTurnActions();
          const tn = userTurn.get(it.id);
          activeTurn = tn;
          out.push(
            <UserMessage key={it.id} text={it.text} failed={it.failed} turn={tn} anchorId={questionAnchorId(it.id)} onEdit={onEditUserMessage} />,
          );
          break;
        }
        case "assistant":
          if (it.streaming) {
            flushProcessItems();
            flushPendingAnswers();
            out.push(<LiveAssistantMessage key={it.id} item={it as AssistantItem} defaultExpanded={defaultExpandThinking} />);
          } else if (it.reasoning) {
            pendingProcessItems.push(it);
            if (it.text.trim() !== "") {
              pendingAnswerNodes.push(<AssistantMessage key={`${it.id}-text`} item={{ ...it, reasoning: "" }} defaultExpanded={defaultExpandThinking} />);
            }
          } else {
            flushProcessItems();
            flushPendingAnswers();
            out.push(<AssistantMessage key={it.id} item={it as AssistantItem} defaultExpanded={defaultExpandThinking} />);
          }
          if (!it.streaming && it.text.trim() !== "") {
            actionText = it.text;
            actionReady = true;
          }
          break;
        case "steer":
          flushProcessItems();
          flushPendingAnswers();
          out.push(<SteerCard key={it.id} text={it.text} />);
          break;
        case "tool":
          if (it.parentId) break;
          if (it.name === "todo_write") break;
          if (it.name === "exit_plan_mode") break;
          if (isProcessItemRunning(it)) {
            flushProcessItems();
            out.push(<ToolCard key={it.id} item={it} subcalls={subcallsByParent.get(it.id)} livePulse={it.id === liveToolID} />);
          } else {
            pendingProcessItems.push(it);
          }
          break;
        case "phase":
        case "notice":
        case "compaction":
          if (isProcessItemRunning(it)) {
            flushProcessItems();
            out.push(renderProcessItem(it, subcallsByParent, liveToolID, defaultExpandThinking));
          } else {
            pendingProcessItems.push(it);
          }
          break;
        case "turn_stats": {
          const processItems = pendingProcessItems;
          pendingProcessItems = [];
          out.push(
            <TurnProcessSummary
              key={it.id}
              stats={it}
              items={processItems}
              subcalls={subcallsByParent}
              liveToolID={liveToolID}
              defaultExpandThinking={defaultExpandThinking}
            />,
          );
          flushPendingAnswers();
          break;
        }
      }
    }
    pushTurnActions();
    return out;
  }, [hotStartIdx, items, openAction, actionPending, rewindDisabled, onRewind, subcallsByParent, userTurn, checkpointsByTurn, liveToolID, defaultExpandThinking]);

  // ── Assemble rendered output ──────────────────────────────────────────────
  // Warm/cold zone is a separate memo'd WarmZone component so streaming tokens
  // don't rebuild it. The hot zone uses LiveAssistantMessage (reads live from
  // LiveStreamContext) so streaming updates are captured immediately.
  return (
    <div className="transcript-shell">
      <div
        className={`transcript${empty ? " transcript--empty" : ""}`}
        ref={scrollRef}
        onScroll={onScroll}
      >
        {empty && <Welcome onPrompt={onPrompt} />}

        {!empty && showQuestionNav && (
          <QuestionJumpBar questions={questions} activeTurn={activeJumpTurn} onJump={handleJumpToQuestion} />
        )}

        <LiveStreamContext.Provider value={live}>
          {turnGroups.length > HOT_TURNS && (
            <WarmZone
              turnGroups={turnGroups}
              expandedWarmTurns={expandedWarmTurns}
              shownWarmStart={shownWarmStart}
              coldTurnCount={coldTurnCount}
              scrollRef={scrollRef}
              warmItems={items}
              warmSubcalls={subcallsByParent}
              warmUserTurn={userTurn}
              warmCheckpoints={checkpointsByTurn}
              warmOpenAction={openAction}
              warmActionPending={actionPending}
              warmRewindDisabled={rewindDisabled}
              warmOnRewind={onRewind}
              warmOnEditUserMessage={onEditUserMessage}
              warmSetOpenAction={setOpenAction}
              defaultExpandThinking={defaultExpandThinking}
              liveToolID={liveToolID}
              onToggleColdPage={() => setColdPage((p) => p + 1)}
              onToggleWarmTurn={(g, expand) => {
                setExpandedWarmTurns((prev) => {
                  const next = new Set(prev);
                  if (expand) next.add(g); else next.delete(g);
                  return next;
                });
              }}
            />
          )}
          {hotZoneNodes}
          {backgroundTaskJobs.length > 0 && <BackgroundJobsCard jobs={backgroundTaskJobs} />}
        </LiveStreamContext.Provider>
      </div>

      {!empty && followButton && showFollowButton && (
        <button
          type="button"
          className="transcript-follow"
          aria-label={t("transcript.followLatest")}
          title={t("transcript.followLatest")}
          onClick={() => scrollToBottom(true)}
        >
          <ArrowDown size={18} aria-hidden="true" />
        </button>
      )}
    </div>
  );
}

// ── WarmZone sub-component (React.memo for streaming isolation) ────────────
// Receives structural props only; reads streaming state (items, live) via refs
// so it never invalidates on streaming token arrival.

const WarmZone = memo(function WarmZone({
  turnGroups,
  expandedWarmTurns,
  shownWarmStart,
  coldTurnCount,
  scrollRef,
  warmItems,
  warmSubcalls,
  warmUserTurn,
  warmCheckpoints,
  warmOpenAction,
  warmActionPending,
  warmRewindDisabled,
  warmOnRewind,
  warmOnEditUserMessage,
  warmSetOpenAction,
  defaultExpandThinking = false,
  liveToolID = "",
  onToggleColdPage,
  onToggleWarmTurn,
}: {
  turnGroups: TurnGroup[];
  expandedWarmTurns: ReadonlySet<number>;
  shownWarmStart: number;
  coldTurnCount: number;
  scrollRef: React.RefObject<HTMLDivElement | null>;
  warmItems: readonly Item[];
  warmSubcalls: ReadonlyMap<string, ToolItem[]>;
  warmUserTurn: ReadonlyMap<string, number>;
  warmCheckpoints: ReadonlyMap<number, CheckpointMeta>;
  warmOpenAction: OpenTurnAction | null;
  warmActionPending: boolean;
  warmRewindDisabled: boolean;
  warmOnRewind: ((turn: number, scope: string) => void) | undefined;
  warmOnEditUserMessage?: (text: string) => void;
  warmSetOpenAction: (action: OpenTurnAction | null) => void;
  defaultExpandThinking?: boolean;
  liveToolID?: string;
  onToggleColdPage: () => void;
  onToggleWarmTurn: (g: number, expand: boolean) => void;
}) {
  const t = useT();
  const out: React.ReactNode[] = [];

  // 1. Cold zone: paginated warm turns (show more button).
  if (coldTurnCount > 0) {
    out.push(
      <button
        key="cold-load-more"
        type="button"
        className="warm-collapse"
        onClick={onToggleColdPage}
      >
        {t("transcript.showEarlierHistory", { n: coldTurnCount })}
      </button>,
    );
  }

  // 2. Warm zone: collapsed/expanded warm turn cards.
  let warmStartTurn = 0;
  if (turnGroups.length > HOT_TURNS) {
    warmStartTurn = turnGroups.length - HOT_TURNS - shownWarmStart;
    for (let g = warmStartTurn; g < turnGroups.length - HOT_TURNS; g++) {
      const group = turnGroups[g];
      if (!group) continue;
      const expanded = expandedWarmTurns.has(g);

      if (expanded) {
        const userText = group.userItem.kind === "user" ? group.userItem.text : "";
        out.push(
          <WarmTurnCard
            key={`warm-${g}`}
            userText={warmUserPreview(userText)}
            assistantPreview={group.assistantPreview}
            toolCount={group.toolCount}
            expanded={true}
            onToggle={() => onToggleWarmTurn(g, false)}
          >
            {/* Expanded warm turns render items that are stable (never the
                streaming turn), so this captures items/live via a ref. */}
            <WarmTurnItems
              startIdx={group.startIdx}
              endIdx={group.endIdx}
              items={warmItems}
              subcalls={warmSubcalls}
              userTurnMap={warmUserTurn}
              checkpoints={warmCheckpoints}
              openAction={warmOpenAction}
              actionPending={warmActionPending}
              rewindDisabled={warmRewindDisabled}
              onRewind={warmOnRewind}
              onEditUserMessage={warmOnEditUserMessage}
              setOpenAction={warmSetOpenAction}
              defaultExpandThinking={defaultExpandThinking}
              liveToolID={liveToolID}
            />
          </WarmTurnCard>,
        );
      } else {
        const userText = group.userItem.kind === "user" ? group.userItem.text : "";
        out.push(
          <WarmTurnCard
            key={`warm-${g}`}
            userText={warmUserPreview(userText)}
            assistantPreview={group.assistantPreview}
            toolCount={group.toolCount}
            expanded={false}
            onToggle={() => {
              onToggleWarmTurn(g, true);
              const el = scrollRef.current;
              const node = document.getElementById(questionAnchorId(group.userItem.id));
              if (el && node) {
                requestAnimationFrame(() => {
                  el.scrollTo({ top: node.offsetTop - el.offsetTop - 80, behavior: "smooth" });
                });
              }
            }}
          />,
        );
      }
    }
  }

  return out;
});

function WarmTurnItems({
  startIdx,
  endIdx,
  items,
  subcalls,
  userTurnMap,
  checkpoints,
  openAction,
  actionPending,
  rewindDisabled,
  onRewind,
  onEditUserMessage,
  setOpenAction,
  defaultExpandThinking = false,
  liveToolID = "",
}: {
  startIdx: number;
  endIdx: number;
  items: readonly Item[];
  subcalls: ReadonlyMap<string, ToolItem[]>;
  userTurnMap: ReadonlyMap<string, number>;
  checkpoints: ReadonlyMap<number, CheckpointMeta>;
  openAction: OpenTurnAction | null;
  actionPending: boolean;
  rewindDisabled: boolean;
  onRewind: ((turn: number, scope: string) => void) | undefined;
  onEditUserMessage?: (text: string) => void;
  setOpenAction: (action: OpenTurnAction | null) => void;
  defaultExpandThinking?: boolean;
  liveToolID?: string;
}) {
  const nodes: React.ReactNode[] = [];
  let actionText = "";
  let actionReady = false;
  let activeTurn: number | undefined;
  const pushTurnActions = () => {
    if (activeTurn == null || !actionReady || actionText.trim() === "") return;
    const turn = activeTurn;
    const openMenu = openAction && openAction.turn === turn ? openAction.menu : null;
    nodes.push(
      <TurnActions
        key={`ta-${turn}`}
        text={actionText}
        turn={turn}
        openMenu={openMenu}
        onOpenMenu={(menu) => setOpenAction(menu ? { turn, menu } : null)}
        checkpoint={checkpoints.get(turn)}
        actionPending={actionPending}
        rewindDisabled={rewindDisabled}
        onRewind={(targetTurn, scope) => {
          onRewind?.(targetTurn, scope);
          setOpenAction(null);
        }}
      />,
    );
    actionText = "";
    actionReady = false;
  };

  for (let i = startIdx; i < endIdx && i < items.length; i++) {
    const it = items[i];
    switch (it.kind) {
      case "user": {
        pushTurnActions();
        const tn = userTurnMap.get(it.id);
        activeTurn = tn;
        nodes.push(
          <UserMessage key={it.id} text={it.text} failed={it.failed} turn={tn} anchorId={questionAnchorId(it.id)} onEdit={onEditUserMessage} />,
        );
        break;
      }
      case "assistant": {
        nodes.push(<AssistantMessage key={it.id} item={it} defaultExpanded={defaultExpandThinking} />);
        if (!it.streaming && it.text.trim() !== "") {
          actionText = it.text;
          actionReady = true;
        }
        break;
      }
      case "tool": {
        if (it.parentId) break;
        if (it.name === "todo_write") break;
        if (it.name === "exit_plan_mode") break;
        nodes.push(<ToolCard key={it.id} item={it} subcalls={subcalls.get(it.id)} livePulse={it.id === liveToolID} />);
        break;
      }
      case "phase": nodes.push(<PhaseCard key={it.id} text={it.text} />); break;
      case "steer": nodes.push(<SteerCard key={it.id} text={it.text} />); break;
      case "notice": nodes.push(<NoticeCard key={it.id} level={it.level} text={it.text} />); break;
      case "compaction": nodes.push(<CompactionCard key={it.id} item={it} />); break;
      case "turn_stats": break;
    }
  }
  pushTurnActions();
  return nodes;
}

// ── Warm turn summary card ────────────────────────────────────────────────────

function WarmTurnCard({
  userText,
  assistantPreview,
  toolCount,
  expanded,
  onToggle,
  children,
}: {
  userText: string;
  assistantPreview: string;
  toolCount: number;
  expanded: boolean;
  onToggle: () => void;
  children?: React.ReactNode;
}) {
  const t = useT();
  return (
    <div className={`warm-turn${expanded ? " warm-turn--expanded" : ""}`}>
      <button
        type="button"
        className="warm-turn__head"
        onClick={onToggle}
        aria-expanded={expanded}
      >
        <span className="warm-turn__chevron">
          <ChevronRight className={expanded ? "warm-turn__chevron--open" : ""} size={13} />
        </span>
        <span className="warm-turn__preview">{userText}</span>
        <span className="warm-turn__meta">
          {toolCount > 0 && <span>{t("transcript.toolCount", { n: toolCount })}</span>}
        </span>
      </button>
      {expanded ? (
        <div className="warm-turn__body">{children}</div>
      ) : (
        assistantPreview && <div className="warm-turn__assistant">{assistantPreview}</div>
      )}
    </div>
  );
}

// ── JumpBar, PhaseCard, NoticeCard, CompactionCard ────────────────────────────

function QuestionJumpBar({
  questions,
  activeTurn,
  onJump,
}: {
  questions: QuestionAnchor[];
  activeTurn: number | null;
  onJump: (question: QuestionAnchor) => void;
}) {
  const t = useT();
  const [hovered, setHovered] = useState<number | null>(null);
  const barRef = useRef<HTMLDivElement>(null);
  const previewTop = useRef(0);
  const [showPreview, setShowPreview] = useState(false);

  useEffect(() => {
    if (activeTurn === null) return;
    const el = barRef.current?.querySelector(`[data-turn="${activeTurn}"]`);
    el?.scrollIntoView({ block: "nearest" });
  }, [activeTurn]);

  const hoverIdx = hovered !== null ? questions.findIndex((question) => question.turn === hovered) : -1;
  const hoveredQuestion = hovered !== null ? questions.find((question) => question.turn === hovered) : undefined;

  const closestQuestionFromY = (clientY: number): { question: QuestionAnchor; previewY: number } | null => {
    const el = barRef.current;
    if (!el) return null;
    const markers = el.querySelectorAll<HTMLElement>(".jump-item");
    const barRect = el.getBoundingClientRect();
    let closest = -1;
    let closestDist = Infinity;
    let closestY = 0;
    markers.forEach((item, index) => {
      const rect = item.getBoundingClientRect();
      const midY = rect.top + rect.height / 2;
      const dist = Math.abs(clientY - midY);
      if (dist < closestDist) {
        closestDist = dist;
        closest = index;
        closestY = midY - barRect.top;
      }
    });
    const question = questions[closest];
    if (!question) return null;
    return { question, previewY: closestY };
  };

  const onMove = (e: ReactMouseEvent<HTMLDivElement>) => {
    const closest = closestQuestionFromY(e.clientY);
    if (!closest) return;
    previewTop.current = closest.previewY;
    setHovered(closest.question.turn);
    setShowPreview(true);
  };

  const scrollTo = (question: QuestionAnchor) => {
    onJump(question);
  };

  const onRailMouseDown = (e: ReactMouseEvent<HTMLDivElement>) => {
    const closest = closestQuestionFromY(e.clientY);
    if (!closest) return;
    e.preventDefault();
    previewTop.current = closest.previewY;
    setHovered(closest.question.turn);
    setShowPreview(true);
    scrollTo(closest.question);
  };

  const onItemMouseDown = (e: ReactMouseEvent<HTMLButtonElement>, question: QuestionAnchor) => {
    e.preventDefault();
    scrollTo(question);
  };

  const dotProps = (
    idx: number,
    turn: number,
  ): { style: CSSProperties; "data-d"?: string } => {
    const isActive = activeTurn === turn;
    if (hoverIdx < 0) {
      return { style: { width: isActive ? 18 : 12, background: isActive ? "var(--accent)" : undefined } };
    }
    const d = Math.abs(idx - hoverIdx);
    const width = d === 0 ? 32 : d === 1 ? 20 : d === 2 ? 14 : isActive ? 18 : 12;
    const background = d <= 2 ? undefined : isActive ? "var(--accent)" : undefined;
    return {
      style: { width, transitionDelay: `${d * 20}ms`, background },
      "data-d": d <= 2 ? String(d) : undefined,
    };
  };

  return (
    <nav
      className="jump-bar"
      ref={barRef}
      aria-label={t("questionNav.label")}
      onMouseMove={onMove}
      onMouseLeave={() => {
        setHovered(null);
        setShowPreview(false);
      }}
    >
      <div className="jump-scroll" onMouseDown={onRailMouseDown} onClick={onRailMouseDown}>
        {questions.map((question, index) => (
          <button
            className="jump-item"
            key={question.id}
            type="button"
            data-turn={question.turn}
            aria-label={t("questionNav.jump", { n: question.turn + 1 })}
            onMouseDown={(e) => onItemMouseDown(e, question)}
            onClick={(e) => {
              e.stopPropagation();
              if (e.detail === 0) scrollTo(question);
            }}
          >
            <span className="jump-dot" {...dotProps(index, question.turn)} />
          </button>
        ))}
      </div>
      {showPreview && hoveredQuestion && (
        <div className="jump-preview" style={{ top: previewTop.current }} role="tooltip">
          <span className="jump-text">{hoveredQuestion.text}</span>
        </div>
      )}
    </nav>
  );
}

type CompactionItem = Extract<Item, { kind: "compaction" }>;
type NoticeItem = Extract<Item, { kind: "notice" }>;

function PhaseCard({ text }: { text: string }) {
  return (
    <ProcessCard
      tone="accent"
      icon={<ProcessPhaseIcon size={12} />}
      kind="phase"
      name={text}
      className="phase process-card--phase"
    />
  );
}

function NoticeCard({ level, text }: { level: NoticeItem["level"]; text: string }) {
  const t = useT();
  const warning = level === "warn";
  return (
    <ProcessCard
      tone={warning ? "warning" : "default"}
      icon={<ProcessInfoIcon size={12} />}
      kind="notice"
      name={t(warning ? "notice.warning" : "notice.info")}
      meta={warning ? <ProcessStatusIcon state="waiting" label={t("notice.warning")} /> : undefined}
      className={`notice notice--${level}`}
    >
      <div className="notice__body">{text}</div>
    </ProcessCard>
  );
}

function SteerCard({ text }: { text: string }) {
  const t = useT();
  return (
    <ProcessCard
      tone="accent"
      icon={<ChevronRight size={12} />}
      kind={t("steer.kind")}
      name={t("steer.title")}
      defaultOpen
      className="steer"
    >
      <div className="steer__body">{text}</div>
    </ProcessCard>
  );
}

function CompactionCard({ item }: { item: CompactionItem }) {
  const t = useT();
  const trigger = item.trigger === "manual" ? "manual" : item.trigger === "auto" ? "auto" : item.trigger;
  if (item.pending) {
    return (
      <ProcessCard
        tone="accent"
        icon={<ProcessCompactIcon size={12} />}
        kind="context"
        name={t("compaction.working")}
        meta={<ProcessStatusIcon state="running" label={t("compaction.working")} />}
        className="compaction compaction--pending"
      />
    );
  }
  return (
    <ProcessCard
      tone="accent"
      icon={<ProcessCompactIcon size={12} />}
      kind="context"
      name={t("compaction.title")}
      meta={`${t("compaction.messages", { n: item.messages })}${trigger ? ` - ${trigger}` : ""}`}
      className="compaction"
    >
      <pre className="compaction__summary">{item.summary}</pre>
    </ProcessCard>
  );
}
