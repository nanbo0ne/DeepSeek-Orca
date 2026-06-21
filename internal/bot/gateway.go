package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"deepseek-orca/internal/agent"
	"deepseek-orca/internal/boot"
	"deepseek-orca/internal/control"
	"deepseek-orca/internal/event"
)

// GatewayConfig configures the multi-channel IM bot gateway.
type GatewayConfig struct {
	Model         string
	PromptMode    string
	MaxSteps      int
	WorkspaceRoot string
	Allowlist     AllowlistConfig
	Enabled       map[Platform]bool
	Debounce      time.Duration
	SessionLister botSessionLister
	CreateSession SessionCreator
	BuildSession  botControllerFactory
	MirrorEvent   EventMirror
	MirrorUser    UserMirror
	AfterTurn     AfterTurnHook
}

// AllowlistConfig controls which remote users/groups can invoke the bot.
type AllowlistConfig struct {
	Enabled  bool
	AllowAll bool
	Users    map[Platform][]string
	Groups   map[Platform][]string
}

// BotGateway manages adapters, controllers, session selection, and event
// rendering for IM platforms.
type BotGateway struct {
	cfg      GatewayConfig
	adapters map[Platform]Adapter
	sessions *SessionManager

	mu             sync.Mutex
	controllers    map[string]*sessionState
	remoteStates   map[string]*remoteSessionState
	allowlist      map[Platform]map[string]bool
	groupAllowlist map[Platform]map[string]bool

	logger *slog.Logger
}

type sessionState struct {
	ctrl          *control.Controller
	sink          *sessionEventSink
	cancel        context.CancelFunc
	pendingAsks   map[string][]event.AskQuestion
	createdAt     time.Time
	lastActive    time.Time
	sessionPath   string
	workspaceRoot string
	title         string
}

type botControllerFactory func(ctx context.Context, choice sessionChoice, sink event.Sink) (*control.Controller, error)
type SessionCreator func(ctx context.Context, remoteKey string, msg InboundMessage) (SessionChoice, error)
type EventMirror func(sessionPath string, e event.Event)
type UserMirror func(sessionPath, text string)
type AfterTurnHook func(sessionPath string)

type remoteMode string

const (
	remoteModeListing   remoteMode = "listing"
	remoteModeInSession remoteMode = "in_session"
)

type remoteSessionState struct {
	mode          remoteMode
	selectedKey   string
	selectedPath  string
	workspaceRoot string
	title         string
	lastListed    []sessionChoice
}

type sessionEventSink struct {
	mu     sync.RWMutex
	target event.Sink
}

func (s *sessionEventSink) setTarget(target event.Sink) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.target = target
}

func (s *sessionEventSink) Emit(e event.Event) {
	s.mu.RLock()
	target := s.target
	s.mu.RUnlock()
	if target != nil {
		target.Emit(e)
	}
}

func NewGateway(cfg GatewayConfig, adapters map[Platform]Adapter, logger *slog.Logger) *BotGateway {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.Debounce <= 0 {
		cfg.Debounce = 1500 * time.Millisecond
	}
	if cfg.SessionLister == nil {
		cfg.SessionLister = defaultBotSessionLister
	}
	if cfg.CreateSession == nil {
		cfg.CreateSession = defaultBotSessionCreator
	}
	if cfg.BuildSession == nil {
		cfg.BuildSession = func(ctx context.Context, choice sessionChoice, sink event.Sink) (*control.Controller, error) {
			return boot.Build(ctx, boot.Options{
				Model:         cfg.Model,
				PromptMode:    cfg.PromptMode,
				MaxSteps:      cfg.MaxSteps,
				RequireKey:    true,
				Sink:          sink,
				WorkspaceRoot: strings.TrimSpace(choice.WorkspaceRoot),
				SessionDir:    filepath.Dir(choice.Path),
			})
		}
	}
	gw := &BotGateway{
		cfg:            cfg,
		adapters:       adapters,
		sessions:       NewSessionManager(cfg.Debounce),
		controllers:    make(map[string]*sessionState),
		remoteStates:   make(map[string]*remoteSessionState),
		allowlist:      make(map[Platform]map[string]bool),
		groupAllowlist: make(map[Platform]map[string]bool),
		logger:         logger.With("component", "bot_gateway"),
	}
	gw.buildAllowlist()
	return gw
}

