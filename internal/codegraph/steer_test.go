package codegraph

import (
	"strings"
	"testing"
)

func TestSteerTextForToolsUsesRegisteredNames(t *testing.T) {
	text := SteerTextForTools([]string{
		"mcp__codegraph__context",
		"mcp__codegraph__search",
		"mcp__codegraph__callers",
		"mcp__codegraph__callees",
	})
	if !strings.Contains(text, "mcp__codegraph__context") {
		t.Fatalf("steer text missing registered context tool:\n%s", text)
	}
	if !strings.Contains(text, "mcp__codegraph__search") {
		t.Fatalf("steer text missing registered search tool:\n%s", text)
	}
	if strings.Contains(text, "- codegraph_search") || strings.Contains(text, "call bare codegraph_search") {
		t.Fatalf("steer text should not recommend bare codegraph_search:\n%s", text)
	}
}
