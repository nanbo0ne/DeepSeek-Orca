package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/desktop/computeruse"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/boot"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/config"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/localai"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/provider"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/tool"
)

const computerControlMaxSteps = 40

type computerTaskTool struct {
	app   *App
	tabID string
}
type computerStatusTool struct{ app *App }
type computerStopTool struct{ app *App }

func (a *App) computerTools(tabID string) []tool.Tool {
	return []tool.Tool{computerTaskTool{app: a, tabID: tabID}, computerStatusTool{app: a}, computerStopTool{app: a}}
}

func (computerTaskTool) Name() string { return "computer_task" }
func (computerTaskTool) Description() string {
	return "Control Windows to complete a concrete desktop task. Provide the goal, an observable success condition, and any restrictions. O.R.C.A. uses an isolated visual control agent and re-observes after every action."
}
func (computerTaskTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","required":["goal"],"properties":{"goal":{"type":"string"},"success_criteria":{"type":"string"},"restrictions":{"type":"string"}},"additionalProperties":false}`)
}
func (computerTaskTool) ReadOnly() bool { return false }
func (t computerTaskTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var input struct {
		Goal            string `json:"goal"`
		SuccessCriteria string `json:"success_criteria"`
		Restrictions    string `json:"restrictions"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return "", err
	}
	return t.app.runComputerTask(ctx, computeruse.StartRequest{TabID: t.tabID, Goal: strings.TrimSpace(input.Goal), SuccessCriteria: strings.TrimSpace(input.SuccessCriteria), Restrictions: strings.TrimSpace(input.Restrictions)})
}

func (computerStatusTool) Name() string { return "computer_status" }
func (computerStatusTool) Description() string {
	return "Return the current Windows computer-control session state and recent action summary."
}
func (computerStatusTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{},"additionalProperties":false}`)
}
func (computerStatusTool) ReadOnly() bool { return true }
func (t computerStatusTool) Execute(context.Context, json.RawMessage) (string, error) {
	body, err := json.Marshal(t.app.GetComputerUseState())
	return string(body), err
}

