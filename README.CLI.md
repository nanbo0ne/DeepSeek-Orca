# DeepSeek-Orca CLI V2.0.9

DeepSeek-Orca CLI is the terminal entry point for the DeepSeek-Orca coding agent. It keeps the core Reasonix-derived agent loop, tools, MCP, skills, memory, permission control, session resume, rollback, and compaction features.

The Windows desktop installer is the recommended package for most users:

[DeepSeek-Orca-Setup-2.0.9-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.9/DeepSeek-Orca-Setup-2.0.9-windows-amd64.exe)

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
