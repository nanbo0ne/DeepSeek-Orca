package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/agent"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/provider"
)

// Older Windows test runs did not isolate APPDATA and could leave one exact
// boot-test transcript in the real desktop history. Remove only that complete
// fixture signature; ordinary conversations titled "first review" are kept.
func (a *App) cleanupLeakedBootReviewSessions() map[string]struct{} {
	return a.cleanupLeakedBootReviewSessionsInDirs(a.knownSessionDirs())
}

func (a *App) cleanupLeakedBootReviewSessionsInDirs(dirs []string) map[string]struct{} {
	topicIDs := map[string]struct{}{}
	type topicLocation struct {
		scope string
		root  string
		id    string
	}
	locations := []topicLocation{}

	for _, dir := range dirs {
		infos, err := agent.ListSessions(dir)
		if err != nil {
			continue
		}
		for _, info := range infos {
			session, err := agent.LoadSession(info.Path)
			if err != nil || !isLeakedBootReviewFixture(session.Snapshot()) {
				continue
			}
			if meta, ok, err := agent.LoadBranchMeta(info.Path); err == nil && ok && strings.TrimSpace(meta.TopicID) != "" {
				id := strings.TrimSpace(meta.TopicID)
				topicIDs[id] = struct{}{}
				locations = append(locations, topicLocation{scope: strings.TrimSpace(meta.Scope), root: strings.TrimSpace(meta.WorkspaceRoot), id: id})
			}
			if err := purgeLeakedBootReviewSession(dir, info.Path); err != nil {
				slog.Warn("desktop: failed to remove leaked boot test session", "path", info.Path, "err", err)
			}
		}
	}

	if len(locations) == 0 {
		return topicIDs
	}
	f := loadProjectsFile()
	for _, location := range locations {
		root := location.root
		if location.scope != "project" {
			root = ""
			f.GlobalTopics = removeString(f.GlobalTopics, location.id)
		} else {
			for i := range f.Projects {
				if normalizeProjectRoot(f.Projects[i].Root) == normalizeProjectRoot(root) {
					f.Projects[i].Topics = removeString(f.Projects[i].Topics, location.id)
				}
			}
		}
		titles := loadTopicTitles(root)
		delete(titles, location.id)
		_ = saveTopicTitles(root, titles)
		sources := loadTopicTitleSources(root)
		delete(sources, location.id)
		_ = saveTopicTitleSources(root, sources)
		deleteTopicCreatedAt(root, location.id)
	}
	_ = saveProjectsFile(f)
	return topicIDs
}

func isLeakedBootReviewFixture(messages []provider.Message) bool {
	users := []string{}
	finals := map[string]bool{}
	tasks := map[string]bool{}
	for _, message := range messages {
		switch message.Role {
		case provider.RoleUser:
			users = append(users, strings.TrimSpace(message.Content))
		case provider.RoleAssistant:
			content := strings.TrimSpace(message.Content)
			if content == "parent first done" || content == "parent second done" {
				finals[content] = true
			}
			for _, call := range message.ToolCalls {
				if call.Name != "review" {
					continue
				}
				var args struct {
					Task string `json:"task"`
				}
				if json.Unmarshal([]byte(call.Arguments), &args) == nil {
					tasks[strings.TrimSpace(args.Task)] = true
				}
			}
		}
	}
	return len(users) == 2 && users[0] == "first review" && users[1] == "continue review" &&
		finals["parent first done"] && finals["parent second done"] &&
		tasks["first skill task"] && tasks["second skill task"]
}

func purgeLeakedBootReviewSession(dir, path string) error {
	path, key, err := validateSessionPath(dir, path)
	if err != nil {
		return err
	}
	if err := agent.DeleteSubagentsByParent(dir, agent.BranchID(path)); err != nil {
		return err
	}
	for _, artifact := range []string{path, path + ".meta"} {
		if err := os.Remove(artifact); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.RemoveAll(strings.TrimSuffix(path, ".jsonl") + ".ckpt"); err != nil {
		return err
	}
	titles := loadSessionTitles(dir)
	delete(titles, key)
	if err := saveSessionTitles(dir, titles); err != nil {
		return err
	}
	displays := loadSessionDisplays(dir)
	delete(displays, key)
	if err := saveSessionDisplays(dir, displays); err != nil {
		return err
	}
	_ = os.Remove(filepath.Join(dir, strings.TrimSuffix(key, ".jsonl")+".events.jsonl"))
	return nil
}
