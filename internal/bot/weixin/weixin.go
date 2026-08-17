// Package weixin implements the WeChat iLink/OpenClaw Bot adapter.
package weixin

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/bot"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/config"
)

const (
	defaultWeixinAPI = "https://ilinkai.weixin.qq.com"
	getUpdatesPath   = "/ilink/bot/getupdates"
	sendMessagePath  = "/ilink/bot/sendmessage"
	sendTypingPath   = "/ilink/bot/sendtyping"
	getConfigPath    = "/ilink/bot/getconfig"
	uploadMediaPath  = "/ilink/bot/getuploadurl"
	getBotQRPath     = "/ilink/bot/get_bot_qrcode"
	getQRStatusPath  = "/ilink/bot/get_qrcode_status"

	ilinkAppID          = "bot"
	ilinkClientVersion  = (2 << 16) | (4 << 8) | 3
	ilinkChannelVersion = "2.4.3"
	weixinItemText      = 1
	weixinMsgTypeBot    = 2
	weixinMsgStateDone  = 2
)

type ilinkUpdate struct {
	UpdateID   int64  `json:"update_id"`
	UpdateType string `json:"update_type"`
	Message    struct {
		MessageID flexString `json:"message_id"`
		ChatID    flexString `json:"chat_id"`
		ChatType  string     `json:"chat_type"`
		From      struct {
			UserID   flexString `json:"user_id"`
			UserName string     `json:"user_name"`
		} `json:"from"`
		Text      string `json:"text"`
		Timestamp int64  `json:"timestamp"`
	} `json:"message"`
}

type ilinkMessage struct {
	MessageID    flexString `json:"message_id"`
	FromUserID   flexString `json:"from_user_id"`
	ToUserID     flexString `json:"to_user_id"`
	RoomID       flexString `json:"room_id"`
	ChatRoomID   flexString `json:"chat_room_id"`
	ContextToken flexString `json:"context_token"`
	MsgType      int        `json:"msg_type"`
	CreateTimeMS int64      `json:"create_time_ms"`
	ItemList     []struct {
		Type     int `json:"type"`
		TextItem struct {
			Text string `json:"text"`
		} `json:"text_item"`
	} `json:"item_list"`
}

type ilinkResponse struct {
	Ret                  int            `json:"ret"`
	Errcode              int            `json:"errcode"`
	Errmsg               string         `json:"errmsg"`
	Updates              []ilinkUpdate  `json:"updates"`
	Msgs                 []ilinkMessage `json:"msgs"`
	HasMore              bool           `json:"has_more"`
	ContextToken         string         `json:"context_token"`
	GetUpdatesBuf        string         `json:"get_updates_buf"`
	LongpollingTimeoutMs int            `json:"longpolling_timeout_ms"`
}

type flexString string

func (s *flexString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = ""
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = flexString(str)
		return nil
	}
	var num json.Number
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&num); err == nil {
		*s = flexString(num.String())
		return nil
	}
	return fmt.Errorf("expected string or number, got %s", string(data))
}

func (s flexString) String() string { return string(s) }

type adapter struct {
	cfg    config.WeixinBotConfig
	logger *slog.Logger
	msgCh  chan bot.InboundMessage
	cancel context.CancelFunc

	mu            sync.Mutex
	contextTokens map[string]string
	syncBuf       string
	lastUpdateID  int64
	pollReadyOnce sync.Once
	lastPollLog   time.Time
}

// New creates a WeChat Bot adapter.
func New(cfg config.WeixinBotConfig, logger *slog.Logger) bot.Adapter {
	return &adapter{
		cfg:           cfg,
		logger:        logger.With("platform", "weixin"),
		contextTokens: make(map[string]string),
	}
}

func (a *adapter) Platform() bot.Platform { return bot.PlatformWeixin }
func (a *adapter) Name() string           { return "weixin" }

