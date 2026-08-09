package main

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"deepseek-orca/internal/agent"
	"deepseek-orca/internal/provider"
)

func TestAutomationHistoryExcludesCurrentOrcaSegment(t *testing.T) {
	isolateDesktopUserDirs(t)
	root := automationWorkspaceRoot()
	dir := desktopSessionDir(root)
	topicID := "orca-main"
	oldPath := filepath.Join(dir, "old.jsonl")
	currentPath := filepath.Join(dir, "current.jsonl")
	for i, item := range []struct {
		path    string
		content string
	}{
		{oldPath, "older Orca context"},
		{currentPath, "current Orca context"},
	} {
		session := agent.NewSession("system")
		session.Add(provider.Message{Role: provider.RoleUser, Content: item.content})
		if err := session.Save(item.path); err != nil {
			t.Fatal(err)
		}
		stamp := time.Now().Add(time.Duration(i) * time.Minute)
		if err := agent.SaveBranchMetaPreserveUpdated(item.path, agent.BranchMeta{
			ID: agent.BranchID(item.path), Scope: scopeAutomation, WorkspaceRoot: root,
			TopicID: topicID, TopicTitle: automationMainTopicTitle, CreatedAt: stamp, UpdatedAt: stamp,
		}); err != nil {
			t.Fatal(err)
		}
	}

	tool := automationHistoryTool{topicID: topicID, currentPath: func() string { return currentPath }}
	out, err := tool.Execute(context.Background(), json.RawMessage(`{"limit":20}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "older Orca context") {
		t.Fatalf("history output omitted older segment: %s", out)
	}
	if strings.Contains(out, "current Orca context") {
		t.Fatalf("history output leaked current segment: %s", out)
	}
}
