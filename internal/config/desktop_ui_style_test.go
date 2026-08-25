package config

import (
	"strings"
	"testing"
)

func TestDesktopUIStyleDefaultsModernAndPreservesClassic(t *testing.T) {
	c := Default()
	if got := c.DesktopUIStyle(); got != DesktopUIStyleModern {
		t.Fatalf("default DesktopUIStyle() = %q, want modern", got)
	}
	if err := c.SetDesktopUIStyle(DesktopUIStyleClassic); err != nil {
		t.Fatal(err)
	}
	if got := c.DesktopUIStyle(); got != DesktopUIStyleClassic {
		t.Fatalf("DesktopUIStyle() = %q, want classic", got)
	}
	if err := c.SetDesktopUIStyle("bulky"); err == nil {
		t.Fatal("SetDesktopUIStyle accepted an unknown style")
	}
}

func TestDesktopUIStyleRendersInUserConfig(t *testing.T) {
	c := Default()
	if err := c.SetDesktopUIStyle(DesktopUIStyleClassic); err != nil {
		t.Fatal(err)
	}
	body := RenderTOMLForScope(c, RenderScopeUser)
	if !strings.Contains(body, `ui_style = "classic"`) {
		t.Fatalf("rendered config is missing classic UI style:\n%s", body)
	}
}
