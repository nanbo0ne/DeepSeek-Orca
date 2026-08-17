package weixin

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/bot"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/config"
)

func TestIlinkResponseAcceptsNumericMessageIDs(t *testing.T) {
	var got ilinkResponse
	err := json.Unmarshal([]byte(`{
		"ret": 0,
		"errcode": 0,
		"msgs": [{
			"message_id": 123456,
			"from_user_id": 987,
			"to_user_id": "bot",
			"context_token": 456,
			"msg_type": 1,
			"item_list": [{
				"type": 1,
				"text_item": {"text": "hello"}
			}]
		}],
		"updates": [{
			"update_id": 1,
			"update_type": "message",
			"message": {
				"message_id": 234567,
				"chat_id": 345678,
				"chat_type": "private",
				"from": {"user_id": 456789, "user_name": "tester"},
				"text": "ping"
			}
		}]
	}`), &got)
	if err != nil {
		t.Fatalf("Unmarshal ilinkResponse: %v", err)
	}
	if got.Msgs[0].MessageID.String() != "123456" {
		t.Fatalf("message_id = %q, want 123456", got.Msgs[0].MessageID.String())
	}
	if got.Msgs[0].FromUserID.String() != "987" {
		t.Fatalf("from_user_id = %q, want 987", got.Msgs[0].FromUserID.String())
	}
	if got.Msgs[0].ContextToken.String() != "456" {
		t.Fatalf("context_token = %q, want 456", got.Msgs[0].ContextToken.String())
	}
	if got.Updates[0].Message.ChatID.String() != "345678" {
		t.Fatalf("update chat_id = %q, want 345678", got.Updates[0].Message.ChatID.String())
	}
}

func TestGetUpdatesQueuesMessageAndSendUsesOpenClawShape(t *testing.T) {
	t.Setenv("WEIXIN_BOT_TOKEN", "token")
	var getUpdatesBody map[string]any
	var sendBody map[string]any
	var getConfigBody map[string]any
	var sendTypingBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Fatalf("authorization = %q, want bearer token", got)
		}
		switch r.URL.Path {
		case "/ilink/bot/getupdates":
			decodeJSON(t, r, &getUpdatesBody)
			writeJSON(t, w, map[string]any{
				"ret":             0,
				"errcode":         0,
				"get_updates_buf": "next-buf",
				"msgs": []map[string]any{{
					"message_id":    "m1",
					"from_user_id":  "user-1",
					"to_user_id":    "bot-account",
					"context_token": "ctx-1",
					"msg_type":      1,
					"item_list": []map[string]any{{
						"type":      1,
						"text_item": map[string]any{"text": "hello"},
					}},
				}},
			})
		case "/ilink/bot/sendmessage":
			decodeJSON(t, r, &sendBody)
			writeJSON(t, w, map[string]any{"ret": 0, "errcode": 0, "message_id": "sent-1"})
		case "/ilink/bot/getconfig":
			decodeJSON(t, r, &getConfigBody)
			writeJSON(t, w, map[string]any{"ret": 0, "errcode": 0, "typing_ticket": "ticket-1"})
		case "/ilink/bot/sendtyping":
			decodeJSON(t, r, &sendTypingBody)
			writeJSON(t, w, map[string]any{"ret": 0, "errcode": 0})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	a := New(config.WeixinBotConfig{
		AccountID: "bot-account",
		TokenEnv:  "WEIXIN_BOT_TOKEN",
		APIBase:   server.URL,
	}, slog.Default()).(*adapter)
	a.msgCh = make(chan bot.InboundMessage, 1)

	updates, err := a.getUpdates(context.Background())
	if err != nil {
		t.Fatalf("getUpdates: %v", err)
	}
	if len(updates) != 0 {
		t.Fatalf("updates = %d, want 0 direct update records", len(updates))
	}
	msg := <-a.msgCh
	if msg.ChatID != "user-1" || msg.UserID != "user-1" || msg.Text != "hello" {
		t.Fatalf("inbound = %#v", msg)
	}
	if getUpdatesBody["get_updates_buf"] != "" {
		t.Fatalf("get_updates_buf = %#v, want blank", getUpdatesBody["get_updates_buf"])
	}
	if _, ok := getUpdatesBody["base_info"].(map[string]any); !ok {
		t.Fatalf("getupdates missing base_info: %#v", getUpdatesBody)
	}

	res, err := a.sendMessage(context.Background(), bot.OutboundMessage{ChatID: "user-1", Text: "reply"})
	if err != nil {
		t.Fatalf("sendMessage: %v", err)
	}
	if res.MessageID != "sent-1" {
		t.Fatalf("message id = %q", res.MessageID)
	}
	msgPayload := sendBody["msg"].(map[string]any)
	if msgPayload["to_user_id"] != "user-1" || msgPayload["context_token"] != "ctx-1" {
		t.Fatalf("send msg payload = %#v", msgPayload)
	}
	if msgPayload["message_type"] != float64(2) || msgPayload["message_state"] != float64(2) {
		t.Fatalf("send message type/state = %#v", msgPayload)
	}
	if _, ok := sendBody["base_info"].(map[string]any); !ok {
		t.Fatalf("send missing base_info: %#v", sendBody)
	}

	if err := a.sendTyping(context.Background(), "user-1"); err != nil {
		t.Fatalf("sendTyping: %v", err)
	}
	if getConfigBody["context_token"] != "ctx-1" {
		t.Fatalf("getconfig context token = %#v", getConfigBody)
	}
	if sendTypingBody["typing_ticket"] != "ticket-1" || sendTypingBody["context_token"] != "ctx-1" {
		t.Fatalf("sendtyping body = %#v", sendTypingBody)
	}
}

