# DeepSeek-Orca Desktop V2.0.22

DeepSeek-Orca Desktop is the primary Windows experience for DeepSeek-Orca. It brings project conversations, file changes, rollback, context statistics, model settings, MCP, skills, memory, and bot connections into one desktop workspace.

Release notes are no longer kept in README files. They are stored in [DESKTOP_CHANGELOG.md](DESKTOP_CHANGELOG.md).

## Install

Download the Windows installer:

[DeepSeek-Orca-Setup-2.0.22-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.22/DeepSeek-Orca-Setup-2.0.22-windows-amd64.exe)

## What It Is For

DeepSeek-Orca Desktop is designed for local development, long-running AI collaboration, and desktop-first agent workflows:

- Read, edit, test, and roll back code in a project workspace.
- Manage conversations by project, independent workspace, pinned topic, history, and trash.
- Extend the agent through MCP, skills, memory files, CodeGraph, and slash commands.
- Control tool permissions through approval modes.
- Use host tools for web search, local commands, process management, runtime checks, and document extraction.
- Recover older intent after context compression through long-term conversation search.

## Prompt Modes

The composer mode menu provides three prompt profiles:

- Assistant mode: general help, daily questions, light explanation, and casual support. It runs as `Orca` and reads assistant memory only.
- Normal mode: regular conversations, search, ordinary development work, and lighter engineering collaboration. It runs as `DeepSeek-Orca`.
- Enhanced mode: complex coding, long tasks, higher-quality reasoning, architecture work, review, and agentic coding. It runs as `DeepSeek-Orca` and may use more tokens.

Normal and Enhanced modes read all memory but write to the shared agent memory profile. Assistant mode reads and writes only assistant memory.

## Workspaces And Conversations

DeepSeek-Orca supports both project workspaces and independent workspaces:

- Project workspaces are bound to real code folders and are best for long-lived repository work.
- Independent workspaces are useful for temporary chats, experiments, and tasks that should not touch a project folder.
- Each independent conversation gets its own small workspace root, so files, attachments, config, memory, tool cwd, and session data do not leak across independent conversations.
- New conversations inherit the last active conversation's model, thinking effort, approval level, and prompt mode. Temporary collaboration options reset by default.

## Host Tool Library

The host tool library is enabled by default and can be managed from the Tool Library panel. Disabled tool groups are removed from both the registered tool schema and the model-visible routing policy.

Main tool groups include:

- System and host tools: native commands, system information, process listing, process termination, app launch, clipboard, and notifications.
- Web search: `web_search` for unknown URLs and `web_fetch` for known URLs.
- Node / Python runtimes: script validation, data processing, dependency checks, and structured output.
- Document tools: basic inspection and text extraction for Word, PowerPoint, Excel, and PDF files.
- Thread management: local DeepSeek-Orca session and topic inspection.
- Long-term conversation search: search older local conversations after compaction, then read fuller nearby context by locator.

Screenshot recognition, OCR, coordinate clicking, keyboard input, and visual desktop automation are intentionally not included yet.

## Approval Modes

DeepSeek-Orca uses approval modes to control tool behavior:

- Ask: risky or confirmation-worthy tool operations ask first.
- Auto approve: ordinary tool permission prompts are approved automatically.
- Full access: tool operations can run without extra approval gates.

Use full access only when the task boundary, workspace, and risk are clear.

## Planning, Todo, And Side Chat

- Complex tasks can use automatic Todo tracking to split work into visible steps and update status as execution progresses.
- Plan mode presents a dedicated plan proposal card. You can approve the plan or request a complete replacement plan before execution.
- Side chat is a read-only right-dock conversation that can reference the main transcript, with recent turns prioritized. It does not write to the main history or participate in main token statistics, compaction, or title generation.

## Automation

Automation is reserved for clearly recurring, continuous, or background-monitoring tasks, such as periodic build checks, daily reminders, or log monitoring.

One-off timer-style requests are not persisted as automations. Automation records are stored locally and can be viewed, paused, resumed, cancelled, or cleared after completion.

## Getting Started

1. Run the Windows installer.
2. Start DeepSeek-Orca Desktop.
3. Add a DeepSeek API key or OpenAI-compatible provider in Settings.
4. Create a new conversation from the left sidebar.
5. Choose a project folder or an independent workspace.
6. Select Assistant, Normal, or Enhanced mode based on the task.

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
