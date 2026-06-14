package weixin

import (
	"encoding/json"
	"testing"
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