func (gw *BotGateway) buildAllowlist() {
	for _, plat := range []Platform{PlatformQQ, PlatformFeishu, PlatformWeixin} {
		gw.allowlist[plat] = make(map[string]bool)
		if !gw.cfg.Allowlist.Enabled {
			continue
		}
		for _, uid := range gw.cfg.Allowlist.Users[plat] {
			gw.allowlist[plat][uid] = true
		}
		gw.groupAllowlist[plat] = make(map[string]bool)
		for _, gid := range gw.cfg.Allowlist.Groups[plat] {
			gw.groupAllowlist[plat][gid] = true
		}
	}
}

func (gw *BotGateway) Start(ctx context.Context) error {
	for plat, adapter := range gw.adapters {
		if !gw.cfg.Enabled[plat] {
			gw.logger.Info("platform disabled, skipping", "platform", plat)
			continue
		}
		gw.logger.Info("starting adapter", "platform", plat)
		if err := adapter.Start(ctx); err != nil {
			return fmt.Errorf("start adapter %s: %w", plat, err)
		}
	}

	for plat, adapter := range gw.adapters {
		if !gw.cfg.Enabled[plat] {
			continue
		}
		go gw.dispatchLoop(ctx, plat, adapter)
	}
	return nil
}

func (gw *BotGateway) Stop() {
	gw.mu.Lock()
	for key, state := range gw.controllers {
		if state.cancel != nil {
			state.cancel()
		}
		state.ctrl.Close()
		delete(gw.controllers, key)
	}
	gw.mu.Unlock()

	for _, adapter := range gw.adapters {
		if err := adapter.Stop(); err != nil {
			gw.logger.Warn("error stopping adapter", "err", err)
		}
	}
}

func (gw *BotGateway) dispatchLoop(ctx context.Context, plat Platform, adapter Adapter) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-adapter.Messages():
			if !ok {
				return
			}
			gw.handleMessage(ctx, plat, adapter, msg)
		}
	}
}

func (gw *BotGateway) handleMessage(ctx context.Context, plat Platform, adapter Adapter, msg InboundMessage) {
	msg.Platform = plat
	if !gw.checkAllowlist(plat, msg) {
		gw.logger.Info("user not in allowlist", "platform", plat, "user", hashID(msg.UserID))
		_ = gw.sendText(ctx, adapter, msg, "抱歉，您没有使用此 bot 的权限。")
		return
	}

	remoteKey := BuildSessionKey(msg.Session())
	text := strings.TrimSpace(msg.Text)
	if strings.HasPrefix(text, "/start") {
		gw.showSessionList(ctx, adapter, remoteKey, msg)
		return
	}
	if gw.trySelectListedSession(ctx, adapter, remoteKey, msg) {
		return
	}
	if IsSlashBypass(text) {
		gw.handleSlashCommand(ctx, adapter, remoteKey, msg)
		return
	}

	sessionKey, ok := gw.activeControllerKey(remoteKey)
	if !ok {
		_ = gw.sendText(ctx, adapter, msg, "请先发送 /start 选择要进入的对话。")
		return
	}

	acquired, merged := gw.sessions.TryAcquire(sessionKey, msg)
	if merged {
		gw.logger.Debug("message merged to pending queue", "session", sessionKey)
		return
	}
	if !acquired {
		gw.logger.Debug("session busy, queued", "session", sessionKey)
		return
	}
	gw.runTurn(ctx, adapter, remoteKey, sessionKey, msg)
}

func (gw *BotGateway) checkAllowlist(plat Platform, msg InboundMessage) bool {
	if gw.cfg.Allowlist.AllowAll {
		return true
	}
	if !gw.cfg.Allowlist.Enabled {
		return false
	}
	if !gw.allowlist[plat][msg.UserID] {
		return false
	}
	groups := gw.groupAllowlist[plat]
	if chatUsesGroupAllowlist(msg.ChatType) && len(groups) > 0 && !groups[msg.ChatID] {
		return false
	}
	return true
}

func chatUsesGroupAllowlist(chatType ChatType) bool {
	switch chatType {
	case ChatGroup, ChatGuild, ChatThread:
		return true
	default:
		return false
	}
}

func (gw *BotGateway) showSessionList(ctx context.Context, adapter Adapter, remoteKey string, msg InboundMessage) {
	choices, err := gw.cfg.SessionLister(maxBotSessionChoices)
	if err != nil {
		gw.logger.Warn("list sessions failed", "err", err)
		_ = gw.sendText(ctx, adapter, msg, "读取会话列表失败："+err.Error())
		return
	}
	gw.mu.Lock()
	gw.remoteStates[remoteKey] = &remoteSessionState{mode: remoteModeListing, lastListed: choices}
	gw.mu.Unlock()
	_ = gw.sendText(ctx, adapter, msg, formatSessionChoices(choices))
}

