package qq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"deepseek-orca/internal/bot"
	"deepseek-orca/internal/config"

	"golang.org/x/net/websocket"
)

const (
	qqTokenURL               = "https://bots.qq.com/app/getAppAccessToken"
	qqProductionRESTBase     = "https://api.sgroup.qq.com"
	qqProductionGatewayURL   = "wss://api.sgroup.qq.com/websocket"
	qqSandboxRESTBase        = "https://sandbox.api.sgroup.qq.com"
	qqSandboxGatewayURL      = "wss://sandbox.api.sgroup.qq.com/websocket"
	qqSendMsgPathTemplate    = "/v2/users/%s/messages"
	qqSendGroupPathTemplate  = "/v2/groups/%s/messages"
	qqSendGuildPathTemplate  = "/v2/channels/%s/messages"
	qqSendDirectPathTemplate = "/v2/dms/%s/messages"

	opDispatch     = 0
	opHeartbeat    = 1
	opIdentify     = 2
	opResume       = 6
	opReconnect    = 7
	opHello        = 10
	opHeartbeatAck = 11
)

var qqTokenEndpoint = qqTokenURL

// gatewayPayload QQ WebSocket 消息载荷。
type gatewayPayload struct {
	Op int             `json:"op"`
	D  json.RawMessage `json:"d,omitempty"`
	S  int64           `json:"s,omitempty"`
	T  string          `json:"t,omitempty"`
}

type helloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type identifyData struct {
	Token      string     `json:"token"`
	Intents    int        `json:"intents"`
	Shard      [2]int     `json:"shard"`
	Properties properties `json:"properties"`
}

type properties struct {
	OS      string `json:"$os"`
	Browser string `json:"$browser"`
	Device  string `json:"$device"`
}

type dispatchEvent struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	Author    struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"author"`
	ChannelID   string `json:"channel_id"`
	GuildID     string `json:"guild_id"`
	GroupOpenID string `json:"group_openid"`
}

// wsClient 管理 QQ WebSocket 连接。
type wsClient struct {
	mu          sync.Mutex
	conn        *websocket.Conn
	heartbeatMs int
	sessionID   string
	lastSeq     int64
	token       string
	logger      *slog.Logger
}

func (a *adapter) gatewayLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		token, err := a.getAccessToken(ctx)
		if err != nil {
			a.logger.Error("get access token failed", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if err := a.connectGateway(ctx, token); err != nil {
			a.logger.Error("gateway connection failed", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}
	}
}

func (a *adapter) getAccessToken(ctx context.Context) (string, error) {
	a.tokenMu.Lock()
	if a.token != "" && time.Now().Before(a.tokenExpiry) {
		token := a.token
		a.tokenMu.Unlock()
		return token, nil
	}
	a.tokenMu.Unlock()

	body := fmt.Sprintf(`{"appId":"%s","clientSecret":"%s"}`,
		a.cfg.AppID, os.Getenv(a.cfg.AppSecretEnv))

	req, err := http.NewRequestWithContext(ctx, "POST", qqTokenEndpoint, bytes.NewBufferString(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("QQ token HTTP %d: %s", resp.StatusCode, trimQQBody(data))
	}
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Code        int    `json:"code"`
		ErrCode     int    `json:"errcode"`
		Message     string `json:"message"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	if result.AccessToken == "" {
		if result.Message != "" || result.ErrMsg != "" || result.Code != 0 || result.ErrCode != 0 {
			return "", fmt.Errorf("QQ token error code=%d errcode=%d: %s", result.Code, result.ErrCode, firstNonEmptyQQ(result.Message, result.ErrMsg))
		}
		return "", fmt.Errorf("QQ token response did not include access_token")
	}
	a.tokenMu.Lock()
	a.token = result.AccessToken
	if result.ExpiresIn > 60 {
		a.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)
	} else {
		a.tokenExpiry = time.Now().Add(5 * time.Minute)
	}
	a.tokenMu.Unlock()
	return result.AccessToken, nil
}

// Validate verifies QQ App ID / Secret by requesting an access token.
func Validate(ctx context.Context, cfg config.QQBotConfig) error {
	if cfg.AppID == "" {
		return fmt.Errorf("QQ App ID is empty")
	}
	if cfg.AppSecretEnv == "" || os.Getenv(cfg.AppSecretEnv) == "" {
		return fmt.Errorf("%s is not set", firstNonEmptyQQ(cfg.AppSecretEnv, "QQ_BOT_APP_SECRET"))
	}
	a := &adapter{cfg: cfg, logger: slog.Default().With("platform", "qq")}
	_, err := a.getAccessToken(ctx)
	return err
}

