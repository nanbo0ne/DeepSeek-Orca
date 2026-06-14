package cli

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"deepseek-orca/internal/bot"
	"deepseek-orca/internal/bot/feishu"
	"deepseek-orca/internal/bot/qq"
	"deepseek-orca/internal/bot/weixin"
	"deepseek-orca/internal/config"
)

func botCommand(args []string, version string) int {
	if len(args) < 1 {
		botUsage()
		return 2
	}

	sub := args[0]
	rest := args[1:]

	switch sub {
	case "start":
		return botStart(rest, version)
	case "doctor":
		return botDoctor(rest)
	case "weixin-login":
		return botWeixinLogin(rest)
	case "help", "--help", "-h":
		botUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown bot subcommand %q\n\n", sub)
		botUsage()
		return 2
	}
}

func botStart(args []string, version string) int {
	fs := flag.NewFlagSet("bot start", flag.ContinueOnError)
	channels := fs.String("channels", "", "enabled channels, comma-separated: qq,feishu,weixin")
	dir := fs.String("dir", "", "workspace directory; empty uses [bot] workspace_root or the isolated bot workspace")
	model := fs.String("model", "", "model name; empty uses [bot] model or default_model")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load config: %v\n", err)
		return 1
	}

	if !cfg.Bot.Enabled {
		fmt.Fprintln(os.Stderr, "error: bot is not enabled in config; set [bot] enabled = true")
		return 1
	}
	if !cfg.Bot.Allowlist.AllowAll && (!cfg.Bot.Allowlist.Enabled || botAllowlistUserCount(cfg.Bot.Allowlist) == 0) {
		fmt.Fprintln(os.Stderr, "error: bot requires an explicit allowlist; set [bot.allowlist] enabled = true with platform user ids, or set allow_all = true intentionally")
		return 1
	}

	workspaceRoot := strings.TrimSpace(*dir)
	if workspaceRoot == "" {
		workspaceRoot = strings.TrimSpace(cfg.Bot.WorkspaceRoot)
	}
	if workspaceRoot == "" {
		workspaceRoot = config.BotWorkspaceDir()
	}
	if workspaceRoot == "" {
		if wd, err := os.Getwd(); err == nil {
			workspaceRoot = wd
		}
	}
	if workspaceRoot != "" {
		if err := os.MkdirAll(workspaceRoot, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "error: create bot workspace %q: %v\n", workspaceRoot, err)
			return 1
		}
	}

	// Determine enabled platforms.
	enabledPlatforms := make(map[bot.Platform]bool)
	if *channels != "" {
		for _, ch := range strings.Split(*channels, ",") {
			ch = strings.TrimSpace(ch)
			switch bot.Platform(ch) {
			case bot.PlatformQQ:
				enabledPlatforms[bot.PlatformQQ] = cfg.Bot.QQ.Enabled
			case bot.PlatformFeishu:
				enabledPlatforms[bot.PlatformFeishu] = cfg.Bot.Feishu.Enabled
			case bot.PlatformWeixin:
				enabledPlatforms[bot.PlatformWeixin] = cfg.Bot.Weixin.Enabled
			default:
				fmt.Fprintf(os.Stderr, "warning: unknown channel %q\n", ch)
			}
		}
	} else {
		enabledPlatforms[bot.PlatformQQ] = cfg.Bot.QQ.Enabled
		enabledPlatforms[bot.PlatformFeishu] = cfg.Bot.Feishu.Enabled
		enabledPlatforms[bot.PlatformWeixin] = cfg.Bot.Weixin.Enabled
	}

	hasEnabled := false
	for _, v := range enabledPlatforms {
		if v {
			hasEnabled = true
			break
		}
	}
	if !hasEnabled {
		fmt.Fprintln(os.Stderr, "error: no bot channels enabled; enable at least one in config")
		return 1
	}

	modelName := strings.TrimSpace(*model)
	if modelName == "" {
		modelName = cfg.Bot.Model
	}
	if modelName == "" {
		modelName = cfg.DefaultModel
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Build gateway config.
	gwCfg := bot.GatewayConfig{
		Model:         modelName,
		MaxSteps:      cfg.Bot.MaxSteps,
		WorkspaceRoot: workspaceRoot,
		Enabled:       enabledPlatforms,
		Allowlist: bot.AllowlistConfig{
			Enabled:  cfg.Bot.Allowlist.Enabled,
			AllowAll: cfg.Bot.Allowlist.AllowAll,
			Users: map[bot.Platform][]string{
				bot.PlatformQQ:     cfg.Bot.Allowlist.QQUsers,
				bot.PlatformFeishu: cfg.Bot.Allowlist.FeishuUsers,
				bot.PlatformWeixin: cfg.Bot.Allowlist.WeixinUsers,
			},
			Groups: map[bot.Platform][]string{
				bot.PlatformQQ:     cfg.Bot.Allowlist.QQGroups,
				bot.PlatformFeishu: cfg.Bot.Allowlist.FeishuGroups,
				bot.PlatformWeixin: cfg.Bot.Allowlist.WeixinGroups,
			},
		},
		Debounce: time.Duration(cfg.Bot.DebounceMs) * time.Millisecond,
	}

	// Create channel adapters.
	adapters := make(map[bot.Platform]bot.Adapter)
	if enabledPlatforms[bot.PlatformQQ] {
		adapters[bot.PlatformQQ] = qq.New(cfg.Bot.QQ, logger)
	}
	if enabledPlatforms[bot.PlatformFeishu] {
		adapters[bot.PlatformFeishu] = feishu.New(cfg.Bot.Feishu, logger)
	}
	if enabledPlatforms[bot.PlatformWeixin] {
		adapters[bot.PlatformWeixin] = weixin.New(cfg.Bot.Weixin, logger)
	}

	gw := bot.NewGateway(gwCfg, adapters, logger)

	// Handle shutdown signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "\nshutting down...")
		cancel()
		gw.Stop()
	}()

	fmt.Fprintf(os.Stderr, "deepseek-orca bot starting (model: %s, channels: %s)...\n", modelName, *channels)
	fmt.Fprintf(os.Stderr, "workspace: %s\n", workspaceRoot)
	fmt.Fprintf(os.Stderr, "version: %s\n", version)

	if err := gw.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: start gateway: %v\n", err)
		return 1
	}

	// Wait for a signal or context cancellation.
	<-ctx.Done()
	return 0
}

