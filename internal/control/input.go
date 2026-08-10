package control

import (
	"context"
	"regexp"
	"strings"

	"deepseek-orca/internal/promptprofile"
	"deepseek-orca/internal/skill"
)

var reComposeBlock = regexp.MustCompile(`(?s)^\s*(?:<host_context\b[^>]*>.*?</host_context>|<(?:memory-update|background-jobs|system-reminder|workflow-reminder|automation-context)>.*?</(?:memory-update|background-jobs|system-reminder|workflow-reminder|automation-context)>)\s*(?:\n|$)`)

var reservedHostTag = regexp.MustCompile(`(?i)</?host_context\b`)

func escapeReservedHostTags(text string) string {
	return reservedHostTag.ReplaceAllStringFunc(text, func(tag string) string {
		return strings.Replace(tag, "<", "&lt;", 1)
	})
}

// PlanModeMarker is prepended to every user turn while plan mode is on. It rides
// in the user message (not the system prompt or tools), so the cache-stable
// prompt prefix is left untouched and the toggle costs nothing in cache hits.
const PlanModeMarker = "[Plan 模式：只读。请先探索代码库（可用 read_file、ls、grep、glob、web_fetch、task；写入工具会被宿主拒绝），然后以回复形式给出分层计划并停止；不要写文件、编辑文件或运行有副作用的 bash。计划必须是两层 markdown 列表：每个阶段使用顶层编号列表项（一个连贯里程碑，例如 \"1. 添加配置加载器\"），每个阶段下方用缩进 bullet 写出具体、可验证的子步骤（例如 \"   - 将 TOML 解析进 Config\"）。阶段必须使用普通编号列表项，不要写成 markdown 标题（##、###），这样两层都能被解析。阶段数量保持少量（约 2-6 个）。任何修改发生前都会请求用户批准。]"

const (
	activeGoalOpen  = "<active-goal>"
	activeGoalClose = "</active-goal>"
)

const (
	GoalStatusRunning  = "running"
	GoalStatusComplete = "complete"
	GoalStatusBlocked  = "blocked"
	GoalStatusStopped  = "stopped"
)

// StripComposePrefixes removes controller-injected prefixes from a composed
// user message so that the display text matches what the user actually typed.
// It strips the PlanModeMarker, <memory-update>…</memory-update>, and
// <background-jobs>…</background-jobs> blocks that Compose prepends to user
// turns. This is used as a fallback when no .display.json sidecar recording
// exists (e.g. sessions created before the display-recording feature, or
// synthetic user messages injected by the controller).
func StripComposePrefixes(content string) string {
	s := content
	for {
		next := reComposeBlock.ReplaceAllStringFunc(s, func(match string) string {
			return ""
		})
		if next == s {
			break
		}
		s = next
	}
	s = strings.TrimPrefix(s, PlanModeMarker+"\n\n")
	s = strings.TrimPrefix(s, PlanModeMarker)
	s = strings.TrimSpace(s)
	return s
}

// IsSyntheticUserMessage returns true if the content matches one of the known
// synthetic user messages injected by the controller or agent loop (plan
// approval, stream recovery, readiness retry, etc.). These should not be shown
// in the chat UI.
func IsSyntheticUserMessage(content string) bool {
	trimmed := strings.TrimSpace(content)
	if trimmed == planApprovedMessage {
		return true
	}
	for _, prefix := range syntheticPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}

// syntheticPrefixes must be kept in sync with the synthetic user messages
// injected by the controller (planApprovedMessage), agent loop
// (streamRecoveryMessage, finalReadinessRetryMessage, emptyFinalRetryMessage,
// executorHandoffRetryMessage in internal/agent/agent.go), and compaction
// folds (internal/agent/compact.go), which store checkpoints as user-role
// messages the chat UI must never render as user bubbles (#3653).
var syntheticPrefixes = []string{
	"Plan approved — plan mode is off",
	"计划已批准，Plan 模式已关闭",
	"Host final-answer readiness check failed",
	"宿主的最终回复就绪检查未通过",
	"You are already in the executor phase",
	"你已经处于执行阶段",
	"The previous assistant response was interrupted while a tool call",
	"The previous assistant response was interrupted during streaming",
	"The previous assistant response was interrupted before visible",
	"The previous assistant response finished without any visible answer",
	"上一条助手回复在工具调用流式输出期间被中断",
	"上一条助手回复在流式输出期间被中断",
	"上一条助手回复在完成可见答案文本前被流式中断",
	"上一条助手回复结束时没有任何可见答案文本",
	"<context-checkpoint>",
	"<compaction-summary>",
	"Summary of the later conversation (compacted from here on):",
	"Summary of earlier conversation (compacted up to here):",
	"[Mid-turn steer queued by the user.",
	"[用户已排队一条中途引导。",
}

