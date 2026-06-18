package hosttools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"deepseek-orca/internal/config"
	"deepseek-orca/internal/notify"
	"deepseek-orca/internal/tool"
)

// Tools returns the default host-tool library. It intentionally excludes visual
// desktop control, OCR, screenshots, and coordinate input.
func Tools(workDir string) []tool.Tool {
	rt := newRuntimeManager(workDir)
	return []tool.Tool{
		hostCommand{workDir: workDir},
		hostSystemInfo{workDir: workDir},
		hostListProcesses{},
		hostKillProcess{},
		hostOpenApp{workDir: workDir},
		hostClipboard{},
		notifyUser{},
		automationCreate{workDir: workDir},
		automationList{},
		automationCancel{},
		threadList{},
		webSearch{},
		nodeRepl{rt: rt},
		pythonRepl{rt: rt},
		documentInspect{workDir: workDir},
		documentExtract{workDir: workDir},
	}
}

var automations = newAutomationStore()

type automationStore struct {
	mu    sync.Mutex
	next  int
	items map[string]*automationItem
}

type automationItem struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	Action    string    `json:"action"`
	Command   string    `json:"command,omitempty"`
	Message   string    `json:"message,omitempty"`
	RunAt     time.Time `json:"runAt"`
	Status    string    `json:"status"`
	Result    string    `json:"result,omitempty"`
	cancel    context.CancelFunc
	createdAt time.Time
}

func newAutomationStore() *automationStore {
	return &automationStore{items: map[string]*automationItem{}}
}

func (s *automationStore) add(item *automationItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	item.ID = fmt.Sprintf("automation-%d", s.next)
	item.createdAt = time.Now()
	s.items[item.ID] = item
}

func (s *automationStore) list() []automationItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]automationItem, 0, len(s.items))
	for _, item := range s.items {
		copy := *item
		copy.cancel = nil
		out = append(out, copy)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].createdAt.Before(out[j].createdAt) })
	return out
}

func (s *automationStore) cancel(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := s.items[id]
	if item == nil || item.Status != "scheduled" {
		return false
	}
	item.Status = "cancelled"
	if item.cancel != nil {
		item.cancel()
	}
	return true
}

type automationCreate struct{ workDir string }

func (automationCreate) Name() string { return "automation_create" }
func (automationCreate) Description() string {
	return "Create an unattended timed automation. It can show a reminder notification or run a native host command at a future time. The automation must not ask the user questions while running; it should finish, self-check obvious failures, and notify success or failure."
}
func (automationCreate) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"label":{"type":"string"},"delay_seconds":{"type":"integer","minimum":1,"description":"Delay before running. Use either delay_seconds or run_at."},"run_at":{"type":"string","description":"RFC3339 timestamp for when to run."},"action":{"type":"string","enum":["notify","host_command"],"description":"Automation action."},"message":{"type":"string","description":"Notification body for notify action."},"command":{"type":"string","description":"Native host command for host_command action."}},"required":["action"]}`)
}
func (automationCreate) ReadOnly() bool { return false }
func (a automationCreate) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Label        string `json:"label"`
		DelaySeconds int    `json:"delay_seconds"`
		RunAt        string `json:"run_at"`
		Action       string `json:"action"`
		Message      string `json:"message"`
		Command      string `json:"command"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", err
	}
	runAt := time.Now().Add(time.Duration(p.DelaySeconds) * time.Second)
	if strings.TrimSpace(p.RunAt) != "" {
		t, err := time.Parse(time.RFC3339, p.RunAt)
		if err != nil {
			return "", fmt.Errorf("run_at must be RFC3339: %w", err)
		}
		runAt = t
	}
	if !runAt.After(time.Now()) {
		return "", fmt.Errorf("scheduled time must be in the future")
	}
	if p.Label == "" {
		p.Label = p.Action
	}
	if p.Action == "notify" && strings.TrimSpace(p.Message) == "" {
		return "", fmt.Errorf("message is required for notify automation")
	}
	if p.Action == "host_command" && strings.TrimSpace(p.Command) == "" {
		return "", fmt.Errorf("command is required for host_command automation")
	}
	runCtx, cancel := context.WithCancel(context.Background())
	item := &automationItem{Label: p.Label, Action: p.Action, Command: p.Command, Message: p.Message, RunAt: runAt, Status: "scheduled", cancel: cancel}
	automations.add(item)
	go runAutomation(runCtx, a.workDir, item)
	_ = ctx
	return fmt.Sprintf("Scheduled %s for %s (id: %s).", item.Label, item.RunAt.Format(time.RFC3339), item.ID), nil
}

