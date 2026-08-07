# DeepSeek-Orca Desktop Changelog

## V2.0.34 - 2026-08-07

- Moved the Todo expansion into the footer's normal layout flow so its height is measured and the transcript always moves clear of the task list instead of being covered by a fixed popover.
- Consolidated Todo, queued prompts, approvals, plan confirmation, questions, and context cleanup into one visual surface per bottom panel with flat internal rows and dividers.
- Aggregated each turn's reasoning, progress text, tools, notices, compaction, and subagent activity into one process panel while keeping the final answer independent.
- Added explicit panel/content surface markers and regression coverage for nested-panel prevention, transparent process rows, stable compact/detailed behavior, and responsive overflow safeguards.

## V2.0.33 - 2026-08-07

- Added automatic interface scaling based on the DPI-adjusted usable display size plus an immediate 80%-125% manual scale control, and rebuilt Composer permission controls so they remain usable at narrow content widths.
- Unified file selection, Explorer clipboard paths, browser paste, native image paste, and drag-and-drop into ordered attachment batches. Added DOCX, XLSX, PPTX, PDF, CSV, TSV, and text extraction without blocking successful files when one item fails.
- Fixed user and project provider configuration merging and protected provider/model saves from stale background refreshes, empty catalog responses, and concurrent settings snapshots.
- Improved vision capability detection with provider metadata, protocol-specific probes, tolerant verification parsing, and per-model automatic/supported/unsupported overrides.
- Separated image thumbnails and file chips from the blue user text bubble, enlarged the completed-turn summary, and replaced nested process cards with a flat chronological activity rail.
- Synchronized Compact and Detailed switching across completed turns and reasoning rows, and fixed the question navigation rail so cold history is paged in the correct direction, mounted before an immediate stable jump, and triggered only once per pointer action.
- Added image support to the custom right-click Paste command and kept native text, file, mixed clipboard, attachment failure, retry, and ordering behavior consistent.
- Stopped the boot subagent test from writing `first review` fixtures into the real Windows profile and added a strict startup cleanup that removes only the exact leaked fixture signature.
- Muted the empty Composer send arrow across themed styles, kept failed-only attachment batches disabled, and made automatic proxy mode bypass loopback addresses so local model endpoints and provider tests are never routed through the Windows system proxy.

## V2.0.31 - 2026-08-02

- Added the Automation Workspace beside project and independent workspaces. Automation conversations use the retained Assistant prompt and canonical assistant profile while ordinary engineering conversations remain Normal or Enhanced.
- Added conversation routing tools for listing, reading, dispatching, waiting, status, cancellation, and ordinary-session creation. Routing stays inside the Assistant model's normal tool loop; only the 30-minute continuity check is a separate low-token request.
- Added QQ and Weixin first-message automation session creation, persisted remote-to-topic restoration, `/new`, `/continue`, `/hi`, and one-time persisted onboarding guidance.
- Added per-session execution leases so desktop and mobile controllers cannot concurrently write the same transcript, and waiting controllers refresh newer history before continuing.
- Preserved legacy Assistant sessions and memory stores while migrating them into a canonical profile without deleting the source data. Fixed automation-root indexing so it is not treated as an ordinary project.
- Updated automation capability tests, routing prompt coverage, broker ownership checks, and desktop/frontend layout regression checks.
- Fixed the running Composer action so typed or attached follow-up content replaces Stop with a same-size blue queued-send button, then restores Stop after the draft is queued.

## V2.0.30 - 2026-08-01

- Reduced process presentation to Compact and Detailed, made Compact the default, and migrated legacy Standard or collapsed-thinking settings to Compact.
- Fixed compact process rows so stable segment IDs preserve explicit open/closed state across streaming updates and closed details leave no residual layout or hit area.
- Added an optional, off-by-default blue spinner below the current turn. It rotates clockwise for model activity and counterclockwise for tool activity, with pause/completion handling and reduced-motion support.
- Removed the full-width footer surface behind Todo, plan, approval, queued prompt, and context cards while retaining normal-flow height and individual card surfaces.
- Split conversation loading into first-paint meta/history and background auxiliary hydration, added immediate switching for already-open topics, cached transcript derivations, and suppressed restored-history entrance replay.
- Changed plain-text paste to remain direct editable text and made the Composer grow automatically for up to ten lines before scrolling.
- Updated the desktop version and download documentation to V2.0.30.

## V2.0.29 - 2026-08-01

- Collapsed every explicitly successful completed turn to the user message, elapsed-time row, and final answer while preserving failed, cancelled, interrupted, and active diagnostics.
- Restored the full chronological process on demand and replaced nested process cards with a flat activity rail for reasoning, tools, notices, and subagents.
- Migrated multimodal vision to Off, Auto, and On modes, with new installations defaulting to per-model automatic capability detection.
- Added isolated visual probes, persisted model/endpoint capability status, bounded retries, manual rechecks, and expected text-only handling for DeepSeek.
- Added current-turn image delegation to confirmed vision-capable subagents without placing image base64 in transcripts, titles, memory, or compaction.
- Ensured Normal mode always includes the complete built-in engineering prompt, appends user instructions, and requires relevant verification after the latest write.
- Updated the desktop version and download documentation to V2.0.29.

