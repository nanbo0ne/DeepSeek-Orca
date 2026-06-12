package memory

import (
	"context"
	"encoding/json"
	"fmt"

	"deepseek-orca/internal/tool"
)

// forgetTool deletes a saved memory the model judges wrong or stale. Like
// rememberTool it is stateful (bound to one project's Store), so boot constructs
// it and adds it to the registry.
type forgetTool struct{ store Store }

// NewForgetTool returns the `forget` tool bound to store.
func NewForgetTool(store Store) tool.Tool { return forgetTool{store: store} }

func (forgetTool) Name() string { return "forget" }

func (forgetTool) Description() string {
	return "当一条已保存记忆错误、过期或被替代时，按名称删除它，使其不再加载到未来会话中。" +
		"请使用记忆索引中的 slug，也就是 \"[label](<name>.md)\" 里的 \"<name>\"。" +
		"优先用 `remember` 更新记忆（复用原名称），不要先 forget 再重建；只有当该事实应彻底不存在时才使用 forget。"
}

func (forgetTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "要删除的记忆 slug，即索引中 \"[label](<name>.md)\" 里的 \"<name>\"。"}
		},
		"required": ["name"]
	}`)
}

func (t forgetTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var in struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	if in.Name == "" {
		return "", fmt.Errorf("name is required")
	}
	if err := t.store.Delete(in.Name); err != nil {
		return "", err
	}
	if q, ok := QueueFromContext(ctx); ok {
		q.QueueMemory("Deleted memory \"" + slug(in.Name) + "\" — disregard its line still shown in the saved-memories index until next session.")
	}
	return fmt.Sprintf("已忘记记忆 %q（它不再适用，也不会加载到未来会话中）。", in.Name), nil
}

func (forgetTool) ReadOnly() bool { return false }
