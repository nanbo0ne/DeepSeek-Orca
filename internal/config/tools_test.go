package config

import (
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestBashTimeoutSecondsDefaultsToSafetyCap(t *testing.T) {
	cfg := Default()
	if cfg.Tools.BashTimeoutSeconds != nil {
		t.Fatalf("default raw bash timeout = %v, want nil", *cfg.Tools.BashTimeoutSeconds)
	}
	if got := cfg.BashTimeoutSeconds(); got != 120 {
		t.Fatalf("BashTimeoutSeconds() = %d, want 120", got)
	}
}

func TestBashTimeoutSecondsAllowsExplicitZero(t *testing.T) {
	cfg := Default()
	cfg.Tools.BashTimeoutSeconds = intPtr(0)
	if got := cfg.BashTimeoutSeconds(); got != 0 {
		t.Fatalf("BashTimeoutSeconds() = %d, want 0", got)
	}
}

func TestBashTimeoutSecondsParsesExplicitZero(t *testing.T) {
	cfg := Default()
	if _, err := toml.Decode("[tools]\nbash_timeout_seconds = 0\n", cfg); err != nil {
		t.Fatalf("decode explicit zero: %v", err)
	}
	if cfg.Tools.BashTimeoutSeconds == nil {
		t.Fatal("explicit zero decoded as nil")
	}
	if got := cfg.BashTimeoutSeconds(); got != 0 {
		t.Fatalf("BashTimeoutSeconds() = %d, want 0", got)
	}
}

func TestBashTimeoutSecondsFallsBackForNegative(t *testing.T) {
	cfg := Default()
	cfg.Tools.BashTimeoutSeconds = intPtr(-1)
	if got := cfg.BashTimeoutSeconds(); got != 120 {
		t.Fatalf("BashTimeoutSeconds() = %d, want 120", got)
	}
}

func TestBuildActiveToolRoutingPolicyRespectsToolLibrarySettings(t *testing.T) {
	settings := DefaultToolLibrarySettings()
	settings.WebSearchEnabled = false
	settings.ProactiveToolUseEnabled = false
	policy := BuildActiveToolRoutingPolicy(settings)
	for _, forbidden := range []string{"Be evidence-first", "Use web_search"} {
		if strings.Contains(policy, forbidden) {
			t.Fatalf("policy should not contain %q when disabled:\n%s", forbidden, policy)
		}
	}
	if !strings.Contains(policy, "Use web_fetch only when you already have a specific URL") {
		t.Fatalf("policy missing web_fetch fallback when web search disabled:\n%s", policy)
	}
	if !strings.Contains(policy, "host_command") {
		t.Fatalf("policy should still include enabled host tools:\n%s", policy)
	}
}

func TestDefaultToolLibrarySettingsEnableExtendedTools(t *testing.T) {
	cfg := Default()
	if cfg.ToolLibrary != DefaultToolLibrarySettings() {
		t.Fatalf("default tool library = %+v, want %+v", cfg.ToolLibrary, DefaultToolLibrarySettings())
	}
}