## V2.0.28 - 2026-07-31

- Defined the engineering edition boundary: the desktop UI exposes only Normal and Enhanced while retaining Assistant prompt and memory code for the future standalone Orca application.
- Migrated restored Assistant tab and recent/bot preferences to Normal without deleting sessions, attachments, Assistant memories, or pending-memory state.
- Disabled Assistant automatic-memory scheduling and processing in the engineering edition and forced engineering controllers and bot sessions to use shared-agent memory.
- Removed prompt-mode descriptions and the separate help button from the engineering mode menu; compact layouts now show one mode icon with a tooltip.
- Added a capability-driven prompt-mode interface and removed hardcoded Assistant choices from the engineering Composer and bot settings.
- Consolidated the final Composer responsive layout layer across 720, 580, 460, 380, and 320px container widths while preserving model selection and preventing control overlap.
- Rewrote the engineering README documentation and added the Orca assistant application handoff document.

## V2.0.27 - 2026-07-25

- Added fail-closed SignPath Authenticode signing for the Windows application and final NSIS installer, including trusted timestamp verification before release publication.
- Added an off-by-default manual override for the temporary unsigned V2.0.27 Windows release while SignPath Foundation enrollment is pending; signed assets will replace it in place.
- Added a public Windows code-signing policy and verification instructions for official GitHub downloads.
- Centered the compact Todo control and made its width follow the active task text while retaining progress, pinning, hover, keyboard, and narrow-window behavior.
- Reworked automatic topic titles to summarize the complete first user/assistant turn across tool calls instead of stopping at the first assistant fragment.
- Restored original display text before title generation so `Referenced context`, reminders, handoffs, and attachment payloads cannot become conversation titles.
- Added lazy repair for legacy auto-generated wrapper titles without overwriting manually renamed conversations.

## V2.0.26 - 2026-07-24

- Hotfix (2026-07-25): fixed mojibake in the Windows install-options page by compiling the customized NSIS script as BOM-marked UTF-8 Unicode.
- Hotfix (2026-07-25): stopped cumulative session usage from being mistaken for a full 1,000K current context after prompt-mode or controller rebuilds.
- Hotfix (2026-07-25): moved the manual update check beside the automatic update toggle instead of using a separate settings row.
- Replaced the always-open Todo list with a compact progress bar that expands upward on hover or keyboard focus and can be pinned without changing footer height.
- Added systematic narrow-window priorities for the app chrome, topic actions, composer controls, status bar, and bottom panels at the 760px minimum window width.
- Restored safe update detection against official GitHub Releases, filtered to stable `desktop-v*` tags, with a 24-hour cache and notification-only download entry.
- Added General settings for automatic update checks and manual checks; in-app downloading, installation, and forced updates remain disabled.
- Added Windows installer choices for desktop shortcut creation and launch-after-install, including upgrade-state preservation and silent-install behavior.
- Migrated configuration to version 3 so update checks default on once while preserving a user's later explicit opt-out.

## V2.0.25 - 2026-07-24

- Rebuilt the transcript as a chronological event timeline so assistant text, reasoning, tools, notices, image reads, and compaction stay at their real positions.
- Added stable viewport anchoring while streaming, loading images, and growing tool output, with explicit follow-latest behavior.
- Restored Compact process display to single-line collapsed white rows and kept Standard/Detailed process behavior chronological.
- Delayed completed-turn actions until `TurnDone` and moved elapsed/token statistics to a lightweight turn header.
- Unified queued prompts, Todo, approval/plan, questions, and context confirmation into a non-overlapping footer shelf layout.
- Strengthened background bash and subagent guidance so dependent work must be collected with `wait`; background completion no longer implies a new model turn.
- Rewrote current-product README documentation in Chinese and English and updated release links for V2.0.25.

## V2.0.24 - 2026-07-20

- Added an opt-in multimodal vision setting for PNG, JPEG, WebP, and GIF images attached from the composer or referenced from the workspace.
- Added OpenAI-compatible and Anthropic-native image request serialization while keeping image base64 out of session JSONL, memory, compaction, and conversation search.
- Added image count and size limits, workspace image snapshots, independent attachment copies for global forks, and clear handling for missing or unsupported images.
- Added Compact, Standard, and Detailed process display modes. Compact mode keeps live reasoning and tool activity on one low-emphasis expandable line without changing final answer length.
- Preserved prompt mode, approval mode, ask workflow, step thinking, and goal state when settings rebuild the active controller.
- Decoupled balance requests from conversation loading, reduced high-frequency polling, and lowered streaming Markdown and transcript rendering overhead.
- Deferred Assistant memory generation to idle time so background profile updates do not compete with active conversations.
- Fixed the slash skill list subagent label and cleaned overlapping process/composer presentation rules.

## V2.0.23 - 2026-06-21

