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
	return "Delete a saved memory by name when it is wrong, stale, or should no longer be loaded in future sessions. " +
		"Use the slug from the memory index: the <name> in \"[label](<name>.md)\". " +
		"Prefer updating an existing memory with remember using the same name; use forget only when the fact should be removed entirely."
}

func (forgetTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "Memory slug to delete: the <name> in \"[label](<name>.md)\" from the loaded memory index."}
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
		q.QueueMemory("Deleted memory \"" + slug(in.Name) + "\": disregard its line still shown in the saved-memories index until next session.")
	}
	return fmt.Sprintf("Deleted memory %q; it no longer applies and will not load in future sessions.", in.Name), nil
}

func (forgetTool) ReadOnly() bool { return false }
