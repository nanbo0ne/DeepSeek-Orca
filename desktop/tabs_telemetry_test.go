package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"deepseek-orca/internal/agent"
	"deepseek-orca/internal/control"
	"deepseek-orca/internal/event"
	"deepseek-orca/internal/provider"
	"deepseek-orca/internal/tool"
)

type usageProvider struct {
	usage *provider.Usage
}

func (p usageProvider) Name() string { return "usage" }

func (p usageProvider) Stream(_ context.Context, _ provider.Request) (<-chan provider.Chunk, error) {
	ch := make(chan provider.Chunk, 2)
	ch <- provider.Chunk{Type: provider.ChunkText, Text: "ok"}
	ch <- provider.Chunk{Type: provider.ChunkUsage, Usage: p.usage}
	close(ch)
	return ch, nil
}

func TestTelemetryLoadsLegacyReadFileArray(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl.telemetry.json")
	if err := os.WriteFile(path, []byte(`[{"path":"README.md","turn":2,"time":1000}]`), 0o644); err != nil {
		t.Fatalf("write legacy telemetry: %v", err)
	}

	got := loadTelemetry(path)
	if len(got.ReadFiles) != 1 || got.ReadFiles[0].Path != "README.md" {
		t.Fatalf("legacy read files = %+v", got.ReadFiles)
	}
	if got.Usage.RequestCount != 0 {
		t.Fatalf("legacy usage request count = %d, want 0", got.Usage.RequestCount)
	}
}

func TestWorkspaceTabAggregatesSessionUsageTelemetry(t *testing.T) {
	tab := &WorkspaceTab{}
	start := time.Now().Add(-2 * time.Second).UnixMilli()
	tab.recordTurnStarted(0, start)
	tab.recordUsage(event.Event{
		Usage:       &provider.Usage{PromptTokens: 100, CompletionTokens: 40, TotalTokens: 140, CacheHitTokens: 70, CacheMissTokens: 30, ReasoningTokens: 10},
		SessionHit:  70,
		SessionMiss: 30,
		Pricing:     &provider.Pricing{CacheHit: 1, Input: 2, Output: 3, Currency: "¥"},
	})
	tab.recordTurnDone(start + 1500)

	got := tab.telemetrySnapshot().Usage
	if got.RequestCount != 1 || got.PromptTokens != 100 || got.CompletionTokens != 40 || got.TotalTokens != 140 || got.ReasoningTokens != 10 {
		t.Fatalf("usage tokens = %+v", got)
	}
	if got.CacheHitTokens != 70 || got.CacheMissTokens != 30 {
		t.Fatalf("cache tokens = hit %d miss %d", got.CacheHitTokens, got.CacheMissTokens)
	}
	if got.ElapsedMs != 1500 {
		t.Fatalf("elapsed = %d, want 1500", got.ElapsedMs)
	}
	if got.SessionCost <= 0 || got.SessionCurrency != "¥" {
		t.Fatalf("cost = %f %q, want positive ¥", got.SessionCost, got.SessionCurrency)
	}

	app := &App{tabs: map[string]*WorkspaceTab{"tab": tab}}
	if context := app.ContextUsageForTab("tab"); context.SessionTokens != 140 {
		t.Fatalf("context usage session tokens = %d, want 140", context.SessionTokens)
	} else if context.SessionCacheHitTokens != 70 || context.SessionCacheMissTokens != 30 {
		t.Fatalf("context cache tokens = hit %d miss %d, want 70/30", context.SessionCacheHitTokens, context.SessionCacheMissTokens)
	} else if context.SessionCost <= 0 || context.SessionCurrency == "" {
		t.Fatalf("context cost = %f %q, want positive non-empty currency", context.SessionCost, context.SessionCurrency)
	}
	if panel := app.ContextPanel("tab"); panel.TotalTokens != 140 {
		t.Fatalf("context panel total tokens = %d, want 140", panel.TotalTokens)
	}
}

func TestContextUsageRestoresLastUsageEventAfterRestart(t *testing.T) {
	tab := &WorkspaceTab{}
	tab.recordUsage(event.Event{
		Usage: &provider.Usage{
			PromptTokens:     100,
			CompletionTokens: 40,
			TotalTokens:      140,
			CacheHitTokens:   70,
			CacheMissTokens:  30,
			ReasoningTokens:  10,
		},
		SessionHit:  70,
		SessionMiss: 30,
		Pricing:     &provider.Pricing{CacheHit: 1, Input: 2, Output: 3, Currency: "¥"},
	})
	snapshot := tab.telemetrySnapshot()

	restored := &WorkspaceTab{
		ID:                   "tab",
		readTelemetry:        snapshot.ReadFiles,
		usageTelemetry:       snapshot.Usage,
		usageTelemetryEvents: snapshot.UsageEvents,
	}
	app := &App{tabs: map[string]*WorkspaceTab{"tab": restored}}
	context := app.ContextUsageForTab("tab")
	if context.TotalTokens != 140 || context.PromptTokens != 100 || context.CompletionTokens != 40 || context.ReasoningTokens != 10 {
		t.Fatalf("restored last usage = %+v, want latest event tokens", context)
	}
	if context.CacheHitTokens != 70 || context.CacheMissTokens != 30 || context.SessionCacheHitTokens != 70 || context.SessionCacheMissTokens != 30 {
		t.Fatalf("restored cache = %+v, want latest and session cache", context)
	}
	if context.SessionCost <= 0 || context.SessionCurrency == "" {
		t.Fatalf("restored cost = %f %q, want persisted cost with currency", context.SessionCost, context.SessionCurrency)
	}
}

