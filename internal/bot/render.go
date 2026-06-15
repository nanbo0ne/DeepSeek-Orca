package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"deepseek-orca/internal/control"
	"deepseek-orca/internal/event"
)

// renderSink renders DeepSeek-Orca event streams into compact IM messages.
// It buffers assistant text until TurnDone so mobile clients receive a coherent
// reply instead of many tiny streaming bubbles.
type renderSink struct {
	ctx      context.Context
	adapter  Adapter
	chatID   string
	chatType ChatType
	replyTo  string
	logger   *slog.Logger
	ctrl     *control.Controller
	onAsk    func(event.Ask)

	buf             strings.Builder
	finalText       string
	toolNames       map[string]string
	sentProgress    map[string]bool
	lastProgressAt  time.Time
	progressCounter int
}

func newRenderSink(ctx context.Context, adapter Adapter, chatID string, chatType ChatType, replyTo string, logger *slog.Logger, onAsk func(event.Ask)) *renderSink {
	return &renderSink{
		ctx:          ctx,
		adapter:      adapter,
		chatID:       chatID,
		chatType:     chatType,
		replyTo:      replyTo,
		logger:       logger,
		onAsk:        onAsk,
		toolNames:    make(map[string]string),
		sentProgress: make(map[string]bool),
	}
}

func (s *renderSink) Emit(e event.Event) {
	switch e.Kind {
	case event.TurnStarted:
		s.buf.Reset()
		s.finalText = ""
		s.toolNames = make(map[string]string)
		s.sentProgress = make(map[string]bool)
		s.lastProgressAt = time.Time{}
		s.progressCounter = 0

	case event.Text:
		s.buf.WriteString(e.Text)

	case event.Message:
		if strings.TrimSpace(e.Text) != "" {
			s.finalText = e.Text
		}

	case event.ToolDispatch:
		s.toolNames[e.Tool.ID] = e.Tool.Name
		s.sendProgress(toolProgressText(e.Tool.Name, e.Tool.ReadOnly))

	case event.ToolResult:
		name := firstNonEmptyBotString(s.toolNames[e.Tool.ID], e.Tool.Name, e.Tool.ID)
		if strings.TrimSpace(e.Tool.Err) != "" {
			s.sendImmediate(fmt.Sprintf("工具 %s 执行失败：%s", name, e.Tool.Err))
		} else {
			s.sendProgress(toolDoneText(name))
		}

	case event.ApprovalRequest:
		approvalText := fmt.Sprintf("需要批准操作\n工具：%s\n操作：%s\n\nID：`%s`\n使用 `/approve %s` 批准，或 `/deny %s` 拒绝。",
			e.Approval.Tool, e.Approval.Subject, e.Approval.ID, e.Approval.ID, e.Approval.ID)
		msg := OutboundMessage{
			ChatID:       s.chatID,
			ChatType:     s.chatType,
			Text:         approvalText,
			ReplyToMsgID: s.replyTo,
		}
		switch s.adapter.Platform() {
		case PlatformQQ:
			msg.Keyboard = approvalKeyboard(e.Approval.ID)
		case PlatformFeishu:
			msg.Card = approvalCard(e.Approval, s.chatType)
		}
		_ = s.send(msg)

	case event.AskRequest:
		if s.onAsk != nil {
			s.onAsk(e.Ask)
		}
		text := askRequestText(e.Ask)
		msg := OutboundMessage{
			ChatID:       s.chatID,
			ChatType:     s.chatType,
			Text:         text,
			ReplyToMsgID: s.replyTo,
		}
		if s.adapter.Platform() == PlatformFeishu {
			msg.Card = askCard(e.Ask, text)
		}
		_ = s.send(msg)

	case event.TurnDone:
		s.flushFinal()
		if e.Err != nil && !strings.Contains(e.Err.Error(), "context canceled") {
			s.sendImmediate(fmt.Sprintf("执行出错：%v", e.Err))
		}

	case event.Notice:
		if e.Level == event.LevelWarn {
			s.sendImmediate("提示：" + e.Text)
		}

	case event.CompactionStarted:
		s.sendProgress("正在整理上下文…")
	}
}