func runAutomation(ctx context.Context, workDir string, item *automationItem) {
	timer := time.NewTimer(time.Until(item.RunAt))
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
		return
	}
	automations.mu.Lock()
	if item.Status != "scheduled" {
		automations.mu.Unlock()
		return
	}
	item.Status = "running"
	automations.mu.Unlock()
	var result string
	var err error
	switch item.Action {
	case "notify":
		err = notify.NewPlatformSender().Send(notify.Message{Title: firstNonEmpty(item.Label, "DeepSeek-Orca 自动化"), Body: item.Message})
		result = "notification sent"
	case "host_command":
		cmd := exec.CommandContext(ctx, nativeShellArgv("auto", item.Command)[0], nativeShellArgv("auto", item.Command)[1:]...)
		cmd.Dir = workDir
		out, runErr := cmd.CombinedOutput()
		result = decodeOutput(out)
		err = runErr
		if err == nil {
			_ = notify.NewPlatformSender().Send(notify.Message{Title: "DeepSeek-Orca 自动化完成", Body: item.Label})
		} else {
			_ = notify.NewPlatformSender().Send(notify.Message{Title: "DeepSeek-Orca 自动化失败", Body: item.Label + ": " + err.Error()})
		}
	}
	automations.mu.Lock()
	defer automations.mu.Unlock()
	if err != nil {
		item.Status = "failed"
		item.Result = strings.TrimSpace(result + "\n" + err.Error())
		return
	}
	item.Status = "done"
	item.Result = result
}

type automationList struct{}

func (automationList) Name() string { return "automation_list" }
func (automationList) Description() string {
	return "List timed automations created in this app process."
}
func (automationList) Schema() json.RawMessage { return json.RawMessage(`{"type":"object","properties":{}}`) }
func (automationList) ReadOnly() bool          { return true }
func (automationList) Execute(context.Context, json.RawMessage) (string, error) {
	b, _ := json.MarshalIndent(automations.list(), "", "  ")
	return string(b), nil
}

type automationCancel struct{}

func (automationCancel) Name() string { return "automation_cancel" }
func (automationCancel) Description() string {
	return "Cancel a scheduled automation that has not started yet."
}
func (automationCancel) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"id":{"type":"string"}},"required":["id"]}`)
}
func (automationCancel) ReadOnly() bool { return false }
func (automationCancel) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var p struct{ ID string `json:"id"` }
	if err := json.Unmarshal(args, &p); err != nil {
		return "", err
	}
	if automations.cancel(p.ID) {
		return "Automation cancelled.", nil
	}
	return "Automation was not scheduled or was not found.", nil
}

type threadList struct{}

func (threadList) Name() string { return "thread_list" }
func (threadList) Description() string {
	return "List saved DeepSeek-Orca conversation threads/topics from the local session store. This manages Orca sessions, not Codex threads."
}
func (threadList) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"limit":{"type":"integer","minimum":1,"maximum":200}}}`)
}
func (threadList) ReadOnly() bool { return true }
func (threadList) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var p struct{ Limit int `json:"limit"` }
	_ = json.Unmarshal(args, &p)
	if p.Limit <= 0 || p.Limit > 200 {
		p.Limit = 50
	}
	root := config.SessionDir()
	var rows []map[string]any
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".jsonl") {
			return nil
		}
		st, err := d.Info()
		if err != nil {
			return nil
		}
		rows = append(rows, map[string]any{"path": path, "modified": st.ModTime().Format(time.RFC3339), "sizeBytes": st.Size()})
		return nil
	})
	sort.Slice(rows, func(i, j int) bool {
		return fmt.Sprint(rows[i]["modified"]) > fmt.Sprint(rows[j]["modified"])
	})
	if len(rows) > p.Limit {
		rows = rows[:p.Limit]
	}
	b, _ := json.MarshalIndent(rows, "", "  ")
	return string(b), nil
}

