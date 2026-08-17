[简体中文](README.md) | **English**

# O.R.C.A for Windows

**O.R.C.A** means **Open Reasoning & Computing Agent**: a local desktop workspace that keeps reasoning, engineering execution, and everyday Work under the user's control. It combines multi-provider chat, project-scoped sessions, inspectable tools, planning, memory, MCP, skills, CodeGraph, artifacts, and reversible execution in one Windows, macOS, and Linux application.

Ordinary conversations expose two profiles: **Coding** for repository work and **Assistant** for everyday Work tasks. A fixed, higher-priority **Orca** conversation coordinates desktop and phone automation.

## Download

- [Windows installer](https://github.com/nanbo0ne/O.R.C.A-for-Windows/releases/download/desktop-v3.0.0/O.R.C.A-for-Windows-windows-amd64-installer.exe)
- [Windows portable package](https://github.com/nanbo0ne/O.R.C.A-for-Windows/releases/download/desktop-v3.0.0/O.R.C.A-for-Windows-windows-amd64.zip)
- [macOS universal DMG](https://github.com/nanbo0ne/O.R.C.A-for-Windows/releases/download/desktop-v3.0.0/O.R.C.A-macos-universal.dmg)
- [Linux amd64 DEB](https://github.com/nanbo0ne/O.R.C.A-for-Windows/releases/download/desktop-v3.0.0/O.R.C.A-linux-amd64.deb)
- [All release assets](https://github.com/nanbo0ne/O.R.C.A-for-Windows/releases/tag/desktop-v3.0.0)

The application never downloads or installs an update without user action. Update detection only displays a small release-page entry when a stable `desktop-v*` release is available. The Windows installer can create a desktop shortcut and can optionally launch the application after installation.

## What It Provides

### Coding and Assistant modes

- **Coding mode**: repositories, shell, Git, tests, LSP, CodeGraph, code review, and sustained engineering execution. Legacy Normal and Enhanced conversations migrate here.
- **Assistant mode**: everyday questions, research, information organization, office documents, automations, and computer work. It has Work instructions and personal memory without coding-only intelligence or Orca dispatch tools.

The mode is stored per conversation. A conversation with history shows a cache-invalidation warning before its system prompt, tools, and memory profile change. Visible non-system history is preserved, running turns apply a confirmed switch afterward, and a failed rebuild keeps the original controller and mode. The first ordinary conversation after upgrade defaults to Assistant; later conversations inherit the latest choice.

### Providers and models

First launch no longer depends on a DeepSeek key. The built-in catalog covers OpenAI, Anthropic, OpenRouter, DeepSeek, DashScope, Zhipu, Kimi, MiniMax, Volcano Ark, Baidu Qianfan, Tencent Hunyuan, StepFun, Xiaomi MiMo, SiliconFlow, and isolated subscription-plan endpoints, plus custom OpenAI- and Anthropic-compatible services. Every endpoint has its own provider ID, credential slot, and fully qualified model references, so same-name models never cross provider boundaries.

### Workspaces and sessions

- Project workspaces bind conversations to real repository directories.
- Independent topics receive isolated working directories and session data.
- Sessions support tabs, pinning, renaming, previews, restore, forks, checkpoints, rollback, trash, and export.
- Long sessions can be compacted while preserving a durable summary and searchable local history.
- Main transcript JSONL stays local and does not contain provider secrets.

### Vision attachments

Vision under Settings > Models offers **Off**, **Auto**, and **On**. New installations default to Auto: provider model metadata is used when available, otherwise each newly configured model is checked by a history-free request containing the Orca icon and a random verification code. Image bytes are sent only to models confirmed to support vision. Capability is tracked independently per provider type, endpoint, and model, with automatic, manually supported, and manually unsupported choices plus recheck controls in Settings. DeepSeek being marked unsupported is expected and does not affect text use.

Pasted, dropped, or workspace-referenced PNG, JPEG, WebP, and GIF files remain limited to 8 images per turn, 20 MB total, and 10 MB per image. Windows also includes optional Computer Use: an isolated control subagent re-observes after every action and can use UI Automation, guarded mouse/keyboard input, windows, and scrolling. Screenshots stay in memory by default and `Esc` is an emergency stop. UAC secure desktop, lock screen, passwords, verification codes, CAPTCHAs, and higher-integrity processes are always blocked. The V3 macOS/Linux builds keep chat functionality but do not provide local inference or Computer Use.

### Local AI and model library

Windows users can install the O.R.C.A-managed `llama.cpp` sidecar and Qwen GGUF models independently. On 16GB-class GPUs the first recommendation is Qwen3.8-27B IQ3_XXS; startup fits context, batch size, and GPU layers to current free VRAM while preserving headroom. Downloads support verified domestic mirrors, resumable transfers, live speed/ETA, and SHA-256 verification. Uninstalling the runtime leaves model files intact, and models never live inside the application install directory.

### Tools and local execution

The Tool Library controls groups for web search, host operations, Node/Python runtimes, document inspection, thread utilities, and conversation search. A disabled group is removed from both the registry and the model-visible routing policy. The built-in shell remains available for builds, tests, Git, and package managers. Tool approval modes make permission boundaries visible before execution; sandbox and write-root settings further constrain local operations.

### MCP, Skills, and CodeGraph

- MCP settings manage servers, authorization, connection status, retries, and exposed tools.
- Skills provide reusable workflows that load on demand.
- CodeGraph exposes project symbol and call-relationship tools through the configured MCP server.
- The slash menu combines built-in commands, skills, and MCP prompts.

### Memory

Coding uses the `shared-agent` engineering profile. Assistant, Orca, QQ, and Weixin share the canonical personal profile. Legacy Assistant stores are imported once without deleting their source files; proactive updates run only after foreground conversations and dispatched work are idle.

### Planning and process visibility

- Plan mode presents a reviewable plan before execution.
- Todo uses a compact centered progress control that expands upward on hover or keyboard focus and can be pinned without pushing the composer.
- Goal mode can continue a multi-step task and stops on its internal complete or blocked signal.
- Process display offers Compact and Detailed levels, with Compact as the default. Reasoning, tools, notices, and subagents use a flat chronological activity rail instead of nested cards; compact rows can be repeatedly expanded and collapsed.
- A theme-blue spinner is enabled by default and can be disabled in Settings. It rotates clockwise while the model is active and counterclockwise while tools are running, respects reduced-motion settings, and disappears when work pauses or ends.
- Coding and Assistant share an explicit Turn/Item lifecycle. Stage updates, reasoning, and tools remain visible while work runs; only a readiness-checked, explicitly committed visible answer can complete a turn successfully. Successful turns collapse to their user message, elapsed-time row, and complete final answer. Expanding restores every intermediate event in order; failed, cancelled, interrupted, active, and uncertain legacy turns keep their text visible for diagnosis.
- The transcript preserves the actual order of assistant text, reasoning, tool calls, tool results, notices, images, and compaction.
- Conversation switches reuse cached transcript structure, paint history before auxiliary status, and skip replaying entrance animations for restored messages. Plain-text paste remains directly editable, right-click paste accepts images, multi-file paste/drop keeps the complete ordered batch, and the composer grows automatically for up to ten lines before scrolling.

### Work artifacts

Assistant and Orca provide built-in `artifact_create`, `artifact_edit`, `artifact_preview`, and `artifact_validate` tools for DOCX, XLSX, PPTX, and PDF without requiring Python, Office, or LibreOffice. Orca-created files carry a structural sidecar for reliable later edits. Complex third-party Office files without that sidecar are rejected explicitly instead of being presented as lossless edits. See the [Work artifact runtime notes](docs/ARTIFACT_RUNTIME.md) for scope, font behavior, and limitations.

### Orca and remote connections

A fixed **Orca** entry appears directly below search, ahead of every project; there is no Automation Workspace folder or ordinary mode selector. Desktop, QQ, and Weixin all write to Orca and share its dedicated model and canonical memory. During upgrade, legacy automation topics move to ordinary independent workspaces. Orca answers directly unless a request genuinely depends on existing context, in which case it can selectively read or dispatch to Coding and Assistant conversations without recursively dispatching Orca.

After 30 minutes of inactivity, a small no-tool check decides only whether the next turn should load the previous segment's context. Related and unrelated turns remain visible in the same Orca transcript. `/new` starts a clean logical segment inside Orca and `/continue` forces previous-context continuation; `/hi`, `/status`, `/stop`, `/approve`, `/deny`, and `/answer` remain for compatibility. Every channel shares the current segment and one serial execution queue, so phone use never creates another sidebar conversation.

Trusted automation access is confirmed once on desktop and can be revoked in Settings. Before confirmation Orca can still chat, but protected tools are declined without per-command approval cards. After confirmation, policy-allowed tools and plan execution proceed automatically; Ask questions and explicit deny rules still apply. The four home suggestions are refreshed every 24 hours from limited local summaries and sanitized samples of the user's own phrasing, producing questions in the user's voice with cached fallback and no hidden conversation or memory side effects.

## Privacy and Safety

O.R.C.A is a local desktop shell, but model requests are sent to the provider selected by the user. Read the active provider, endpoint, proxy, vision, tool, and approval settings before using sensitive data. Image bytes are transmitted only when the selected vision mode permits the target model and a turn includes the image. Vision probing itself makes one very small request to the configured provider. API keys are read from configured environment variables or the local credentials flow rather than written into conversation messages.

Tool execution is visible in the transcript. Approval, sandbox, workspace roots, read-only tools, and background-task status are separate controls. Cancel, pause, rollback, checkpoints, and session trash provide recovery paths for long tasks.

## Configuration

After first launch, open Settings and configure models and providers, workspace and sandbox behavior, Tool Library and MCP servers, vision, process display, update checks, language, appearance, permissions, and approval defaults. Interface scaling follows Windows and its PerMonitorV2 DPI handling by default; an optional 80%-125% manual scale is applied relative to that native size. The app exposes the active configuration path in Settings. Project-local instructions can be supplied through the supported instruction files in the workspace.

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
