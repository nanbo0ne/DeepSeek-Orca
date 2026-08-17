package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/agent"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/provider"
)

func TestParseWelcomeSuggestions(t *testing.T) {
	got := parseWelcomeSuggestions("result:\n[\"继续完成发布检查\",\"复盘最近的实现\",\"整理尚未完成的事项\",\"规划下一阶段优化\"]")
	if len(got) != 4 || got[0] != "继续完成发布检查" || got[3] != "规划下一阶段优化" {
		t.Fatalf("parseWelcomeSuggestions() = %#v", got)
	}
}

func TestWelcomeSuggestionTextRedactsSecretsPathsAndAttachments(t *testing.T) {
	raw := "继续 D:\\private\\project\\main.go\nAPI_KEY=sk-supersecret123456\nAuthorization: Bearer abc.def.ghi\nReferenced context: .orca/attachments/file.txt"
	got := welcomeSuggestionText(raw)
	for _, forbidden := range []string{"D:\\private", "supersecret", "abc.def.ghi", "attachments/file.txt"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("welcome suggestion input leaked %q: %q", forbidden, got)
		}
	}
	if !strings.Contains(got, "[REDACTED]") || !strings.Contains(got, "[本地路径]") {
		t.Fatalf("welcome suggestion redaction markers missing: %q", got)
	}
}

func TestParseWelcomeSuggestionsRejectsInvalidSets(t *testing.T) {
	for _, raw := range []string{
		`["one","two","three"]`,
		`["same","same","three","four"]`,
		`["建议你继续项目","帮我检查发布","把文档整理一下","看看还有什么问题"]`,
		`not json`,
	} {
		if got := parseWelcomeSuggestions(raw); got != nil {
			t.Fatalf("parseWelcomeSuggestions(%q) = %#v, want nil", raw, got)
		}
	}
}

func TestWelcomeSuggestionCacheInvalidatesOldVoiceSchema(t *testing.T) {
	isolateDesktopUserDirs(t)
	old := welcomeSuggestionCache{
		Schema:      welcomeSuggestionSchema - 1,
		Prompts:     []string{"one", "two", "three", "four"},
		GeneratedAt: 123,
	}
	data, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(desktopConfigDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(welcomeSuggestionPath(), data, 0o600); err != nil {
		t.Fatal(err)
	}
	got := loadWelcomeSuggestionCache()
	if len(got.Prompts) != 0 || got.GeneratedAt != 0 {
		t.Fatalf("old suggestion cache should be invalidated, got %+v", got)
	}
}

func TestWelcomeSuggestionPromptRequiresUserVoice(t *testing.T) {
	prompt := welcomeSuggestionPrompt(`{"conversations":[],"style_samples":[{"text":"帮我仔细看看这个问题"}]}`)
	for _, want := range []string{"style_samples", "模仿", "用户的话", "建议你", "JSON"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("welcome suggestion prompt missing %q: %s", want, prompt)
		}
	}
}

func TestWelcomeUserStyleSamplesUseDisplayedUserTextOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "style.jsonl")
	wrapped := "Referenced context: .orca/attachments/example.txt\n\nInternal compose wrapper"
	session := agent.NewSession("")
	session.Add(provider.Message{Role: provider.RoleUser, Content: wrapped})
	session.Add(provider.Message{Role: provider.RoleAssistant, Content: "建议用户下一步复盘项目"})
	if err := session.Save(path); err != nil {
		t.Fatal(err)
	}
	if err := recordSessionDisplay(dir, path, wrapped, "你仔细看看这个问题到底出在哪"); err != nil {
		t.Fatal(err)
	}

	got := welcomeUserStyleSamples(dir, path, 2)
	if len(got) != 1 || got[0] != "你仔细看看这个问题到底出在哪" {
		t.Fatalf("style samples = %#v", got)
	}
}