type hostCommand struct{ workDir string }

func (hostCommand) Name() string { return "host_command" }
func (hostCommand) Description() string {
	return "Execute a native host command directly through the operating system shell, bypassing Git Bash argument rewriting. On Windows, use this for native commands such as shutdown, taskkill, start, sc, reg, and PowerShell. This can alter the computer and requires approval unless tool approval is auto/yolo."
}
func (hostCommand) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"command":{"type":"string","description":"Command text to execute in the native host shell."},"shell":{"type":"string","enum":["auto","cmd","powershell"],"description":"Windows shell choice. auto uses cmd on Windows and sh elsewhere."},"timeout_seconds":{"type":"integer","minimum":1,"maximum":3600,"description":"Optional timeout in seconds. Defaults to 120."}},"required":["command"]}`)
}
func (hostCommand) ReadOnly() bool { return false }
func (h hostCommand) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Command        string `json:"command"`
		Shell          string `json:"shell"`
		TimeoutSeconds int    `json:"timeout_seconds"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	p.Command = strings.TrimSpace(p.Command)
	if p.Command == "" {
		return "", fmt.Errorf("command is required")
	}
	timeout := time.Duration(p.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	argv := nativeShellArgv(p.Shell, p.Command)
	cmd := exec.CommandContext(runCtx, argv[0], argv[1:]...)
	cmd.Dir = h.workDir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	out := decodeOutput(buf.Bytes())
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return out, fmt.Errorf("host command timed out after %s", timeout)
	}
	if err != nil {
		return out, fmt.Errorf("%s", hostExitSummary(p.Command, err, out))
	}
	if strings.TrimSpace(out) == "" {
		return "Command completed with no output.", nil
	}
	return out, nil
}

func nativeShellArgv(shell, command string) []string {
	if runtime.GOOS == "windows" {
		switch strings.ToLower(strings.TrimSpace(shell)) {
		case "powershell", "pwsh":
			return []string{"powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", "$OutputEncoding=[Console]::OutputEncoding=[System.Text.Encoding]::UTF8;" + command}
		default:
			return []string{"cmd.exe", "/c", command}
		}
	}
	return []string{"sh", "-c", command}
}

func hostExitSummary(command string, err error, output string) string {
	code := -1
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code = exitErr.ExitCode()
	}
	lowerCmd := strings.ToLower(strings.TrimSpace(command))
	lowerOut := strings.ToLower(output)
	reason := ""
	if strings.HasPrefix(lowerCmd, "shutdown") {
		switch {
		case strings.Contains(lowerOut, "access is denied") || strings.Contains(lowerOut, "拒绝访问"):
			reason = "shutdown failed: current process is not elevated or local policy denied the operation"
		case strings.Contains(lowerOut, "usage:") || strings.Contains(lowerOut, "用法"):
			reason = "shutdown failed: arguments were rejected by Windows shutdown.exe"
		default:
			reason = "shutdown failed: Windows rejected the request; check administrator permission, policy, or security software"
		}
	}
	if reason == "" && code >= 0 {
		reason = "process returned non-zero exit code " + strconv.Itoa(code)
	}
	if code >= 0 {
		return fmt.Sprintf("host command exited with code %d: %s", code, reason)
	}
	return fmt.Sprintf("host command failed: %v", err)
}

type hostSystemInfo struct{ workDir string }

