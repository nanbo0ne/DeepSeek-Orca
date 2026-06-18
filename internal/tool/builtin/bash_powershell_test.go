package builtin

import (
	"context"
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"deepseek-orca/internal/sandbox"
)

func powershellPath(t *testing.T) string {
	t.Helper()
	for _, n := range []string{"pwsh", "powershell"} {
		if p, err := exec.LookPath(n); err == nil {
			return p
		}
	}
	t.Skip("no PowerShell on PATH")
	return ""
}

func runPS(t *testing.T, command string) (string, error) {
	t.Helper()
	b := bash{shell: sandbox.Shell{Kind: sandbox.ShellPowerShell, Path: powershellPath(t)}}
	args, _ := json.Marshal(map[string]string{"command": command})
	return b.Execute(context.Background(), args)
}

func TestBashPowerShellRunsNativeCommand(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("powershell e2e is windows-only")
	}
	out, err := runPS(t, "Write-Output deepseek-orca-ok")
	if err != nil {
		t.Fatalf("powershell command failed: %v (out=%q)", err, out)
	}
	if !strings.Contains(out, "deepseek-orca-ok") {
		t.Fatalf("output = %q, want it to contain deepseek-orca-ok", out)
	}
}

func TestBashPowerShellSurfacesNonZeroExit(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("powershell e2e is windows-only")
	}
	if _, err := runPS(t, "exit 3"); err == nil {
		t.Fatal("non-zero exit should surface as an error")
	}
}

func TestBashPowerShellRejectsChaining(t *testing.T) {
	b := bash{shell: sandbox.Shell{Kind: sandbox.ShellPowerShell, Path: "powershell"}}
	for _, cmd := range []string{"echo a && echo b", "echo a || echo b"} {
		args, _ := json.Marshal(map[string]string{"command": cmd})
		out, err := b.Execute(context.Background(), args)
		if err == nil {
			t.Errorf("%q should be rejected on powershell, got out=%q", cmd, out)
		} else if !strings.Contains(err.Error(), "PowerShell") {
			t.Errorf("%q error should explain PowerShell, got %v", cmd, err)
		}
	}
}

func TestBashPowerShellAllowsQuotedOperator(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("runs a real powershell command")
	}
	// "&&" inside a string literal is data, not chaining — must not be rejected.
	out, err := runPS(t, `Write-Output "a && b"`)
	if err != nil {
		t.Fatalf("quoted && should run: %v (out=%q)", err, out)
	}
	if !strings.Contains(out, "a && b") {
		t.Fatalf("output = %q", out)
	}
}

func TestBashPwshAllowsChaining(t *testing.T) {
	// pwsh (PowerShell 7+) parses && — the guard must not block it.
	b := bash{shell: sandbox.Shell{Kind: sandbox.ShellPowerShell, Path: "pwsh"}}
	args, _ := json.Marshal(map[string]string{"command": "echo a && echo b"})
	_, err := b.Execute(context.Background(), args)
	if err != nil && strings.Contains(err.Error(), "does not parse") {
		t.Errorf("pwsh should not be blocked by the chaining guard: %v", err)
	}
}

func TestBashPowerShellOutputIsUTF8(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("powershell e2e is windows-only")
	}
	out, err := runPS(t, "Write-Output 'AB-中文-CD'")
	if err != nil {
		t.Fatalf("command failed: %v (out=%q)", err, out)
	}
	if !strings.Contains(out, "中文") {
		t.Fatalf("non-ASCII output mojibake — got %q (want it to contain 中文)", out)
	}
}

func TestProgressWriterKeepsUTF8RuneBoundaries(t *testing.T) {
	var chunks []string
	w := newProgressWriter(func(chunk string) {
		chunks = append(chunks, chunk)
	})
	want := "\u6587\u4ef6\u540d.txt\n"
	raw := []byte(want)
	if _, err := w.Write(raw[:1]); err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 0 {
		t.Fatalf("emitted partial rune chunk: %q", chunks)
	}
	if _, err := w.Write(raw[1:5]); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(raw[5:]); err != nil {
		t.Fatal(err)
	}
	w.Flush()
	if got := strings.Join(chunks, ""); got != want {
		t.Fatalf("progress chunks = %q, want %q", got, want)
	}
}

func TestBashDescriptionReflectsShell(t *testing.T) {
	ps := bash{shell: sandbox.Shell{Kind: sandbox.ShellPowerShell, Path: "powershell"}}
	if !strings.Contains(ps.Description(), "PowerShell") {
		t.Errorf("powershell description should warn about PowerShell: %q", ps.Description())
	}
	sh := bash{shell: sandbox.Shell{Kind: sandbox.ShellBash, Path: "bash"}}
	if strings.Contains(sh.Description(), "PowerShell") {
		t.Errorf("bash description should not mention PowerShell: %q", sh.Description())
	}
}

func TestShellExitErrorIncludesExitCodeAndShell(t *testing.T) {
	err := shellExitError(
		sandbox.Shell{Kind: sandbox.ShellPowerShell, Path: "powershell"},
		"exit 1",
		"exit 1",
		"",
		&exec.ExitError{},
	)
	msg := err.Error()
	if !strings.Contains(msg, "shell: powershell") {
		t.Fatalf("error should include shell kind, got %q", msg)
	}
	if !strings.Contains(msg, "command exited") {
		t.Fatalf("error should explain command exit, got %q", msg)
	}
}

func TestShutdownFailureReasonIsActionable(t *testing.T) {
	reason := shellExitReason(
		sandbox.Shell{Kind: sandbox.ShellBash, Path: "bash"},
		`shutdown /s /t 60 /c "DeepSeek-Orca test"`,
		"用法: C:\\Windows\\System32\\shutdown.exe [/i | /l | /s]",
		&exec.ExitError{},
	)
	if !strings.Contains(reason, "shutdown") || !strings.Contains(reason, "host_command") || !strings.Contains(reason, "Windows-native") {
		t.Fatalf("shutdown reason should be actionable, got %q", reason)
	}
}