func botDoctor(args []string) int {
	fs := flag.NewFlagSet("bot doctor", flag.ContinueOnError)
	jsonOut := fs.Bool("json", false, "output JSON")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load config: %v\n", err)
		return 1
	}

	bc := cfg.Bot
	type checkResult struct {
		Name   string `json:"name"`
		Status string `json:"status"`
		Detail string `json:"detail,omitempty"`
	}
	var results []checkResult
	addCheck := func(name, status, detail string) {
		results = append(results, checkResult{Name: name, Status: status, Detail: detail})
	}

	if bc.Enabled {
		addCheck("bot.enabled", "ok", "")
	} else {
		addCheck("bot.enabled", "disabled", "bot is not enabled in config")
	}
	if strings.TrimSpace(bc.WorkspaceRoot) != "" {
		addCheck("bot.workspace_root", "ok", bc.WorkspaceRoot)
	} else if def := config.BotWorkspaceDir(); def != "" {
		addCheck("bot.workspace_root", "default", def)
	} else {
		addCheck("bot.workspace_root", "warning", "unable to resolve default bot workspace")
	}

	if bc.QQ.Enabled {
		addCheck("bot.qq.enabled", "ok", "")
		secret := os.Getenv(bc.QQ.AppSecretEnv)
		if secret == "" {
			addCheck("bot.qq.app_secret", "missing", bc.QQ.AppSecretEnv+" is not set")
		} else {
			addCheck("bot.qq.app_secret", "ok", bc.QQ.AppSecretEnv+" is set")
		}
		if bc.QQ.AppID == "" {
			addCheck("bot.qq.app_id", "missing", "app_id is empty")
		} else {
			addCheck("bot.qq.app_id", "ok", "app_id configured")
		}
	} else {
		addCheck("bot.qq", "disabled", "")
	}

	if bc.Feishu.Enabled {
		addCheck("bot.feishu.enabled", "ok", "")
		secret := os.Getenv(bc.Feishu.AppSecretEnv)
		if secret == "" {
			addCheck("bot.feishu.app_secret", "missing", bc.Feishu.AppSecretEnv+" is not set")
		} else {
			addCheck("bot.feishu.app_secret", "ok", bc.Feishu.AppSecretEnv+" is set")
		}
		if bc.Feishu.AppID == "" {
			addCheck("bot.feishu.app_id", "missing", "app_id is empty")
		} else {
			addCheck("bot.feishu.app_id", "ok", "app_id configured")
		}
		mode := bc.Feishu.Mode
		if mode == "" {
			mode = "webhook"
		}
		addCheck("bot.feishu.mode", "ok", mode)
	} else {
		addCheck("bot.feishu", "disabled", "")
	}

	if bc.Weixin.Enabled {
		addCheck("bot.weixin.enabled", "ok", "")
		token := os.Getenv(bc.Weixin.TokenEnv)
		if token != "" {
			addCheck("bot.weixin.token", "ok", bc.Weixin.TokenEnv+" is set")
		} else if weixin.HasSavedAccount(bc.Weixin.AccountID) {
			addCheck("bot.weixin.token", "ok", "saved iLink account is available")
		} else {
			addCheck("bot.weixin.token", "missing", bc.Weixin.TokenEnv+" is not set; run `deepseek-orca bot weixin-login` to save an iLink account")
		}
	} else {
		addCheck("bot.weixin", "disabled", "")
	}

	if bc.Allowlist.AllowAll {
		addCheck("bot.allowlist", "open", "allow_all=true; every reachable user can trigger local tools")
	} else if bc.Allowlist.Enabled {
		addCheck("bot.allowlist", "enabled",
			fmt.Sprintf("qq=%d feishu=%d weixin=%d users", len(bc.Allowlist.QQUsers), len(bc.Allowlist.FeishuUsers), len(bc.Allowlist.WeixinUsers)))
	} else {
		addCheck("bot.allowlist", "missing", "bot start will refuse without allowlist or allow_all=true")
	}

	if *jsonOut {
		fmt.Println("[")
		for i, r := range results {
			comma := ","
			if i == len(results)-1 {
				comma = ""
			}
			fmt.Printf("  {\"name\":%q,\"status\":%q,\"detail\":%q}%s\n", r.Name, r.Status, r.Detail, comma)
		}
		fmt.Println("]")
	} else {
		for _, r := range results {
			marker := "OK"
			if r.Status == "missing" || r.Status == "disabled" {
				marker = "!!"
			}
			fmt.Printf("  %s %s: %s", marker, r.Name, r.Status)
			if r.Detail != "" {
				fmt.Printf(" - %s", r.Detail)
			}
			fmt.Println()
		}
	}
	return 0
}