func (hostSystemInfo) Name() string { return "host_system_info" }
func (hostSystemInfo) Description() string {
	return "Return host OS, architecture, current user, working directory, PATH shell hints, and administrator/elevation hint. Read-only."
}
func (hostSystemInfo) Schema() json.RawMessage { return json.RawMessage(`{"type":"object","properties":{}}`) }
func (hostSystemInfo) ReadOnly() bool          { return true }
func (h hostSystemInfo) Execute(context.Context, json.RawMessage) (string, error) {
	user := os.Getenv("USERNAME")
	if user == "" {
		user = os.Getenv("USER")
	}
	info := map[string]any{
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"user":      user,
		"workDir":   h.workDir,
		"path":      os.Getenv("PATH"),
		"elevated":  elevationHint(),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	b, _ := json.MarshalIndent(info, "", "  ")
	return string(b), nil
}

func elevationHint() string {
	if runtime.GOOS != "windows" {
		if os.Geteuid() == 0 {
			return "root"
		}
		return "not-root"
	}
	cmd := exec.Command("cmd.exe", "/c", "net session >nul 2>nul")
	if err := cmd.Run(); err == nil {
		return "administrator"
	}
	return "not-administrator-or-unknown"
}

type hostListProcesses struct{}

func (hostListProcesses) Name() string { return "host_list_processes" }
func (hostListProcesses) Description() string {
	return "List running processes as a compact text table. Read-only."
}
func (hostListProcesses) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"filter":{"type":"string","description":"Optional case-insensitive substring filter."},"limit":{"type":"integer","minimum":1,"maximum":500,"description":"Maximum rows. Defaults to 80."}}}`)
}
func (hostListProcesses) ReadOnly() bool { return true }
func (hostListProcesses) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Filter string `json:"filter"`
		Limit  int    `json:"limit"`
	}
	_ = json.Unmarshal(args, &p)
	if p.Limit <= 0 || p.Limit > 500 {
		p.Limit = 80
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", "Get-Process | Select-Object Id,ProcessName,CPU,WS | ConvertTo-Csv -NoTypeInformation")
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", "ps -eo pid,comm,%cpu,rss | head -n 500")
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(strings.TrimSpace(decodeOutput(out)), "\n")
	filter := strings.ToLower(strings.TrimSpace(p.Filter))
	var kept []string
	for _, line := range lines {
		if filter == "" || strings.Contains(strings.ToLower(line), filter) || len(kept) == 0 {
			kept = append(kept, line)
			if len(kept) >= p.Limit+1 {
				break
			}
		}
	}
	return strings.Join(kept, "\n"), nil
}

type hostKillProcess struct{}

func (hostKillProcess) Name() string { return "host_kill_process" }
func (hostKillProcess) Description() string {
	return "Terminate a host process by pid or process name. This is destructive and requires approval unless auto/yolo is active."
}
func (hostKillProcess) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"pid":{"type":"integer","description":"Process ID to terminate."},"name":{"type":"string","description":"Process name to terminate when pid is not set."},"force":{"type":"boolean","description":"Force termination. Defaults true."}}}`)
}
func (hostKillProcess) ReadOnly() bool { return false }
func (hostKillProcess) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		PID   int    `json:"pid"`
		Name  string `json:"name"`
		Force bool   `json:"force"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", err
	}
	if p.PID <= 0 && strings.TrimSpace(p.Name) == "" {
		return "", fmt.Errorf("pid or name is required")
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		if p.PID > 0 {
			args := []string{"/PID", strconv.Itoa(p.PID)}
			if p.Force || !strings.Contains(string(argsJSON(args)), "force:false") {
				args = append([]string{"/F"}, args...)
			}
			cmd = exec.CommandContext(ctx, "taskkill", args...)
		} else {
			args := []string{"/IM", p.Name}
			if p.Force || !strings.Contains(string(argsJSON(args)), "force:false") {
				args = append([]string{"/F"}, args...)
			}
			cmd = exec.CommandContext(ctx, "taskkill", args...)
		}
	} else if p.PID > 0 {
		cmd = exec.CommandContext(ctx, "kill", "-TERM", strconv.Itoa(p.PID))
	} else {
		cmd = exec.CommandContext(ctx, "pkill", p.Name)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return decodeOutput(out), err
	}
	return firstNonEmpty(decodeOutput(out), "Process termination requested."), nil
}

func argsJSON(v any) []byte { b, _ := json.Marshal(v); return b }

type hostOpenApp struct{ workDir string }

func (hostOpenApp) Name() string { return "host_open_app" }
func (hostOpenApp) Description() string {
	return "Launch an installed application or executable path. Does not automate the UI after launch."
}
func (hostOpenApp) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"target":{"type":"string","description":"Executable path, document path, URL, or app command to open."},"args":{"type":"array","items":{"type":"string"},"description":"Optional arguments when launching an executable."}},"required":["target"]}`)
}
func (hostOpenApp) ReadOnly() bool { return false }
func (h hostOpenApp) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Target string   `json:"target"`
		Args   []string `json:"args"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", err
	}
	if strings.TrimSpace(p.Target) == "" {
		return "", fmt.Errorf("target is required")
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		if len(p.Args) == 0 {
			cmd = exec.CommandContext(ctx, "cmd.exe", "/c", "start", "", p.Target)
		} else {
			cmd = exec.CommandContext(ctx, p.Target, p.Args...)
		}
	} else if runtime.GOOS == "darwin" && len(p.Args) == 0 {
		cmd = exec.CommandContext(ctx, "open", p.Target)
	} else if len(p.Args) == 0 {
		cmd = exec.CommandContext(ctx, "xdg-open", p.Target)
	} else {
		cmd = exec.CommandContext(ctx, p.Target, p.Args...)
	}
	cmd.Dir = h.workDir
	if err := cmd.Start(); err != nil {
		return "", err
	}
	return "Launch requested.", nil
}

type hostClipboard struct{}

func (hostClipboard) Name() string { return "host_clipboard" }
func (hostClipboard) Description() string {
	return "Read or write plain text clipboard content. Does not support images."
}
func (hostClipboard) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"op":{"type":"string","enum":["read","write"],"description":"Clipboard operation."},"text":{"type":"string","description":"Text to write when op=write."}},"required":["op"]}`)
}
func (hostClipboard) ReadOnly() bool { return false }
func (hostClipboard) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Op   string `json:"op"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", err
	}
	switch strings.ToLower(p.Op) {
	case "read":
		out, err := clipboardRead(ctx)
		if err != nil {
			return "", err
		}
		return out, nil
	case "write":
		if err := clipboardWrite(ctx, p.Text); err != nil {
			return "", err
		}
		return "Clipboard text updated.", nil
	default:
		return "", fmt.Errorf("op must be read or write")
	}
}

func clipboardRead(ctx context.Context) (string, error) {
	if runtime.GOOS == "windows" {
		out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", "Get-Clipboard -Raw").Output()
		return decodeOutput(out), err
	}
	if runtime.GOOS == "darwin" {
		out, err := exec.CommandContext(ctx, "pbpaste").Output()
		return decodeOutput(out), err
	}
	out, err := exec.CommandContext(ctx, "sh", "-c", "xclip -selection clipboard -o 2>/dev/null || xsel --clipboard --output 2>/dev/null").Output()
	return decodeOutput(out), err
}

func clipboardWrite(ctx context.Context, text string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", "Set-Clipboard -Value $args[0]", text)
		return cmd.Run()
	}
	if runtime.GOOS == "darwin" {
		cmd := exec.CommandContext(ctx, "pbcopy")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}
	cmd := exec.CommandContext(ctx, "sh", "-c", "xclip -selection clipboard 2>/dev/null || xsel --clipboard --input 2>/dev/null")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

type notifyUser struct{}

func (notifyUser) Name() string { return "notify_user" }
func (notifyUser) Description() string {
	return "Show a best-effort desktop notification to the user. Use when a long task, automation, or background job finishes or fails."
}
func (notifyUser) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"title":{"type":"string"},"body":{"type":"string"}},"required":["body"]}`)
}
func (notifyUser) ReadOnly() bool { return false }
func (notifyUser) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", err
	}
	if p.Title == "" {
		p.Title = "DeepSeek-Orca"
	}
	if err := notify.NewPlatformSender().Send(notify.Message{Title: p.Title, Body: p.Body}); err != nil {
		return "", err
	}
	return "Notification sent.", nil
}

