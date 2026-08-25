package main

import (
	"os"
	"testing"
	"time"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/config"
)

// TestMain isolates os.UserConfigDir() for the whole package. On Windows it
// reads %AppData%, which the per-test HOME / XDG_CONFIG_HOME overrides do not
// cover — without this, tests that persist desktop state (saveWorkspace,
// session/cache writes) leak into the developer's real DeepSeek-Orca config dir.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "deepseek-orca-desktop-test")
	if err != nil {
		os.Exit(1)
	}
	os.Setenv("AppData", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestDesktopFramelessFollowsStyleOnlyOnWindows(t *testing.T) {
	tests := []struct {
		goos, style string
		want        bool
	}{
		{"windows", config.DesktopUIStyleModern, true},
		{"windows", "", true},
		{"windows", config.DesktopUIStyleClassic, false},
		{"darwin", config.DesktopUIStyleModern, false},
		{"linux", config.DesktopUIStyleModern, false},
	}
	for _, tt := range tests {
		if got := desktopFrameless(tt.goos, tt.style); got != tt.want {
			t.Fatalf("desktopFrameless(%q, %q) = %v, want %v", tt.goos, tt.style, got, tt.want)
		}
	}
}

func TestConsumeRestartWaitArg(t *testing.T) {
	args, wait := consumeRestartWaitArg([]string{"Orca.exe", "--flag", restartWaitArgPrefix + "1800"})
	if wait != 1800*time.Millisecond {
		t.Fatalf("wait = %s", wait)
	}
	if len(args) != 2 || args[1] != "--flag" {
		t.Fatalf("clean args = %#v", args)
	}
}

func TestDesktopBackgroundMatchesPresentation(t *testing.T) {
	modern := desktopBackground(config.DesktopUIStyleModern)
	classic := desktopBackground(config.DesktopUIStyleClassic)
	if modern.R != 255 || modern.G != 255 || modern.B != 255 {
		t.Fatalf("modern background = %#v", modern)
	}
	if classic.R != 251 || classic.G != 252 || classic.B != 255 {
		t.Fatalf("classic background = %#v", classic)
	}
}

func TestWindowsWebview2GPUDisabled(t *testing.T) {
	oldChannel := channel
	t.Cleanup(func() {
		channel = oldChannel
		os.Unsetenv(disableWebview2GPUEnv)
	})

	tests := []struct {
		name    string
		channel string
		env     string
		want    bool
	}{
		{name: "stable default keeps gpu", channel: "stable", want: false},
		{name: "canary default disables gpu", channel: "canary", want: true},
		{name: "env enables fallback", channel: "stable", env: "1", want: true},
		{name: "env disables canary fallback", channel: "canary", env: "0", want: false},
		{name: "truthy env", channel: "stable", env: "yes", want: true},
		{name: "falsey env", channel: "canary", env: "off", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel = tt.channel
			if tt.env == "" {
				os.Unsetenv(disableWebview2GPUEnv)
			} else {
				os.Setenv(disableWebview2GPUEnv, tt.env)
			}
			if got := windowsWebview2GPUDisabled(); got != tt.want {
				t.Fatalf("windowsWebview2GPUDisabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
