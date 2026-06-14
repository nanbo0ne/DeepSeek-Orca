# DeepSeek-Orca Desktop

DeepSeek-Orca Desktop 是一个 Windows 桌面 AI 编程 Agent。它把模型对话、项目文件、工具调用、上下文统计、回滚记录、MCP、Skill、记忆和模型配置放在同一个桌面工作台里，适合用图形界面完成代码阅读、修改、调试和项目维护。

本项目基于 Reasonix fork 改造而来，保留原有 Agent、工具、MCP、Skill、记忆、权限控制和会话恢复能力，并重构为 DeepSeek-Orca 的桌面体验。

## How to use

1. 打开 [GitHub Releases](https://github.com/nanbo0ne/DeepSeek-Orca/releases/latest)。
2. 下载最新版 Windows 安装器：`DeepSeek-Orca-Setup-<version>-windows-amd64.exe`。
3. 双击安装器，阅读并同意许可协议，选择安装目录，然后完成安装。
4. 启动 DeepSeek-Orca Desktop，按引导填写 DeepSeek API Key，或在设置中配置 OpenAI-compatible provider。
5. 点击左侧的新建会话，选择项目文件夹或独立工作区。
6. 在底部输入任务，例如“解释这个项目结构”或“修复当前测试失败”。
7. 运行中继续输入并按 Enter 会加入待发送队列；按 Ctrl+Enter 会把这句话作为“引导”发给正在思考的 Agent。

## Features

- 多会话侧边栏：置顶会话、项目会话和独立工作区分组展示。
- 项目工作区：会话可以绑定项目文件夹，也可以在独立工作区中运行。
- 文件读写与命令执行：Agent 可以读取项目、修改文件、运行 shell，并按权限模式请求确认。
- 权限模式：支持写操作前询问、自动执行写操作和完全访问权限。
- Plan 模式：Plan 开关与权限模式独立，可先制定方案再执行。
- 运行中引导：模型思考时可把新消息排队，或用 Ctrl+Enter 直接引导当前轮。
- workflow 折叠：思考、工具调用、审批和后台任务默认折叠，必要时可展开查看。
- 事务级回滚：可删除 AI 回复并回滚该回复关联的文件改动。
- 上下文统计：显示当前上下文窗口、会话 tokens、缓存命中、请求数、耗时和费用。
- MCP 与 Skill：支持外部 MCP 工具和可复用 Skill。
- 记忆与项目指令：可读取项目指令文件和用户记忆，让 Agent 按项目规则工作。
- 设置面板：支持模型、provider、权限、主题、字体、关闭行为等桌面偏好。

## Installed Files

常见安装目录：

```text
C:\Users\<你的用户名>\AppData\Local\Programs\DeepSeek-Orca
```

主要文件：

- `deepseek-orca-desktop.exe`：桌面端主程序。
- `uninstall.exe`：卸载入口。
- `uninstall.bat`：备用卸载脚本。
- `node.exe`：随程序携带的 Node 运行时。
- `dist/`：桌面前端静态资源。
- `.deepseek-orca/`：配置、凭据引用、Skill、MCP、缓存等本地数据。
- `data/`：会话、历史、索引和工作区元数据。

卸载时可以选择是否删除已保存数据。若不勾选，配置、会话、记忆和缓存会尽量保留。
