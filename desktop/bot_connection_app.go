package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"deepseek-orca/internal/bot/qq"
	"deepseek-orca/internal/bot/weixin"
	"deepseek-orca/internal/config"
)

type BotConnectionCredentialView struct {
	AppID        string `json:"appId"`
	AppSecretEnv string `json:"appSecretEnv"`
	AccountID    string `json:"accountId"`
	TokenEnv     string `json:"tokenEnv"`
	Environment  string `json:"environment"`
	SecretSet    bool   `json:"secretSet"`
}

type QQBotConnectRequest struct {
	AppID       string `json:"appId"`
	AppSecret   string `json:"appSecret"`
	Environment string `json:"environment"`
}

type BotConnectionSessionMappingView struct {
	RemoteID  string `json:"remoteId"`
	SessionID string `json:"sessionId"`
	UpdatedAt string `json:"updatedAt"`
}

type BotConnectionView struct {
	ID              string                            `json:"id"`
	Provider        string                            `json:"provider"`
	Domain          string                            `json:"domain"`
	Label           string                            `json:"label"`
	Enabled         bool                              `json:"enabled"`
	Status          string                            `json:"status"`
	Credential      BotConnectionCredentialView       `json:"credential"`
	SessionMappings []BotConnectionSessionMappingView `json:"sessionMappings"`
	LastError       string                            `json:"lastError"`
	CreatedAt       string                            `json:"createdAt"`
	UpdatedAt       string                            `json:"updatedAt"`
}

type BotInstallStartResult struct {
	OK         bool   `json:"ok"`
	Provider   string `json:"provider"`
	Domain     string `json:"domain"`
	InstallID  string `json:"installId"`
	URL        string `json:"url"`
	DeviceCode string `json:"deviceCode"`
	UserCode   string `json:"userCode"`
	Interval   int    `json:"interval"`
	ExpireIn   int    `json:"expireIn"`
	Message    string `json:"message"`
}

type BotInstallPollResult struct {
	Done       bool              `json:"done"`
	Connection BotConnectionView `json:"connection"`
	Status     string            `json:"status"`
	Message    string            `json:"message"`
	Error      string            `json:"error"`
}

type BotConnectionDiagnostic struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	MessageID string `json:"messageId"`
}

type botInstallSession struct {
	Provider   string
	Domain     string
	DeviceCode string
	UserCode   string
	StartedAt  time.Time
	ExpireAt   time.Time
	Weixin     *weixin.LoginSession
}

