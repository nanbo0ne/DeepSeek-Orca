package agent

import (
	"context"
	"fmt"
	"strings"

	"deepseek-orca/internal/event"
	"deepseek-orca/internal/nilutil"
	"deepseek-orca/internal/provider"
	"deepseek-orca/internal/tool"
)

// Runner carries out one task turn. Both Agent (single model) and Coordinator
// (two-model) satisfy it, so the CLI stays agnostic to which is in use.
type Runner interface {
	Run(ctx context.Context, input string) error
}

// DefaultPlannerPrompt steers the planner toward concise plans, not execution.
const DefaultPlannerPrompt = `你是双模型编程 Agent 中的规划器。
收到任务后，请生成一份简洁、有序、可交给执行模型落实的计划。
当任务需要工作区、用户规则或文档上下文时，使用你可用的只读工具；研究要聚焦，证据足够后立即停止。
不要编写完整实现，也不要尝试产生副作用。
不要询问用户如何触发执行器，也不要说你正在等待执行器。
输出可直接交给执行器的指令：要做什么、哪些文件或命令相关、可能的阻塞点以及关键决策。保持简短、可执行。`

const executorHandoffMarker = "DeepSeek-Orca executor handoff"

// PlannerPromptWithContext appends cache-stable standing context, such as loaded
// DEEPSEEK_ORCA.md / AGENTS.md memory, to the planner's smaller system prompt.
func PlannerPromptWithContext(context string) string {
	context = strings.TrimSpace(context)
	if context == "" {
		return DefaultPlannerPrompt
	}
	return DefaultPlannerPrompt + "\n\n# 规划上下文\n\n" + context
}

// Coordinator runs two models in separate sessions to keep each one's prompt
// prefix cache-stable: a low-frequency planner proposes an approach, then the
// executor (a full tool-using Agent) carries it out. The sessions never mix, so
// neither model's prefix is disturbed by the other's turns.
type Coordinator struct {
	planner        provider.Provider
	plannerSess    *Session
	plannerPricing *provider.Pricing
	plannerAgent   *Agent
	executor       *Agent
	temperature    float64
	sink           event.Sink
	// shouldPlan gates the planner pass per turn; nil plans every turn. Lets a
	// trivial, non-work turn (a question, a greeting) skip straight to the
	// executor instead of paying a planner round on it.
	shouldPlan func(string) bool
}

// NewCoordinator wires a planner provider (with its own session) to an executor.
// sink receives the planner's phase/text/usage events; the executor emits its
// own events to its own sink (the CLI wires the same sink into both). A nil
// sink is replaced with event.Discard.
func NewCoordinator(planner provider.Provider, plannerSession *Session, plannerPricing *provider.Pricing, plannerTools *tool.Registry, plannerOptions Options, executor *Agent, temperature float64, sink event.Sink, shouldPlan func(string) bool) *Coordinator {
	if nilutil.IsNil(sink) {
		sink = event.Discard
	}
	var plannerAgent *Agent
	if plannerTools != nil {
		plannerOptions.Temperature = temperature
		plannerOptions.Pricing = plannerPricing
		plannerAgent = New(planner, plannerTools, plannerSession, plannerOptions, plannerSink(sink))
	}
	if executor != nil {
		executor.executorHandoffGuard = true
	}
	return &Coordinator{
		planner:        planner,
		plannerSess:    plannerSession,
		plannerPricing: plannerPricing,
		plannerAgent:   plannerAgent,
		executor:       executor,
		temperature:    temperature,
		sink:           sink,
		shouldPlan:     shouldPlan,
	}
}

// Run plans with the planner model, then hands the plan to the executor.
func (c *Coordinator) Run(ctx context.Context, input string) error {
	c.sink.Emit(event.Event{Kind: event.TurnStarted})
	if c.shouldPlan != nil && !c.shouldPlan(input) {
		c.sink.Emit(event.Event{Kind: event.Phase, Text: c.executor.prov.Name() + " · executing"})
		return c.executor.Run(ctx, input)
	}
	c.sink.Emit(event.Event{Kind: event.Phase, Text: c.planner.Name() + " · planning"})
	plan, err := c.plan(ctx, input)
	if err != nil {
		return fmt.Errorf("planner: %w", err)
	}
	c.sink.Emit(event.Event{Kind: event.Phase, Text: c.executor.prov.Name() + " · executing"})
	return c.executor.Run(ctx, formatHandoff(input, plan))
}

