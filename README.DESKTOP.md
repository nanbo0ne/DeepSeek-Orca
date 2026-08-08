# DeepSeek-Orca Desktop Edition

This is the desktop-focused English entry point. The current build is the engineering edition and exposes only **Normal** and **Enhanced** modes. The standalone assistant extraction plan is documented in [docs/ORCA_ASSISTANT_APP_HANDOFF.md](docs/ORCA_ASSISTANT_APP_HANDOFF.md).

## Downloads

- [Windows installer](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.35/DeepSeek-Orca-windows-amd64-installer.exe)
- [Windows portable](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.35/DeepSeek-Orca-windows-amd64.zip)
- [macOS universal DMG](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.35/DeepSeek-Orca-darwin-universal.dmg)
- [Linux amd64 DEB](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.35/DeepSeek-Orca-linux-amd64.deb)
- [Release page](https://github.com/nanbo0ne/DeepSeek-Orca/releases/tag/desktop-v2.0.35)

## Scope

DeepSeek-Orca Desktop provides project, independent, and Automation workspaces, persistent sessions, forks, checkpoints, rollback, multi-provider models, optional vision attachments, tool approval and sandbox controls, MCP, Skills, CodeGraph, memory, Plan/Todo/Goal workflows, background jobs, bots, and chronological process display.

Normal mode handles routine conversation, research, writing, analysis, and everyday development. Enhanced mode handles complex coding, architecture, long tasks, reviews, and active repository changes. Both use the `DeepSeek-Orca` engineering identity. Automation Workspace conversations use the retained Assistant profile with a locked mode selector and a canonical profile shared with QQ and Weixin. They can selectively read or dispatch to ordinary engineering conversations through explicit conversation tools without changing the target's model, workspace, prompt, or approval policy.

Phone messages automatically create or restore an automation topic. A low-token no-tool check separates conversations after 30 minutes of inactivity, while `/continue` and `/new` provide explicit control. Remote mappings, one-time onboarding state, and shared session history persist across restarts; desktop and mobile turns targeting one transcript are serialized.

Vision defaults to Auto. Each newly configured model is probed independently with a history-free Orca-icon verification image, images are sent only to confirmed visual models, and selected current-turn images can be routed from a text-only main model to a visual subagent. DeepSeek being detected as unsupported is expected. Session files keep references, not image base64. Interface scaling follows Windows DPI by default, with an optional relative 80%-125% override.

Process display can be Compact or Detailed and defaults to Compact. Compact process rows can be repeatedly expanded and collapsed, while an optional theme-blue spinner appears below the current turn when enabled. It rotates clockwise for model activity and counterclockwise for tool activity. Successful turns collapse to elapsed time and the final answer; expanding restores the flat chronological activity rail. Failed, cancelled, interrupted, and active turns remain expanded. Conversation history paints before auxiliary status during switches, plain text pastes directly into the composer, and the input grows automatically for up to ten lines. Normal mode always retains the complete built-in engineering prompt and requires relevant post-write verification.

Updates are notification-only: the app checks stable official releases, caches successful checks for 24 hours, and opens the release page in a browser after the user clicks the update entry. It does not auto-download, auto-install, or force upgrades.

See [README.en.md](README.en.md) for the complete feature, safety, configuration, troubleshooting, and development documentation.
