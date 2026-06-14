package bot

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"deepseek-orca/internal/agent"
	"deepseek-orca/internal/control"
	"deepseek-orca/internal/event"
	"deepseek-orca/internal/provider"
)

// fakeAdapter 是一个内存中的假适配器，用于测试 BotGateway。
type fakeAdapter struct {
	mu       sync.Mutex
	platform Platform
	name     string
	msgCh    chan InboundMessage
	sent     []OutboundMessage
	started  bool
}

func newFakeAdapter(platform Platform, name string) *fakeAdapter {
	return &fakeAdapter{
		platform: platform,
		name:     name,
		msgCh:    make(chan InboundMessage, 16),
	}
}

func (f *fakeAdapter) Platform() Platform              { return f.platform }
func (f *fakeAdapter) Name() string                    { return f.name }
func (f *fakeAdapter) Messages() <-chan InboundMessage { return f.msgCh }

func (f *fakeAdapter) Start(ctx context.Context) error {
	f.mu.Lock()
	f.started = true
	f.mu.Unlock()
	return nil
}

func (f *fakeAdapter) Stop() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.msgCh != nil {
		close(f.msgCh)
		f.msgCh = nil
	}
	return nil
}

func (f *fakeAdapter) Send(ctx context.Context, msg OutboundMessage) (SendResult, error) {
	f.mu.Lock()
	f.sent = append(f.sent, msg)
	f.mu.Unlock()
	return SendResult{MessageID: "fake_msg_1"}, nil
}

func (f *fakeAdapter) SendTyping(ctx context.Context, chatID string) error { return nil }

func (f *fakeAdapter) sentMessages() []OutboundMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]OutboundMessage, len(f.sent))
	copy(out, f.sent)
	return out
}

func TestFakeAdapterInterface(t *testing.T) {
	fa := newFakeAdapter(PlatformQQ, "fake-qq")

	if fa.Platform() != PlatformQQ {
		t.Error("wrong platform")
	}
	if fa.Name() != "fake-qq" {
		t.Error("wrong name")
	}

	ctx := context.Background()
	if err := fa.Start(ctx); err != nil {
		t.Fatal("start:", err)
	}
	if !fa.started {
		t.Error("should be started")
	}

	_, err := fa.Send(ctx, OutboundMessage{ChatID: "c1", Text: "hello"})
	if err != nil {
		t.Fatal("send:", err)
	}

	sent := fa.sentMessages()
	if len(sent) != 1 {
		t.Fatalf("sent count = %d, want 1", len(sent))
	}
	if sent[0].Text != "hello" {
		t.Errorf("sent text = %q, want %q", sent[0].Text, "hello")
	}

	if err := fa.Stop(); err != nil {
		t.Fatal("stop:", err)
	}
}

