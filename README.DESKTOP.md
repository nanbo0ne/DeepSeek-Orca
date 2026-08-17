# O.R.C.A Desktop Edition

This desktop build exposes **Coding** and **Assistant** conversation modes plus one fixed internal **Orca** control conversation.

## Downloads

- [Windows installer](https://github.com/nanbo0ne/O.R.C.A-for-Windows/releases/download/desktop-v3.0.0/O.R.C.A-for-Windows-windows-amd64-installer.exe)
- [Windows portable](https://github.com/nanbo0ne/O.R.C.A-for-Windows/releases/download/desktop-v3.0.0/O.R.C.A-for-Windows-windows-amd64.zip)
- [macOS universal DMG](https://github.com/nanbo0ne/O.R.C.A-for-Windows/releases/download/desktop-v3.0.0/O.R.C.A-macos-universal.dmg)
- [Linux amd64 DEB](https://github.com/nanbo0ne/O.R.C.A-for-Windows/releases/download/desktop-v3.0.0/O.R.C.A-linux-amd64.deb)
- [Release page](https://github.com/nanbo0ne/O.R.C.A-for-Windows/releases/tag/desktop-v3.0.0)

## Scope

O.R.C.A Desktop provides project and independent workspaces, a fixed Orca entry, persistent sessions, forks, checkpoints, rollback, multi-provider models, optional vision attachments, tool approval and sandbox controls, MCP, Skills, CodeGraph, memory, Plan/Todo/Goal workflows, Work artifacts, background jobs, bots, and chronological process display.

Coding mode handles repositories and engineering with shell, Git, LSP, CodeGraph, review, testing, and verification. Assistant mode handles research, writing, office artifacts, automations, and computer work with a personal memory profile. The top-level **Orca** entry is shared by desktop, QQ, and Weixin, adds selective conversation dispatch, and has no ordinary mode selector. Legacy automation topics migrate to ordinary independent workspaces.

The 30-minute continuity check, `/new`, and `/continue` create or select logical context segments inside Orca without creating sidebar conversations. Every channel shares one current segment and execution queue. Trusted automation access is confirmed once on desktop and can be revoked; unconfirmed sessions can chat without per-tool approval cards, while confirmed sessions automatically continue policy-allowed tools and plans. Four cached home suggestions are refreshed every 24 hours from limited summaries and sanitized examples of the user's own wording.

Vision defaults to Auto. Each newly configured model is probed independently with a history-free Orca-icon verification image, images are sent only to confirmed visual models, and selected current-turn images can be routed from a text-only main model to a visual subagent. DeepSeek being detected as unsupported is expected. Session files keep references, not image base64. Interface scaling follows Windows DPI by default, with an optional relative 80%-125% override.

Process display can be Compact or Detailed and defaults to Compact. Compact process rows can be repeatedly expanded and collapsed, while the default-on theme-blue spinner rotates clockwise for model output and counterclockwise for tool activity. Explicit Turn/Item identities keep stage updates in the process timeline and the committed final answer outside it. Successful turns collapse only after readiness and final-answer commitment; failed, cancelled, interrupted, active, and uncertain legacy turns remain expanded. Conversation history paints before auxiliary status during switches, plain text pastes directly into the composer, and the elastic input grows and contracts automatically for up to ten lines. Coding mode always retains the complete built-in engineering prompt and requires relevant post-write verification.

Updates are notification-only: the app checks stable official releases, caches successful checks for 24 hours, and opens the release page in a browser after the user clicks the update entry. It does not auto-download, auto-install, or force upgrades.

See [README.en.md](README.en.md) for the complete feature, safety, configuration, troubleshooting, and development documentation.
