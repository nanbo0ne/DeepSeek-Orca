# DeepSeek-Orca Desktop Edition

This is the desktop-focused English entry point. The current build is the engineering edition and exposes only **Normal** and **Enhanced** modes. The standalone assistant extraction plan is documented in [docs/ORCA_ASSISTANT_APP_HANDOFF.md](docs/ORCA_ASSISTANT_APP_HANDOFF.md).

## Downloads

- [Windows installer](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.28/DeepSeek-Orca-windows-amd64-installer.exe)
- [Windows portable](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.28/DeepSeek-Orca-windows-amd64.zip)
- [macOS universal DMG](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.28/DeepSeek-Orca-darwin-universal.dmg)
- [Linux amd64 DEB](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.28/DeepSeek-Orca-linux-amd64.deb)
- [Release page](https://github.com/nanbo0ne/DeepSeek-Orca/releases/tag/desktop-v2.0.28)

## Scope

DeepSeek-Orca Desktop provides project and independent workspaces, persistent sessions, forks, checkpoints, rollback, multi-provider models, optional vision attachments, tool approval and sandbox controls, MCP, Skills, CodeGraph, memory, Plan/Todo/Goal workflows, automation, background jobs, bots, and chronological process display.

Normal mode handles routine conversation, research, writing, analysis, and everyday development. Enhanced mode handles complex coding, architecture, long tasks, reviews, and active repository changes. Both use the `DeepSeek-Orca` engineering identity. The hidden Assistant prompt and memory implementation remain reserved for the future `Orca` application and are not active in this build.

Vision is disabled by default. When enabled, PNG/JPEG/WebP/GIF attachments are sent to the selected provider within the documented count and size limits. Session files keep references, not image base64. Process display can be Compact, Standard, or Detailed, and does not change model output or token accounting.

Updates are notification-only: the app checks stable official releases, caches successful checks for 24 hours, and opens the release page in a browser after the user clicks the update entry. It does not auto-download, auto-install, or force upgrades.

See [README.md](README.md) for the complete feature, safety, configuration, troubleshooting, and development documentation.