// ProbeGateway performs a short QQ gateway handshake and then closes the socket.
func ProbeGateway(ctx context.Context, cfg config.QQBotConfig) error {
	if err := Validate(ctx, cfg); err != nil {
		return err
	}
	a := &adapter{cfg: cfg, logger: slog.Default().With("platform", "qq")}
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return err
	}
	probeCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- a.connectGateway(probeCtx, token) }()
	select {
	case err := <-done:
		return err
	case <-time.After(8 * time.Second):
		cancel()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func trimQQBody(data []byte) string {
	if len(data) > 300 {
		data = data[:300]
	}
	return string(data)
}

func firstNonEmptyQQ(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func (a *adapter) connectGateway(ctx context.Context, token string) error {
	gatewayURL := qqGatewayURL(a.cfg.Environment)
	cfg, err := websocket.NewConfig(gatewayURL, gatewayURL)
	if err != nil {
		return err
	}
	cfg.Header = http.Header{}
	cfg.Header.Set("Authorization", "QQBot "+token)

	conn, err := websocket.DialConfig(cfg)
	if err != nil {
		return fmt.Errorf("dial gateway: %w", err)
	}
	defer conn.Close()

	ws := &wsClient{conn: conn, token: token, logger: a.logger}
	a.ws = ws

	var msg gatewayPayload
	decoder := json.NewDecoder(conn)

	// 第一次读取必须是 Hello
	if err := decoder.Decode(&msg); err != nil {
		return fmt.Errorf("read hello: %w", err)
	}
	if msg.Op != opHello {
		return fmt.Errorf("expected op=%d hello, got op=%d", opHello, msg.Op)
	}
	var hello helloData
	if err := json.Unmarshal(msg.D, &hello); err != nil {
		return err
	}
	ws.heartbeatMs = hello.HeartbeatInterval

	// Identify
	identify := identifyData{
		Token:   fmt.Sprintf("QQBot %s", token),
		Intents: 1<<0 | 1<<1 | 1<<9 | 1<<10 | 1<<12 | 1<<25 | 1<<26,
		Shard:   [2]int{0, 1},
		Properties: properties{
			OS:      "linux",
			Browser: "deepseek-orca",
			Device:  "deepseek-orca-bot",
		},
	}
	identifyJSON, _ := json.Marshal(identify)
	if err := ws.send(opIdentify, identifyJSON); err != nil {
		return fmt.Errorf("send identify: %w", err)
	}

	// 读取 Ready
	if err := decoder.Decode(&msg); err != nil {
		return fmt.Errorf("read ready: %w", err)
	}
	if msg.Op == opDispatch && msg.T == "READY" {
		var ready struct {
			SessionID string `json:"session_id"`
		}
		if err := json.Unmarshal(msg.D, &ready); err != nil {
			return fmt.Errorf("decode ready: %w", err)
		}
		ws.sessionID = ready.SessionID
		a.sessionID = ready.SessionID
		a.seq = msg.S
	}

	// 启动 heartbeat
	heartbeatDone := make(chan struct{})
	go func() {
		defer close(heartbeatDone)
		ticker := time.NewTicker(time.Duration(ws.heartbeatMs) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ws.mu.Lock()
				if err := ws.send(opHeartbeat, json.RawMessage(fmt.Sprintf(`%d`, ws.lastSeq))); err != nil {
					ws.logger.Error("heartbeat failed", "err", err)
					ws.mu.Unlock()
					return
				}
				ws.mu.Unlock()
			}
		}
	}()

	// 主循环：读取 dispatch 事件
	for {
		if err := decoder.Decode(&msg); err != nil {
			a.logger.Error("decode gateway message", "err", err)
			<-heartbeatDone
			return err
		}
		ws.lastSeq = msg.S
		a.seq = msg.S

		switch msg.Op {
		case opDispatch:
			a.handleDispatch(msg)
		case opHeartbeatAck:
		case opReconnect:
			a.logger.Info("gateway requested reconnect")
			<-heartbeatDone
			return nil
		}
	}
}

func (ws *wsClient) send(op int, d json.RawMessage) error {
	payload := gatewayPayload{Op: op, D: d}
	data, _ := json.Marshal(payload)
	_, err := ws.conn.Write(data)
	return err
}

func (a *adapter) handleDispatch(msg gatewayPayload) {
	var evt dispatchEvent
	if err := json.Unmarshal(msg.D, &evt); err != nil {
		a.logger.Error("parse dispatch", "err", err)
		return
	}

	ib := bot.InboundMessage{
		Platform:  bot.PlatformQQ,
		UserID:    evt.Author.ID,
		UserName:  evt.Author.Username,
		Text:      evt.Content,
		MessageID: evt.ID,
	}

	switch msg.T {
	case "C2C_MESSAGE_CREATE":
		ib.ChatType = bot.ChatDM
		ib.ChatID = evt.Author.ID
	case "GROUP_AT_MESSAGE_CREATE":
		ib.ChatType = bot.ChatGroup
		ib.ChatID = evt.GroupOpenID
	case "AT_MESSAGE_CREATE":
		ib.ChatType = bot.ChatGuild
		ib.ChatID = evt.ChannelID
	case "DIRECT_MESSAGE_CREATE":
		ib.ChatType = bot.ChatDirect
		ib.ChatID = evt.GuildID
	case "MESSAGE_CREATE":
		ib.ChatType = bot.ChatDM
		ib.ChatID = evt.ChannelID
	default:
		return // 忽略其他事件
	}

	select {
	case a.msgCh <- ib:
	default:
		a.logger.Warn("message channel full, dropping message")
	}
}

// sendMessage 使用 QQ REST API 发送消息。
func (a *adapter) sendMessage(ctx context.Context, msg bot.OutboundMessage) (bot.SendResult, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return bot.SendResult{}, err
	}

	payload := map[string]interface{}{
		"content":  msg.Text,
		"msg_type": 0,
	}

	// inline keyboard（审批用）
	if msg.Keyboard != nil {
		rows := make([]map[string]interface{}, 0)
		for _, row := range msg.Keyboard.Rows {
			buttons := make([]map[string]interface{}, 0)
			for _, btn := range row.Buttons {
				buttons = append(buttons, map[string]interface{}{
					"id": btn.ID,
					"render_data": map[string]interface{}{
						"label": btn.Label,
						"style": btn.Style,
					},
					"action": map[string]interface{}{
						"type": 2,
						"data": btn.CallbackID,
					},
				})
			}
			rows = append(rows, map[string]interface{}{"buttons": buttons})
		}
		payload["keyboard"] = map[string]interface{}{
			"content": rows,
		}
		payload["msg_type"] = 2
	}

	if msg.ReplyToMsgID != "" {
		payload["msg_id"] = msg.ReplyToMsgID
	}

	url := qqSendURL(a.cfg.Environment, msg)

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return bot.SendResult{}, err
	}
	req.Header.Set("Authorization", "QQBot "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return bot.SendResult{}, err
	}
	defer resp.Body.Close()

	var result struct {
		ID        string `json:"id"`
		Timestamp string `json:"timestamp"`
	}
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return bot.SendResult{}, fmt.Errorf("qq api error %d: %s", resp.StatusCode, string(respBody))
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return bot.SendResult{}, fmt.Errorf("decode send response: %w", err)
	}

	return bot.SendResult{MessageID: result.ID}, nil
}

func qqGatewayURL(environment string) string {
	if qqEnvironment(environment) == "sandbox" {
		return qqSandboxGatewayURL
	}
	return qqProductionGatewayURL
}

func qqRESTBase(environment string) string {
	if qqEnvironment(environment) == "sandbox" {
		return qqSandboxRESTBase
	}
	return qqProductionRESTBase
}

func qqEnvironment(environment string) string {
	if environment == "sandbox" {
		return "sandbox"
	}
	return "production"
}

func qqSendURL(environment string, msg bot.OutboundMessage) string {
	base := qqRESTBase(environment)
	switch msg.ChatType {
	case bot.ChatGroup:
		return base + fmt.Sprintf(qqSendGroupPathTemplate, msg.ChatID)
	case bot.ChatGuild, bot.ChatThread:
		return base + fmt.Sprintf(qqSendGuildPathTemplate, msg.ChatID)
	case bot.ChatDirect:
		return base + fmt.Sprintf(qqSendDirectPathTemplate, msg.ChatID)
	default:
		return base + fmt.Sprintf(qqSendMsgPathTemplate, msg.ChatID)
	}
}