func botAllowlistUserCount(a config.BotAllowlist) int {
	return len(a.QQUsers) + len(a.FeishuUsers) + len(a.WeixinUsers)
}

func botWeixinLogin(args []string) int {
	fs := flag.NewFlagSet("bot weixin-login", flag.ContinueOnError)
	timeoutSeconds := fs.Int("timeout", 480, "login timeout in seconds")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: load config: %v\n", err)
		return 1
	}

	if !cfg.Bot.Weixin.Enabled {
		fmt.Fprintln(os.Stderr, "error: weixin bot is not enabled in config")
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutSeconds)*time.Second)
	defer cancel()
	result, err := weixin.Login(ctx, os.Stdout, time.Duration(*timeoutSeconds)*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: weixin login failed: %v\n", err)
		return 1
	}
	fmt.Printf("\nWeChat login succeeded: account_id=%s user_id=%s base_url=%s\n", result.AccountID, result.UserID, result.BaseURL)
	fmt.Println("Credential saved to the DeepSeek-Orca user config directory; you can also set [bot.weixin] account_id to this account_id.")
	return 0
}

func botUsage() {
	fmt.Print(`deepseek-orca bot - multi-channel IM bot gateway (QQ / Feishu / WeChat)

Usage:
  deepseek-orca bot start   [--channels qq,feishu,weixin] [--dir PATH] [--model NAME]
  deepseek-orca bot doctor  [--json]
  deepseek-orca bot weixin-login [--timeout SECONDS]

Subcommands:
  start         Start the bot gateway
  doctor        Diagnose bot configuration and connectivity
  weixin-login  Log in to WeChat iLink by QR code

Examples:
  deepseek-orca bot start --channels qq,feishu
  deepseek-orca bot start --dir /path/to/project --model deepseek-pro
  deepseek-orca bot doctor --json

Configuration:
  Edit deepseek-orca.toml:
    [bot]           enabled / model / workspace_root / max_steps
    [bot.allowlist]  enabled / qq_users / feishu_users / weixin_users
    [bot.qq]         enabled / app_id / app_secret_env
    [bot.feishu]     enabled / app_id / app_secret_env / verification_token / mode
    [bot.weixin]     enabled / account_id / token_env / api_base

  All secrets are read from environment variables; never put keys in config files.
`)
}
