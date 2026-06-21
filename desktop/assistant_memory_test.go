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
