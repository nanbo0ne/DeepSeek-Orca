package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"deepseek-orca/internal/agent"
	"deepseek-orca/internal/boot"
	"deepseek-orca/internal/config"
	"deepseek-orca/internal/provider"
)

const welcomeSuggestionInterval = 24 * time.Hour

var (
	welcomeSecretAssignment = regexp.MustCompile(`(?i)(api[_ -]?key|access[_ -]?token|authorization|password|secret)\s*[:=]\s*[^\s,;]+`)
	welcomeBearerToken       = regexp.MustCompile(`(?i)bearer\s+[a-z0-9._~+/=-]+`)
	welcomePrivateKey        = regexp.MustCompile(`(?i)\b(sk|pk|key)-[a-z0-9_-]{12,}\b`)
	welcomeWindowsPath       = regexp.MustCompile(`(?i)\b[a-z]:\\[^\r\n\t"']+`)
	welcomeUnixPath          = regexp.MustCompile(`(?:^|\s)/(?:[^\s/]+/)+[^\s,;]+`)
)

type welcomeSuggestionCache struct {
	Prompts     []string `json:"prompts"`
	GeneratedAt int64    `json:"generatedAt"`
	ModelRef    string   `json:"modelRef"`
	Fingerprint string   `json:"fingerprint"`
}

func welcomeSuggestionPath() string {
	return filepath.Join(desktopConfigDir(), "welcome-suggestions.json")
}

func loadWelcomeSuggestionCache() welcomeSuggestionCache {
	var cache welcomeSuggestionCache
	if data, err := os.ReadFile(welcomeSuggestionPath()); err == nil {
		_ = json.Unmarshal(data, &cache)
	}
	if len(cache.Prompts) != 4 {
		cache.Prompts = nil
	}
	return cache
}

func saveWelcomeSuggestionCache(cache welcomeSuggestionCache) error {
	if err := os.MkdirAll(desktopConfigDir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	tmp := welcomeSuggestionPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, welcomeSuggestionPath())
}

func (a *App) GetWelcomeSuggestions() []string {
	return append([]string(nil), loadWelcomeSuggestionCache().Prompts...)
}

func (a *App) runWelcomeSuggestionScheduler() {
	refresh := func() {
		cache := loadWelcomeSuggestionCache()
		if cache.GeneratedAt > 0 && time.Since(time.UnixMilli(cache.GeneratedAt)) < welcomeSuggestionInterval {
			return
		}
		if next, err := a.generateWelcomeSuggestions(); err == nil && len(next.Prompts) == 4 {
			_ = saveWelcomeSuggestionCache(next)
			if a.ctx != nil {
				wruntime.EventsEmit(a.ctx, "welcome:suggestions", next.Prompts)
			}
		}
	}
	refresh()
	ticker := time.NewTicker(welcomeSuggestionInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			refresh()
		case <-a.bootContext().Done():
			return
		}
	}
}