- Added Assistant mode proactive profile memory.
- Assistant memory generation now uses the same model selected for the source conversation, falling back to the default model only for legacy pending items.
- Assistant memory update failures now use bounded retry state and become ignored after 5 failed attempts unless new conversation messages arrive.
- Added Bot prompt mode selection under the Bot model setting, with Assistant, Normal, and Enhanced modes wired through the desktop and CLI bot gateway.
- Assistant mode now silently marks conversations for memory updates when switching, opening, creating, or closing conversations.
- Assistant memory generation runs as a separate lightweight provider call and does not write into the main transcript, title generation, compression, or token accounting.
- Failed assistant memory generation is recorded as pending/failed state and retried later without blocking the main conversation or app shutdown.
- Assistant mode can inject assistant memories before a turn for relevant recall, with a safe size budget for memory bodies.
- Added settings for Assistant auto memory, proactive assistant memory recall, and clearing only the Assistant memory profile.
- Memory entries now carry optional source, timestamp, confidence, and evidence metadata; the UI can show auto-generated memories.
- Normal and Enhanced modes keep the existing tool-based memory behavior.

## V2.0.22 - 2026-06-21

- Fixed the composer textarea alignment after manual resize so the caret and placeholder stay pinned to the top of the expanded input area.
- Removed FangSong and KaiTi from the appearance font picker.
- Added three easier-reading Windows-friendly font choices: DengXian, SimSun, and Microsoft YaHei UI.
- Kept Heiti as the default display option, implemented with Microsoft YaHei first and SimHei fallback.
- Moved README release notes into this changelog so README files can stay focused on download links and stable product documentation.

## V2.0.21 - Process Statistics And Font Settings

- Folded completed-turn process rows now show only elapsed thinking time by default.
- Token usage is shown in smaller text after expanding the process row.
- Mode menu descriptions were shortened and made more task-specific.
- The font picker was simplified to Chinese display fonts.
- Legacy saved font preferences fall back to the default display font.
- Code, terminal, and tool-output areas keep a monospace primary font to preserve alignment.

## V2.0.20 - Prompt And Memory Profiles

- Added three prompt profiles: Assistant, Normal, and Enhanced.
- Assistant mode runs as `Orca` for general help.
- Normal and Enhanced modes run as `DeepSeek-Orca` for engineering collaboration and stronger agentic coding workflows.
- Prompt profiles were adapted from user-provided Claude, GPT, and Claude Code style references while preserving English model-visible prompt structure.
- Platform-specific tool names were mapped to DeepSeek-Orca's real desktop tools.
- Memory was partitioned by mode: Assistant reads and writes assistant memory only; Normal and Enhanced read all memory but write to the shared agent memory profile.
- The Memory panel added filters for all profiles, Assistant mode, and Normal/Enhanced memory.

## V2.0.19 - Side Chat And Right Dock Fixes

- Fixed a React crash when opening Side Chat with empty or null side-chat history.
- `ListSideChat` now returns an empty list for empty history instead of allowing null bridge results.
- Restored the right dock navigation to a single-row four-tab layout with tighter font, icon, and spacing rules.

## V2.0.18 - Top Bar, Right Dock, Composer, And CodeGraph

- Improved top toolbar layout so normal-width windows have enough room for main action labels before compacting controls.
- Adjusted right dock navigation so Overview, Files, Changes, and Side Chat remain visible.
- Tightened running composer controls so Pause, status text, and Stop no longer push the input row out of alignment.
- Rendered queued prompts as a small floating rounded panel instead of a full-width rectangular row.
- Fixed Tool Library enabled-row colors so blue is reserved for hover emphasis.
- Updated CodeGraph steering to use actual MCP tool names such as `mcp__codegraph__context` and `mcp__codegraph__search`.
- Removed guidance that caused the model to call unknown bare `codegraph_search`.

## V2.0.17 - Tool Library And Long-Term Conversation Search

- Added a Tool Library button next to Automation in the top bar.
- Added switches for thread management, web search, Node/Python REPL, document tools, system/host tools, long-term conversation search, and proactive tool-use steering.
- Disabled tool groups are removed from both the registered tool schema and model-visible routing policy.
- Added read-only `conversation_search` and `conversation_read` tools for finding older local transcript information after context compression.

## V2.0.16 - Search, Automation, Todo, Plan Mode, And Workspace Isolation

- Changed `web_search` to avoid DuckDuckGo and use China-accessible search sources first.
- Added an evidence-first tool policy for Normal and Enhanced modes.
- Added shared automatic Todo tracking policy for complex multi-step work in Normal and Enhanced modes.
- Added Codex-style plan proposal cards.
- Revision requests now include the previous complete plan and ask for a complete replacement plan while staying in plan mode.
- Independent conversations now use separate per-topic workspace roots, session directories, attachment areas, memory/config scopes, and tool cwd.
- Fixed bottom status bar approval indicators so Ask, Auto approve, and Full access remain visible and color-coded.
- Made model switching update visible short model labels immediately while controller rebuild happens in the background.
- Reserved automation for explicit recurring, continuous, or background-monitoring tasks.
- Added the Automation manager next to the Bot button.
- Added Side Chat as a read-only right-dock conversation that can reference the main transcript without writing to the main history.
- Localized slash menu descriptions and browser-demo capability text more consistently in the Chinese UI.