func (gw *BotGateway) trySelectListedSession(ctx context.Context, adapter Adapter, remoteKey string, msg InboundMessage) bool {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return false
	}
	n, err := strconv.Atoi(text)
	if err != nil {
		return false
	}

	gw.mu.Lock()
	remote := gw.remoteStates[remoteKey]
	if remote == nil || remote.mode != remoteModeListing {
		gw.mu.Unlock()
		return false
	}
	choices := append([]sessionChoice(nil), remote.lastListed...)
	gw.mu.Unlock()

	if n < 1 || n > len(choices) {
		_ = gw.sendText(ctx, adapter, msg, fmt.Sprintf("请输入 1 到 %d 之间的数字，或发送 /start 刷新列表。", len(choices)))
		return true
	}
	choice := choices[n-1]
	if err := gw.selectSession(ctx, remoteKey, choice); err != nil {
		gw.logger.Warn("select session failed", "err", err, "path", choice.Path)
		_ = gw.sendText(ctx, adapter, msg, "无法打开该会话："+err.Error()+"\n请发送 /start 重新选择。")
		gw.showSessionList(ctx, adapter, remoteKey, msg)
		return true
	}
	_ = gw.sendText(ctx, adapter, msg, formatSessionEntered(choice))
	return true
}

func (gw *BotGateway) selectSession(ctx context.Context, remoteKey string, choice sessionChoice) error {
	if strings.TrimSpace(choice.Path) == "" {
		return fmt.Errorf("会话路径为空")
	}
	if _, err := os.Stat(choice.Path); err != nil {
		return err
	}
	loaded, err := agent.LoadSession(choice.Path)
	if err != nil {
		return err
	}
	workspaceRoot := strings.TrimSpace(choice.WorkspaceRoot)
	sessionKey := controllerKeyForSession(choice.Path)

	gw.mu.Lock()
	if old := gw.remoteStates[remoteKey]; old != nil && old.selectedKey != "" && old.selectedKey != sessionKey {
		gw.sessions.ForceRelease(old.selectedKey)
	}
	state := gw.controllers[sessionKey]
	gw.mu.Unlock()

	if state == nil {
		sessionSink := &sessionEventSink{}
		ctrl, err := gw.cfg.BuildSession(ctx, choice, sessionSink)
		if err != nil {
			return err
		}
		ctrl.EnableInteractiveApproval()
		ctrl.Resume(loaded, choice.Path)
		state = &sessionState{
			ctrl:          ctrl,
			sink:          sessionSink,
			pendingAsks:   make(map[string][]event.AskQuestion),
			createdAt:     time.Now(),
			lastActive:    time.Now(),
			sessionPath:   choice.Path,
			workspaceRoot: workspaceRoot,
			title:         choice.Title,
		}
	} else {
		if state.ctrl != nil {
			state.ctrl.Resume(loaded, choice.Path)
		}
		state.sessionPath = choice.Path
		state.workspaceRoot = workspaceRoot
		state.title = choice.Title
		state.lastActive = time.Now()
	}

	gw.mu.Lock()
	gw.controllers[sessionKey] = state
	gw.remoteStates[remoteKey] = &remoteSessionState{
		mode:          remoteModeInSession,
		selectedKey:   sessionKey,
		selectedPath:  choice.Path,
		workspaceRoot: workspaceRoot,
		title:         choice.Title,
		lastListed:    nil,
	}
	gw.mu.Unlock()
	return nil
}

func controllerKeyForSession(path string) string {
	return "session:" + path
}

func (gw *BotGateway) activeControllerKey(remoteKey string) (string, bool) {
	gw.mu.Lock()
	defer gw.mu.Unlock()
	state := gw.remoteStates[remoteKey]
	if state == nil || state.mode != remoteModeInSession || state.selectedKey == "" {
		return "", false
	}
	return state.selectedKey, true
}