type webSearch struct{}

func (webSearch) Name() string { return "web_search" }
func (webSearch) Description() string {
	return "Search the web for current information and return structured title/url/snippet results. Use when you do not already know the URL. Follow up with web_fetch to read a selected result."
}
func (webSearch) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"},"limit":{"type":"integer","minimum":1,"maximum":10}},"required":["query"]}`)
}
func (webSearch) ReadOnly() bool { return true }
func (webSearch) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", err
	}
	if strings.TrimSpace(p.Query) == "" {
		return "", fmt.Errorf("query is required")
	}
	if p.Limit <= 0 || p.Limit > 10 {
		p.Limit = 5
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://duckduckgo.com/html/?q="+url.QueryEscape(p.Query), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 DeepSeek-Orca")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	results := parseDuckDuckGoHTML(string(body), p.Limit)
	if len(results) == 0 {
		return "No search results parsed. Try a more specific query or use web_fetch with a known URL.", nil
	}
	b, _ := json.MarshalIndent(results, "", "  ")
	return string(b), nil
}

type searchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet,omitempty"`
}

func parseDuckDuckGoHTML(html string, limit int) []searchResult {
	var out []searchResult
	parts := strings.Split(html, `class="result__a"`)
	for _, part := range parts[1:] {
		if len(out) >= limit {
			break
		}
		href := extractBetween(part, `href="`, `"`)
		title := stripHTML(extractBetween(part, ">", "</a>"))
		snippet := stripHTML(extractBetween(part, `class="result__snippet"`, "</a>"))
		if i := strings.Index(snippet, ">"); i >= 0 {
			snippet = strings.TrimSpace(snippet[i+1:])
		}
		if strings.Contains(href, "uddg=") {
			if u, err := url.Parse(href); err == nil {
				if decoded := u.Query().Get("uddg"); decoded != "" {
					href = decoded
				}
			}
		}
		if title != "" && href != "" {
			out = append(out, searchResult{Title: htmlUnescape(title), URL: htmlUnescape(href), Snippet: htmlUnescape(snippet)})
		}
	}
	return out
}

