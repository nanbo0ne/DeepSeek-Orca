# DeepSeek-Orca Desktop V2.0.17

DeepSeek-Orca Desktop is the primary Windows experience for DeepSeek-Orca. It brings project conversations, file changes, rollback, context statistics, model settings, MCP, skills, memory, and bot connections into one desktop workspace.

## Install

Download the Windows installer:

[DeepSeek-Orca-Setup-2.0.17-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.17/DeepSeek-Orca-Setup-2.0.17-windows-amd64.exe)

## V2.0.17 Tool Library Update

V2.0.17 adds a Tool Library button next to Automation in the top bar. It manages the newer extended host-tool groups: thread management, web search, Node/Python REPL, document tools, system/host tools, long-term conversation search, and the proactive tool-use steering switch.

Disabled tool groups are removed from both the registered tool schema and the model-visible routing policy. Bash remains available for development shell tasks, but it no longer advertises disabled dedicated tools in its own description.

The new read-only `conversation_search` and `conversation_read` tools let the model search older local transcripts after context compression, then read fuller nearby context by locator.

## V2.0.16 Interaction And Automation Update

V2.0.16 changes `web_search` to avoid DuckDuckGo. The desktop host search tool now tries China-accessible search sources first: Bing China, then Baidu as fallback. Known URLs should still be read directly with `web_fetch`.

Normal mode and enhanced mode now share a clean English evidence-first tool policy. When an answer depends on current facts, external information, repository state, file contents, command output, runtime behavior, or the local environment, the model is explicitly steered to use the appropriate tool before answering instead of guessing from memory.

V2.0.16 completes the automatic Todo, plan-mode, and independent workspace isolation work. Normal mode and enhanced mode now share the same short English task-tracking policy, so complex multi-step work can trigger `todo_write` consistently without localizing tool schemas or the enhanced core prompt.

Plan mode now renders a dedicated rounded plan proposal card. Revision requests include the previous complete plan plus the user's requested changes and ask the model for a complete replacement plan while staying in plan mode. Once approved, the final plan seeds the Todo panel and execution continues through the existing `todo_write` / `complete_step` flow.

Standalone conversations are isolated more strictly. Each independent topic gets its own workspace root, session directory, attachment area, memory/config scope, and tool cwd. Older global sessions are backfilled into per-topic session directories when first opened, without copying the old shared workspace contents.

V2.0.16 also fixes the bottom status bar approval indicator. Ask, auto approve, and full access now remain visible on light backgrounds and use distinct colors so the current permission mode can be checked at a glance.

V2.0.16 focuses on smoother interaction and safer automation semantics.

Model switching now updates the composer button and status bar immediately with the short model name, such as `deepseek-v4-pro`, while controller rebuild happens in the background. The UI should no longer flash a raw `provider/model` path during that short transition.

Automation is now reserved for explicit recurring, continuous, or background-monitoring tasks. The persistent automation manager is available from the top-left toolbar next to the Bot button. It lists local automations and supports pause, resume, cancel, clear finished, and refresh. One-off timer-style requests are no longer represented as persisted automations.

The right dock adds a fourth tab: Side chat. It is a read-only side conversation that can reference the main conversation, with recent turns prioritized. Side chat history is stored per main session and can be cleared without changing the main transcript, token statistics, compaction, or title generation.

Slash menu descriptions and browser-demo capability text are more consistently localized in Chinese when using the Chinese UI.

## Host Tool Library

The host tool library is available by default, without adding another composer toggle. The model can now use native host commands, system information, process listing and termination, app launch, text clipboard access, desktop notifications, lightweight web search, recurring automations, Orca session listing, Node/Python execution, and basic Word/PPT/Excel/PDF inspection or extraction tools.

Screenshot recognition, OCR, coordinate clicking, keyboard input, and visual desktop automation are intentionally not included yet because the current DeepSeek model path does not provide reliable image understanding.

Approval semantics are clearer in this version: ask mode prompts, auto mode automatically approves tool permission prompts, and yolo mode is full access for tool operations. Use yolo only when you truly want the agent to run any tool action without additional permission gates.

The composer layout was also changed so the bottom mode row is a real layout region instead of being hidden behind the right-side buttons. Long input text should no longer cover the current approval/model/effort row.

Then:

1. Run the installer.
2. Start DeepSeek-Orca Desktop.
3. Add a DeepSeek API key or OpenAI-compatible provider in Settings.
4. Create a new conversation from the left sidebar.
5. Choose a project folder or an independent workspace.

## V2 Desktop Workflow

V2 adds three important workflow controls directly around the composer.

Enhanced Mode sits next to the send button. It uses a separate Claude-like prompt/context profile for complex coding tasks. Memory files are injected as dynamic `<system-reminder>` context instead of being folded into the stable system prompt, and the existing memory chain remains compatible: `DEEPSEEK_ORCA.md`, `AGENTS.md`, `CLAUDE.md`, and local variants.

The small question-mark icon beside Enhanced Mode explains the tradeoff: better answer quality can cost more tokens and lower cache hit rates. If the current conversation is already above 50,000 tokens, switching mode asks for confirmation before rebuilding the controller.

The Ask switch in the plus menu enables a stricter clarification workflow: inspect first, ask only what cannot be discovered locally, lock a plan, run an internal review when useful, and wait for final confirmation before risky edits.

The Step Thinking switch enables a staged workflow for large tasks: explore, brainstorm, design, plan, execute, and review. When Ask is also enabled, Step Thinking skips brainstorming so the two workflows do not duplicate each other.

New conversations inherit the last active conversation's model, thinking effort, approval level, and Enhanced Mode. Temporary plus-menu choices such as Ask, Step Thinking, Plan Mode, and Goal Mode are intentionally reset to off.

## Everyday Features

- Sidebar conversation management by project, independent workspace, pinned topics, history, and trash.
- Composer controls for model, thinking effort, approval level, Enhanced Mode, Ask, Step Thinking, Plan, and Goal.
- File, image, workspace, slash-command, and past-chat references.
- Ask/auto/yolo approval modes.
- Checkpoints, rewind, and file rollback.
- Context panel with token usage, cache hits, request counts, elapsed time, and cost.
- Persistent telemetry across restarts.
- MCP servers, local skills, memory files, CodeGraph, and slash commands.
- QQ/WeChat bot integration.

## Installed Files

Typical install location:

```text
C:\Users\<your-user>\AppData\Local\Programs\DeepSeek-Orca
```

Main files:

- `deepseek-orca-desktop.exe`: desktop executable.
- `uninstall.exe`: generated uninstaller.
- `uninstall.bat`: fallback uninstall script.
- `node.exe`: bundled Node runtime.
- `dist/`: frontend assets.
- `.deepseek-orca/`: config, skill, MCP, and cache data.
- `data/`: conversations, history, indexes, workspace metadata, and telemetry.

License: MIT.
