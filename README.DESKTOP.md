# DeepSeek-Orca Desktop

[中文](README.zh-CN.md)

DeepSeek-Orca is a local desktop AI assistant and engineering agent. It combines multi-provider models, project workspaces, tools, MCP, skills, CodeGraph, long-term memory, planning, and automation while keeping execution inspectable, pausable, and reversible.

## Downloads

- [Windows installer](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.27/DeepSeek-Orca-windows-amd64-installer.exe)
- [Windows portable](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.27/DeepSeek-Orca-windows-amd64.zip)
- [macOS universal DMG](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.27/DeepSeek-Orca-darwin-universal.dmg)
- [Linux amd64 DEB](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.27/DeepSeek-Orca-linux-amd64.deb)
- [All release assets](https://github.com/nanbo0ne/DeepSeek-Orca/releases/tag/desktop-v2.0.27)

Final Windows applications and installers are Authenticode-signed through SignPath Foundation. `Get-AuthenticodeSignature` should report `Valid`; see [SIGNING.md](SIGNING.md) for the complete policy. During SignPath enrollment, a clearly labelled temporary unsigned build may be available and will still show an unknown publisher; its assets will be replaced in place after signing is ready.

Configure an API key, endpoint, and model under Settings > Models after first launch. Multimodal vision is opt-in; images are uploaded to the configured provider only when it is enabled.

## Prompt Modes

- **Assistant** runs as `Orca` for everyday questions, light tasks, and personalized assistance. It can maintain a private Assistant memory profile and recall relevant preferences.
- **Normal** runs as `DeepSeek-Orca` for general conversation, research, writing, analysis, and ordinary development.
- **Enhanced** runs as `DeepSeek-Orca` for complex coding, architecture, long tasks, reviews, and more agentic engineering work.

Prompt mode is stored per conversation. Switching modes rebuilds that controller without dropping history, model, effort, approval, plan, or goal state.

## Models And Providers

DeepSeek-Orca supports DeepSeek, OpenAI-compatible providers, and Anthropic-native providers. The main executor and optional planner can use different providers and models. Model choice, context window, reasoning effort, and pricing metadata are configurable. Balance lookup is asynchronous and does not block conversation loading.

## Workspaces And Conversations

- Project workspaces bind conversations to real directories for persistent repository work.
- Independent workspaces isolate files, attachments, configuration, memory scope, tool cwd, and session data per topic.
- Conversations support pinning, renaming, previews, restore, forks, trash, checkpoints, and rollback.
- Automatic compaction protects long sessions; conversation search can locate and read older local context on demand.

## Multimodal Vision

When enabled, the composer accepts pasted or dropped PNG, JPEG, WebP, and GIF files as well as workspace image references. A turn can contain up to 8 images, 20 MB total, with a 10 MB per-image limit. Session JSONL stores paths, names, and MIME types rather than base64. Unsupported models report an explicit error. Screenshot understanding and full computer-use are not included yet.

## Tool Library

The Tool Library controls groups for web search, host/system operations, Node and Python REPLs, document inspection, local thread management, and long-term conversation search. Disabled groups disappear from both the tool registry and model-visible routing policy. `bash` remains the fallback for builds, tests, Git, and package managers.

## MCP, Skills, And CodeGraph

- MCP management covers servers, authorization, connection status, retries, and tool enablement.
- Agent skills load specialized workflows only when needed.
- CodeGraph exposes real `mcp__codegraph__...` tools for symbols, call relationships, and architectural context.
- The slash menu combines built-in commands, skills, and MCP prompts.

## Memory

Assistant mode reads and writes the Assistant profile. Normal and Enhanced modes read all memories and write shared-agent memory. Assistant auto-memory runs silently when leaving a conversation, never blocks the main turn or app exit, and gives up after five failed retries for the same batch. Users can disable auto-memory or recall and can delete individual or all Assistant memories.

## Plan, Todo, Goal, And Process Display

Plan mode presents a complete proposal before execution. Todo stays as a compact progress bar, expands upward on hover or keyboard focus, and can be pinned open without moving the composer. Goal mode can continue autonomously and stops when its internal complete or blocked marker is received.

## Updates And Installation

The app checks official GitHub Releases after startup and every 24 hours by default. When a newer stable `desktop-v*` release exists, a low-profile download icon opens that exact release in the browser. DeepSeek-Orca never downloads, installs, or forces an update automatically. Settings provides an opt-out and a manual check action.

The Windows installer lets users choose whether to create a desktop shortcut and whether to launch DeepSeek-Orca on completion. Upgrade installs preserve the existing desktop shortcut choice, and silent installs never launch the app.

Process display has three modes:

- **Compact** keeps each process span in a collapsed, single-line white rounded row.
- **Standard** shows live process cards in chronological order and folds completed spans in place.
- **Detailed** preserves chronology and opens process details by default.

Assistant text, reasoning, tools, image reads, and compaction are interleaved by event order. Completion actions appear only after `TurnDone`. Scroll anchoring keeps the user's historical reading position stable while output grows above it.

## Automation, Bots, And Side Chat

Automation is intended for recurring jobs, monitors, and reminders. Bot channels can choose a model and prompt mode. Side Chat can reference the main transcript without writing to main history or affecting its tokens, compaction, or title.

Background bash and subagent jobs remain visible while running. If the current answer depends on a background result, the model must call `wait`; job completion only updates job state and does not create a new model turn.

## Permissions And Privacy

Tool approval modes are Ask, Auto approve, and Full access. File and host operations remain subject to workspace, sandbox, and permission rules. Conversations, memories, indexes, settings, and workspace metadata are local by default. Model requests and opt-in images are sent to the provider configured by the user.

## Troubleshooting

- Vision errors: enable vision and verify that the current model/provider accepts image messages.
- Missing tools: check Tool Library groups, MCP connections, enabled skills, and prompt mode.
- A completed background job does not resume the model automatically: use `wait` when the active answer depends on it.
- Slow Git views in large repositories: open Files/Changes only when needed and reduce unrelated untracked files.

## Development

Go, Node.js, npm, Wails v2, and NSIS for Windows installers are required.

```powershell
git clone https://github.com/nanbo0ne/DeepSeek-Orca.git
cd DeepSeek-Orca
npm install --prefix desktop/frontend
npm run test:all --prefix desktop/frontend
npm run build --prefix desktop/frontend
go test ./...
cd desktop
wails build
```

Desktop configuration lives in `desktop/wails.json`; release automation lives in `.github/workflows/release-desktop.yml`.

License: MIT.
