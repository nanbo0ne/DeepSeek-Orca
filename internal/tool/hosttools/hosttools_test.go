package hosttools

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/agent"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/config"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/provider"

	"golang.org/x/text/encoding/simplifiedchinese"
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
		"conversation_search",
		"conversation_read",
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

func TestToolsRespectToolLibrarySettings(t *testing.T) {
	settings := config.DefaultToolLibrarySettings()
	settings.WebSearchEnabled = false
	settings.REPLRuntimeEnabled = false
	settings.DocumentToolsEnabled = false
	settings.HostSystemToolsEnabled = false
	settings.ThreadManagementEnabled = false
	settings.ConversationSearchEnabled = false
	tools := Tools(t.TempDir(), settings)
	got := map[string]bool{}
	for _, tl := range tools {
		got[tl.Name()] = true
	}
	for _, name := range []string{"automation_create", "automation_list", "automation_cancel"} {
		if !got[name] {
			t.Fatalf("automation tool %q should remain available", name)
		}
	}
	for _, name := range []string{"host_command", "thread_list", "web_search", "node_repl_exec", "python_repl_exec", "document_inspect", "conversation_search", "conversation_read"} {
		if got[name] {
			t.Fatalf("tool %q should be disabled by tool library settings", name)
		}
	}
}

func TestConversationSearchSkipsSyntheticMessagesAndReadLocator(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("AppData", filepath.Join(home, "AppData"))
	workDir := filepath.Join(home, "workspace")
	dir := config.ProjectSessionDir(workDir)
	s := &agent.Session{Messages: []provider.Message{
		{Role: provider.RoleUser, Content: "<system-reminder>needle hidden</system-reminder>"},
		{Role: provider.RoleUser, Content: "用户要求：查找长期对话检索 needle-visible"},
		{Role: provider.RoleAssistant, Content: "我会实现 conversation_search。"},
	}}
	path := filepath.Join(dir, "session.jsonl")
	if err := s.Save(path); err != nil {
		t.Fatalf("save session: %v", err)
	}

	args, _ := json.Marshal(map[string]any{"query": "needle", "limit": 5})
	out, err := (conversationSearch{workDir: workDir}).Execute(context.Background(), args)
	if err != nil {
		t.Fatalf("conversation_search: %v", err)
	}
	if strings.Contains(out, "needle hidden") {
		t.Fatalf("conversation_search should skip synthetic reminders: %s", out)
	}
	if !strings.Contains(out, "needle-visible") {
		t.Fatalf("conversation_search missing real hit: %s", out)
	}
	var decoded struct {
		Hits []struct {
			Locator string `json:"locator"`
		} `json:"hits"`
	}
	if err := json.Unmarshal([]byte(out), &decoded); err != nil || len(decoded.Hits) != 1 {
		t.Fatalf("decode hits (%d): %v\n%s", len(decoded.Hits), err, out)
	}
	readArgs, _ := json.Marshal(map[string]any{"locator": decoded.Hits[0].Locator, "before": 1, "after": 1})
	readOut, err := (conversationRead{workDir: workDir}).Execute(context.Background(), readArgs)
	if err != nil {
		t.Fatalf("conversation_read: %v", err)
	}
	if !strings.Contains(readOut, "conversation_search") || strings.Contains(readOut, "needle hidden") {
		t.Fatalf("conversation_read output mismatch: %s", readOut)
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

func TestDecodeOutputDecodesGB18030(t *testing.T) {
	raw, err := simplifiedchinese.GB18030.NewEncoder().Bytes([]byte("你好世界\n第二行"))
	if err != nil {
		t.Fatalf("encode gb18030: %v", err)
	}
	got := decodeOutput(raw)
	if got != "你好世界\n第二行" {
		t.Fatalf("decodeOutput = %q", got)
	}
	if strings.ContainsRune(got, '\uFFFD') {
		t.Fatalf("decodeOutput left replacement characters: %q", got)
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