type runtimeManager struct {
	workDir string
	mu      sync.Mutex
	node    *replState
	python  *replState
}

type replState struct {
	Vars map[string]string `json:"vars"`
}

func newRuntimeManager(workDir string) *runtimeManager {
	return &runtimeManager{workDir: workDir}
}

type nodeRepl struct{ rt *runtimeManager }

func (nodeRepl) Name() string { return "node_repl_exec" }
func (nodeRepl) Description() string {
	return "Execute JavaScript in a lightweight persistent Node-style REPL. Variables set through the vars object persist across calls. Use for scripts, JSON transforms, package checks, and calculations; not for visual desktop automation."
}
func (nodeRepl) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"code":{"type":"string"},"vars":{"type":"object","additionalProperties":{"type":"string"},"description":"String variables to persist before execution."},"timeout_seconds":{"type":"integer","minimum":1,"maximum":600}},"required":["code"]}`)
}
func (nodeRepl) ReadOnly() bool { return false }
func (n nodeRepl) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Code           string            `json:"code"`
		Vars           map[string]string `json:"vars"`
		TimeoutSeconds int               `json:"timeout_seconds"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", err
	}
	n.rt.mu.Lock()
	if n.rt.node == nil {
		n.rt.node = &replState{Vars: map[string]string{}}
	}
	for k, v := range p.Vars {
		n.rt.node.Vars[k] = v
	}
	envJSON, _ := json.Marshal(n.rt.node.Vars)
	n.rt.mu.Unlock()
	script := "globalThis.vars = " + string(envJSON) + ";\n" + p.Code + "\nif (globalThis.vars) console.error('__VARS__'+JSON.stringify(globalThis.vars));"
	return runInterpreter(ctx, n.rt.workDir, "node", []string{"-e", script}, p.TimeoutSeconds, func(stderr string) {
		if i := strings.LastIndex(stderr, "__VARS__"); i >= 0 {
			var vars map[string]string
			if json.Unmarshal([]byte(strings.TrimSpace(stderr[i+8:])), &vars) == nil {
				n.rt.mu.Lock()
				n.rt.node.Vars = vars
				n.rt.mu.Unlock()
			}
		}
	})
}

type pythonRepl struct{ rt *runtimeManager }