func TestBlankConversationContextUsageDisplaysZero(t *testing.T) {
	ag := agent.New(
		usageProvider{usage: &provider.Usage{PromptTokens: 99, CompletionTokens: 1, TotalTokens: 100}},
		tool.NewRegistry(),
		agent.NewSession("system prompt and tools are loaded"),
		agent.Options{ContextWindow: 200000},
		event.Discard,
	)
	tab := &WorkspaceTab{
		ID:    "tab",
		Ctrl:  control.New(control.Options{Executor: ag, Sink: event.Discard}),
		Scope: "global",
		Ready: true,
	}
	app := &App{tabs: map[string]*WorkspaceTab{"tab": tab}}

	context := app.ContextUsageForTab("tab")
	if context.Used != 0 {
		t.Fatalf("blank context used = %d, want 0", context.Used)
	}
	if context.Window == 0 {
		t.Fatalf("blank context window = 0, want model window retained")
	}
	panel := app.ContextPanel("tab")
	if panel.UsedTokens != 0 {
		t.Fatalf("blank panel used = %d, want 0", panel.UsedTokens)
	}
	if panel.WindowTokens == 0 {
		t.Fatalf("blank panel window = 0, want model window retained")
	}
}

func TestWorkspaceTabRewindTelemetryPrunesUsage(t *testing.T) {
	tab := &WorkspaceTab{}
	tab.recordTurnStarted(0, 1000)
	tab.recordUsage(event.Event{Usage: &provider.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15}})
	tab.recordTurnDone(1100)
	tab.recordTurnStarted(1, 1200)
	tab.recordUsage(event.Event{Usage: &provider.Usage{PromptTokens: 20, CompletionTokens: 6, TotalTokens: 26}})
	tab.recordTurnDone(1300)

	tab.rewindTelemetryBefore(1)

	got := tab.telemetrySnapshot().Usage
	if got.RequestCount != 1 || got.TotalTokens != 15 || got.PromptTokens != 10 || got.CompletionTokens != 5 {
		t.Fatalf("rewound usage = %+v, want only first turn", got)
	}
}

func TestContextPanelUsesLastUsageBreakdownWithTelemetryTotal(t *testing.T) {
	lastUsage := &provider.Usage{
		PromptTokens:     10,
		CompletionTokens: 4,
		TotalTokens:      14,
		CacheHitTokens:   7,
		CacheMissTokens:  3,
		ReasoningTokens:  2,
	}
	ag := agent.New(
		usageProvider{usage: lastUsage},
		tool.NewRegistry(),
		agent.NewSession("system"),
		agent.Options{ContextWindow: 200},
		event.Discard,
	)
	if err := ag.Run(context.Background(), "hello"); err != nil {
		t.Fatal(err)
	}
	tab := &WorkspaceTab{
		ID:    "tab",
		Ctrl:  control.New(control.Options{Executor: ag, Sink: event.Discard}),
		Scope: "global",
		Ready: true,
	}
	tab.recordUsage(event.Event{
		Usage: &provider.Usage{
			PromptTokens:     100,
			CompletionTokens: 40,
			TotalTokens:      140,
			CacheHitTokens:   70,
			CacheMissTokens:  30,
			ReasoningTokens:  10,
		},
	})
	app := &App{tabs: map[string]*WorkspaceTab{"tab": tab}}

	panel := app.ContextPanel("tab")
	if panel.TotalTokens != 140 {
		t.Fatalf("context panel total tokens = %d, want telemetry total 140", panel.TotalTokens)
	}
	if panel.PromptTokens != 10 || panel.CompletionTokens != 4 || panel.ReasoningTokens != 2 {
		t.Fatalf("context panel breakdown = prompt:%d completion:%d reasoning:%d, want last usage 10/4/2",
			panel.PromptTokens, panel.CompletionTokens, panel.ReasoningTokens)
	}
	if panel.CacheHitTokens != 7 || panel.CacheMissTokens != 3 {
		t.Fatalf("context panel cache breakdown = hit:%d miss:%d, want last usage 7/3",
			panel.CacheHitTokens, panel.CacheMissTokens)
	}
}