func (computerStopTool) Name() string { return "computer_stop" }
func (computerStopTool) Description() string {
	return "Immediately stop the active Windows computer-control session and release injected input."
}
func (computerStopTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"reason":{"type":"string"}},"additionalProperties":false}`)
}
func (computerStopTool) ReadOnly() bool { return false }
func (t computerStopTool) Execute(context.Context, json.RawMessage) (string, error) {
	return "Computer control stopped.", t.app.StopComputerUse()
}

var computerControlSchemas = []provider.ToolSchema{
	{Name: "computer_action", Description: "Execute exactly one action against the current observation. Element IDs and coordinates are valid only for the current generation.", Parameters: json.RawMessage(`{"type":"object","required":["type"],"properties":{"type":{"type":"string","enum":["click","double_click","right_click","hover","mouse_down","mouse_up","drag","scroll","key","key_combo","type_text","activate_window","minimize_window","maximize_window","restore_window","close_window","move_window","resize_window","invoke","toggle","select","expand","collapse","set_value","wait"]},"elementId":{"type":"string"},"displayId":{"type":"string"},"x":{"type":"number","minimum":0,"maximum":1},"y":{"type":"number","minimum":0,"maximum":1},"endX":{"type":"number","minimum":0,"maximum":1},"endY":{"type":"number","minimum":0,"maximum":1},"deltaX":{"type":"integer"},"deltaY":{"type":"integer"},"text":{"type":"string"},"key":{"type":"string"},"keys":{"type":"array","items":{"type":"string"}},"windowId":{"type":"string"},"timeoutMs":{"type":"integer","minimum":0,"maximum":30000},"description":{"type":"string"}},"additionalProperties":false}`)},
	{Name: "computer_complete", Description: "Finish only after the observable success condition is satisfied.", Parameters: json.RawMessage(`{"type":"object","required":["summary"],"properties":{"summary":{"type":"string"}},"additionalProperties":false}`)},
	{Name: "computer_escalate", Description: "Return control to the main model when the screen is ambiguous, strategy is needed, or a protected surface blocks progress.", Parameters: json.RawMessage(`{"type":"object","required":["reason"],"properties":{"reason":{"type":"string"},"options":{"type":"array","items":{"type":"string"}}},"additionalProperties":false}`)},
}

func (a *App) runComputerTask(ctx context.Context, request computeruse.StartRequest) (result string, retErr error) {
	if a.computerUse == nil {
		return "", computeruse.ErrNotSupported
	}
	cfg, err := config.LoadForRoot(a.activeWorkspaceRoot())
	if err != nil {
		return "", err
	}
	if !cfg.Desktop.ComputerUseFullAccess || cfg.Desktop.ComputerUseConsent != computerUseConsentVersion {
		return "", fmt.Errorf("computer use full access has not been approved")
	}
	modelRef, err := a.resolveComputerControlModel(cfg, request.ModelRef)
	if err != nil {
		return "", err
	}
	request.ModelRef = modelRef
	entry, err := a.computerProviderEntry(ctx, cfg, modelRef)
	if err != nil {
		return "", err
	}
	prov, err := boot.NewProviderWithProxy(entry, cfg.NetworkProxySpec())
	if err != nil {
		return "", err
	}
	if _, err = a.computerUse.Start(ctx, request); err != nil {
		return "", err
	}
	defer func() {
		if retErr != nil {
			_ = a.computerUse.Stop(retErr.Error())
		}
	}()
	observation, err := a.computerUse.Observe(ctx)
	if err != nil {
		return "", err
	}
	messages := []provider.Message{{Role: provider.RoleSystem, Content: computerControlSystemPrompt(request)}}
	for step := 0; step < computerControlMaxSteps; step++ {
		if err = ctx.Err(); err != nil {
			return "", err
		}
		messages = append(messages, observationMessage(observation, step))
		assistant, calls, err := collectComputerControlResponse(ctx, prov, messages)
		if err != nil {
			return "", err
		}
		if len(calls) == 0 {
			return "", fmt.Errorf("computer control model returned no structured action")
		}
		messages = append(messages, provider.Message{Role: provider.RoleAssistant, Content: assistant, ToolCalls: calls})
		for _, call := range calls {
			switch call.Name {
			case "computer_complete":
				var done struct {
					Summary string `json:"summary"`
				}
				if err = json.Unmarshal([]byte(call.Arguments), &done); err != nil {
					return "", err
				}
				summary := strings.TrimSpace(done.Summary)
				if summary == "" {
					summary = "Computer task completed."
				}
				a.computerUse.Complete(true, summary)
				return summary, nil
			case "computer_escalate":
				var escalation struct {
					Reason  string   `json:"reason"`
					Options []string `json:"options"`
				}
				_ = json.Unmarshal([]byte(call.Arguments), &escalation)
				_, _ = a.computerUse.Pause()
				payload, _ := json.Marshal(escalation)
				return "Computer control paused for main-model judgment: " + string(payload), nil
			case "computer_action":
				var action computeruse.Action
				if err = json.Unmarshal([]byte(call.Arguments), &action); err != nil {
					return "", err
				}
				action.Generation = observation.Generation
				actionResult, actionErr := a.computerUse.Execute(ctx, action)
				toolResult := map[string]any{"success": actionErr == nil, "error": "", "generation": actionResult.Observation.Generation, "summary": actionResult.Observation.Summary}
				if actionErr != nil {
					toolResult["error"] = actionErr.Error()
				} else {
					observation = actionResult.Observation
				}
				encoded, _ := json.Marshal(toolResult)
				messages = append(messages, provider.Message{Role: provider.RoleTool, ToolCallID: call.ID, Name: call.Name, Content: string(encoded)})
			default:
				return "", fmt.Errorf("computer control model requested unknown action %q", call.Name)
			}
		}
	}
	return "", fmt.Errorf("computer control reached the %d-action safety limit", computerControlMaxSteps)
}

func (a *App) computerProviderEntry(ctx context.Context, cfg *config.Config, modelRef string) (*config.ProviderEntry, error) {
	entry, ok := cfg.ResolveModel(modelRef)
	if !ok {
		return nil, fmt.Errorf("unknown computer control model %q", modelRef)
	}
	if entry.Name != localai.ProviderID {
		return entry, nil
	}
	runtimeProviders, err := a.prepareLocalRuntimeProviders(ctx, cfg, modelRef)
	if err != nil {
		return nil, err
	}
	for i := range runtimeProviders {
		if runtimeProviders[i].Name == entry.Name && runtimeProviders[i].Model == entry.Model {
			return &runtimeProviders[i], nil
		}
	}
	return nil, fmt.Errorf("local computer control model did not start")
}

func computerControlSystemPrompt(request computeruse.StartRequest) string {
	return fmt.Sprintf(`You are Orca's isolated Windows control agent. Complete the task through one structured action at a time.
Goal: %s
Success condition: %s
Restrictions: %s

Rules:
- Use UI Automation element IDs when available; otherwise use normalized coordinates in the current crop.
- Every action invalidates the current observation. Never reuse an element ID or coordinate without reading the next observation.
- Never interact with passwords, verification codes, CAPTCHA, UAC, lock screen, payment credentials, or higher-integrity windows.
- Call computer_complete only after observing the success condition.
- Call computer_escalate when choices are ambiguous, strategy is required, or the system blocks access.
- Do not describe an action in prose instead of calling a tool.`, request.Goal, request.SuccessCriteria, request.Restrictions)
}

func observationMessage(observation computeruse.Observation, step int) provider.Message {
	return provider.Message{Role: provider.RoleUser, Content: fmt.Sprintf("Observation %d, generation %d.\n%s", step+1, observation.Generation, observation.Summary), Images: []provider.ImageContent{{Name: "orca-computer-observation.jpg", MediaType: observation.ScreenshotMIME, Data: observation.Screenshot}}}
}

func collectComputerControlResponse(ctx context.Context, prov provider.Provider, messages []provider.Message) (string, []provider.ToolCall, error) {
	stream, err := prov.Stream(ctx, provider.Request{Messages: messages, Tools: computerControlSchemas, Temperature: 0, MaxTokens: 1200})
	if err != nil {
		return "", nil, err
	}
	var text strings.Builder
	var calls []provider.ToolCall
	for chunk := range stream {
		switch chunk.Type {
		case provider.ChunkText:
			text.WriteString(chunk.Text)
		case provider.ChunkToolCall:
			if chunk.ToolCall != nil {
				calls = append(calls, *chunk.ToolCall)
			}
		case provider.ChunkError:
			if chunk.Err != nil {
				return text.String(), calls, chunk.Err
			}
		}
	}
	return strings.TrimSpace(text.String()), calls, nil
}
