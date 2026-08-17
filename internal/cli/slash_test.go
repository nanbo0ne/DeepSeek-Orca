package cli

import (
	"testing"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/command"
)

func TestChatCommandNames(t *testing.T) {
	m := chatTUI{commands: []command.Command{{Name: "review"}, {Name: "git:commit"}}}
	if got := m.commandNames(); got != "/review · /git:commit" {
		t.Errorf("commandNames = %q", got)
	}

	if got := (&chatTUI{}).commandNames(); got != "" {
		t.Errorf("empty commandNames = %q, want \"\"", got)
	}
}