func (a *App) generateWelcomeSuggestions() (welcomeSuggestionCache, error) {
	cfg, err := config.Load()
	if err != nil {
		return welcomeSuggestionCache{}, err
	}
	modelRef, _, ok := cfg.ResolveModelWithFallback(cfg.Bot.Model)
	if !ok {
		modelRef, _, ok = cfg.ResolveModelWithFallback(cfg.DefaultModel)
	}
	if !ok {
		return welcomeSuggestionCache{}, os.ErrNotExist
	}
	entry, ok := cfg.ResolveModel(modelRef)
	if !ok {
		return welcomeSuggestionCache{}, os.ErrNotExist
	}
	prov, err := boot.NewProviderWithProxy(entry, cfg.NetworkProxySpec())
	if err != nil {
		return welcomeSuggestionCache{}, err
	}

	type row struct {
		Title     string `json:"title"`
		Workspace string `json:"workspace,omitempty"`
		Preview   string `json:"preview"`
		Updated   int64  `json:"updated"`
	}
	rows := make([]row, 0, 80)
	seen := map[string]bool{}
	for _, dir := range a.knownSessionDirs() {
		infos, listErr := agent.ListSessions(dir)
		if listErr != nil {
			continue
		}
		for _, info := range infos {
			if seen[info.Path] || len(rows) >= 80 {
				continue
			}
			seen[info.Path] = true
			title := welcomeSuggestionText(info.TopicTitle)
			preview := welcomeSuggestionText(info.Preview)
			if title == "" && preview == "" {
				continue
			}
			if len([]rune(preview)) > 280 {
				preview = string([]rune(preview)[:280])
			}
			workspace := strings.TrimSpace(filepath.Base(strings.TrimSpace(info.WorkspaceRoot)))
			if workspace == "." || workspace == string(filepath.Separator) {
				workspace = ""
			}
			rows = append(rows, row{Title: title, Workspace: welcomeSuggestionText(workspace), Preview: preview, Updated: info.LastActivityAt.UnixMilli()})
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Updated > rows[j].Updated })
	data, _ := json.Marshal(rows)
	fingerprintBytes := sha256.Sum256(data)
	fingerprint := hex.EncodeToString(fingerprintBytes[:])
	prompt := "根据下面这些 DeepSeek-Orca 会话的标题和有限摘要，生成恰好 4 条简短、互不重复、可以直接作为用户消息发送的中文建议。四条应分别偏向：继续当前任务、复盘整理、处理待办、潜在下一步。只返回 JSON 字符串数组，不要解释。不得复述凭据或路径。\n\n" + string(data)
	ctx, cancel := context.WithTimeout(a.bootContext(), 60*time.Second)
	defer cancel()
	stream, err := prov.Stream(ctx, provider.Request{Temperature: 0, MaxTokens: 500, Messages: []provider.Message{{Role: provider.RoleUser, Content: prompt}}})
	if err != nil {
		return welcomeSuggestionCache{}, err
	}
	var output strings.Builder
	for chunk := range stream {
		if chunk.Type == provider.ChunkText {
			output.WriteString(chunk.Text)
		}
		if chunk.Type == provider.ChunkError && chunk.Err != nil {
			return welcomeSuggestionCache{}, chunk.Err
		}
	}
	prompts := parseWelcomeSuggestions(output.String())
	if len(prompts) != 4 {
		return welcomeSuggestionCache{}, os.ErrInvalid
	}
	return welcomeSuggestionCache{Prompts: prompts, GeneratedAt: time.Now().UnixMilli(), ModelRef: modelRef, Fingerprint: fingerprint}, nil
}

func welcomeSuggestionText(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if trimmed == "" || strings.Contains(lower, "referenced context") || strings.Contains(lower, ".deepseek-orca/attachments") || strings.Contains(lower, "snapshot path") || strings.Contains(lower, "<image ") {
			continue
		}
		trimmed = welcomeBearerToken.ReplaceAllString(trimmed, "Bearer [REDACTED]")
		trimmed = welcomeSecretAssignment.ReplaceAllString(trimmed, "$1=[REDACTED]")
		trimmed = welcomePrivateKey.ReplaceAllString(trimmed, "[REDACTED]")
		trimmed = welcomeWindowsPath.ReplaceAllString(trimmed, "[本地路径]")
		trimmed = welcomeUnixPath.ReplaceAllString(trimmed, " [本地路径]")
		kept = append(kept, strings.TrimSpace(trimmed))
	}
	return strings.TrimSpace(strings.Join(kept, " "))
}

func parseWelcomeSuggestions(raw string) []string {
	start, end := strings.Index(raw, "["), strings.LastIndex(raw, "]")
	if start < 0 || end <= start {
		return nil
	}
	var values []string
	if json.Unmarshal([]byte(raw[start:end+1]), &values) != nil || len(values) != 4 {
		return nil
	}
	seen := map[string]bool{}
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
		if values[i] == "" || seen[values[i]] || len([]rune(values[i])) > 60 {
			return nil
		}
		seen[values[i]] = true
	}
	return values
}
