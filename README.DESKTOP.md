# DeepSeek-Orca Desktop Edition

This is the desktop-focused English entry point. The current build is the engineering edition and exposes only **Normal** and **Enhanced** modes. The standalone assistant extraction plan is documented in [docs/ORCA_ASSISTANT_APP_HANDOFF.md](docs/ORCA_ASSISTANT_APP_HANDOFF.md).

## Downloads

- [Windows installer](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.29/DeepSeek-Orca-windows-amd64-installer.exe)
- [Windows portable](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.29/DeepSeek-Orca-windows-amd64.zip)
- [macOS universal DMG](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.29/DeepSeek-Orca-darwin-universal.dmg)
- [Linux amd64 DEB](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.29/DeepSeek-Orca-linux-amd64.deb)
- [Release page](https://github.com/nanbo0ne/DeepSeek-Orca/releases/tag/desktop-v2.0.29)

## Scope

DeepSeek-Orca Desktop provides project and independent workspaces, persistent sessions, forks, checkpoints, rollback, multi-provider models, optional vision attachments, tool approval and sandbox controls, MCP, Skills, CodeGraph, memory, Plan/Todo/Goal workflows, automation, background jobs, bots, and chronological process display.

Normal mode handles routine conversation, research, writing, analysis, and everyday development. Enhanced mode handles complex coding, architecture, long tasks, reviews, and active repository changes. Both use the `DeepSeek-Orca` engineering identity. The hidden Assistant prompt and memory implementation remain reserved for the future `Orca` application and are not active in this build.

Vision defaults to Auto. The app probes capability per model and endpoint with one small isolated request, sends images only to confirmed visual models, and can route selected current-turn images from a text-only main model to a visual subagent. DeepSeek being detected as unsupported is expected. Session files keep references, not image base64.

Process display can be Compact, Standard, or Detailed. Successful turns collapse to elapsed time and the final answer; expanding restores the flat chronological activity rail. Failed, cancelled, interrupted, and active turns remain expanded. Normal mode always retains the complete built-in engineering prompt and requires relevant post-write verification.

Updates are notification-only: the app checks stable official releases, caches successful checks for 24 hours, and opens the release page in a browser after the user clicks the update entry. It does not auto-download, auto-install, or force upgrades.

See [README.md](README.md) for the complete feature, safety, configuration, troubleshooting, and development documentation.
