package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"deepseek-orca/internal/agent"
	"deepseek-orca/internal/provider"
)

type automationSegment struct {
	Path      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func automationTopicSegments(topicID string) []automationSegment {
	dir := desktopSessionDir(automationWorkspaceRoot())
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	segments := make([]automationSegment, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".jsonl") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		meta, ok, err := agent.LoadBranchMeta(path)
		if err != nil || !ok || meta.Scope != scopeAutomation || meta.TopicID != topicID {
			continue
		}
		updated := meta.UpdatedAt
		created := meta.CreatedAt
		if stat, statErr := os.Stat(path); statErr == nil {
			if updated.IsZero() {
				updated = stat.ModTime()
			}
			if created.IsZero() {
				created = stat.ModTime()
			}
		}
		segments = append(segments, automationSegment{Path: path, CreatedAt: created, UpdatedAt: updated})
	}
	sort.SliceStable(segments, func(i, j int) bool {
		ai, aj := segments[i].CreatedAt, segments[j].CreatedAt
		if ai.Equal(aj) {
			return segments[i].Path < segments[j].Path
		}
		return ai.Before(aj)
	})
	return segments
}

type automationHistoryTool struct {
	topicID string
}

func (automationHistoryTool) Name() string { return "automation_history" }
func (automationHistoryTool) Description() string {
	return "Search and read earlier logical segments from the Orca automation conversation. Use only when the current request actually depends on older Orca context. Read-only."
}
func (automationHistoryTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"query":{"type":"string","description":"Optional words to search in earlier user and assistant messages."},"limit":{"type":"integer","minimum":1,"maximum":20}},"additionalProperties":false}`)
}
func (automationHistoryTool) ReadOnly() bool { return true }
func (t automationHistoryTool) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var input struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	_ = json.Unmarshal(args, &input)
	if input.Limit <= 0 || input.Limit > 20 {
		input.Limit = 8
	}
	query := strings.ToLower(strings.TrimSpace(input.Query))
	type result struct {
		Segment   int    `json:"segment"`
		UpdatedAt string `json:"updatedAt,omitempty"`
		Role      string `json:"role"`
		Content   string `json:"content"`
	}
	results := make([]result, 0, input.Limit)
	segments := automationTopicSegments(t.topicID)
	for segmentIndex := len(segments) - 1; segmentIndex >= 0 && len(results) < input.Limit; segmentIndex-- {
		session, err := agent.LoadSession(segments[segmentIndex].Path)
		if err != nil {
			continue
		}
		messages := session.Snapshot()
		for messageIndex := len(messages) - 1; messageIndex >= 0 && len(results) < input.Limit; messageIndex-- {
			message := messages[messageIndex]
			if message.Role != provider.RoleUser && message.Role != provider.RoleAssistant {
				continue
			}
			content := strings.TrimSpace(message.Content)
			if content == "" || (query != "" && !strings.Contains(strings.ToLower(content), query)) {
				continue
			}
			if len([]rune(content)) > 1200 {
				content = string([]rune(content)[:1200]) + "..."
			}
			results = append(results, result{Segment: segmentIndex + 1, UpdatedAt: segments[segmentIndex].UpdatedAt.Format(time.RFC3339), Role: string(message.Role), Content: content})
		}
	}
	payload, err := json.MarshalIndent(map[string]any{"query": input.Query, "results": results}, "", "  ")
	return string(payload), err
}
