package bot

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"deepseek-orca/internal/event"
)

func TestRenderSinkBuffersStreamingTextUntilTurnDone(t *testing.T) {
	fa := newFakeAdapter(PlatformWeixin, "fake-weixin")
	sink := newRenderSink(context.Background(), fa, "chat-1", ChatDM, "msg-1", slog.New(slog.NewTextHandler(io.Discard, nil)), nil)

	sink.Emit(event.Event{Kind: event.TurnStarted})
	sink.Emit(event.Event{Kind: event.Text, Text: "你好"})
	sink.Emit(event.Event{Kind: event.Text, Text: "，我正在"})
	sink.Emit(event.Event{Kind: event.Text, Text: "处理。"})

	if sent := fa.sentMessages(); len(sent) != 0 {
		t.Fatalf("streaming text should be buffered, sent = %#v", sent)
	}

	sink.Emit(event.Event{Kind: event.Message, Text: "你好，我正在处理。"})
	sink.Emit(event.Event{Kind: event.TurnDone})

	sent := fa.sentMessages()
	if len(sent) != 1 {
		t.Fatalf("sent count = %d, want 1: %#v", len(sent), sent)
	}
	if sent[0].Text != "你好，我正在处理。" {
		t.Fatalf("final text = %q", sent[0].Text)
	}
}

func TestRenderSinkSendsCompactToolProgress(t *testing.T) {
	fa := newFakeAdapter(PlatformQQ, "fake-qq")
	sink := newRenderSink(context.Background(), fa, "chat-1", ChatDM, "msg-1", slog.New(slog.NewTextHandler(io.Discard, nil)), nil)

	sink.Emit(event.Event{Kind: event.TurnStarted})
	sink.Emit(event.Event{Kind: event.ToolDispatch, Tool: event.Tool{ID: "t1", Name: "read_file", ReadOnly: true}})
	sink.Emit(event.Event{Kind: event.ToolResult, Tool: event.Tool{ID: "t1", Name: "read_file", Output: strings.Repeat("x", 2000)}})
	sink.Emit(event.Event{Kind: event.Message, Text: "读取完成。"})
	sink.Emit(event.Event{Kind: event.TurnDone})

	sent := fa.sentMessages()
	if len(sent) != 2 {
		t.Fatalf("sent count = %d, want progress + final: %#v", len(sent), sent)
	}
	if sent[0].Text != "正在读取文件…" {
		t.Fatalf("progress text = %q", sent[0].Text)
	}
	if sent[1].Text != "读取完成。" {
		t.Fatalf("final text = %q", sent[1].Text)
	}
}