func (pythonRepl) Name() string { return "python_repl_exec" }
func (pythonRepl) Description() string {
	return "Execute Python in a lightweight persistent REPL. Variables set through vars persist as strings across calls. Use for data processing, document helpers, calculations, and scripts."
}
func (pythonRepl) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"code":{"type":"string"},"vars":{"type":"object","additionalProperties":{"type":"string"}},"timeout_seconds":{"type":"integer","minimum":1,"maximum":600}},"required":["code"]}`)
}
func (pythonRepl) ReadOnly() bool { return false }
func (pyt pythonRepl) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Code           string            `json:"code"`
		Vars           map[string]string `json:"vars"`
		TimeoutSeconds int               `json:"timeout_seconds"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", err
	}
	pyt.rt.mu.Lock()
	if pyt.rt.python == nil {
		pyt.rt.python = &replState{Vars: map[string]string{}}
	}
	for k, v := range p.Vars {
		pyt.rt.python.Vars[k] = v
	}
	envJSON, _ := json.Marshal(pyt.rt.python.Vars)
	pyt.rt.mu.Unlock()
	script := "import json, sys\nvars = json.loads(" + strconv.Quote(string(envJSON)) + ")\n" + p.Code + "\nprint('__VARS__'+json.dumps(vars), file=sys.stderr)"
	return runInterpreter(ctx, pyt.rt.workDir, pythonExe(), []string{"-c", script}, p.TimeoutSeconds, func(stderr string) {
		if i := strings.LastIndex(stderr, "__VARS__"); i >= 0 {
			var vars map[string]string
			if json.Unmarshal([]byte(strings.TrimSpace(stderr[i+8:])), &vars) == nil {
				pyt.rt.mu.Lock()
				pyt.rt.python.Vars = vars
				pyt.rt.mu.Unlock()
			}
		}
	})
}

func runInterpreter(ctx context.Context, dir, exe string, args []string, timeoutSec int, capture func(stderr string)) (string, error) {
	timeout := time.Duration(timeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(runCtx, exe, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	errText := decodeOutput(stderr.Bytes())
	if capture != nil {
		capture(errText)
	}
	cleanErr := strings.TrimSpace(stripVarsMarker(errText))
	out := strings.TrimSpace(decodeOutput(stdout.Bytes()))
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return out, fmt.Errorf("%s timed out after %s", exe, timeout)
	}
	if err != nil {
		if cleanErr != "" {
			return out, fmt.Errorf("%s failed: %s", exe, cleanErr)
		}
		return out, err
	}
	if cleanErr != "" {
		return out + "\nstderr:\n" + cleanErr, nil
	}
	if out == "" {
		return "Execution completed with no output.", nil
	}
	return out, nil
}

func stripVarsMarker(s string) string {
	if i := strings.LastIndex(s, "__VARS__"); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

func pythonExe() string {
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "python3"
}

type documentInspect struct{ workDir string }

func (documentInspect) Name() string { return "document_inspect" }
func (documentInspect) Description() string {
	return "Inspect a Word/PPT/Excel/PDF or text-like file and return metadata, size, type, and available processing hints."
}
func (documentInspect) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (documentInspect) ReadOnly() bool { return true }
func (d documentInspect) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var p struct{ Path string `json:"path"` }
	if err := json.Unmarshal(args, &p); err != nil {
		return "", err
	}
	path := resolvePath(d.workDir, p.Path)
	st, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	info := map[string]any{
		"path":      path,
		"sizeBytes": st.Size(),
		"modified":  st.ModTime().Format(time.RFC3339),
		"extension": strings.ToLower(filepath.Ext(path)),
		"toolHint":  documentHint(path),
	}
	b, _ := json.MarshalIndent(info, "", "  ")
	return string(b), nil
}

type documentExtract struct{ workDir string }

func (documentExtract) Name() string { return "document_extract" }
func (documentExtract) Description() string {
	return "Extract text from text/markdown/json/csv files and best-effort text from PDF/DOCX/XLSX/PPTX using available local runtimes. For complex editing use python_repl_exec with document libraries."
}
func (documentExtract) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"path":{"type":"string"},"max_bytes":{"type":"integer","minimum":1024,"maximum":10485760}},"required":["path"]}`)
}
func (documentExtract) ReadOnly() bool { return true }
func (d documentExtract) Execute(_ context.Context, args json.RawMessage) (string, error) {
	var p struct {
		Path     string `json:"path"`
		MaxBytes int64  `json:"max_bytes"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", err
	}
	if p.MaxBytes <= 0 {
		p.MaxBytes = 1 << 20
	}
	path := resolvePath(d.workDir, p.Path)
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt", ".md", ".json", ".csv", ".tsv", ".xml", ".html", ".log", ".go", ".ts", ".tsx", ".js", ".py":
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		if int64(len(b)) > p.MaxBytes {
			b = b[:p.MaxBytes]
		}
		return decodeOutput(b), nil
	case ".pdf":
		return runDocPython(path, "pdf")
	case ".docx":
		return runDocPython(path, "docx")
	case ".xlsx":
		return runDocPython(path, "xlsx")
	case ".pptx":
		return runDocPython(path, "pptx")
	default:
		return "", fmt.Errorf("unsupported extension %q; use python_repl_exec for custom extraction", ext)
	}
}

