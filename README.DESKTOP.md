# DeepSeek-Orca Desktop V2.0.7

DeepSeek-Orca Desktop is the primary Windows experience for DeepSeek-Orca. It brings project conversations, file changes, rollback, context statistics, model settings, MCP, skills, memory, and bot connections into one desktop workspace.

## Install

Download the Windows installer:

[DeepSeek-Orca-Setup-2.0.7-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.7/DeepSeek-Orca-Setup-2.0.7-windows-amd64.exe)

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
