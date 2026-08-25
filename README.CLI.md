# O.R.C.A CLI

O.R.C.A CLI is the terminal entry point for **Open Reasoning & Computing Agent**. It shares the desktop agent core: model providers, tools, MCP, skills, memory, permissions, session resume, rollback, and context compaction.

The Windows desktop installer is the recommended package for most users:

[O.R.C.A for Windows 3.0.1](https://github.com/nanbo0ne/O.R.C.A-for-Windows/releases/tag/desktop-v3.0.1)

## Build From Source

```powershell
git clone https://github.com/nanbo0ne/O.R.C.A-for-Windows.git
cd O.R.C.A-for-Windows
go build -o bin/orca.exe ./cmd/orca
```

## Configure

```powershell
$env:ORCA_API_KEY="your provider key"
.\bin\orca.exe setup
```

The setup command writes canonical O.R.C.A user configuration. Project configuration is `orca.toml`; provider-specific endpoint and credential settings remain isolated. DeepSeek, OpenAI-compatible services, and other catalog providers are optional choices rather than startup requirements.

## Use

Start an interactive project chat:

```powershell
cd D:\your-project
D:\path\to\O.R.C.A-for-Windows\bin\orca.exe chat
```

Run a one-shot task:

```powershell
D:\path\to\O.R.C.A-for-Windows\bin\orca.exe run "Read this project and summarize the main modules"
```

## Common Commands

- `orca setup`: create or update user configuration.
- `orca chat`: start the interactive TUI.
- `orca run "task"`: run a one-shot task.
- `orca bot start`: start the bot gateway.
- `orca bot doctor`: inspect bot configuration.
- `/init`: create or update `ORCA.md` project instructions.
- `/plan`: toggle Plan Mode.
- `/skill`: manage skills.
- `/mcp`: manage MCP servers.
- `/model`: switch model.
- `/resume`: resume a saved session.
- `/rewind`: rewind conversation context or associated code changes.
- `/compact`: run context compaction.

The desktop app additionally provides Assistant, Coding, Orca, local-model, and Computer Use capabilities. Platform support and permissions are shown by the application before use.

License: MIT.
