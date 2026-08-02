package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"deepseek-orca/internal/agent"
	"deepseek-orca/internal/config"
	"deepseek-orca/internal/provider"
)

const maxBotSessionChoices = 15

type sessionChoice struct {
	TopicID       string
	Number        int
	Path          string
	Title         string
	Preview       string
	Location      string
	WorkspaceRoot string
	LastActivity  time.Time
	Turns         int
	LastAssistant string
}

type SessionChoice = sessionChoice

type botSessionLister func(limit int) ([]sessionChoice, error)

func defaultBotSessionLister(limit int) ([]sessionChoice, error) {
	if limit <= 0 {
		limit = maxBotSessionChoices
	}

	seen := map[string]bool{}
	var choices []sessionChoice
	for _, dir := range knownBotSessionDirs() {
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true

		infos, err := agent.ListSessions(dir)
		if err != nil {
			continue
		}
		titles := loadBotSessionTitles(dir)
		for _, info := range infos {
			choice := sessionChoiceFromInfo(info, titles[filepath.Base(info.Path)])
			if choice.Path == "" {
				continue
			}
			if last, err := lastAssistantTail(info.Path, 220); err == nil {
				choice.LastAssistant = last
			}
			choices = append(choices, choice)
		}
	}

	sort.SliceStable(choices, func(i, j int) bool {
		if choices[i].LastActivity.Equal(choices[j].LastActivity) {
			return choices[i].Path < choices[j].Path
		}
		return choices[i].LastActivity.After(choices[j].LastActivity)
	})
	if len(choices) > limit {
		choices = choices[:limit]
	}
	for i := range choices {
		choices[i].Number = i + 1
	}
	return choices, nil
}

func defaultBotSessionCreator(ctx context.Context, remoteKey string, msg InboundMessage) (sessionChoice, error) {
	root := strings.TrimSpace(config.BotWorkspaceDir())
	if root == "" {
		root = globalBotWorkspaceRoot()
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return sessionChoice{}, err
	}
	dir := config.ProjectSessionDir(root)
	if dir == "" {
		dir = config.SessionDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return sessionChoice{}, err
	}
	path := agent.NewSessionPath(dir, "bot")
	sess := agent.NewSession("")
	if err := sess.Save(path); err != nil {
		return sessionChoice{}, err
	}
	now := time.Now()
	title := "机器人新对话"
	_ = agent.SaveBranchMetaPreserveUpdated(path, agent.BranchMeta{
		ID:            agent.BranchID(path),
		Scope:         "global",
		WorkspaceRoot: root,
		TopicTitle:    title,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	return sessionChoice{
		TopicID:       agent.BranchID(path),
		Path:          path,
		Title:         title,
		Location:      "独立工作区",
		WorkspaceRoot: root,
		LastActivity:  now,
	}, nil
}

func knownBotSessionDirs() []string {
	seen := map[string]bool{}
	var out []string
	add := func(dir string) {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			return
		}
		if abs, err := filepath.Abs(dir); err == nil {
			dir = abs
		}
		if seen[dir] {
			return
		}
		seen[dir] = true
		out = append(out, dir)
	}

	add(config.SessionDir())
	if root := globalBotWorkspaceRoot(); root != "" {
		add(config.ProjectSessionDir(root))
	}
	for _, root := range desktopProjectRoots() {
		add(config.ProjectSessionDir(root))
	}
	return out
}

func globalBotWorkspaceRoot() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		if home == "" {
			return ""
		}
		return filepath.Join(home, ".deepseek-orca", "global-workspace")
	}
	return filepath.Join(dir, "deepseek-orca", "global-workspace")
}

func desktopProjectRoots() []string {
	path := filepath.Join(config.MemoryUserDir(), "desktop-projects.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var f struct {
		Projects []struct {
			Root string `json:"root"`
		} `json:"projects"`
	}
	if json.Unmarshal(b, &f) != nil {
		return nil
	}
	out := make([]string, 0, len(f.Projects))
	for _, p := range f.Projects {
		root := strings.TrimSpace(p.Root)
		if root != "" {
			out = append(out, root)
		}
	}
	return out
}

func loadBotSessionTitles(dir string) map[string]string {
	m := map[string]string{}
	b, err := os.ReadFile(filepath.Join(dir, ".titles.json"))
	if err != nil {
		return m
	}
	_ = json.Unmarshal(b, &m)
	return m
}

