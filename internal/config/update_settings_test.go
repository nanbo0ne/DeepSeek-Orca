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
		t.Fatalf("V2 update migration = version %d, enabled %v; want 3,true before the V4 vision migration", legacy.ConfigVersion, legacy.DesktopCheckUpdates())
	}

	v3 := Default()
	v3.ConfigVersion = 3
	v3.Desktop.CheckUpdates = &disabled
	normalizeDesktopUpdatePreference(v3)
	if v3.DesktopCheckUpdates() {
		t.Fatal("explicit V3 opt-out was not preserved")
	}

	fresh := Default()
	if fresh.ConfigVersion != 7 || !fresh.DesktopCheckUpdates() || fresh.DesktopUIScale() != 0 {
		t.Fatalf("fresh default = version %d, enabled %v, scale %d; want 7,true,0", fresh.ConfigVersion, fresh.DesktopCheckUpdates(), fresh.DesktopUIScale())
	}
	if err := fresh.SetDesktopUIScale(85); err != nil || fresh.DesktopUIScale() != 85 {
		t.Fatalf("valid manual scale was rejected: err=%v scale=%d", err, fresh.DesktopUIScale())
	}
	if err := fresh.SetDesktopUIScale(83); err == nil {
		t.Fatal("non-five-percent UI scale was accepted")
	}
}

func TestDesktopUpdatePreferenceMigrationPersistsOptOut(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("config_version = 2\n[desktop]\ncheck_updates = false\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := LoadForEdit(path)
	if cfg.ConfigVersion != 7 || !cfg.DesktopCheckUpdates() {
		t.Fatalf("loaded V2 config = version %d, enabled %v", cfg.ConfigVersion, cfg.DesktopCheckUpdates())
	}
	if err := cfg.SetDesktopCheckUpdates(false); err != nil {
		t.Fatal(err)
	}
	if err := cfg.SaveToScope(path, RenderScopeUser); err != nil {
		t.Fatal(err)
	}
	reloaded := LoadForEdit(path)
	if reloaded.ConfigVersion != 7 || reloaded.DesktopCheckUpdates() {
		t.Fatalf("reloaded opt-out = version %d, enabled %v", reloaded.ConfigVersion, reloaded.DesktopCheckUpdates())
	}
}

func TestDesktopVisionPreferenceMigration(t *testing.T) {
	for _, tt := range []struct {
		name    string
		enabled bool
		want    string
	}{
		{name: "legacy disabled", enabled: false, want: VisionModeOff},
		{name: "legacy enabled", enabled: true, want: VisionModeOn},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			cfg.ConfigVersion = 3
			cfg.Desktop.VisionMode = ""
			cfg.Desktop.VisionEnabled = tt.enabled

			normalizeDesktopVisionPreference(cfg)

			if cfg.ConfigVersion != 5 || cfg.DesktopVisionMode() != tt.want {
				t.Fatalf("migrated vision = version %d, mode %q; want 5,%q", cfg.ConfigVersion, cfg.DesktopVisionMode(), tt.want)
			}
		})
	}

	fresh := Default()
	if fresh.DesktopVisionMode() != VisionModeAuto {
		t.Fatalf("fresh vision mode = %q, want auto", fresh.DesktopVisionMode())
	}
}