func (a *adapter) Start(ctx context.Context) error {
	a.msgCh = make(chan bot.InboundMessage, 64)
	ctx, a.cancel = context.WithCancel(ctx)
	a.loadContextTokens()
	if a.token() == "" {
		return a.tokenMissingError()
	}

	a.logger.Info("weixin polling started", "account", logHash(a.accountID()), "api_base", a.apiBase())
	go a.pollLoop(ctx)
	return nil
}

func (a *adapter) Stop() error {
	if a.cancel != nil {
		a.cancel()
	}
	return nil
}

func (a *adapter) Send(ctx context.Context, msg bot.OutboundMessage) (bot.SendResult, error) {
	return a.sendMessage(ctx, msg)
}

func (a *adapter) SendTyping(ctx context.Context, chatID string) error {
	return a.sendTyping(ctx, chatID)
}

func (a *adapter) Messages() <-chan bot.InboundMessage {
	return a.msgCh
}

// SendText sends one plain text message to a saved Weixin iLink conversation.
func SendText(ctx context.Context, cfg config.WeixinBotConfig, chatID, text string) (bot.SendResult, error) {
	a := &adapter{cfg: cfg, logger: slog.Default().With("platform", "weixin"), contextTokens: make(map[string]string)}
	return a.sendMessage(ctx, bot.OutboundMessage{ChatID: chatID, Text: text})
}

// Validate verifies that a saved Weixin/OpenClaw account token can poll updates.
func Validate(ctx context.Context, cfg config.WeixinBotConfig) error {
	a := &adapter{cfg: cfg, logger: slog.Default().With("platform", "weixin"), contextTokens: make(map[string]string)}
	_, err := a.getUpdates(ctx)
	return err
}

func (a *adapter) token() string {
	if env := strings.TrimSpace(a.cfg.TokenEnv); env != "" {
		if token := strings.TrimSpace(os.Getenv(env)); token != "" {
			return token
		}
	}
	account, _ := loadSavedAccount(a.accountID())
	if strings.TrimSpace(account.Token) != "" {
		return strings.TrimSpace(account.Token)
	}
	if a.cfg.AccountID == "" {
		account, _ = loadAnySavedAccount()
		return strings.TrimSpace(account.Token)
	}
	return ""
}

func (a *adapter) tokenMissingError() error {
	if strings.TrimSpace(a.cfg.TokenEnv) == "" {
		return fmt.Errorf("weixin token is not configured and no saved weixin account is available")
	}
	return fmt.Errorf("%s is not set and no saved weixin account is available", a.cfg.TokenEnv)
}

func (a *adapter) apiBase() string {
	if strings.TrimSpace(a.cfg.APIBase) != "" {
		return strings.TrimRight(strings.TrimSpace(a.cfg.APIBase), "/")
	}
	account, _ := loadSavedAccount(a.accountID())
	if strings.TrimSpace(account.BaseURL) != "" {
		return strings.TrimRight(strings.TrimSpace(account.BaseURL), "/")
	}
	return defaultWeixinAPI
}

func (a *adapter) accountID() string {
	if strings.TrimSpace(a.cfg.AccountID) != "" {
		return strings.TrimSpace(a.cfg.AccountID)
	}
	return "default"
}

func (a *adapter) contextToken(chatID string) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.contextTokens[chatID]
}

func (a *adapter) setContextToken(chatID, token string) {
	a.mu.Lock()
	if token == "" {
		delete(a.contextTokens, chatID)
	} else {
		a.contextTokens[chatID] = token
	}
	a.mu.Unlock()
	a.saveContextTokens()
}

func (a *adapter) tokenStorePath() string {
	root := config.MemoryUserDir()
	if root == "" {
		return ""
	}
	return filepath.Join(weixinAccountDir(root), a.accountID()+".context-tokens.json")
}

