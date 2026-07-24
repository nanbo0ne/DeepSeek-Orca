package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDesktopUpdatePreferenceMigration(t *testing.T) {
	legacy := Default()
	legacy.ConfigVersion = 2
	disabled := false
	legacy.Desktop.CheckUpdates = &disabled
	normalizeDesktopUpdatePreference(legacy)
	if legacy.ConfigVersion != 3 || !legacy.DesktopCheckUpdates() {
		t.Fatalf("V2 migration = version %d, enabled %v; want 3,true", legacy.ConfigVersion, legacy.DesktopCheckUpdates())
	}

	v3 := Default()
	v3.ConfigVersion = 3
	v3.Desktop.CheckUpdates = &disabled
	normalizeDesktopUpdatePreference(v3)
	if v3.DesktopCheckUpdates() {
		t.Fatal("explicit V3 opt-out was not preserved")
	}

	fresh := Default()
	if fresh.ConfigVersion != 3 || !fresh.DesktopCheckUpdates() {
		t.Fatalf("fresh default = version %d, enabled %v; want 3,true", fresh.ConfigVersion, fresh.DesktopCheckUpdates())
	}
}

func TestDesktopUpdatePreferenceMigrationPersistsOptOut(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("config_version = 2\n[desktop]\ncheck_updates = false\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := LoadForEdit(path)
	if cfg.ConfigVersion != 3 || !cfg.DesktopCheckUpdates() {
		t.Fatalf("loaded V2 config = version %d, enabled %v", cfg.ConfigVersion, cfg.DesktopCheckUpdates())
	}
	if err := cfg.SetDesktopCheckUpdates(false); err != nil {
		t.Fatal(err)
	}
	if err := cfg.SaveToScope(path, RenderScopeUser); err != nil {
		t.Fatal(err)
	}
	reloaded := LoadForEdit(path)
	if reloaded.ConfigVersion != 3 || reloaded.DesktopCheckUpdates() {
		t.Fatalf("reloaded opt-out = version %d, enabled %v", reloaded.ConfigVersion, reloaded.DesktopCheckUpdates())
	}
}
