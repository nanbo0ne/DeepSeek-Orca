package hosttools

import "testing"

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