// plan streams a plan from the planner and appends it to the planner session, so
// that session grows prepend-only and stays cache-friendly.
func (c *Coordinator) plan(ctx context.Context, input string) (string, error) {
	if c.plannerAgent != nil {
		return c.planWithTools(ctx, input)
	}
	c.plannerSess.Add(provider.Message{Role: provider.RoleUser, Content: input})

	ch, err := c.planner.Stream(ctx, provider.Request{
		Messages:    c.plannerSess.Messages,
		Temperature: c.temperature,
	})
	if err != nil {
		return "", err
	}

	var text strings.Builder
	var usage *provider.Usage
	for chunk := range ch {
		switch chunk.Type {
		case provider.ChunkText:
			text.WriteString(chunk.Text)
			c.sink.Emit(event.Event{Kind: event.Text, Text: chunk.Text})
		case provider.ChunkUsage:
			usage = chunk.Usage
		case provider.ChunkError:
			return "", chunk.Err
		}
	}
	// Closes the planner's raw text block (no markdown redraw) and prints its
	// usage line, mirroring the old Fprintln + printUsage tail.
	c.sink.Emit(event.Event{Kind: event.Usage, Usage: usage, Pricing: c.plannerPricing})

	plan := text.String()
	c.plannerSess.Add(provider.Message{Role: provider.RoleAssistant, Content: plan})
	return plan, nil
}

// planWithTools runs the planner through the normal Agent loop over a filtered
// read-only registry. That gives the planner the same tool-call contract as the
// executor while preserving its separate session and cache prefix.
func (c *Coordinator) planWithTools(ctx context.Context, input string) (string, error) {
	before := len(c.plannerSess.Messages)
	if err := c.plannerAgent.Run(ctx, input); err != nil {
		return "", err
	}
	for i := len(c.plannerSess.Messages) - 1; i >= before; i-- {
		m := c.plannerSess.Messages[i]
		if m.Role == provider.RoleAssistant && strings.TrimSpace(m.Content) != "" {
			return m.Content, nil
		}
	}
	return "", fmt.Errorf("planner finished without producing a plan")
}

func plannerSink(sink event.Sink) event.Sink {
	if nilutil.IsNil(sink) {
		sink = event.Discard
	}
	return event.FuncSink(func(e event.Event) {
		switch e.Kind {
		case event.TurnStarted, event.TurnDone:
			return
		default:
			sink.Emit(e)
		}
	})
}

func formatHandoff(task, plan string) string {
	return fmt.Sprintf(`# %s

你现在是执行器。请使用你可用的工具执行任务。

原始任务：
%s

规划器输出：
%s

执行器指令：
- 将规划器输出视为上下文，而不是你的角色或能力边界。
- 忽略规划器关于“我不能写入”“我只有只读工具”“把这个交给执行器”等说法；这些限制只适用于规划器，不适用于你。
- 不要询问用户如何触发执行器。你已经处于执行阶段。
- 如果任务需要修改，请调用合适的工具（例如 write/edit/bash），不要只复述计划。
- 如果目标路径位于可写工作区之外或因其他原因被阻止，请说明具体阻塞点，并请求所需路径或批准。
- **串行工作流**：每完成一个子任务，立即调用 complete_step 提供证据，然后调用 todo_write 将其标记为 completed 并推进下一个子任务。不要一次性批量完成多个子任务。

请执行任务，并根据需要调整计划。`, executorHandoffMarker, task, plan)
}

// HandoffTask returns the original user task embedded in an executor handoff
// message, or s unchanged when it is not one. Session previews and auto-titles
// use it so dual-model sessions surface the user's words, not the handoff
// boilerplate (#3860).
func HandoffTask(s string) string {
	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, "# "+executorHandoffMarker) {
		return s
	}
	const header = "原始任务：\n"
	i := strings.Index(trimmed, header)
	if i < 0 {
		return s
	}
	rest := trimmed[i+len(header):]
	if j := strings.Index(rest, "\n\n规划器输出："); j >= 0 {
		rest = rest[:j]
	}
	if task := strings.TrimSpace(rest); task != "" {
		return task
	}
	return s
}