func TestGatewayConstructAndStop(t *testing.T) {
	cfg := GatewayConfig{
		Model:         "test",
		MaxSteps:      10,
		WorkspaceRoot: ".",
		Enabled:       map[Platform]bool{PlatformQQ: true},
		Allowlist:     AllowlistConfig{Enabled: false},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gw := NewGateway(cfg, map[Platform]Adapter{
		PlatformQQ: newFakeAdapter(PlatformQQ, "fake-qq"),
	}, logger)

	// 网关不应该 panic
	if gw == nil {
		t.Fatal("gateway should not be nil")
	}
	gw.Stop()
}

func TestGatewayAllowlistCheck(t *testing.T) {
	cfg := GatewayConfig{
		Allowlist: AllowlistConfig{
			Enabled: true,
			Users: map[Platform][]string{
				PlatformQQ: {"allowed_user_1"},
			},
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gw := NewGateway(cfg, nil, logger)

	if !gw.checkAllowlist(PlatformQQ, InboundMessage{Platform: PlatformQQ, ChatType: ChatDM, UserID: "allowed_user_1"}) {
		t.Error("allowed user should pass")
	}
	if gw.checkAllowlist(PlatformQQ, InboundMessage{Platform: PlatformQQ, ChatType: ChatDM, UserID: "unknown_user"}) {
		t.Error("unknown user should not pass")
	}
	// 不同平台
	if gw.checkAllowlist(PlatformFeishu, InboundMessage{Platform: PlatformFeishu, ChatType: ChatDM, UserID: "allowed_user_1"}) {
		t.Error("QQ allowlist should not apply to feishu")
	}
}

func TestGatewayAllowlistDoesNotApplyGroupsToDirectMessages(t *testing.T) {
	cfg := GatewayConfig{
		Allowlist: AllowlistConfig{
			Enabled: true,
			Users: map[Platform][]string{
				PlatformQQ: {"allowed_user"},
			},
			Groups: map[Platform][]string{
				PlatformQQ: {"allowed_group"},
			},
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gw := NewGateway(cfg, nil, logger)

	if !gw.checkAllowlist(PlatformQQ, InboundMessage{Platform: PlatformQQ, ChatType: ChatDirect, ChatID: "guild-dm", UserID: "allowed_user"}) {
		t.Error("direct message should not be rejected by group allowlist")
	}
	if gw.checkAllowlist(PlatformQQ, InboundMessage{Platform: PlatformQQ, ChatType: ChatGroup, ChatID: "unknown_group", UserID: "allowed_user"}) {
		t.Error("unknown group should still be rejected by group allowlist")
	}
}

func TestGatewayAllowlistDisabledRejectsByDefault(t *testing.T) {
	cfg := GatewayConfig{
		Allowlist: AllowlistConfig{Enabled: false},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gw := NewGateway(cfg, nil, logger)

	if gw.checkAllowlist(PlatformQQ, InboundMessage{Platform: PlatformQQ, ChatType: ChatDM, UserID: "any_user"}) {
		t.Error("disabled allowlist should reject unless allow_all is explicit")
	}
}

func TestGatewayAllowAll(t *testing.T) {
	cfg := GatewayConfig{
		Allowlist: AllowlistConfig{AllowAll: true},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	gw := NewGateway(cfg, nil, logger)

	if !gw.checkAllowlist(PlatformQQ, InboundMessage{Platform: PlatformQQ, ChatType: ChatDM, UserID: "any_user"}) {
		t.Error("allow_all should allow everyone")
	}
}

func TestGatewayStartListsRecentSessions(t *testing.T) {
	fa := newFakeAdapter(PlatformQQ, "fake-qq")
	cfg := GatewayConfig{
		Enabled:   map[Platform]bool{PlatformQQ: true},
		Allowlist: AllowlistConfig{AllowAll: true},
		SessionLister: func(limit int) ([]sessionChoice, error) {
			return []sessionChoice{
				{Number: 1, Title: "修复登录", Location: "proj-a", Turns: 3},
				{Number: 2, Title: "解释代码", Location: "独立工作区", Turns: 1},
			}, nil
		},
	}
	gw := NewGateway(cfg, map[Platform]Adapter{PlatformQQ: fa}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	gw.handleMessage(context.Background(), PlatformQQ, fa, InboundMessage{
		ChatType:  ChatDM,
		ChatID:    "chat-1",
		UserID:    "user-1",
		Text:      "/start",
		MessageID: "msg-1",
	})

	sent := fa.sentMessages()
	if len(sent) != 1 {
		t.Fatalf("sent count = %d, want 1", len(sent))
	}
	if !strings.Contains(sent[0].Text, "1. 修复登录") || !strings.Contains(sent[0].Text, "2. 解释代码") {
		t.Fatalf("/start list text missing sessions: %q", sent[0].Text)
	}
}

func TestGatewayStartReportsEmptySessions(t *testing.T) {
	fa := newFakeAdapter(PlatformWeixin, "fake-weixin")
	cfg := GatewayConfig{
		Enabled:       map[Platform]bool{PlatformWeixin: true},
		Allowlist:     AllowlistConfig{AllowAll: true},
		SessionLister: func(limit int) ([]sessionChoice, error) { return nil, nil },
	}
	gw := NewGateway(cfg, map[Platform]Adapter{PlatformWeixin: fa}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	gw.handleMessage(context.Background(), PlatformWeixin, fa, InboundMessage{
		ChatType:  ChatDM,
		ChatID:    "chat-empty",
		UserID:    "user-1",
		Text:      "/start",
		MessageID: "msg-1",
	})

	sent := fa.sentMessages()
	if len(sent) != 1 || !strings.Contains(sent[0].Text, "暂无可恢复的对话") {
		t.Fatalf("empty list response = %#v", sent)
	}
}

func TestGatewaySelectsListedSessionAndShowsAssistantTail(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.jsonl")
	sess := agent.NewSession("sys")
	sess.Add(provider.Message{Role: provider.RoleUser, Content: "第一条问题"})
	sess.Add(provider.Message{Role: provider.RoleAssistant, Content: "这是 AI 的一段比较长的回复，用来确认机器人进入历史会话之后，会把最后一段回复的结尾发送给用户。"})
	if err := sess.Save(path); err != nil {
		t.Fatalf("save session: %v", err)
	}

	fa := newFakeAdapter(PlatformFeishu, "fake-feishu")
	buildCalls := 0
	cfg := GatewayConfig{
		Enabled:   map[Platform]bool{PlatformFeishu: true},
		Allowlist: AllowlistConfig{AllowAll: true},
		SessionLister: func(limit int) ([]sessionChoice, error) {
			return []sessionChoice{{
				Number:        1,
				Path:          path,
				Title:         "历史调试",
				Location:      "proj-a",
				WorkspaceRoot: dir,
				Turns:         1,
				LastAssistant: "…最后一段回复的结尾发送给用户。",
			}}, nil
		},
		BuildSession: func(ctx context.Context, choice sessionChoice, sink event.Sink) (*control.Controller, error) {
			buildCalls++
			return control.New(control.Options{SessionDir: dir, SessionPath: path, Sink: sink, Label: "test"}), nil
		},
	}
	gw := NewGateway(cfg, map[Platform]Adapter{PlatformFeishu: fa}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	msg := InboundMessage{ChatType: ChatDM, ChatID: "chat-1", UserID: "user-1", MessageID: "msg-1"}
	msg.Text = "/start"
	gw.handleMessage(context.Background(), PlatformFeishu, fa, msg)
	msg.Text = "1"
	msg.MessageID = "msg-2"
	gw.handleMessage(context.Background(), PlatformFeishu, fa, msg)

	sent := fa.sentMessages()
	if len(sent) != 2 {
		t.Fatalf("sent count = %d, want 2: %#v", len(sent), sent)
	}
	if !strings.Contains(sent[1].Text, "已进入：历史调试") || !strings.Contains(sent[1].Text, "最后一段回复的结尾") {
		t.Fatalf("selection response = %q", sent[1].Text)
	}
	if buildCalls != 1 {
		t.Fatalf("build calls = %d, want 1", buildCalls)
	}
	key := controllerKeyForSession(path)
	if _, ok := gw.controllers[key]; !ok {
		t.Fatalf("selected controller %q not registered", key)
	}
}