func (a *adapter) loadContextTokens() {
	path := a.tokenStorePath()
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var tokens map[string]string
	if err := json.Unmarshal(data, &tokens); err != nil {
		a.logger.Warn("failed to load weixin context tokens", "err", err)
		return
	}
	a.mu.Lock()
	a.contextTokens = tokens
	a.mu.Unlock()
}

func (a *adapter) saveContextTokens() {
	path := a.tokenStorePath()
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		a.logger.Warn("failed to create weixin token dir", "err", err)
		return
	}
	a.mu.Lock()
	data, err := json.MarshalIndent(a.contextTokens, "", "  ")
	a.mu.Unlock()
	if err != nil {
		return
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		a.logger.Warn("failed to save weixin context tokens", "err", err)
	}
}

func ilinkGET(ctx context.Context, baseURL, endpoint string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", strings.TrimRight(baseURL, "/")+"/"+strings.TrimLeft(endpoint, "/"), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("iLink-App-Id", ilinkAppID)
	req.Header.Set("iLink-App-ClientVersion", fmt.Sprintf("%d", ilinkClientVersion))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		if len(data) > 300 {
			data = data[:300]
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (a *adapter) pollLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		updates, err := a.getUpdates(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			a.logger.Error("weixin getupdates failed", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, upd := range updates {
			a.handleUpdate(upd)
		}

		if len(updates) == 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (a *adapter) getUpdates(ctx context.Context) ([]ilinkUpdate, error) {
	tok := a.token()
	if tok == "" {
		return nil, a.tokenMissingError()
	}

	a.mu.Lock()
	payload := map[string]any{
		"get_updates_buf": a.syncBuf,
		"base_info":       weixinBaseInfo(),
	}
	a.mu.Unlock()

	var result ilinkResponse
	if err := a.postJSON(ctx, getUpdatesPath, tok, payload, &result); err != nil {
		return nil, err
	}
	if result.Ret != 0 || result.Errcode != 0 {
		return nil, fmt.Errorf("getupdates error ret=%d errcode=%d: %s", result.Ret, result.Errcode, result.Errmsg)
	}
	a.pollReadyOnce.Do(func() {
		a.logger.Info("weixin getupdates ready", "account", logHash(a.accountID()), "api_base", a.apiBase())
	})
	a.logPollHealth(result)

	a.mu.Lock()
	if result.GetUpdatesBuf != "" {
		a.syncBuf = result.GetUpdatesBuf
	}
	if len(result.Updates) > 0 {
		last := result.Updates[len(result.Updates)-1]
		a.lastUpdateID = last.UpdateID
	}
	a.mu.Unlock()

	for _, msg := range result.Msgs {
		a.handleIlinkMessage(msg)
	}
	return result.Updates, nil
}

func (a *adapter) postJSON(ctx context.Context, endpoint, token string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", a.apiBase()+endpoint, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	setIlinkHeaders(req, token, body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		if len(respBody) > 300 {
			respBody = respBody[:300]
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return err
	}
	return nil
}

func weixinBaseInfo() map[string]string {
	return map[string]string{
		"channel_version": ilinkChannelVersion,
		"bot_agent":       "O.R.C.A/3.0 (OpenClaw-compatible)",
	}
}

func (a *adapter) logPollHealth(result ilinkResponse) {
	shouldLog := len(result.Updates) > 0 || len(result.Msgs) > 0
	a.mu.Lock()
	if !shouldLog && time.Since(a.lastPollLog) >= 5*time.Minute {
		shouldLog = true
	}
	if shouldLog {
		a.lastPollLog = time.Now()
	}
	a.mu.Unlock()
	if !shouldLog {
		return
	}
	a.logger.Info("weixin getupdates heartbeat",
		"updates", len(result.Updates),
		"msgs", len(result.Msgs),
		"has_more", result.HasMore,
		"timeout_ms", result.LongpollingTimeoutMs)
}

func (a *adapter) handleUpdate(upd ilinkUpdate) {
	if upd.UpdateType != "message" {
		a.logger.Info("weixin update ignored", "reason", "non_message", "update_type", upd.UpdateType)
		return
	}

	m := upd.Message
	chatType := bot.ChatDM
	if m.ChatType == "group" {
		chatType = bot.ChatGroup
	}

	ib := bot.InboundMessage{
		Platform:  bot.PlatformWeixin,
		ChatType:  chatType,
		ChatID:    m.ChatID.String(),
		UserID:    m.From.UserID.String(),
		UserName:  m.From.UserName,
		Text:      m.Text,
		MessageID: m.MessageID.String(),
	}

	select {
	case a.msgCh <- ib:
		a.logger.Info("weixin inbound queued", "source", "update", "chat_type", chatType, "chat", logHash(ib.ChatID), "user", logHash(ib.UserID), "message", logHash(ib.MessageID), "text_chars", len([]rune(ib.Text)))
	default:
		a.logger.Warn("weixin message channel full")
	}
}

func (a *adapter) handleIlinkMessage(m ilinkMessage) {
	fromUserID := m.FromUserID.String()
	if fromUserID == "" || fromUserID == a.accountID() {
		a.logger.Info("weixin message ignored", "reason", "self_or_missing_sender", "from", logHash(fromUserID), "message", logHash(m.MessageID.String()))
		return
	}
	text := extractIlinkText(m.ItemList)
	if text == "" {
		a.logger.Info("weixin message ignored", "reason", "empty_text", "from", logHash(fromUserID), "message", logHash(m.MessageID.String()))
		return
	}
	chatType, chatID := guessIlinkChat(m, a.accountID())
	if chatID == "" {
		a.logger.Info("weixin message ignored", "reason", "missing_chat", "from", logHash(fromUserID), "message", logHash(m.MessageID.String()))
		return
	}
	if token := m.ContextToken.String(); token != "" {
		a.setContextToken(chatID, token)
	}
	ib := bot.InboundMessage{
		Platform:  bot.PlatformWeixin,
		ChatType:  chatType,
		ChatID:    chatID,
		UserID:    fromUserID,
		UserName:  fromUserID,
		Text:      text,
		MessageID: m.MessageID.String(),
	}
	select {
	case a.msgCh <- ib:
		a.logger.Info("weixin inbound queued", "source", "message", "chat_type", chatType, "chat", logHash(ib.ChatID), "user", logHash(ib.UserID), "message", logHash(ib.MessageID), "text_chars", len([]rune(ib.Text)))
	default:
		a.logger.Warn("weixin message channel full")
	}
}

func logHash(id string) string {
	if id == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(id))
	return hex.EncodeToString(sum[:])[:12]
}

func extractIlinkText(items []struct {
	Type     int `json:"type"`
	TextItem struct {
		Text string `json:"text"`
	} `json:"text_item"`
}) string {
	var out []string
	for _, item := range items {
		if item.Type == weixinItemText && item.TextItem.Text != "" {
			out = append(out, item.TextItem.Text)
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func guessIlinkChat(m ilinkMessage, accountID string) (bot.ChatType, string) {
	roomID := firstNonEmptyString(m.RoomID.String(), m.ChatRoomID.String())
	if roomID != "" {
		return bot.ChatGroup, roomID
	}
	toUserID := m.ToUserID.String()
	if toUserID != "" && accountID != "" && toUserID != accountID && m.MsgType == 1 {
		return bot.ChatGroup, toUserID
	}
	return bot.ChatDM, m.FromUserID.String()
}

func setIlinkHeaders(req *http.Request, token string, body []byte) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AuthorizationType", "ilink_bot_token")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	req.Header.Set("X-WECHAT-UIN", randomWechatUIN())
	req.Header.Set("iLink-App-Id", ilinkAppID)
	req.Header.Set("iLink-App-ClientVersion", fmt.Sprintf("%d", ilinkClientVersion))
}

func randomWechatUIN() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	}
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", uint32(b[0])<<24|uint32(b[1])<<16|uint32(b[2])<<8|uint32(b[3]))))
}

func firstNonEmptyString(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func (a *adapter) sendMessage(ctx context.Context, msg bot.OutboundMessage) (bot.SendResult, error) {
	tok := a.token()
	if tok == "" {
		return bot.SendResult{}, a.tokenMissingError()
	}

	clientID := fmt.Sprintf("orca-%d", time.Now().UnixNano())
	payload := map[string]any{
		"base_info": weixinBaseInfo(),
		"msg": map[string]any{
			"from_user_id":  "",
			"to_user_id":    msg.ChatID,
			"client_id":     clientID,
			"message_type":  weixinMsgTypeBot,
			"message_state": weixinMsgStateDone,
			"item_list": []map[string]any{
				{"type": weixinItemText, "text_item": map[string]string{"text": msg.Text}},
			},
		},
	}
	if contextToken := a.contextToken(msg.ChatID); contextToken != "" {
		if m, ok := payload["msg"].(map[string]any); ok {
			m["context_token"] = contextToken
		}
	}

	var result struct {
		Ret       int        `json:"ret"`
		Errcode   int        `json:"errcode"`
		Errmsg    string     `json:"errmsg"`
		MessageID flexString `json:"message_id"`
	}
	if err := a.postJSON(ctx, sendMessagePath, tok, payload, &result); err != nil {
		return bot.SendResult{}, err
	}
	if result.Ret != 0 || result.Errcode != 0 {
		if a.contextToken(msg.ChatID) != "" {
			a.setContextToken(msg.ChatID, "")
			return a.sendMessage(ctx, msg)
		}
		return bot.SendResult{}, fmt.Errorf("sendmessage error ret=%d errcode=%d: %s", result.Ret, result.Errcode, result.Errmsg)
	}
	messageID := result.MessageID.String()
	if messageID == "" {
		messageID = clientID
	}
	return bot.SendResult{MessageID: messageID}, nil
}

func (a *adapter) sendTyping(ctx context.Context, chatID string) error {
	tok := a.token()
	if tok == "" {
		return a.tokenMissingError()
	}
	ticket, err := a.getTypingTicket(ctx, chatID)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"base_info":     weixinBaseInfo(),
		"ilink_user_id": chatID,
		"typing_ticket": ticket,
		"status":        1,
	}
	if contextToken := a.contextToken(chatID); contextToken != "" {
		payload["context_token"] = contextToken
	}
	var result struct {
		Ret     int    `json:"ret"`
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	if err := a.postJSON(ctx, sendTypingPath, tok, payload, &result); err != nil {
		return err
	}
	if result.Ret != 0 || result.Errcode != 0 {
		return fmt.Errorf("sendtyping error ret=%d errcode=%d: %s", result.Ret, result.Errcode, result.Errmsg)
	}
	return nil
}

func (a *adapter) getTypingTicket(ctx context.Context, chatID string) (string, error) {
	tok := a.token()
	if tok == "" {
		return "", a.tokenMissingError()
	}
	payload := map[string]any{
		"base_info":     weixinBaseInfo(),
		"ilink_user_id": chatID,
	}
	if contextToken := a.contextToken(chatID); contextToken != "" {
		payload["context_token"] = contextToken
	}
	var result struct {
		Ret          int    `json:"ret"`
		Errcode      int    `json:"errcode"`
		Errmsg       string `json:"errmsg"`
		TypingTicket string `json:"typing_ticket"`
	}
	if err := a.postJSON(ctx, getConfigPath, tok, payload, &result); err != nil {
		return "", err
	}
	if result.Ret != 0 || result.Errcode != 0 {
		return "", fmt.Errorf("getconfig error ret=%d errcode=%d: %s", result.Ret, result.Errcode, result.Errmsg)
	}
	if result.TypingTicket == "" {
		return "", fmt.Errorf("getconfig response missing typing_ticket")
	}
	return result.TypingTicket, nil
}
