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
	topicID := newTopicID()
	root := independentWorkspaceRoot(topicID)
	if ensured, err := ensureIndependentWorkspaceRoot(topicID); err == nil {
		root = ensured
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return bot.SessionChoice{}, err
	}
	dir := desktopSessionDir(root)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return bot.SessionChoice{}, err
	}

	title := "机器人新对话"
	if err := setTopicTitleWithSource("", topicID, title, topicTitleSourceAuto); err != nil {
		return bot.SessionChoice{}, err
	}
	_ = setTopicCreatedAt("", topicID, time.Now().Unix())
	f := loadProjectsFile()
	f.GlobalTopics = prependUniqueString(f.GlobalTopics, topicID)
	_ = saveProjectsFile(f)

	path := agent.NewSessionPath(dir, "bot")
	sess := agent.NewSession("")
	if err := sess.Save(path); err != nil {
		return bot.SessionChoice{}, err
	}
	now := time.Now()
	_ = agent.SaveBranchMetaPreserveUpdated(path, agent.BranchMeta{
		ID:            agent.BranchID(path),
		Scope:         "global",
		WorkspaceRoot: root,
		TopicID:       topicID,
		TopicTitle:    title,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	a.emitProjectTreeChanged()

	return bot.SessionChoice{
		Path:          path,
		Title:         title,
		Location:      "独立工作区",
		WorkspaceRoot: root,
		LastActivity:  now,
	}, nil
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
