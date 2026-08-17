package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/agent"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/provider"
)

func TestLeakedBootReviewFixtureRequiresCompleteSignature(t *testing.T) {
	fixture := leakedBootReviewTestMessages()
	if !isLeakedBootReviewFixture(fixture) {
		t.Fatal("exact leaked boot fixture was not recognized")
	}
	fixture[1].Content = "a real first review conversation"
	if isLeakedBootReviewFixture(fixture) {
		t.Fatal("ordinary conversation was classified as a leaked fixture")
	}
}

func TestCleanupLeakedBootReviewSessionsRemovesFixtureOnly(t *testing.T) {
	isolateDesktopUserDirs(t)
	dir := filepath.Join(t.TempDir(), "sessions")
	fixturePath := filepath.Join(dir, "fixture.jsonl")
	realPath := filepath.Join(dir, "real.jsonl")
	for path, messages := range map[string][]provider.Message{
		fixturePath: leakedBootReviewTestMessages(),
		realPath:    {{Role: provider.RoleUser, Content: "first review"}, {Role: provider.RoleAssistant, Content: "real answer"}},
	} {
		session := agent.NewSession("")
		for _, message := range messages {
			session.Add(message)
		}
		if err := session.Save(path); err != nil {
			t.Fatal(err)
		}
	}
	app := NewApp()
	app.tabs = map[string]*WorkspaceTab{}
	app.cleanupLeakedBootReviewSessionsInDirs([]string{dir})
	if _, err := os.Stat(fixturePath); !os.IsNotExist(err) {
		t.Fatalf("fixture session still exists: %v", err)
	}
	if _, err := os.Stat(realPath); err != nil {
		t.Fatalf("real session was removed: %v", err)
	}
}

func leakedBootReviewTestMessages() []provider.Message {
	return []provider.Message{
		{Role: provider.RoleSystem, Content: "BASE"},
		{Role: provider.RoleUser, Content: "first review"},
		{Role: provider.RoleAssistant, ToolCalls: []provider.ToolCall{{ID: "review-1", Name: "review", Arguments: `{"task":"first skill task"}`}}},
		{Role: provider.RoleAssistant, Content: "parent first done"},
		{Role: provider.RoleUser, Content: "continue review"},
		{Role: provider.RoleAssistant, ToolCalls: []provider.ToolCall{{ID: "review-2", Name: "review", Arguments: `{"task":"second skill task"}`}}},
		{Role: provider.RoleAssistant, Content: "parent second done"},
	}
}
