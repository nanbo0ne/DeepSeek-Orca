package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"deepseek-orca/internal/bot"
	"deepseek-orca/internal/bot/qq"
	"deepseek-orca/internal/bot/weixin"
	"deepseek-orca/internal/config"
)

type BotRuntimeStatusView struct {
	Status   string   `json:"status"`
	Message  string   `json:"message"`
	Channels []string `json:"channels"`
}

func (a *App) restartDesktopBotGateway() {
	cfg, err := config.Load()
	if err != nil {
		a.setBotRuntimeStatus("error", "读取机器人配置失败："+err.Error())
		return
	}
	a.startDesktopBotGateway(cfg)
}

func (a *App) startDesktopBotGateway(cfg *config.Config) {
	a.stopDesktopBotGateway()
	if cfg == nil {
		a.setBotRuntimeStatus("error", "机器人配置为空。")
		return
	}

	cfg.Bot.Enabled = true
	cfg.Bot.Allowlist.Enabled = true
	cfg.Bot.Allowlist.AllowAll = true

	enabled := map[bot.Platform]bool{}
	adapters := map[bot.Platform]bot.Adapter{}
	channels := []string{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if qqConfigured(cfg.Bot.QQ) {
		enabled[bot.PlatformQQ] = true
		adapters[bot.PlatformQQ] = qq.New(cfg.Bot.QQ, logger)
		channels = append(channels, "QQ")
	}
	if weixinConfigured(cfg.Bot.Weixin) {
		enabled[bot.PlatformWeixin] = true
		adapters[bot.PlatformWeixin] = weixin.New(cfg.Bot.Weixin, logger)
		channels = append(channels, "微信")
	}
	if len(adapters) == 0 {
		a.setBotRuntimeStatus("idle", "QQ/微信尚未连接。")
		return
	}

	modelName := strings.TrimSpace(cfg.Bot.Model)
	if modelName == "" {
		modelName = cfg.DefaultModel
	}
	workspaceRoot := strings.TrimSpace(cfg.Bot.WorkspaceRoot)
	if workspaceRoot == "" {
		workspaceRoot = config.BotWorkspaceDir()
	}

	ctx, cancel := context.WithCancel(context.Background())
	gw := bot.NewGateway(bot.GatewayConfig{
		Model:         modelName,
		MaxSteps:      cfg.Bot.MaxSteps,
		WorkspaceRoot: workspaceRoot,
		Enabled:       enabled,
		Allowlist: bot.AllowlistConfig{
			Enabled:  true,
			AllowAll: true,
			Users:    map[bot.Platform][]string{},
			Groups:   map[bot.Platform][]string{},
		},
		Debounce: time.Duration(cfg.Bot.DebounceMs) * time.Millisecond,
	}, adapters, logger)
	if err := gw.Start(ctx); err != nil {
		cancel()
		a.setBotRuntimeStatus("error", "机器人后台网关启动失败："+err.Error())
		return
	}

	a.mu.Lock()
	a.botGateway = gw
	a.botGatewayCancel = cancel
	a.botRuntimeStatus = "running"
	a.botRuntimeErr = ""
	a.mu.Unlock()
	a.setBotRuntimeStatus("running", fmt.Sprintf("后台运行中：%s", strings.Join(channels, "、")))
}

func (a *App) stopDesktopBotGateway() {
	a.mu.Lock()
	gw := a.botGateway
	cancel := a.botGatewayCancel
	a.botGateway = nil
	a.botGatewayCancel = nil
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if gw != nil {
		gw.Stop()
	}
}

func (a *App) setBotRuntimeStatus(status, message string) {
	a.mu.Lock()
	a.botRuntimeStatus = status
	a.botRuntimeErr = message
	a.mu.Unlock()
}

func (a *App) GetBotRuntimeStatus() (BotRuntimeStatusView, error) {
	cfg, _ := config.Load()
	channels := []string{}
	if cfg != nil {
		if qqConfigured(cfg.Bot.QQ) {
			channels = append(channels, "QQ")
		}
		if weixinConfigured(cfg.Bot.Weixin) {
			channels = append(channels, "微信")
		}
	}
	a.mu.RLock()
	status := a.botRuntimeStatus
	message := a.botRuntimeErr
	a.mu.RUnlock()
	if status == "" {
		status = "idle"
		message = "QQ/微信尚未连接。"
	}
	return BotRuntimeStatusView{Status: status, Message: message, Channels: channels}, nil
}

func qqConfigured(cfg config.QQBotConfig) bool {
	return strings.TrimSpace(cfg.AppID) != "" && strings.TrimSpace(cfg.AppSecretEnv) != "" && os.Getenv(cfg.AppSecretEnv) != ""
}

func weixinConfigured(cfg config.WeixinBotConfig) bool {
	if strings.TrimSpace(cfg.TokenEnv) != "" && os.Getenv(cfg.TokenEnv) != "" {
		return true
	}
	return weixin.HasSavedAccount(cfg.AccountID)
}
