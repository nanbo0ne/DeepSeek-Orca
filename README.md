# DeepSeek-Orca Desktop

DeepSeek-Orca is a local desktop engineering workspace for using AI models on real projects. It combines multi-provider chat, project-scoped sessions, inspectable tools, planning, memory, MCP, skills, CodeGraph, and reversible execution in one Windows, macOS, and Linux application.

This repository's desktop product is the **engineering edition**. It exposes two user-facing prompt modes: **Normal** and **Enhanced**. The future Orca personal-assistant edition is kept behind an internal product boundary and is documented separately in [docs/ORCA_ASSISTANT_APP_HANDOFF.md](docs/ORCA_ASSISTANT_APP_HANDOFF.md).

## Download

- [Windows installer](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.30/DeepSeek-Orca-windows-amd64-installer.exe)
- [Windows portable package](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.30/DeepSeek-Orca-windows-amd64.zip)
- [macOS universal DMG](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.30/DeepSeek-Orca-darwin-universal.dmg)
- [Linux amd64 DEB](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.30/DeepSeek-Orca-linux-amd64.deb)
- [All release assets](https://github.com/nanbo0ne/DeepSeek-Orca/releases/tag/desktop-v2.0.30)

The application never downloads or installs an update without user action. Update detection only displays a small release-page entry when a stable `desktop-v*` release is available. The Windows installer can create a desktop shortcut and can optionally launch the application after installation.

## What It Provides

### Engineering modes

- **Normal mode**: general conversation, research, writing, analysis, everyday coding, and regular tasks. Its complete built-in engineering instructions stay active even when custom instructions are configured, and file-changing work requires a relevant verification after the last write.
- **Enhanced mode**: complex coding, architecture work, long-running tasks, reviews, repository changes, and more agentic execution.

The mode is stored per conversation. Switching mode rebuilds only the controller while preserving session history, model, effort, approvals, plan state, goal state, and session path. Both modes identify the product as `DeepSeek-Orca` to the model.

### Providers and models

Use DeepSeek, OpenAI-compatible providers, and Anthropic-compatible providers. Configure endpoints, API keys, model lists, context windows, reasoning effort, pricing metadata, planner models, and optional executor/planner separation from Settings. Balance information loads asynchronously so a slow provider does not block opening a conversation.

### Workspaces and sessions

- Project workspaces bind conversations to real repository directories.
- Independent topics receive isolated working directories and session data.
- Sessions support tabs, pinning, renaming, previews, restore, forks, checkpoints, rollback, trash, and export.
- Long sessions can be compacted while preserving a durable summary and searchable local history.
- Main transcript JSONL stays local and does not contain provider secrets.

### Vision attachments

Vision under Settings > Models offers **Off**, **Auto**, and **On**. New installations default to Auto: each configured model is checked with one small, isolated request, and image bytes are sent only to models confirmed to support vision. Capability is tracked per provider type, endpoint, and model, with status and manual recheck controls in Settings. DeepSeek being marked unsupported is expected and does not affect text use.

Pasted, dropped, or workspace-referenced PNG, JPEG, WebP, and GIF files remain limited to 8 images per turn, 20 MB total, and 10 MB per image. A text-only main model receives references rather than image bytes and may explicitly delegate selected current-turn images to a confirmed vision-capable subagent. Sessions store paths, names, and MIME types, never image base64; an unsupported or unknown target returns a clear error instead of pretending it saw the image. Screenshot automation and full computer-use are outside this edition.

### Tools and local execution

The Tool Library controls groups for web search, host operations, Node/Python runtimes, document inspection, thread utilities, and conversation search. A disabled group is removed from both the registry and the model-visible routing policy. The built-in shell remains available for builds, tests, Git, and package managers. Tool approval modes make permission boundaries visible before execution; sandbox and write-root settings further constrain local operations.

### MCP, Skills, and CodeGraph

- MCP settings manage servers, authorization, connection status, retries, and exposed tools.
- Skills provide reusable workflows that load on demand.
- CodeGraph exposes project symbol and call-relationship tools through the configured MCP server.
- The slash menu combines built-in commands, skills, and MCP prompts.

### Memory

Normal and Enhanced use the `shared-agent` memory profile. Existing memory documents and tool-based `remember` / `forget` workflows remain available. The hidden assistant profile is not read, written, generated, or processed by the engineering edition. Assistant data is retained on disk for the future standalone Orca application and is not deleted during migration.

### Planning and process visibility

- Plan mode presents a reviewable plan before execution.
- Todo uses a compact centered progress control that expands upward on hover or keyboard focus and can be pinned without pushing the composer.
- Goal mode can continue a multi-step task and stops on its internal complete or blocked signal.
- Process display offers Compact and Detailed levels, with Compact as the default. Reasoning, tools, notices, and subagents use a flat chronological activity rail instead of nested cards; compact rows can be repeatedly expanded and collapsed.
- An optional, off-by-default theme-blue spinner can appear below the current turn. It rotates clockwise while the model is active and counterclockwise while tools are running, respects reduced-motion settings, and disappears when work pauses or ends.
- A successfully completed turn automatically collapses to its user message, elapsed-time row, and final answer. Expanding the elapsed row restores every intermediate event in its original order and shows token usage; failed, cancelled, interrupted, and active turns stay open for diagnosis.
- The transcript preserves the actual order of assistant text, reasoning, tool calls, tool results, notices, images, and compaction.
- Conversation switches reuse cached transcript structure, paint history before auxiliary status, and skip replaying entrance animations for restored messages. Plain-text paste remains directly editable, and the composer grows automatically for up to ten lines before scrolling.

### Automation and remote connections

The application includes scheduled automation, task status, local background jobs, and optional bot channel integrations. Background jobs do not silently start a new model turn after completion; work that depends on a result must explicitly collect it. Remote bot configuration is kept separate from the active desktop session and uses the same engineering prompt-mode boundary.

## Privacy and Safety

DeepSeek-Orca is a local desktop shell, but model requests are sent to the provider selected by the user. Read the active provider, endpoint, proxy, vision, tool, and approval settings before using sensitive data. Image bytes are transmitted only when the selected vision mode permits the target model and a turn includes the image. Vision probing itself makes one very small request to the configured provider. API keys are read from configured environment variables or the local credentials flow rather than written into conversation messages.

Tool execution is visible in the transcript. Approval, sandbox, workspace roots, read-only tools, and background-task status are separate controls. Cancel, pause, rollback, checkpoints, and session trash provide recovery paths for long tasks.

## Configuration

After first launch, open Settings and configure models and providers, workspace and sandbox behavior, Tool Library and MCP servers, vision, process display, update checks, language, appearance, permissions, and approval defaults. The app exposes the active configuration path in Settings. Project-local instructions can be supplied through the supported instruction files in the workspace.

## Troubleshooting

- **The model is unavailable**: check the provider endpoint, selected model, API-key environment variable, and proxy settings.
- **A conversation opens slowly**: balance lookup is independent; inspect provider/network status and local workspace size if history itself is slow.
- **An image is rejected**: inspect the model's vision status in Settings. In Auto, choose a confirmed vision-capable model or subagent; DeepSeek being unsupported is normal. On forces an attempt but cannot add capability to a text-only model.
- **A tool asks for approval repeatedly**: review the approval mode, sandbox, and tool-group settings rather than retrying blindly.
- **A window is narrow**: keep the application above its supported minimum width. Composer controls progressively hide labels while keeping model selection and core actions accessible.
- **An update is not shown**: automatic checks are cached for 24 hours and use official stable GitHub Releases. Manual checking is available next to the update toggle in Settings.

## Build From Source

Requirements: a Go toolchain compatible with `go.mod`, Node.js/npm, and Wails CLI v2.

```powershell
cd desktop/frontend
npm install
npm run test:all
npm run build

cd ../..
go test ./...
```

Generate Wails bindings after changing exported `desktop.App` methods:

```powershell
cd desktop
wails generate module
wails build
```

Release-specific procedures and signing configuration are intentionally kept outside this stable product overview.