func (gw *BotGateway) handleSlashCommand(ctx context.Context, adapter Adapter, remoteKey string, msg InboundMessage) {
	key, hasSession := gw.activeControllerKey(remoteKey)
	switch {
	case strings.HasPrefix(msg.Text, "/stop"):
		if hasSession {
			gw.mu.Lock()
			state, ok := gw.controllers[key]
			gw.mu.Unlock()
			if ok && state.cancel != nil {
				state.cancel()
			}
			gw.sessions.ForceRelease(key)
		}
		_ = gw.sendText(ctx, adapter, msg, "已停止当前任务。")

	case strings.HasPrefix(msg.Text, "/new") || strings.HasPrefix(msg.Text, "/reset"):
		choice, err := gw.cfg.CreateSession(ctx, remoteKey, msg)
		if err != nil {
			gw.logger.Warn("create bot session failed", "err", err)
			_ = gw.sendText(ctx, adapter, msg, "创建新对话失败："+err.Error())
			return
		}
		if err := gw.selectSession(ctx, remoteKey, choice); err != nil {
			gw.logger.Warn("select new bot session failed", "err", err, "path", choice.Path)
			_ = gw.sendText(ctx, adapter, msg, "新对话已创建，但无法进入："+err.Error())
			return
		}
		_ = gw.sendText(ctx, adapter, msg, formatSessionCreated(choice))

	case strings.HasPrefix(msg.Text, "/approve"):
		parts := strings.Fields(msg.Text)
		if len(parts) < 2 {
			_ = gw.sendText(ctx, adapter, msg, "用法：/approve <id>")
			return
		}
		state, ok := gw.controllerState(key, hasSession)
		if ok {
			state.ctrl.Approve(parts[1], true, false, false)
			_ = gw.sendText(ctx, adapter, msg, "已批准。")
			return
		}
		_ = gw.sendText(ctx, adapter, msg, "没有找到当前会话。请先发送 /start 选择对话。")

	case strings.HasPrefix(msg.Text, "/deny"):
		parts := strings.Fields(msg.Text)
		if len(parts) < 2 {
			_ = gw.sendText(ctx, adapter, msg, "用法：/deny <id>")
			return
		}
		state, ok := gw.controllerState(key, hasSession)
		if ok {
			state.ctrl.Approve(parts[1], false, false, false)
			_ = gw.sendText(ctx, adapter, msg, "已拒绝。")
			return
		}
		_ = gw.sendText(ctx, adapter, msg, "没有找到当前会话。请先发送 /start 选择对话。")

	case strings.HasPrefix(msg.Text, "/answer"):
		parts := strings.Fields(msg.Text)
		if len(parts) < 3 {
			_ = gw.sendText(ctx, adapter, msg, "用法：/answer <id> <选项或 q1=选项;q2=选项>")
			return
		}
		state, ok := gw.controllerState(key, hasSession)
		if !ok {
			_ = gw.sendText(ctx, adapter, msg, "没有找到当前会话。请先发送 /start 选择对话。")
			return
		}
		askID := parts[1]
		rawAnswer := strings.TrimSpace(strings.Join(parts[2:], " "))
		gw.mu.Lock()
		questions := state.pendingAsks[askID]
		delete(state.pendingAsks, askID)
		gw.mu.Unlock()
		answers := parseAskAnswers(questions, rawAnswer)
		state.ctrl.AnswerQuestion(askID, answers)
		_ = gw.sendText(ctx, adapter, msg, "已提交回答。")

	case strings.HasPrefix(msg.Text, "/status"):
		active := gw.sessions.ActiveCount()
		gw.mu.Lock()
		remote := gw.remoteStates[remoteKey]
		controllers := len(gw.controllers)
		gw.mu.Unlock()
		current := "未选择"
		if remote != nil && remote.title != "" {
			current = remote.title
		}
		_ = gw.sendText(ctx, adapter, msg, fmt.Sprintf("当前对话：%s\n活跃任务数：%d\n已加载会话数：%d", current, active, controllers))

	case strings.HasPrefix(msg.Text, "/help"):
		_ = gw.sendText(ctx, adapter, msg, botHelpText())
	}
}

func (gw *BotGateway) controllerState(key string, hasSession bool) (*sessionState, bool) {
	if !hasSession || key == "" {
		return nil, false
	}
	gw.mu.Lock()
	defer gw.mu.Unlock()
	state, ok := gw.controllers[key]
	return state, ok && state != nil && state.ctrl != nil
}

func botHelpText() string {
	return "可用命令：\n" +
		"/start - 回到会话列表，选择最近 15 条对话\n" +
		"/new - 创建新的独立工作区对话\n" +
		"/stop - 停止当前任务\n" +
		"/approve <id> - 批准操作\n" +
		"/deny <id> - 拒绝操作\n" +
		"/answer <id> <选项> - 回答 ask 问题\n" +
		"/status - 查看当前状态\n" +
		"/help - 显示帮助"
}

