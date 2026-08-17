// Package qq implements the official QQ Bot API v2 adapter.
package qq

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/bot"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/config"
)

// New creates a QQ Bot adapter.
func New(cfg config.QQBotConfig, logger *slog.Logger) bot.Adapter {
	return &adapter{
		cfg:    cfg,
		logger: logger.With("platform", "qq"),
	}
}

type adapter struct {
	cfg    config.QQBotConfig
	logger *slog.Logger
	msgCh  chan bot.InboundMessage
	cancel context.CancelFunc

	ws          *wsClient
	sessionID   string
	seq         int64
	token       string
	tokenExpiry time.Time
	tokenMu     sync.Mutex

	sendMu             sync.Mutex
	nextOutboundMsgSeq int
	markdownDisabled   bool
}

func (a *adapter) Platform() bot.Platform { return bot.PlatformQQ }
func (a *adapter) Name() string           { return "qq" }

func (a *adapter) Start(ctx context.Context) error {
	a.msgCh = make(chan bot.InboundMessage, 64)
	startupCtx, startupCancel := context.WithTimeout(ctx, qqStartupValidationTimeout)
	defer startupCancel()
	token, err := a.getAccessToken(startupCtx)
	if err != nil {
		return err
	}
	if _, err := a.getGatewayURL(startupCtx, token); err != nil {
		return err
	}
	ctx, a.cancel = context.WithCancel(ctx)

	go a.gatewayLoop(ctx)
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
	return nil // QQ Bot does not support typing indicators.
}

func (a *adapter) Messages() <-chan bot.InboundMessage {
	return a.msgCh
}
