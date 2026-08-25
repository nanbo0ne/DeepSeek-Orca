package permission

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type fakeAutoReviewer struct {
	allow bool
	err   error
	seen  RiskInput
	calls int
}

func (f *fakeAutoReviewer) Review(_ context.Context, tool, subject string, args json.RawMessage, readOnly bool) (bool, error) {
	f.calls++
	f.seen = RedactedRiskInput(tool, subject, args, readOnly)
	return f.allow, f.err
}

type rejectingApprover struct{}

func (rejectingApprover) Approve(context.Context, string, string, json.RawMessage) (bool, bool, error) {
	return false, false, nil
}

func TestGateAutoReviewPreservesHostPrecedence(t *testing.T) {
	reviewer := &fakeAutoReviewer{allow: true}
	g := NewGate(New("allow", nil, []string{"bash(git commit*)"}, []string{"bash(rm*)"}), nil)
	g.AutoReviewer = reviewer

	allow, _, err := g.Check(context.Background(), "bash", json.RawMessage(`{"command":"go test ./..."}`), false)
	if err != nil || !allow {
		t.Fatalf("fallback auto review allow = %v, err=%v", allow, err)
	}
	if reviewer.seen.Tool != "bash" {
		t.Fatalf("reviewer tool = %q", reviewer.seen.Tool)
	}
	if reviewer.calls != 1 {
		t.Fatalf("fallback review calls = %d, want 1", reviewer.calls)
	}

	g.Approver = rejectingApprover{}
	allow, _, err = g.Check(context.Background(), "bash", json.RawMessage(`{"command":"git commit -m x"}`), false)
	if err != nil || allow {
		t.Fatalf("explicit ask rule should remain manual, allow=%v err=%v", allow, err)
	}
	if reviewer.calls != 1 {
		t.Fatalf("explicit ask called automatic reviewer; calls=%d", reviewer.calls)
	}
	allow, _, err = g.Check(context.Background(), "bash", json.RawMessage(`{"command":"rm -rf build"}`), false)
	if err != nil || allow {
		t.Fatalf("deny rule should remain authoritative, allow=%v err=%v", allow, err)
	}
	if reviewer.calls != 1 {
		t.Fatalf("deny rule called automatic reviewer; calls=%d", reviewer.calls)
	}
}

func TestRedactedRiskInputDropsSecretsAndImages(t *testing.T) {
	input := RedactedRiskInput("host_command", "curl -H 'Authorization: Bearer sk-secret-token'", json.RawMessage(`{"command":"curl","api_key":"sk-123456789","screenshot":"data:image/png;base64,AAAA","nested":{"password":"pw"}}`), false)
	if string(input.Arguments) == "" {
		t.Fatal("redacted arguments are empty")
	}
	text := string(input.Arguments) + input.Operation
	for _, secret := range []string{"sk-123456789", "data:image/png", "pw", "sk-secret-token"} {
		if strings.Contains(text, secret) {
			t.Fatalf("risk payload still contains %q: %s", secret, text)
		}
	}
	if !strings.Contains(text, "REDACTED") || !strings.Contains(text, "OMITTED_IMAGE") {
		t.Fatalf("redaction markers missing: %s", text)
	}
}

func TestGateClassifierFailureKeepsAutoFallback(t *testing.T) {
	reviewer := &fakeAutoReviewer{err: errors.New("invalid json")}
	g := NewGate(New("allow", nil, nil, nil), nil)
	g.AutoReviewer = reviewer
	allow, _, err := g.Check(context.Background(), "write_file", json.RawMessage(`{"path":"a.txt"}`), false)
	if err != nil || !allow {
		t.Fatalf("classifier failure should preserve auto fallback, allow=%v err=%v", allow, err)
	}
}
