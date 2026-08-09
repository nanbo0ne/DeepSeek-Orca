package main

import (
	"strings"
	"testing"
)

func TestParseWelcomeSuggestions(t *testing.T) {
	got := parseWelcomeSuggestions("result:\n[\"继续完成发布检查\",\"复盘最近的实现\",\"整理尚未完成的事项\",\"规划下一阶段优化\"]")
	if len(got) != 4 || got[0] != "继续完成发布检查" || got[3] != "规划下一阶段优化" {
		t.Fatalf("parseWelcomeSuggestions() = %#v", got)
	}
}

func TestWelcomeSuggestionTextRedactsSecretsPathsAndAttachments(t *testing.T) {
	raw := "继续 D:\\private\\project\\main.go\nAPI_KEY=sk-supersecret123456\nAuthorization: Bearer abc.def.ghi\nReferenced context: .deepseek-orca/attachments/file.txt"
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
		`not json`,
	} {
		if got := parseWelcomeSuggestions(raw); got != nil {
			t.Fatalf("parseWelcomeSuggestions(%q) = %#v, want nil", raw, got)
		}
	}
}