// Compose applies the plan-mode marker to a turn's text when plan mode is on,
// returning the message to actually send to the model. The frontend keeps
// showing the raw text as the user bubble.
func (c *Controller) Compose(text string) string {
	c.mu.Lock()
	plan := c.planMode
	goal := c.goal
	goalStatus := c.goalStatus
	notes := c.pendingMemory
	enhanced := c.enhancedMode
	memoryReminder := c.memoryReminder
	askWorkflow := c.askWorkflow
	stepThinking := c.stepThinking
	mem := c.mem
	turnContext := c.turnContext
	c.pendingMemory = nil
	c.mu.Unlock()

	// The model can distinguish genuine host context only because raw user text
	// cannot introduce the reserved wrapper verbatim.
	text = escapeReservedHostTags(text)

	if turnContext != nil {
		if contextBlock := strings.TrimSpace(turnContext()); contextBlock != "" {
			text = `<host_context type="automation" trust="host">` + "\n" + contextBlock + "\n</host_context>\n\n" + text
		}
	}

	if reminder := promptprofile.WorkflowReminder(askWorkflow, stepThinking); reminder != "" {
		text = reminder + "\n\n" + text
	}
	if enhanced || memoryReminder {
		if reminder := promptprofile.MemoryReminder(mem); reminder != "" {
			text = reminder + "\n\n" + text
		}
	}

	if strings.TrimSpace(goal) != "" && goalStatus == GoalStatusRunning {
		text = activeGoalBlock(goal) + "\n\n" + text
	}
	if plan {
		text = PlanModeMarker + "\n\n" + text
	}

	// Memory added mid-session rides the turn (never the cached system prefix),
	// so it takes effect now without invalidating the prompt cache. It folds into
	// the system prefix on the next session, where it costs nothing per turn.
	if len(notes) > 0 {
		var b strings.Builder
		b.WriteString("<memory-update>\n")
		b.WriteString("以下项目记忆变更刚刚完成，并从现在开始生效：\n")
		for _, n := range notes {
			b.WriteString("- " + n + "\n")
		}
		b.WriteString("</memory-update>\n\n")
		text = b.String() + text
	}

	// Background jobs that finished since the last turn ride the turn too, so the
	// model learns of completions even though the user-facing notices don't reach
	// its context. Like memory, this never touches the cache-stable prefix.
	if c.jobs != nil {
		if note := c.jobs.DrainCompletedNote(); note != "" {
			text = "<background-jobs>\n" + note + "\n</background-jobs>\n\n" + text
		}
	}
	return text
}

func activeGoalBlock(goal string) string {
	goal = strings.TrimSpace(goal)
	goal = strings.ReplaceAll(goal, activeGoalClose, "<\\/active-goal>")
	var b strings.Builder
	b.WriteString(activeGoalOpen)
	b.WriteString("\n")
	b.WriteString(goal)
	b.WriteString("\n\n")
	b.WriteString("Goal 模式：请自主推进这个目标。跨轮持续工作，直到目标完成。优先采用合理默认值，不要轻易询问用户；只有在真正被用户拥有的决策阻塞时才使用 ask。不要在描述计划后停止，而要执行下一个有用步骤。每条 Goal 模式助手回复的最后一行必须且只能包含一个状态标记：[goal:continue]、[goal:complete] 或 [goal:blocked:<short reason>]。")
	b.WriteString("\n")
	b.WriteString(activeGoalClose)
	return b.String()
}