func (s *renderSink) sendProgress(text string) {
	text = strings.TrimSpace(text)
	if text == "" || s.sentProgress[text] {
		return
	}
	if s.progressCounter >= 4 && time.Since(s.lastProgressAt) < 8*time.Second {
		return
	}
	if !s.lastProgressAt.IsZero() && time.Since(s.lastProgressAt) < 1200*time.Millisecond {
		return
	}
	s.sentProgress[text] = true
	s.progressCounter++
	s.lastProgressAt = time.Now()
	s.sendImmediate(text)
}

func (s *renderSink) flushFinal() {
	text := strings.TrimSpace(s.finalText)
	if text == "" {
		text = strings.TrimSpace(s.buf.String())
	}
	if text == "" {
		return
	}
	s.sendImmediate(text)
	s.buf.Reset()
	s.finalText = ""
}

func (s *renderSink) sendImmediate(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	_ = s.send(OutboundMessage{
		ChatID:       s.chatID,
		ChatType:     s.chatType,
		Text:         text,
		ReplyToMsgID: s.replyTo,
	})
}

func (s *renderSink) send(msg OutboundMessage) error {
	_, err := s.adapter.Send(s.ctx, msg)
	return err
}

func toolProgressText(name string, readOnly bool) string {
	name = strings.TrimSpace(name)
	switch name {
	case "read_file", "list_files", "glob", "grep", "search", "rg":
		return "正在读取文件…"
	case "apply_patch", "write_file", "edit_file":
		return "正在修改文件…"
	case "bash", "shell", "run_command":
		return "正在运行命令…"
	case "task":
		return "正在调用 subagent…"
	}
	if readOnly {
		return "正在查看信息…"
	}
	if name == "" {
		return "正在处理…"
	}
	return "正在执行：" + name
}

func toolDoneText(name string) string {
	switch strings.TrimSpace(name) {
	case "apply_patch", "write_file", "edit_file":
		return "文件修改已完成。"
	case "bash", "shell", "run_command":
		return "命令执行已完成。"
	}
	return ""
}

func askRequestText(ask event.Ask) string {
	var qb strings.Builder
	qb.WriteString("请回答以下问题：\n")
	for i, q := range ask.Questions {
		fmt.Fprintf(&qb, "\n%d. %s\n", i+1, q.Prompt)
		for j, opt := range q.Options {
			fmt.Fprintf(&qb, "  %d. %s", j+1, opt.Label)
			if opt.Description != "" {
				fmt.Fprintf(&qb, " - %s", opt.Description)
			}
			qb.WriteString("\n")
		}
		if q.Multi {
			qb.WriteString("  （可多选，用逗号分隔）\n")
		}
	}
	fmt.Fprintf(&qb, "\nID：`%s`", ask.ID)
	fmt.Fprintf(&qb, "\n使用 `/answer %s <选项编号或文本>` 回答。", ask.ID)
	return qb.String()
}

func approvalKeyboard(id string) *InlineKeyboard {
	return &InlineKeyboard{Rows: []InlineKeyboardRow{{
		Buttons: []InlineKeyboardButton{
			{ID: "allow_once", Label: "允许一次", Style: 1, CallbackID: "/approve " + id},
			{ID: "deny", Label: "拒绝", Style: 2, CallbackID: "/deny " + id},
		},
	}}}
}

func approvalCard(a event.Approval, chatType ChatType) *InteractiveCard {
	return &InteractiveCard{
		Header: "需要批准操作",
		Elements: []InteractiveCardElement{
			{Tag: "markdown", Content: fmt.Sprintf("**工具**：%s\n\n**操作**：%s\n\nID：`%s`", a.Tool, a.Subject, a.ID)},
			{Tag: "action", Extra: map[string]any{
				"actions": []map[string]any{
					{"tag": "button", "text": map[string]string{"tag": "plain_text", "content": "允许一次"}, "type": "primary", "value": cardActionValue("/approve "+a.ID, chatType)},
					{"tag": "button", "text": map[string]string{"tag": "plain_text", "content": "拒绝"}, "type": "danger", "value": cardActionValue("/deny "+a.ID, chatType)},
				},
			}},
		},
	}
}

func cardActionValue(command string, chatType ChatType) map[string]string {
	return map[string]string{
		"command":   command,
		"chat_type": string(chatType),
	}
}

func askCard(ask event.Ask, fallback string) *InteractiveCard {
	return &InteractiveCard{
		Header: "需要回答问题",
		Elements: []InteractiveCardElement{
			{Tag: "markdown", Content: fallback},
		},
	}
}

func firstNonEmptyBotString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