func sessionChoiceFromInfo(info agent.SessionInfo, explicitTitle string) sessionChoice {
	title := firstNonEmpty(
		strings.TrimSpace(explicitTitle),
		strings.TrimSpace(info.TopicTitle),
		strings.TrimSpace(info.Preview),
		strings.TrimSuffix(filepath.Base(info.Path), filepath.Ext(info.Path)),
	)
	if title == "" {
		title = "未命名会话"
	}
	workspaceRoot := strings.TrimSpace(info.WorkspaceRoot)
	if info.Scope == "global" {
		workspaceRoot = globalBotWorkspaceRoot()
	}
	return sessionChoice{
		TopicID:       info.TopicID,
		Path:          info.Path,
		Title:         oneLine(title, 42),
		Preview:       oneLine(info.Preview, 64),
		Location:      botSessionLocation(info),
		WorkspaceRoot: workspaceRoot,
		LastActivity:  info.LastActivityAt,
		Turns:         info.Turns,
	}
}

func botSessionLocation(info agent.SessionInfo) string {
	if strings.TrimSpace(info.WorkspaceRoot) != "" {
		base := filepath.Base(strings.TrimRight(info.WorkspaceRoot, string(os.PathSeparator)))
		if base != "" && base != "." {
			return base
		}
	}
	if info.Scope == "global" {
		return "独立工作区"
	}
	return "全局会话"
}

func lastAssistantTail(path string, maxRunes int) (string, error) {
	sess, err := agent.LoadSession(path)
	if err != nil {
		return "", err
	}
	msgs := sess.Snapshot()
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role != provider.RoleAssistant {
			continue
		}
		text := strings.TrimSpace(msgs[i].Content)
		if text == "" {
			continue
		}
		return tailText(text, maxRunes), nil
	}
	return "", nil
}

func formatSessionChoices(choices []sessionChoice) string {
	if len(choices) == 0 {
		return "暂无可恢复的对话。请先在桌面端、CLI 或手机端创建至少一轮对话。"
	}

	var b strings.Builder
	b.WriteString("请选择要进入的对话（回复数字即可）：\n")
	for _, c := range choices {
		meta := sessionChoiceMeta(c)
		fmt.Fprintf(&b, "%d. %s", c.Number, c.Title)
		if len(meta) > 0 {
			fmt.Fprintf(&b, "（%s）", strings.Join(meta, " / "))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n发送 /start 可随时重新打开列表，发送 /new 可创建新的独立工作区对话。")
	return b.String()
}

func sessionChoiceMeta(c sessionChoice) []string {
	meta := []string{}
	if c.Location != "" {
		meta = append(meta, c.Location)
	}
	if !c.LastActivity.IsZero() {
		meta = append(meta, c.LastActivity.Local().Format("01-02 15:04"))
	}
	if c.Turns > 0 {
		meta = append(meta, fmt.Sprintf("%d 轮", c.Turns))
	}
	return meta
}

func formatSessionEntered(choice sessionChoice) string {
	var b strings.Builder
	fmt.Fprintf(&b, "已进入：%s", choice.Title)
	if choice.Location != "" {
		fmt.Fprintf(&b, "\n位置：%s", choice.Location)
	}
	if strings.TrimSpace(choice.LastAssistant) != "" {
		fmt.Fprintf(&b, "\n\n上次 AI 回复结尾：\n%s", choice.LastAssistant)
	} else {
		b.WriteString("\n\n这个对话里还没有可展示的 AI 回复。")
	}
	return b.String()
}

func formatSessionCreated(choice sessionChoice) string {
	var b strings.Builder
	fmt.Fprintf(&b, "已创建新对话：%s", choice.Title)
	if choice.Location != "" {
		fmt.Fprintf(&b, "\n位置：%s", choice.Location)
	}
	b.WriteString("\n你现在可以直接发送消息。")
	return b.String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func oneLine(text string, maxRunes int) string {
	text = strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if maxRunes <= 0 {
		return text
	}
	r := []rune(text)
	if len(r) <= maxRunes {
		return text
	}
	return string(r[:maxRunes-1]) + "…"
}

func tailText(text string, maxRunes int) string {
	text = strings.TrimSpace(text)
	if maxRunes <= 0 {
		return text
	}
	r := []rune(text)
	if len(r) <= maxRunes {
		return text
	}
	return "…" + string(r[len(r)-maxRunes+1:])
}
