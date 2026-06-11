import type { ContextInfo, ContextPanelInfo, WireUsage } from "./types";

export interface ContextPanelUsageSummary {
  usedTokens: number;
  windowTokens: number;
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
  reasoningTokens: number;
  cacheHitTokens: number;
  cacheMissTokens: number;
}

interface ContextPanelUsageInput {
  context?: ContextInfo;
  info?: ContextPanelInfo | null;
  usage?: WireUsage;
  sessionTokens?: number;
}

function positive(value?: number): number {
  return typeof value === "number" && value > 0 ? value : 0;
}

function inputTokensFromUsage(usage?: WireUsage): number {
  const promptTokens = positive(usage?.promptTokens);
  if (promptTokens > 0) return promptTokens;
  return positive(usage?.cacheHitTokens) + positive(usage?.cacheMissTokens);
}

export function computeContextPanelUsage({
  context,
  info,
  usage,
  sessionTokens,
}: ContextPanelUsageInput): ContextPanelUsageSummary {
  const hasPanelBreakdown = Boolean(
    positive(info?.promptTokens) > 0 ||
    positive(info?.completionTokens) > 0 ||
    positive(info?.reasoningTokens) > 0 ||
    positive(info?.cacheHitTokens) > 0 ||
    positive(info?.cacheMissTokens) > 0
  );

  const promptTokens = hasPanelBreakdown ? positive(info?.promptTokens) : inputTokensFromUsage(usage);
  const completionTokens = hasPanelBreakdown ? positive(info?.completionTokens) : positive(usage?.completionTokens);
  const reasoningTokens = hasPanelBreakdown ? positive(info?.reasoningTokens) : positive(usage?.reasoningTokens);
  const cacheHitTokens = hasPanelBreakdown ? positive(info?.cacheHitTokens) : positive(usage?.cacheHitTokens);
  const cacheMissTokens = hasPanelBreakdown ? positive(info?.cacheMissTokens) : positive(usage?.cacheMissTokens);

  const totalTokens =
    positive(info?.totalTokens) ||
    positive(sessionTokens) ||
    positive(usage?.totalTokens) ||
    promptTokens + completionTokens;

  const windowTokens = positive(context?.window) || positive(info?.windowTokens);
  const knownContextUse = positive(context?.used) || positive(info?.usedTokens) || promptTokens;
  const approximateContextUse =
    knownContextUse ||
    (windowTokens > 0 && totalTokens > 0 ? Math.min(totalTokens, windowTokens) : 0);

  return {
    usedTokens: approximateContextUse,
    windowTokens,
    promptTokens,
    completionTokens,
    totalTokens,
    reasoningTokens,
    cacheHitTokens,
    cacheMissTokens,
  };
}