// MemoryQuickAddNote parses the legacy "# <note>" memory shortcut. The space
// after "#" is intentional: "#7", "#issue", and "#标题" are ordinary user
// prompts, not memory writes.
func MemoryQuickAddNote(input string) (note string, ok bool) {
	trimmed := strings.TrimSpace(input)
	if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "#\t") {
		return strings.TrimSpace(trimmed[1:]), true
	}
	return "", false
}

// RememberCommandNote parses the explicit "/remember <note>" memory command.
func RememberCommandNote(input string) (note string, ok bool) {
	trimmed := strings.TrimSpace(input)
	switch {
	case trimmed == "/remember":
		return "", true
	case strings.HasPrefix(trimmed, "/remember ") || strings.HasPrefix(trimmed, "/remember\t"):
		return strings.TrimSpace(trimmed[len("/remember"):]), true
	default:
		return "", false
	}
}

type GoalCommandAction int

const (
	GoalCommandStatus GoalCommandAction = iota + 1
	GoalCommandSet
	GoalCommandClear
)

type GoalCommand struct {
	Action GoalCommandAction
	Text   string
}

func ParseGoalCommand(input string) (GoalCommand, bool) {
	trimmed := strings.TrimSpace(input)
	if trimmed != "/goal" && !strings.HasPrefix(trimmed, "/goal ") && !strings.HasPrefix(trimmed, "/goal\t") {
		return GoalCommand{}, false
	}
	args := strings.TrimSpace(trimmed[len("/goal"):])
	switch strings.ToLower(args) {
	case "", "status":
		return GoalCommand{Action: GoalCommandStatus}, true
	case "clear", "off", "stop", "done":
		return GoalCommand{Action: GoalCommandClear}, true
	default:
		return GoalCommand{Action: GoalCommandSet, Text: args}, true
	}
}

// CustomCommand resolves a "/name args…" line against the loaded custom slash
// commands, returning the rendered prompt to send (found=false when no command
// matches). It does not apply the plan-mode marker — call Compose for that.
func (c *Controller) CustomCommand(input string) (sent string, found bool) {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return "", false
	}
	name := strings.TrimPrefix(fields[0], "/")
	for _, cmd := range c.commands {
		if cmd.Name == name {
			return cmd.Render(fields[1:]), true
		}
	}
	return "", false
}

// RunSkill resolves a "/<name> args…" line against the loaded skills, returning
// the skill's rendered body to send as a turn (found=false when no skill
// matches). Invoking a skill by slash always inlines its body — the model reads
// and follows the playbook in the main loop; a subagent skill's isolation is
// only engaged when the model calls it via run_skill / the dedicated tool. The
// caller applies Compose for plan-mode/memory framing.
func (c *Controller) RunSkill(input string) (sent string, found bool) {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return "", false
	}
	name := strings.TrimPrefix(fields[0], "/")
	if sk, ok := c.skillByName(name); ok {
		return skill.Render(sk, strings.Join(fields[1:], " ")), true
	}
	return "", false
}

func (c *Controller) skillByName(name string) (skill.Skill, bool) {
	if c.skillStore != nil {
		return c.skillStore.Read(name)
	}
	for _, sk := range c.skills {
		if sk.Name == name {
			return sk, true
		}
	}
	return skill.Skill{}, false
}

// MCPPrompt resolves a "/mcp__server__prompt args…" line: it maps the positional
// args onto the prompt's declared arguments and fetches the rendered prompt from
// the MCP server (an async prompts/get). found is false when no such prompt
// exists; err carries a fetch failure. Honours ctx.
func (c *Controller) MCPPrompt(ctx context.Context, input string) (sent string, found bool, err error) {
	if c.host == nil {
		return "", false, nil
	}
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return "", false, nil
	}
	name := strings.TrimPrefix(fields[0], "/")

	prompts := c.host.Prompts()
	idx := -1
	for i := range prompts {
		if prompts[i].Name == name {
			idx = i
			break
		}
	}
	if idx < 0 {
		return "", false, nil
	}

	args := map[string]string{}
	for i, a := range prompts[idx].Args {
		if i+1 < len(fields) {
			args[a.Name] = fields[i+1]
		}
	}
	text, err := prompts[idx].Get(ctx, args)
	if err != nil {
		return "", true, err
	}
	return text, true, nil
}