func TestStartLoginPostsLocalTokenListAndPollLoginSavesConfirmedAccount(t *testing.T) {
	oldTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = oldTransport }()
	tempHome := t.TempDir()
	t.Setenv("APPDATA", tempHome)
	t.Setenv("XDG_CONFIG_HOME", tempHome)

	var qrBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ilink/bot/get_bot_qrcode":
			decodeJSON(t, r, &qrBody)
			writeJSON(t, w, map[string]any{"qrcode": "qr-1", "qrcode_img_content": "https://qr.example/1"})
		case "/ilink/bot/get_qrcode_status":
			if r.URL.Query().Get("qrcode") != "qr-1" {
				t.Fatalf("qrcode query = %q", r.URL.RawQuery)
			}
			writeJSON(t, w, map[string]any{
				"status":        "confirmed",
				"ilink_bot_id":  "bot-account",
				"bot_token":     "bot-token",
				"ilink_user_id": "scanner",
				"baseurl":       serverBaseURL(r),
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	http.DefaultTransport = rewriteWeixinHostTransport{base: server.URL, rt: oldTransport}
	session, err := StartLogin(context.Background())
	if err != nil {
		t.Fatalf("StartLogin: %v", err)
	}
	if session.QRCode != "qr-1" || session.QRCodeURL == "" {
		t.Fatalf("session = %#v", session)
	}
	if _, ok := qrBody["local_token_list"].([]any); !ok {
		t.Fatalf("qr body missing local_token_list: %#v", qrBody)
	}

	result, status, err := PollLogin(context.Background(), session)
	if err != nil {
		t.Fatalf("PollLogin: %v", err)
	}
	if status != "confirmed" || result == nil || result.AccountID != "bot-account" || result.Token != "bot-token" {
		t.Fatalf("result=%#v status=%q", result, status)
	}
	if !HasSavedAccount("bot-account") {
		t.Fatal("confirmed account was not saved")
	}
}

func decodeJSON(t *testing.T, r *http.Request, out any) {
	t.Helper()
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatal(err)
	}
}

type rewriteWeixinHostTransport struct {
	base string
	rt   http.RoundTripper
}

func (t rewriteWeixinHostTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.EqualFold(req.URL.Host, "ilinkai.weixin.qq.com") {
		baseReq, _ := http.NewRequest(req.Method, t.base, nil)
		req.URL.Scheme = baseReq.URL.Scheme
		req.URL.Host = baseReq.URL.Host
	}
	return t.rt.RoundTrip(req)
}

func serverBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

var _ = os.ErrNotExist