func (gw *BotGateway) runTurn(ctx context.Context, adapter Adapter, remoteKey, sessionKey string, msg InboundMessage) {
	defer func() {
		next := gw.sessions.Release(sessionKey)
		if next != nil {
			gw.runTurn(ctx, adapter, remoteKey, sessionKey, *next)
			return
		}
	}()

	input := msg.Text
	if msg.ChatType == ChatGroup {
		input = fmt.Sprintf("[%s] %s", msg.UserName, msg.Text)
	}

	state, ok := gw.controllerState(sessionKey, true)
	if !ok {
		_ = gw.sendText(ctx, adapter, msg, "内部错误：无法找到已选择的会话。请发送 /start 重新选择。")
		return
	}
	if gw.cfg.MirrorUser != nil {
		gw.cfg.MirrorUser(state.sessionPath, input)
	}
	_ = adapter.SendTyping(ctx, msg.ChatID)

	renderSink := newRenderSink(ctx, adapter, msg.ChatID, msg.ChatType, msg.MessageID, gw.logger, func(ask event.Ask) {
		gw.mu.Lock()
		if state.pendingAsks == nil {
			state.pendingAsks = make(map[string][]event.AskQuestion)
		}
		state.pendingAsks[ask.ID] = ask.Questions
		gw.mu.Unlock()
	})
	var target event.Sink = renderSink
	if gw.cfg.MirrorEvent != nil {
		target = multiEventSink{
			renderSink,
			botMirrorSink{sessionPath: state.sessionPath, mirror: gw.cfg.MirrorEvent},
		}
	}
	state.sink.setTarget(target)
	defer state.sink.setTarget(nil)

	turnCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	gw.mu.Lock()
	state.cancel = cancel
	state.lastActive = time.Now()
	gw.mu.Unlock()

	renderSink.ctrl = state.ctrl
	err := state.ctrl.RunTurn(turnCtx, input)
	target.Emit(event.Event{Kind: event.TurnDone, Err: err})
	if gw.cfg.AfterTurn != nil {
		gw.cfg.AfterTurn(state.sessionPath)
	}
	if err != nil {
		gw.logger.Warn("turn error", "remote", remoteKey, "session", sessionKey, "err", err)
	}
}

type multiEventSink []event.Sink

func (s multiEventSink) Emit(e event.Event) {
	for _, sink := range s {
		if sink != nil {
			sink.Emit(e)
		}
	}
}

type botMirrorSink struct {
	sessionPath string
	mirror      EventMirror
}

func (s botMirrorSink) Emit(e event.Event) {
	if s.mirror != nil && strings.TrimSpace(s.sessionPath) != "" {
		s.mirror(s.sessionPath, e)
	}
}

func (gw *BotGateway) sendText(ctx context.Context, adapter Adapter, msg InboundMessage, text string) error {
	_, err := adapter.Send(ctx, OutboundMessage{
		ChatID:       msg.ChatID,
		ChatType:     msg.ChatType,
		Text:         text,
		ReplyToMsgID: msg.MessageID,
	})
	return err
}

func parseAskAnswers(questions []event.AskQuestion, raw string) []event.AskAnswer {
	raw = strings.TrimSpace(raw)
	if len(questions) == 0 {
		return []event.AskAnswer{{Selected: []string{raw}}}
	}
	byID := make(map[string]*event.AskQuestion, len(questions))
	for i := range questions {
		q := &questions[i]
		byID[q.ID] = q
		byID[fmt.Sprintf("%d", i+1)] = q
	}
	answerMap := make(map[string][]string, len(questions))
	if strings.Contains(raw, "=") {
		for _, part := range strings.Split(raw, ";") {
			k, v, ok := strings.Cut(part, "=")
			if !ok {
				continue
			}
			q := byID[strings.TrimSpace(k)]
			if q == nil {
				continue
			}
			answerMap[q.ID] = normalizeAskSelection(*q, strings.TrimSpace(v))
		}
	} else if len(questions) == 1 {
		answerMap[questions[0].ID] = normalizeAskSelection(questions[0], raw)
	}
	out := make([]event.AskAnswer, 0, len(questions))
	for _, q := range questions {
		out = append(out, event.AskAnswer{QuestionID: q.ID, Selected: answerMap[q.ID]})
	}
	return out
}

func normalizeAskSelection(q event.AskQuestion, raw string) []string {
	parts := []string{raw}
	if q.Multi && strings.Contains(raw, ",") {
		parts = strings.Split(raw, ",")
	}
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if idx, err := strconv.Atoi(part); err == nil && idx >= 1 && idx <= len(q.Options) {
			out = append(out, q.Options[idx-1].Label)
			continue
		}
		out = append(out, part)
	}
	return out
}
