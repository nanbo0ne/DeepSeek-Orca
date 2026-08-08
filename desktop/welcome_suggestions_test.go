package main

import "testing"

func TestParseWelcomeSuggestions(t *testing.T) {
	got := parseWelcomeSuggestions("result:\n[\"继续完成发布检查\",\"复盘最近的实现\",\"整理尚未完成的事项\",\"规划下一阶段优化\"]")
	if len(got) != 4 || got[0] != "继续完成发布检查" || got[3] != "规划下一阶段优化" {
		t.Fatalf("parseWelcomeSuggestions() = %#v", got)
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
