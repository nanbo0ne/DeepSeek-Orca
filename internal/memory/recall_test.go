package memory

import (
	"strings"
	"testing"
)

func TestAssistantRecallBlockIncludesDetailsAndBudgetNotice(t *testing.T) {
	userDir := t.TempDir()
	cwd := t.TempDir()
	store := AssistantStoreFor(userDir, cwd)
	if _, err := store.Save(Memory{
		Name:        "short-style",
		Description: "User prefers short answers",
		Body:        "Prefer concise answers unless the task needs detail.",
		Confidence:  0.9,
		UpdatedAt:   "2026-06-21T02:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(Memory{
		Name:        "long-style",
		Description: "Long detail",
		Body:        strings.Repeat("detail ", 200),
		Confidence:  0.8,
		UpdatedAt:   "2026-06-21T01:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	set := Load(Options{CWD: cwd, UserDir: userDir, Profile: ProfileAssistant})
	block := set.AssistantRecallBlock(260)
	if !strings.Contains(block, "short-style.md") || !strings.Contains(block, "Prefer concise answers") {
		t.Fatalf("recall block should include index and fitting body:\n%s", block)
	}
	if !strings.Contains(block, "omitted") {
		t.Fatalf("recall block should mention omitted details when over budget:\n%s", block)
	}
}
