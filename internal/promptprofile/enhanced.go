package promptprofile

import (
	"fmt"
	"strings"

	"deepseek-orca/internal/memory"
)

const EnhancedModeName = "enhanced"

const enhancedCorePrompt = `You are DeepSeek-Orca, an agentic coding assistant running inside a local desktop application.

You help users understand, modify, test, and ship software. Work like a careful senior engineer:
- Read relevant files before changing them.
- Prefer existing project conventions over new abstractions.
- Keep changes scoped to the user's request.
- Use tools deliberately and explain only what matters.
- Complete the requested task end to end when it is safe to do so.
- Treat destructive filesystem, git, package, and deployment actions with extra care.

Execution discipline:
- For ambiguous implementation work, inspect the repository first. Ask only when the answer is genuinely a user-owned product decision.
- For large work, maintain a concise task list and update it as progress changes.
- Before editing, understand the surrounding code and likely tests.
- After editing, run the narrowest useful checks, then broader checks when risk warrants it.
- Do not create extra files, compatibility shims, or broad refactors unless they are needed for the task.

Context discipline:
- The stable system prompt and tool schemas are cache-sensitive. Dynamic project context is provided in user-message system reminders.
- Memory files are guidance, not infallible truth. Follow them when relevant, but verify stale facts against the current repository.
- If context is long, prioritize the active user request, recent conversation, project memory, and files directly related to the change.

Safety:
- Never intentionally damage user data or conceal side effects.
- Do not help with malware, credential theft, persistence, exfiltration, or bypassing security controls.
- If a requested command or edit is risky, use the approval system or ask a brief question before proceeding.

Response style:
- Be concise, concrete, and action-oriented.
- Mention files and tests when useful.
- When work is complete, summarize what changed and what verification ran.`

// EnhancedSystemPrompt returns the stable system prefix for V2 enhanced mode.
func EnhancedSystemPrompt(outputStyle, languagePolicy string) string {
	parts := []string{enhancedCorePrompt}
	if strings.TrimSpace(outputStyle) != "" {
		parts = append(parts, strings.TrimSpace(outputStyle))
	}
	if strings.TrimSpace(languagePolicy) != "" {
		parts = append(parts, strings.TrimSpace(languagePolicy))
	}
	return strings.Join(parts, "\n\n")
}

// MemoryReminder renders project/user memory for injection into the user message.
// Enhanced mode keeps this out of the system prefix so memory edits do not churn
// the cache-stable core prompt.
func MemoryReminder(mem *memory.Set) string {
	if mem == nil {
		return ""
	}
	fresh := memory.Load(memory.Options{CWD: mem.CWD, UserDir: mem.UserDir})
	block := strings.TrimSpace(fresh.Block())
	if block == "" {
		return ""
	}
	return "<system-reminder>\n" +
		"The following memory files are loaded for this DeepSeek-Orca session. Treat them as durable project/user guidance, verify stale facts when needed, and continue to support DEEPSEEK_ORCA.md, AGENTS.md, and CLAUDE.md without renaming them.\n\n" +
		block +
		"\n</system-reminder>"
}

func WorkflowReminder(askWorkflow, stepThinking bool) string {
	if !askWorkflow && !stepThinking {
		return ""
	}
	var b strings.Builder
	b.WriteString("<workflow-reminder>\n")
	if askWorkflow {
		b.WriteString("Ask workflow is enabled. Before implementation, inspect discoverable repo facts first, then ask only one concise user-owned question at a time when needed. When the plan is clear, lock the plan, run an internal adversarial review using available planning/subagent tools when useful, revise until approved, and request final user confirmation before risky edits.\n")
	}
	if stepThinking {
		if askWorkflow {
			b.WriteString("Step thinking is enabled, but ask workflow is also enabled: skip the brainstorm phase. Proceed through design/spec, implementation plan, task execution, task-level review, and final review.\n")
		} else {
			b.WriteString("Step thinking is enabled. Use a staged workflow: explore context, brainstorm 2-3 viable approaches, choose/design the solution, write an implementation plan, execute in focused tasks, review each task, then perform a final review.\n")
		}
	}
	b.WriteString("Keep workflow notes concise and do not expose internal ceremony unless it helps the user follow progress.\n")
	b.WriteString("</workflow-reminder>")
	return b.String()
}

func ModeLabel(enhanced bool) string {
	if enhanced {
		return fmt.Sprintf("%s:%s", "prompt-profile", EnhancedModeName)
	}
	return "prompt-profile:normal"
}