func (a *App) StartBotConnectionInstall(provider, domain string) (BotInstallStartResult, error) {
	provider, domain = normalizeBotInstallTarget(provider, domain)
	if provider != "weixin" {
		return BotInstallStartResult{OK: false, Provider: provider, Domain: domain, Message: "该机器人渠道暂未开放。"}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	session, err := weixin.StartLogin(ctx)
	if err != nil {
		return BotInstallStartResult{OK: false, Provider: provider, Domain: domain, Message: readableBotError(err)}, nil
	}

	installID := randomInstallID()
	a.mu.Lock()
	if a.botInstalls == nil {
		a.botInstalls = map[string]*botInstallSession{}
	}
	a.botInstalls[installID] = &botInstallSession{
		Provider:   provider,
		Domain:     domain,
		DeviceCode: session.QRCode,
		StartedAt:  session.StartedAt,
		ExpireAt:   time.Now().Add(5 * time.Minute),
		Weixin:     session,
	}
	a.mu.Unlock()

	return BotInstallStartResult{
		OK:         true,
		Provider:   provider,
		Domain:     domain,
		InstallID:  installID,
		URL:        firstNonEmptyBot(session.QRCodeURL, session.QRCode),
		DeviceCode: session.QRCode,
		Interval:   3,
		ExpireIn:   300,
		Message:    "请使用微信扫码完成连接。",
	}, nil
}

func (a *App) ConnectQQBot(req QQBotConnectRequest) (BotConnectionView, error) {
	appID := strings.TrimSpace(req.AppID)
	appSecret := strings.TrimSpace(req.AppSecret)
	env := qqEnvironmentOrDefault(req.Environment)
	if appID == "" {
		return BotConnectionView{}, fmt.Errorf("请填写 QQ Bot App ID。")
	}

	secretEnv := "QQ_BOT_APP_SECRET"
	if appSecret != "" {
		if err := upsertDotEnv(secretEnv, appSecret); err != nil {
			return BotConnectionView{}, err
		}
	} else if !envIsSet(secretEnv) {
		return BotConnectionView{}, fmt.Errorf("请填写 QQ Bot App Secret。")
	}

	qqCfg := config.QQBotConfig{Enabled: true, AppID: appID, AppSecretEnv: secretEnv, Environment: env}
	status := "connected"
	lastError := ""
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	if err := qq.Validate(ctx, qqCfg); err != nil {
		status = "warning"
		lastError = "QQ 凭据已保存，但 access token 验证未通过：" + readableBotError(err)
	} else if err := qq.ProbeGateway(ctx, qqCfg); err != nil {
		status = "warning"
		lastError = "QQ 凭据已保存，但网关暂时不可达：" + readableBotError(err)
	}
	cancel()

	conn, err := a.upsertBotConnection(config.BotConnectionConfig{
		ID:       connectionID("qq", "qq"),
		Provider: "qq",
		Domain:   "qq",
		Label:    "QQ",
		Enabled:  true,
		Status:   status,
		Credential: config.BotConnectionCredential{
			AppID:        appID,
			AppSecretEnv: secretEnv,
			Environment:  env,
		},
		LastError: lastError,
	}, func(c *config.Config) {
		c.Bot.Enabled = true
		c.Bot.Allowlist.Enabled = true
		c.Bot.Allowlist.AllowAll = true
		c.Bot.QQ.Enabled = true
		c.Bot.QQ.AppID = appID
		c.Bot.QQ.AppSecretEnv = secretEnv
		c.Bot.QQ.Environment = env
	})
	if err != nil {
		return BotConnectionView{}, err
	}
	a.restartDesktopBotGateway()
	return conn, nil
}

func (a *App) PollBotConnectionInstall(installID string) (BotInstallPollResult, error) {
	installID = strings.TrimSpace(installID)
	a.mu.RLock()
	session := a.botInstalls[installID]
	a.mu.RUnlock()
	if session == nil {
		return BotInstallPollResult{Error: "安装会话不存在，请重新开始连接。"}, nil
	}
	if time.Now().After(session.ExpireAt) {
		a.deleteBotInstall(installID)
		return BotInstallPollResult{Status: "expired", Error: "二维码已过期，请重新扫码。"}, nil
	}
	if session.Provider != "weixin" {
		return BotInstallPollResult{Status: "unavailable", Error: "该机器人渠道暂未开放。"}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	result, status, err := weixin.PollLogin(ctx, session.Weixin)
	cancel()
	if err != nil {
		return BotInstallPollResult{Status: status, Error: readableBotError(err)}, nil
	}
	if result == nil {
		return BotInstallPollResult{Status: status, Message: weixinInstallStatusMessage(status)}, nil
	}

	a.deleteBotInstall(installID)
	conn, err := a.upsertBotConnection(config.BotConnectionConfig{
		ID:         connectionID("weixin", "weixin"),
		Provider:   "weixin",
		Domain:     "weixin",
		Label:      "微信",
		Enabled:    true,
		Status:     "connected",
		Credential: config.BotConnectionCredential{AccountID: result.AccountID, TokenEnv: "WEIXIN_BOT_TOKEN"},
	}, func(c *config.Config) {
		c.Bot.Enabled = true
		c.Bot.Allowlist.Enabled = true
		c.Bot.Allowlist.AllowAll = true
		c.Bot.Weixin.Enabled = true
		c.Bot.Weixin.AccountID = result.AccountID
		c.Bot.Weixin.APIBase = result.BaseURL
		if c.Bot.Weixin.TokenEnv == "" {
			c.Bot.Weixin.TokenEnv = "WEIXIN_BOT_TOKEN"
		}
	})
	if err != nil {
		return BotInstallPollResult{Status: "error", Error: err.Error()}, nil
	}
	a.restartDesktopBotGateway()
	return BotInstallPollResult{Done: true, Status: "connected", Connection: conn, Message: "微信已连接，后台正在启动。"}, nil
}

func (a *App) DiagnoseBotConnection(id string) (BotConnectionDiagnostic, error) {
	cfg, err := config.Load()
	if err != nil {
		return BotConnectionDiagnostic{ID: id, Status: "error", Message: err.Error()}, nil
	}
	for _, conn := range cfg.Bot.Connections {
		if conn.ID != strings.TrimSpace(id) {
			continue
		}
		if !conn.Enabled {
			return BotConnectionDiagnostic{ID: conn.ID, Label: conn.Label, Status: "disabled", Message: "连接已保存但未启用。"}, nil
		}
		switch conn.Provider {
		case "qq":
			qqCfg := cfg.Bot.QQ
			qqCfg.Enabled = true
			qqCfg.AppID = firstNonEmptyBot(conn.Credential.AppID, qqCfg.AppID)
			qqCfg.AppSecretEnv = firstNonEmptyBot(conn.Credential.AppSecretEnv, qqCfg.AppSecretEnv)
			qqCfg.Environment = firstNonEmptyBot(conn.Credential.Environment, qqCfg.Environment)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			err := qq.Validate(ctx, qqCfg)
			if err == nil {
				err = qq.ProbeGateway(ctx, qqCfg)
			}
			cancel()
			if err != nil {
				return BotConnectionDiagnostic{ID: conn.ID, Label: conn.Label, Status: "warning", Message: "QQ 配置已保存，但当前诊断未完全通过：" + readableBotError(err)}, nil
			}
			return BotConnectionDiagnostic{ID: conn.ID, Label: conn.Label, Status: "ok", Message: "QQ 凭据和网关验证通过，桌面端会自动在后台接收消息。"}, nil
		case "weixin":
			wxCfg := cfg.Bot.Weixin
			wxCfg.Enabled = true
			wxCfg.AccountID = firstNonEmptyBot(conn.Credential.AccountID, wxCfg.AccountID)
			wxCfg.TokenEnv = firstNonEmptyBot(conn.Credential.TokenEnv, wxCfg.TokenEnv)
			if !weixin.HasSavedAccount(wxCfg.AccountID) && !envIsSet(wxCfg.TokenEnv) {
				return BotConnectionDiagnostic{ID: conn.ID, Label: conn.Label, Status: "warning", Message: "未找到微信登录凭据，请重新扫码连接。"}, nil
			}
			return BotConnectionDiagnostic{ID: conn.ID, Label: conn.Label, Status: "ok", Message: "微信登录凭据已保存，桌面端会自动在后台接收消息。"}, nil
		default:
			return BotConnectionDiagnostic{ID: conn.ID, Label: conn.Label, Status: "unavailable", Message: "该渠道暂未开放。"}, nil
		}
	}
	return BotConnectionDiagnostic{ID: id, Status: "missing", Message: "未找到连接。"}, nil
}

func (a *App) TestBotConnection(id, target string) (BotConnectionDiagnostic, error) {
	cfg, err := config.Load()
	if err != nil {
		return BotConnectionDiagnostic{ID: id, Status: "error", Message: err.Error()}, nil
	}
	var conn *config.BotConnectionConfig
	for i := range cfg.Bot.Connections {
		if cfg.Bot.Connections[i].ID == strings.TrimSpace(id) {
			conn = &cfg.Bot.Connections[i]
			break
		}
	}
	if conn == nil {
		return BotConnectionDiagnostic{ID: id, Status: "missing", Message: "未找到连接。"}, nil
	}
	if conn.Provider == "qq" {
		return a.DiagnoseBotConnection(conn.ID)
	}
	if conn.Provider != "weixin" {
		return BotConnectionDiagnostic{ID: conn.ID, Label: conn.Label, Status: "unavailable", Message: "该渠道暂未开放。"}, nil
	}

	target = firstNonEmptyBot(strings.TrimSpace(target), firstSessionRemoteID(conn.SessionMappings))
	if target == "" {
		return BotConnectionDiagnostic{ID: conn.ID, Label: conn.Label, Status: "warning", Message: "请输入测试会话 ID 后再发送测试消息。"}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	weixinCfg := cfg.Bot.Weixin
	weixinCfg.Enabled = true
	weixinCfg.AccountID = firstNonEmptyBot(conn.Credential.AccountID, weixinCfg.AccountID)
	weixinCfg.TokenEnv = firstNonEmptyBot(conn.Credential.TokenEnv, weixinCfg.TokenEnv)
	result, err := weixin.SendText(ctx, weixinCfg, target, "DeepSeek-Orca bot 测试消息：连接和发送链路可用。")
	if err != nil {
		return BotConnectionDiagnostic{ID: conn.ID, Label: conn.Label, Status: "error", Message: readableBotError(err)}, nil
	}
	_ = a.rememberBotConnectionRemote(conn.ID, target)
	msg := "测试消息已发送。"
	if result.MessageID != "" {
		msg += " Message ID: " + result.MessageID
	}
	return BotConnectionDiagnostic{ID: conn.ID, Label: conn.Label, Status: "ok", Message: msg, MessageID: result.MessageID}, nil
}

func (a *App) upsertBotConnection(conn config.BotConnectionConfig, updateLegacy func(*config.Config)) (BotConnectionView, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	if conn.CreatedAt == "" {
		conn.CreatedAt = now
	}
	conn.UpdatedAt = now
	if conn.Status == "" {
		conn.Status = "connected"
	}
	if conn.ID == "" {
		conn.ID = connectionID(conn.Provider, conn.Domain)
	}
	err := a.applyConfigOnly(func(c *config.Config) error {
		if updateLegacy != nil {
			updateLegacy(c)
		}
		replaced := false
		for i, existing := range c.Bot.Connections {
			if existing.ID == conn.ID {
				conn.CreatedAt = firstNonEmptyBot(existing.CreatedAt, conn.CreatedAt)
				c.Bot.Connections[i] = conn
				replaced = true
				break
			}
		}
		if !replaced {
			c.Bot.Connections = append(c.Bot.Connections, conn)
		}
		return nil
	})
	return botConnectionView(conn), err
}

func (a *App) rememberBotConnectionRemote(id, remoteID string) error {
	id = strings.TrimSpace(id)
	remoteID = strings.TrimSpace(remoteID)
	if id == "" || remoteID == "" {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339)
	return a.applyConfigOnly(func(c *config.Config) error {
		for i := range c.Bot.Connections {
			if c.Bot.Connections[i].ID != id {
				continue
			}
			for j := range c.Bot.Connections[i].SessionMappings {
				if c.Bot.Connections[i].SessionMappings[j].RemoteID == remoteID {
					c.Bot.Connections[i].SessionMappings[j].UpdatedAt = now
					c.Bot.Connections[i].UpdatedAt = now
					return nil
				}
			}
			c.Bot.Connections[i].SessionMappings = append(c.Bot.Connections[i].SessionMappings, config.BotConnectionSessionMapping{
				RemoteID:  remoteID,
				SessionID: "",
				UpdatedAt: now,
			})
			c.Bot.Connections[i].UpdatedAt = now
			return nil
		}
		return nil
	})
}

func firstSessionRemoteID(mappings []config.BotConnectionSessionMapping) string {
	for _, mapping := range mappings {
		if strings.TrimSpace(mapping.RemoteID) != "" {
			return strings.TrimSpace(mapping.RemoteID)
		}
	}
	return ""
}

func (a *App) deleteBotInstall(installID string) {
	a.mu.Lock()
	delete(a.botInstalls, installID)
	a.mu.Unlock()
}

func normalizeBotInstallTarget(provider, domain string) (string, string) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	domain = strings.ToLower(strings.TrimSpace(domain))
	if provider == "weixin" || provider == "wechat" {
		return "weixin", "weixin"
	}
	if provider == "qq" || domain == "qq" {
		return "qq", "qq"
	}
	if provider == "lark" || domain == "lark" {
		return "feishu", "lark"
	}
	return "feishu", "feishu"
}

func botConnectionView(conn config.BotConnectionConfig) BotConnectionView {
	secretSet := envIsSet(firstNonEmptyBot(conn.Credential.AppSecretEnv, conn.Credential.TokenEnv))
	if conn.Provider == "weixin" && !secretSet {
		secretSet = weixin.HasSavedAccount(conn.Credential.AccountID)
	}
	return BotConnectionView{
		ID: conn.ID, Provider: conn.Provider, Domain: conn.Domain, Label: conn.Label, Enabled: conn.Enabled, Status: conn.Status,
		Credential: BotConnectionCredentialView{
			AppID: conn.Credential.AppID, AppSecretEnv: conn.Credential.AppSecretEnv, AccountID: conn.Credential.AccountID, TokenEnv: conn.Credential.TokenEnv, Environment: conn.Credential.Environment,
			SecretSet: secretSet,
		},
		SessionMappings: botSessionMappingViews(conn.SessionMappings),
		LastError:       conn.LastError, CreatedAt: conn.CreatedAt, UpdatedAt: conn.UpdatedAt,
	}
}

func botConnectionViews(connections []config.BotConnectionConfig) []BotConnectionView {
	if connections == nil {
		return []BotConnectionView{}
	}
	out := make([]BotConnectionView, 0, len(connections))
	for _, conn := range connections {
		out = append(out, botConnectionView(conn))
	}
	return out
}

func botConnectionConfig(view BotConnectionView) config.BotConnectionConfig {
	return config.BotConnectionConfig{
		ID:       strings.TrimSpace(view.ID),
		Provider: strings.TrimSpace(view.Provider),
		Domain:   strings.TrimSpace(view.Domain),
		Label:    strings.TrimSpace(view.Label),
		Enabled:  view.Enabled,
		Status:   strings.TrimSpace(view.Status),
		Credential: config.BotConnectionCredential{
			AppID:        strings.TrimSpace(view.Credential.AppID),
			AppSecretEnv: strings.TrimSpace(view.Credential.AppSecretEnv),
			AccountID:    strings.TrimSpace(view.Credential.AccountID),
			TokenEnv:     strings.TrimSpace(view.Credential.TokenEnv),
			Environment:  strings.TrimSpace(view.Credential.Environment),
		},
		SessionMappings: botSessionMappingConfigs(view.SessionMappings),
		LastError:       strings.TrimSpace(view.LastError),
		CreatedAt:       strings.TrimSpace(view.CreatedAt),
		UpdatedAt:       strings.TrimSpace(view.UpdatedAt),
	}
}

func botConnectionConfigs(views []BotConnectionView) []config.BotConnectionConfig {
	if views == nil {
		return nil
	}
	out := make([]config.BotConnectionConfig, 0, len(views))
	for _, view := range views {
		cfg := botConnectionConfig(view)
		if cfg.ID == "" || cfg.Provider == "" {
			continue
		}
		out = append(out, cfg)
	}
	return out
}

func botSessionMappingViews(mappings []config.BotConnectionSessionMapping) []BotConnectionSessionMappingView {
	if mappings == nil {
		return []BotConnectionSessionMappingView{}
	}
	out := make([]BotConnectionSessionMappingView, 0, len(mappings))
	for _, m := range mappings {
		out = append(out, BotConnectionSessionMappingView{RemoteID: m.RemoteID, SessionID: m.SessionID, UpdatedAt: m.UpdatedAt})
	}
	return out
}

func botSessionMappingConfigs(mappings []BotConnectionSessionMappingView) []config.BotConnectionSessionMapping {
	if mappings == nil {
		return nil
	}
	out := make([]config.BotConnectionSessionMapping, 0, len(mappings))
	for _, m := range mappings {
		out = append(out, config.BotConnectionSessionMapping{
			RemoteID:  strings.TrimSpace(m.RemoteID),
			SessionID: strings.TrimSpace(m.SessionID),
			UpdatedAt: strings.TrimSpace(m.UpdatedAt),
		})
	}
	return out
}

func connectionID(provider, domain string) string {
	return strings.Trim(strings.ToLower(provider+"-"+domain), "-")
}

func randomInstallID() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("install-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func envIsSet(name string) bool {
	return strings.TrimSpace(name) != "" && strings.TrimSpace(os.Getenv(name)) != ""
}

func firstNonEmptyBot(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func weixinInstallStatusMessage(status string) string {
	switch status {
	case "scaned":
		return "已扫码，请在微信里确认。"
	case "scaned_but_redirect":
		return "已扫码，正在切换微信授权节点。"
	case "need_verifycode", "verify_code_blocked":
		return "微信要求输入配对码，请按微信提示完成验证。"
	case "binded_redirect":
		return "该微信账号已绑定，正在恢复连接。"
	case "expired":
		return "二维码已过期，请重新扫码。"
	default:
		return "等待扫码。"
	}
}

func readableBotError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	replacements := map[string]string{
		"qq app_id is empty":                         "QQ App ID 为空",
		"qq app secret is empty":                     "QQ App Secret 为空",
		"empty access token":                         "QQ 平台没有返回 access token，请检查 App ID / Secret 和机器人审核状态",
		"weixin token is not configured":             "微信登录凭据不存在，请重新扫码",
		"no saved weixin account":                    "没有找到已保存的微信登录账号",
		"weixin qr response missing qrcode":          "微信服务器没有返回二维码，请稍后重试",
		"weixin qr confirmed but credential payload": "微信已确认扫码，但没有返回完整登录凭据，请重新扫码",
	}
	for needle, replacement := range replacements {
		if strings.Contains(msg, needle) {
			return replacement + "。"
		}
	}
	return msg
}
