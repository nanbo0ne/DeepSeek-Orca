package main

import (
	"testing"
	"time"

	"deepseek-orca/internal/provider"
)

func TestAssistantMemoryRealMessageSkipsInjectedContext(t *testing.T) {
	cases := []provider.Message{
		{Role: provider.RoleUser, Content: "<system-reminder>\nsecret\n</system-reminder>"},
		{Role: provider.RoleUser, Content: "<workflow-reminder>\nsteps\n</workflow-reminder>"},
		{Role: provider.RoleUser, Content: "<context-checkpoint>\nsummary\n</context-checkpoint>"},
		{Role: provider.RoleTool, Content: "tool output"},
	}
	for _, msg := range cases {
		if assistantMemoryRealMessage(msg) {
			t.Fatalf("message should be skipped: %+v", msg)
		}
	}
	if !assistantMemoryRealMessage(provider.Message{Role: provider.RoleUser, Content: "I prefer concise replies."}) {
		t.Fatal("real user message should be included")
	}
	if !assistantMemoryRealMessage(provider.Message{Role: provider.RoleAssistant, Content: "Noted."}) {
		t.Fatal("real assistant message should be included")
	}
}

func TestAssistantMemoryUpdateAllowedFiltersLowConfidenceAndSensitive(t *testing.T) {
	good := assistantMemoryUpdate{
		Action: "create", Description: "User prefers concise replies", Body: "Keep future replies concise.", Confidence: 0.82,
	}
	if !assistantMemoryUpdateAllowed(good) {
		t.Fatal("high-confidence durable preference should be allowed")
	}
	low := good
	low.Confidence = 0.5
	if assistantMemoryUpdateAllowed(low) {
		t.Fatal("low-confidence memory should be rejected")
	}
	sensitive := good
	sensitive.Body = "The user's API key is abc."
	if assistantMemoryUpdateAllowed(sensitive) {
		t.Fatal("sensitive memory should be rejected")
	}
}

func TestFinishAssistantMemoryPendingPreservesNewerPendingAndAdvancesCursor(t *testing.T) {
	isolateDesktopUserDirs(t)
	app := NewApp()
	key := "session.jsonl"
	newer := time.Now().UnixMilli()
	if err := saveAssistantMemoryPendingFile(assistantMemoryPendingFile{Items: map[string]assistantMemoryPendingItem{
		key: {
			SessionPath:           key,
			LastProcessedMessages: 2,
			MarkedAt:              newer,
			Status:                "pending",
		},
	}}); err != nil {
		t.Fatal(err)
	}
	item := assistantMemoryPendingItem{
		SessionPath:           key,
		LastProcessedMessages: 5,
		MarkedAt:              newer - 1000,
		Status:                "running",
	}
	app.finishAssistantMemoryPending(key, item, nil)
	got := loadAssistantMemoryPendingFile().Items[key]
	if got.Status != "pending" {
		t.Fatalf("newer pending marker should remain pending, got %+v", got)
	}
	if got.LastProcessedMessages != 5 {
		t.Fatalf("cursor should advance to avoid duplicate processing, got %+v", got)
	}
}

func TestAssistantMemoryFailedRetryBackoffAndLimit(t *testing.T) {
	isolateDesktopUserDirs(t)
	app := NewApp()
	key := "session.jsonl"
	now := time.Now()
	item := assistantMemoryPendingItem{
		SessionPath:   key,
		Status:        "failed",
		RetryCount:    1,
		LastErrorAt:   now.UnixMilli(),
		LastAttemptAt: now.Add(-time.Minute).UnixMilli(),
	}
	if err := saveAssistantMemoryPendingFile(assistantMemoryPendingFile{Items: map[string]assistantMemoryPendingItem{key: item}}); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := app.claimNextAssistantMemoryPending(); ok {
		t.Fatal("failed item should not be claimed before retry delay")
	}
	item.LastErrorAt = now.Add(-11 * time.Minute).UnixMilli()
	if err := saveAssistantMemoryPendingFile(assistantMemoryPendingFile{Items: map[string]assistantMemoryPendingItem{key: item}}); err != nil {
		t.Fatal(err)
	}
	_, claimed, ok := app.claimNextAssistantMemoryPending()
	if !ok || claimed.Status != "running" {
		t.Fatalf("failed item should be claimed after retry delay, got ok=%v item=%+v", ok, claimed)
	}
	app.finishAssistantMemoryPending(key, claimed, assertErr("still broken"))
	got := loadAssistantMemoryPendingFile().Items[key]
	if got.Status != "failed" || got.RetryCount != 2 || got.LastErrorAt == 0 {
		t.Fatalf("failed retry state = %+v, want failed retryCount=2 with lastErrorAt", got)
	}

	got.RetryCount = assistantMemoryMaxRetries - 1
	got.Status = "running"
	app.finishAssistantMemoryPending(key, got, assertErr("final failure"))
	final := loadAssistantMemoryPendingFile().Items[key]
	if final.Status != "ignored" || final.RetryCount != assistantMemoryMaxRetries {
		t.Fatalf("final retry state = %+v, want ignored after max retries", final)
	}
}

type assertErr string

func (e assertErr) Error() string { return string(e) }