func runDocPython(path, kind string) (string, error) {
	code := ""
	switch kind {
	case "pdf":
		code = `import sys
try:
 import PyPDF2
 r=PyPDF2.PdfReader(sys.argv[1])
 print("\n".join((p.extract_text() or "") for p in r.pages))
except Exception as e:
 print("PDF extraction requires PyPDF2 or another PDF library: "+str(e), file=sys.stderr); sys.exit(1)`
	case "docx":
		code = `import sys, zipfile, re
with zipfile.ZipFile(sys.argv[1]) as z:
 data=z.read("word/document.xml").decode("utf-8", "ignore")
 print(re.sub(r"<[^>]+>", " ", data))`
	case "xlsx":
		code = `import sys, zipfile, re
with zipfile.ZipFile(sys.argv[1]) as z:
 names=[n for n in z.namelist() if n.startswith("xl/worksheets/") and n.endswith(".xml")]
 for n in names[:20]:
  print("## "+n)
  print(re.sub(r"<[^>]+>", " ", z.read(n).decode("utf-8","ignore"))[:20000])`
	case "pptx":
		code = `import sys, zipfile, re
with zipfile.ZipFile(sys.argv[1]) as z:
 names=[n for n in z.namelist() if n.startswith("ppt/slides/") and n.endswith(".xml")]
 for n in names:
  print("## "+n)
  print(re.sub(r"<[^>]+>", " ", z.read(n).decode("utf-8","ignore")))`
	}
	cmd := exec.Command(pythonExe(), "-c", code, path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return decodeOutput(out), err
	}
	return decodeOutput(out), nil
}

func documentHint(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".docx":
		return "Word DOCX: document_extract can read XML text; use python_repl_exec for structured editing."
	case ".pptx":
		return "PowerPoint PPTX: document_extract can read slide XML text; use python_repl_exec for structured editing."
	case ".xlsx":
		return "Excel XLSX: document_extract can read worksheet XML; use python_repl_exec for formulas/tables."
	case ".pdf":
		return "PDF: document_extract tries PyPDF2 when installed."
	default:
		return "Text-like file or unsupported binary."
	}
}

func resolvePath(workDir, p string) string {
	if filepath.IsAbs(p) || workDir == "" {
		return filepath.Clean(p)
	}
	return filepath.Join(workDir, p)
}

func decodeOutput(b []byte) string {
	return strings.ReplaceAll(string(b), "\r\n", "\n")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func extractBetween(s, start, end string) string {
	i := strings.Index(s, start)
	if i < 0 {
		return ""
	}
	s = s[i+len(start):]
	j := strings.Index(s, end)
	if j < 0 {
		return ""
	}
	return s[:j]
}

func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				b.WriteRune(r)
			}
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func htmlUnescape(s string) string {
	r := strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'", "&#x27;", "'")
	return r.Replace(s)
}

func init() {
	// Keep deterministic tool order in tests that inspect registered names.
	_ = sort.Strings
}
