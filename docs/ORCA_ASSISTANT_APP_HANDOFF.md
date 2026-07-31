# Orca Assistant Application Handoff

This document defines the extraction boundary for the future standalone personal-assistant application. The current DeepSeek-Orca engineering release does not build a second installer and does not expose the assistant profile in its UI.

## Product Identity

- Product name: `Orca`.
- Edition value: `assistant`.
- User-facing prompt mode: `assistant` only.
- Model identity in the prompt: `Orca`, not `DeepSeek-Orca`.
- Engineering product: `DeepSeek-Orca`, edition `engineering`, visible modes `normal` and `enhanced`.

The boundary is returned by `GetProductCapabilities()`. Engineering returns `engineering`, `[normal, enhanced]`, and `assistantMemoryEnabled=false`. A future assistant build should return `assistant`, `[assistant]`, and `assistantMemoryEnabled=true`.

## Reusable Kernel

The standalone app should reuse the internal Go kernel rather than copy prompt or provider logic:

- `internal/boot` keeps `PromptModeAssistant` and accepts `boot.Options{PromptMode: "assistant"}`.
- `internal/promptprofile` contains `AssistantSystemPrompt(...)` and its tool-routing, language, vision, and task policy composition.
- `internal/provider` handles OpenAI-compatible and Anthropic-compatible requests.
- `internal/control` owns turn execution, approvals, questions, plans, Todo, Goal, tools, session resume, and events.
- `internal/memory` owns the assistant store, shared-agent store, index format, frontmatter, and profile normalization.
- `desktop/assistant_memory.go` contains the automatic assistant-memory pending queue, cursor handling, bounded retry behavior, idle worker, and provider call.
- Shared React components may be reused where their copy and capability gates are neutral. The assistant shell must not render the engineering mode switch or engineering-only settings.

The engineering edition passes `MemoryProfile: "shared-agent"` to every desktop controller and skips the assistant-memory worker. Generic `boot.Build` remains compatible with assistant mode so the future app can opt into the assistant profile without changing session JSONL or provider protocols.

## Assistant Prompt and Memory

`AssistantSystemPrompt` should remain a stable system prefix so provider cache behavior is preserved. It describes Orca as a general assistant and keeps only tool schemas that are actually registered by the selected tool library.

Assistant memory rules:

- Assistant mode reads and writes only the `assistant` profile.
- Normal and Enhanced read and write the `shared-agent` profile.
- Existing unclassified memories remain shared-agent data.
- Assistant recall is hidden context before a turn and never enters the visible transcript.
- Automatic profile extraction runs after leaving a conversation or on next launch when a pending marker exists.
- Extraction uses the same model selected for the source conversation, performs no tools, and does not add messages to the main JSONL.
- Keep durable preferences, working style, stable interests, projects, goals, and recurring needs; exclude secrets, credentials, sensitive profiling, one-off questions, and full chat transcripts.
- Preserve confidence filtering, similarity updates, explicit forget requests, and user deletion.

The pending state file records the session path, topic, workspace root, model, message cursor, status, timestamps, and retry count. Failed batches retry at most five times. After the fifth failed attempt they become ignored until new conversation messages create a new pending batch. The engineering build leaves old pending state untouched so the future assistant app can resume it.

## Data Migration

The assistant app must use independent writable roots. It must not concurrently write the engineering app's files.

Recommended identities:

- Windows config: `%APPDATA%\\Orca`.
- Windows local data: `%LOCALAPPDATA%\\Orca`.
- macOS config/data: `~/Library/Application Support/Orca`.
- Linux config/data: `${XDG_CONFIG_HOME:-~/.config}/orca` and `${XDG_DATA_HOME:-~/.local/share}/orca`.
- Executable: `Orca` / `Orca.exe`.
- Installer/product identifier: `com.deepseek-orca.orca`.
- Release tag: `orca-vX.Y.Z`.

On first launch, perform an idempotent copy migration:

1. Read old tab metadata and identify entries whose prompt mode was `assistant`.
2. Copy assistant memory documents and assistant auto-memory files into the Orca data root.
3. Copy pending assistant-memory state, preserving cursors, statuses, retry counts, and model names.
4. Copy assistant-mode session JSONL and display sidecars into the Orca session root.
5. Copy assistant-specific settings and normalize them against assistant-only capabilities.
6. Never move or delete source files during the first migration.
7. Record a migration marker and source root so restart cannot duplicate entries.
8. If a session is also present in engineering, the assistant app owns its copied version; the products must never share a writable session file.

Engineering sessions that were assistant sessions are restored as Normal for continuity. This changes only the engineering tab preference and leaves JSONL, title, attachments, assistant memories, and pending state intact.

## UI and Settings

The assistant app shows one assistant identity and no Normal/Enhanced selector. It may expose assistant automatic memory, assistant recall, memory list/filter/delete/clear, model/provider, vision, process display, tools, permissions, and update settings valid for that product. Personalization copy should promise relevant recall without claiming perfect memory. Memory activity stays quiet unless the user asks why Orca knows something or opens memory settings.

## Release and Updates

The assistant app needs its own version line, application identifier, installer name, update tag, release notes, and signing identity. It must not consume engineering `desktop-v*` releases. Build artifacts should be signed before publication, and the update checker must filter to the assistant release namespace.

Build from a clean checkout, package executable and installer separately, verify platform signatures, publish checksums, and create the release only after independent data roots and migration tests pass.

## Acceptance Tests

- Assistant exposes only `assistant` and uses the Orca identity.
- Existing assistant sessions and memories migrate by copy and the migration is idempotent.
- The assistant worker retries no more than five times and ignores a permanently failing batch until new evidence exists.
- Assistant recall never enters visible transcript or main token statistics.
- Engineering cannot select assistant, cannot process assistant pending work, and cannot read assistant memories.
- No two products write the same session, memory, config, or pending-state file.
