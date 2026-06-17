# DeepSeek-Orca V2.0.0

DeepSeek-Orca is a Windows desktop and CLI coding agent based on the Reasonix fork, rebuilt around DeepSeek and OpenAI-compatible providers. It keeps the core agent loop, tools, MCP, skills, memory, permission control, session resume, checkpoints, context compression, and rollback features, while adding a V2 desktop workflow focused on higher-quality long-context coding work.

## Download

Windows installer:

[DeepSeek-Orca-Setup-2.0.0-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.0/DeepSeek-Orca-Setup-2.0.0-windows-amd64.exe)

After installation, open DeepSeek-Orca Desktop and add a DeepSeek API key or an OpenAI-compatible provider in Settings.

## V2 Highlights

### Enhanced Mode

V2 adds an Enhanced Mode button beside the send button. It switches the session to a separate Claude-like prompt and context assembly profile:

- The stable system prompt is kept separate from dynamic project memory for better cache hygiene.
- Project memory is injected as a `<system-reminder>` user-message block instead of being folded into the system prompt.
- Memory is refreshed from disk each turn, so edits to project guidance can affect the current session sooner.
- The existing memory file chain remains supported: `DEEPSEEK_ORCA.md`, `AGENTS.md`, `CLAUDE.md`, and their local variants.

Enhanced Mode is intended for complex coding tasks, long-running refactors, careful reviews, and conversations where answer quality matters more than the smallest possible token footprint. It can use more tokens and may reduce prompt-cache hits. If a conversation is already over 50,000 tokens, DeepSeek-Orca warns before switching because mid-conversation mode changes may miss the cache and may not let the model perform at its best.

### New Conversation Preference Inheritance

New conversations now inherit the model, thinking effort, tool approval level, and Enhanced Mode from the most recent conversation that actually ran. This keeps the common workflow smooth: choose the model and approval strength once, then keep moving.

Temporary collaboration options do not carry over. Ask workflow, Step Thinking, Plan Mode, and Goal Mode start off for every new conversation so a previous planning-heavy task does not surprise the next one.

### Ask Workflow

The new Ask switch in the composer menu enables a more deliberate planning workflow inspired by `grill-me-codex`:

- Explore the repository first instead of asking questions that can be answered from code.
- Ask only user-owned decisions, one concise question at a time.
- Lock a plan before implementation.
- Run an internal adversarial review using the existing planner/subagent capabilities when useful.
- Ask for final confirmation before risky edits.

This is useful when the request has product tradeoffs, unclear acceptance criteria, or a high blast radius.

### Step Thinking

The Step Thinking switch enables a staged workflow inspired by `superpowers`:

- Explore relevant context.
- Brainstorm viable approaches.
- Choose/design the implementation.
- Produce an implementation plan.
- Execute focused tasks.
- Review each task and then perform a final review.

When Ask workflow and Step Thinking are both enabled, Step Thinking skips its brainstorm phase to avoid duplicate planning loops.

## Desktop Features

- Project and independent-workspace conversations in the left sidebar.
- Pinned conversations, history, trash, and topic management.
- Model, thinking effort, and approval controls directly in the composer.
- Ask, Step Thinking, Plan, Goal, and Enhanced Mode controls for different task styles.
- File, image, workspace, command, and past-chat references from the composer.
- Tool approval modes: ask, auto, and yolo/full access.
- Checkpoints, conversation rewind, and file rollback with hash checks.
- Context panel with window usage, token totals, cache hit rate, elapsed time, request count, and cost.
- Persistent telemetry so token/cost statistics survive restart.
- Manual and automatic CONTEXT CHECKPOINT compaction.
- MCP, local skills, memory files, CodeGraph support, and slash commands.
- QQ/WeChat bot integration from the desktop settings panel.

## CLI

Build from source:

```powershell
git clone https://github.com/nanbo0ne/DeepSeek-Orca.git
cd DeepSeek-Orca
go build -o bin/deepseek-orca.exe ./cmd/deepseek-orca
```

Configure a key and start chat:

```powershell
$env:DEEPSEEK_API_KEY="your DeepSeek API Key"
.\bin\deepseek-orca.exe setup

cd D:\your-project
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe chat
```

Run a one-shot task:

```powershell
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe run "Read this project and summarize the main modules"
```

## Installed Files

Typical Windows install location:

```text
C:\Users\<your-user>\AppData\Local\Programs\DeepSeek-Orca
```

Main files:

- `deepseek-orca-desktop.exe`: desktop application.
- `uninstall.exe`: generated uninstaller.
- `uninstall.bat`: fallback uninstall script.
- `node.exe`: bundled Node runtime.
- `dist/`: desktop frontend assets.
- `.deepseek-orca/`: config, credentials references, skills, MCP, cache, and local data.
- `data/`: conversations, history indexes, workspace metadata, and telemetry.

DeepSeek-Orca is based on the Reasonix fork. License: MIT.
