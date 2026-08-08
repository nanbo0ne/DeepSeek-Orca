package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"deepseek-orca/internal/agent"
	"deepseek-orca/internal/bot"
	"deepseek-orca/internal/event"
)

type botExternalUserEvent struct {
	Kind  string `json:"kind"`
	Text  string `json:"text,omitempty"`
	TabID string `json:"tabId"`
}

type botSessionUpdatedEvent struct {
	Kind  string `json:"kind"`
	TabID string `json:"tabId"`
}

func (a *App) createDesktopBotSession(ctx context.Context, remoteKey string, msg bot.InboundMessage) (bot.SessionChoice, error) {
	mainTopic, err := a.ensureAutomationMainTopic()
	if err != nil {
		return bot.SessionChoice{}, err
	}
	topicID := mainTopic.ID
	root := automationWorkspaceRoot()
	if err := os.MkdirAll(root, 0o755); err != nil {
		return bot.SessionChoice{}, err
	}
	dir := desktopSessionDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return bot.SessionChoice{}, err
	}

	title := automationMainTopicTitle

	path := agent.NewSessionPath(dir, "bot")
	sess := agent.NewSession("")
	if err := sess.Save(path); err != nil {
		return bot.SessionChoice{}, err
	}
	now := time.Now()
	_ = agent.SaveBranchMetaPreserveUpdated(path, agent.BranchMeta{
		ID:            agent.BranchID(path),
		Scope:         scopeAutomation,
		WorkspaceRoot: root,
		TopicID:       topicID,
		TopicTitle:    title,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	a.emitProjectTreeChanged()
	_ = a.rememberBotAutomationSession(msg.Platform, remoteKey, topicID)

	return bot.SessionChoice{
		TopicID:       topicID,
		Path:          path,
		Title:         title,
		Location:      "自动化工作区",
		WorkspaceRoot: root,
		LastActivity:  now,
	}, nil
}

func (a *App) restoreDesktopBotSession(_ context.Context, remoteKey string, msg bot.InboundMessage) (bot.SessionChoice, bool, error) {
	mainTopic, err := a.ensureAutomationMainTopic()
	if err != nil {
		return bot.SessionChoice{}, false, err
	}
	topicID := mainTopic.ID
	path, _ := a.findKnownTopicSession(topicID)
	if strings.TrimSpace(path) == "" {
		return bot.SessionChoice{}, false, nil
	}
	meta, ok, err := agent.LoadBranchMeta(path)
	if err != nil || !ok || meta.Scope != scopeAutomation {
		return bot.SessionChoice{}, false, err
	}
	lastActivity := meta.UpdatedAt
	if lastActivity.IsZero() {
		if stat, statErr := os.Stat(path); statErr == nil {
			lastActivity = stat.ModTime()
		}
	}
	title := automationMainTopicTitle
	_ = a.rememberBotAutomationSession(msg.Platform, remoteKey, topicID)
	return bot.SessionChoice{
		TopicID: topicID, Path: path, Title: title, Location: "自动化工作区",
		WorkspaceRoot: automationWorkspaceRoot(), LastActivity: lastActivity,
	}, true, nil
}

func (a *App) afterDesktopBotTurn(sessionPath, model string) {
	a.refreshOpenTabForBotSession(sessionPath)
	meta, ok, _ := agent.LoadBranchMeta(sessionPath)
	if !ok || meta.Scope != scopeAutomation {
		return
	}
	a.markAssistantMemoryPendingForCandidate(assistantMemoryCandidate{
		SessionPath: sessionPath, TopicID: meta.TopicID, WorkspaceRoot: meta.WorkspaceRoot,
		PromptMode: promptModeAssistant, Model: model,
	}, true)
}

func (a *App) mirrorBotUserToOpenTab(sessionPath, text string) {
	tabID := a.openTabIDForSession(sessionPath)
	if tabID == "" || a.ctx == nil {
		return
	}
	wruntime.EventsEmit(a.ctx, eventChannel, botExternalUserEvent{
		Kind:  "external_user",
		Text:  text,
		TabID: tabID,
	})
}

func (a *App) mirrorBotEventToOpenTab(sessionPath string, e event.Event) {
	tabID := a.openTabIDForSession(sessionPath)
	if tabID == "" || a.ctx == nil {
		return
	}
	wruntime.EventsEmit(a.ctx, eventChannel, toWireTab(e, tabID))
}

func (a *App) refreshOpenTabForBotSession(sessionPath string) {
	tab := a.openTabForSession(sessionPath)
	if tab == nil || tab.Ctrl == nil || tab.Ctrl.Running() {
		return
	}
	loaded, err := agent.LoadSession(sessionPath)
	if err != nil {
		return
	}
	tab.Ctrl.Resume(loaded, sessionPath)
	a.rememberTabSessionPath(tab, sessionPath)
	a.emitProjectTreeChanged()
	if a.ctx != nil {
		wruntime.EventsEmit(a.ctx, eventChannel, botSessionUpdatedEvent{
			Kind:  "session_updated",
			TabID: tab.ID,
		})
	}
}

func (a *App) openTabIDForSession(sessionPath string) string {
	if tab := a.openTabForSession(sessionPath); tab != nil {
		return tab.ID
	}
	return ""
}

func (a *App) openTabForSession(sessionPath string) *WorkspaceTab {
	want := canonicalBotSessionPath(sessionPath)
	if want == "" {
		return nil
	}
	topicID := ""
	if meta, ok, _ := agent.LoadBranchMeta(sessionPath); ok && meta.Scope == scopeAutomation {
		topicID = strings.TrimSpace(meta.TopicID)
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	for _, tab := range a.tabs {
		if tab == nil {
			continue
		}
		if canonicalBotSessionPath(tab.currentSessionPath()) == want {
			return tab
		}
	}
	if topicID != "" {
		for _, tab := range a.tabs {
			if tab != nil && tab.Scope == scopeAutomation && strings.TrimSpace(tab.TopicID) == topicID {
				return tab
			}
		}
	}
	return nil
}

func canonicalBotSessionPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return filepath.Clean(path)
}
