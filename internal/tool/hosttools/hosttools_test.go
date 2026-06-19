package hosttools

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestToolsExposeDefaultHostLibraryWithoutVisualTools(t *testing.T) {
	tools := Tools(t.TempDir())
	got := map[string]bool{}
	for _, tl := range tools {
		got[tl.Name()] = true
	}
	for _, name := range []string{
		"host_command",
		"host_system_info",
		"host_list_processes",
		"host_kill_process",
		"host_open_app",
		"host_clipboard",
		"notify_user",
		"automation_create",
		"automation_list",
		"automation_cancel",
		"thread_list",
		"web_search",
		"node_repl_exec",
		"python_repl_exec",
		"document_inspect",
		"document_extract",
	} {
		if !got[name] {
			t.Fatalf("missing host tool %q", name)
		}
	}
	for _, name := range []string{"screenshot", "ocr", "computer_use", "click", "type_text"} {
		if got[name] {
			t.Fatalf("visual/control tool %q should not be registered in v2.0.11 host library", name)
		}
	}
}

func TestWebSearchUsesChinaAccessibleSourcesInDescription(t *testing.T) {
	desc := webSearch{}.Description()
	if strings.Contains(strings.ToLower(desc), "duckduckgo") {
		t.Fatalf("web_search description should not mention DuckDuckGo: %q", desc)
	}
	if !strings.Contains(desc, "China-accessible") {
		t.Fatalf("web_search description should advertise China-accessible sources: %q", desc)
	}
}

func TestParseBingHTML(t *testing.T) {
	html := `<html><body>
<li class="b_algo"><h2><a href="https://example.com/a">Example <b>Title</b></a></h2><p>Example snippet</p></li>
<li class="b_algo"><h2><a href="https://example.com/b">Second</a></h2><p>Second snippet</p></li>
</body></html>`
	got := parseBingHTML(html, 1)
	if len(got) != 1 {
		t.Fatalf("parseBingHTML len = %d, want 1: %#v", len(got), got)
	}
	if got[0].Title != "Example Title" || got[0].URL != "https://example.com/a" || got[0].Snippet != "Example snippet" {
		t.Fatalf("parseBingHTML result = %#v", got[0])
	}
}

func TestParseBaiduHTML(t *testing.T) {
	html := `<html><body>
<div class="result c-container"><h3 class="t"><a href="https://example.com/a">百度 <em>标题</em></a></h3><div class="c-abstract">摘要内容</div></div>
</body></html>`
	got := parseBaiduHTML(html, 5)
	if len(got) != 1 {
		t.Fatalf("parseBaiduHTML len = %d, want 1: %#v", len(got), got)
	}
	if got[0].Title != "百度 标题" || got[0].URL != "https://example.com/a" || got[0].Snippet != "摘要内容" {
		t.Fatalf("parseBaiduHTML result = %#v", got[0])
	}
}

func TestHostCommandSuccessUsesStructuredStatus(t *testing.T) {
	tl := hostCommand{workDir: t.TempDir()}
	command := "printf ok"
	shell := "auto"
	if runtime.GOOS == "windows" {
		command = "echo ok"
		shell = "cmd"
	}
	args, _ := json.Marshal(map[string]any{"command": command, "shell": shell})
	out, err := tl.Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("host_command failed: %v (out=%q)", err, out)
	}
	for _, want := range []string{"status=done", "tool=host_command", "ok"} {
		if !strings.Contains(out, want) {
			t.Fatalf("host_command output missing %q: %q", want, out)
		}
	}
}

func TestHostCommandFailureUsesStructuredStatus(t *testing.T) {
	tl := hostCommand{workDir: t.TempDir()}
	command := "exit 7"
	shell := "auto"
	if runtime.GOOS == "windows" {
		command = "exit /b 7"
		shell = "cmd"
	}
	args, _ := json.Marshal(map[string]any{"command": command, "shell": shell})
	out, err := tl.Execute(context.Background(), args)
	if err == nil {
		t.Fatalf("host_command should fail, got out=%q", out)
	}
	msg := err.Error()
	for _, want := range []string{"status=failed", "tool=host_command", "exit_code=7", "reason=process_returned_non_zero", "advice="} {
		if !strings.Contains(msg, want) {
			t.Fatalf("host_command error missing %q: %q", want, msg)
		}
	}
}

func TestHostExitSummaryShutdownCases(t *testing.T) {
	msg := hostExitSummaryV2("shutdown /s /t 1800", exitErrWithCode(t, 1190), "System shutdown is already scheduled (1190)", "cmd")
	if !strings.Contains(msg, "status=failed") || !strings.Contains(msg, "reason=shutdown_already_scheduled") {
		t.Fatalf("1190 should be classified as already scheduled, got %q", msg)
	}
	msg = hostExitSummaryV2("shutdown /a", exitErrWithCode(t, 1116), "No shutdown is pending", "cmd")
	if !strings.Contains(msg, "reason=shutdown_none_pending") {
		t.Fatalf("no pending shutdown should be classified, got %q", msg)
	}
	msg = hostExitSummaryV2("shutdown /s /t 60", exitErrWithCode(t, 5), "Access is denied.", "cmd")
	if !strings.Contains(msg, "reason=shutdown_access_denied") {
		t.Fatalf("access denied should be classified, got %q", msg)
	}
}

func exitErrWithCode(t *testing.T, code int) error {
	t.Helper()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/c", "exit", "/b", strconv.Itoa(code))
	} else {
		cmd = exec.Command("sh", "-c", "exit "+strconv.Itoa(code))
	}
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected exit error for code %d", code)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *exec.ExitError, got %T", err)
	}
	return err
}
