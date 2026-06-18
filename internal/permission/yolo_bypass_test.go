package permission

import (
	"context"
	"testing"
)

func TestGateBypassAllowsDenyRule(t *testing.T) {
	g := NewGate(New("ask", nil, nil, []string{"host_command(*)"}), nil)
	g.Bypass = true
	allow, reason, err := g.Check(context.Background(), "host_command", []byte(`{"command":"shutdown /s /t 60"}`), false)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !allow {
		t.Fatalf("Bypass gate should allow deny-listed tool, reason=%q", reason)
	}
}
