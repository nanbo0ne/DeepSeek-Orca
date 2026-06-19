# DeepSeek-Orca CLI V2.0.17

DeepSeek-Orca CLI is the terminal entry point for the DeepSeek-Orca coding agent. It keeps the core Reasonix-derived agent loop, tools, MCP, skills, memory, permission control, session resume, rollback, and compaction features.

The Windows desktop installer is the recommended package for most users:

[DeepSeek-Orca-Setup-2.0.17-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.17/DeepSeek-Orca-Setup-2.0.17-windows-amd64.exe)

## V2.0.17 Notes

V2.0.17 adds configurable Tool Library groups for the desktop host-tool layer. The model-visible tool routing policy is now generated from those settings, and the new read-only conversation search tools can retrieve older local transcript details after context compression.

## V2.0.16 Notes

V2.0.16 changes `web_search` to avoid DuckDuckGo. Host web search now tries China-accessible sources first: Bing China, then Baidu as fallback. Known URLs should still be read with `web_fetch`.

Normal mode and enhanced mode now share a clean English evidence-first tool policy. When an answer depends on current facts, external information, repository state, file contents, command output, runtime behavior, or the local environment, the model is explicitly steered to use the appropriate tool before answering instead of guessing from memory.

V2.0.16 completes the shared automatic Todo policy for both normal and enhanced prompt profiles. The added model-visible policy is intentionally short and English-only: complex multi-step tasks should use `todo_write`, while simple one-step work should not create ceremonial lists.

Desktop plan mode now uses a dedicated plan proposal card and revised-plan flow, and standalone desktop conversations now get per-topic independent workspace roots and session directories. These are desktop-first changes, but they preserve the shared `todo_write` / `complete_step` behavior used by the core agent.

V2.0.16 includes a desktop status-bar polish fix: the approval mode indicator now uses distinct visible colors for ask, auto approve, and full access.

V2.0.16 keeps the shared host tool library and tightens automation semantics: persisted automations are intended for explicit recurring, continuous, or background-monitoring tasks. The desktop app also adds an automation manager and a read-only side chat panel; those are desktop-first UI features.

The shared host tools include native host commands, process/system helpers, notifications, web search, recurring automations, Orca session listing, Node/Python execution, and basic document extraction helpers. Visual desktop control and screenshot recognition are not included in this release.

## Build From Source

```powershell
git clone https://github.com/nanbo0ne/DeepSeek-Orca.git
cd DeepSeek-Orca
go build -o bin/deepseek-orca.exe ./cmd/deepseek-orca
```

## Configure

```powershell
$env:DEEPSEEK_API_KEY="your DeepSeek API Key"
.\bin\deepseek-orca.exe setup
```

You can also configure OpenAI-compatible providers in `deepseek-orca.toml`.

## Use

Start an interactive project chat:

```powershell
cd D:\your-project
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe chat
```

Run a one-shot task:

```powershell
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe run "Read this project and summarize the main modules"
```

## Common Commands

- `deepseek-orca setup`: create or update user config.
- `deepseek-orca chat`: start the interactive TUI.
- `deepseek-orca run "task"`: run a one-shot task.
- `deepseek-orca bot start`: start the bot gateway.
- `deepseek-orca bot doctor`: inspect bot configuration.
- `/init`: create or update project instructions.
- `/plan`: toggle Plan Mode.
- `/skill`: manage skills.
- `/mcp`: manage MCP servers.
- `/model`: switch model.
- `/resume`: resume a saved session.
- `/rewind`: rewind conversation context or associated code changes.
- `/compact`: run CONTEXT CHECKPOINT compaction.

## Notes

V2's Enhanced Mode, Ask workflow, Step Thinking controls, and new-conversation preference inheritance are desktop-first features. The CLI continues to use the established terminal workflow and the same memory file conventions: `DEEPSEEK_ORCA.md`, `AGENTS.md`, and `CLAUDE.md`.

License: MIT.
