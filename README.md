# DeepSeek-Orca V1.0

DeepSeek-Orca 是一个面向代码项目的 AI 编程 Agent，基于 Reasonix fork 改造而来。它提供 Windows 桌面端和 CLI 两种入口，保留 Agent、工具调用、MCP、Skill、Memory、权限控制、会话恢复和文件回滚等核心能力，并围绕 DeepSeek / OpenAI-compatible 模型重新整理了交互、统计和上下文管理体验。

## 快速开始

### 桌面端

下载 Windows 安装器：

[DeepSeek-Orca-Setup-1.0.20-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v1.0.20/DeepSeek-Orca-Setup-1.0.20-windows-amd64.exe)

双击安装器，阅读并同意许可协议，选择安装目录，然后启动 DeepSeek-Orca Desktop。首次启动时填写 DeepSeek API Key；如果你使用代理服务或其他兼容接口，也可以在设置里添加 OpenAI-compatible provider。

### CLI

```powershell
git clone https://github.com/nanbo0ne/DeepSeek-Orca.git
cd DeepSeek-Orca
go build -o bin/deepseek-orca.exe ./cmd/deepseek-orca

$env:DEEPSEEK_API_KEY="你的 DeepSeek API Key"
.\bin\deepseek-orca.exe setup

cd D:\your-project
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe chat
```

也可以执行一次性任务：

```powershell
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe run "阅读这个项目并总结主要模块"
```

## 桌面体验

DeepSeek-Orca Desktop 使用左侧会话栏作为主入口。会话按置顶、项目文件夹和独立工作区分组；置顶只影响排序，不改变会话所属项目。点击新建对话时，可以选择进入某个项目，也可以直接使用独立工作区。

对话区默认保留编码 Agent 需要的信息密度：用户消息、AI 最终回复和文件改动会直接显示，推理、工具调用、审批、后台任务和 checkpoint 会折叠成一行，必要时再展开。模型生成回复时不会强制锁住滚动位置；如果你想回到最新输出，可以点击对话区里的向下圆形按钮重新跟随。

底部输入器支持普通消息、文件和图片附件、`@` 文件引用、slash command、模型/effort 选择、Plan 模式和权限模式。Agent 正在运行时，按 `Enter` 会把新输入加入待发送队列；按 `Ctrl+Enter` 会把输入作为“引导”直接送给当前正在运行的 Agent。

## Agent 能力

- 读取、搜索、编辑和创建项目文件。
- 执行 shell 命令，用于构建、测试、诊断和自动化任务。
- 使用事务级回滚：AI 回复关联的文件改动会被记录，删除并回退时会校验文件 hash，避免覆盖用户后续修改。
- 支持 review、auto、yolo 三种权限模式，其中 yolo 在界面中解释为完全访问权限。
- Plan 模式独立于权限模式，可先让 Agent 只读规划，再决定是否执行。
- 支持排队消息和运行中引导，长任务期间输入不会丢失。
- 支持 MCP 外部工具、Skill 工作流、Memory 长期偏好、CodeGraph 代码结构辅助和项目指令文件。
- 支持 DeepSeek 官方 thinking mode 约定，也支持 OpenAI-compatible provider。

## 上下文与统计

DeepSeek-Orca 会显示当前上下文窗口占用、会话累计 token、缓存命中率、请求数量、耗时和费用。统计信息会持久化到会话 telemetry，关闭并重新打开应用后仍会恢复。

当上下文接近阈值时，DeepSeek-Orca 会执行 CONTEXT CHECKPOINT：它不是普通摘要，而是为接续模型准备的交接说明，包含当前进度、关键决策、用户偏好、约束、未完成事项和继续工作所需数据。手动输入 `/compact` 或 `/compact <focus>` 也会触发同样的交接式压缩。压缩后右侧上下文窗口会重新计算，避免把旧的累计 token 错当作当前上下文占用。

## CLI 适合谁

CLI 面向更熟悉终端的开发者：可以在本地项目、远程服务器、脚本和 CI 辅助流程中使用。它保留交互式 TUI、一次性 `run`、配置向导、MCP、Skill、Memory、Plan、权限、回滚和会话恢复能力。

常用命令：

- `deepseek-orca setup`：创建或更新配置。
- `deepseek-orca chat`：启动交互式 TUI。
- `deepseek-orca run "task"`：执行一次性任务。
- `deepseek-orca doctor`：检查配置和运行环境。

会话中常用 slash command：

- `/init`：创建或更新项目指令。
- `/plan`：切换 Plan 模式。
- `/skill`：管理 Skill。
- `/mcp`：管理 MCP 服务。
- `/model`：切换模型。
- `/resume`：恢复历史会话。
- `/rewind`：回滚会话上下文或相关改动。
- `/compact`：手动执行 CONTEXT CHECKPOINT。

## 安装后目录

默认安装位置：

```text
C:\Users\<你的用户名>\AppData\Local\Programs\DeepSeek-Orca
```

主要文件和目录：

- `deepseek-orca-desktop.exe`：桌面端主程序。
- `uninstall.exe`：安装器生成的卸载程序。
- `uninstall.bat`：备用卸载脚本。
- `node.exe`：随程序携带的 Node 运行时。
- `dist/`：桌面端前端静态资源。
- `.deepseek-orca/`：配置、凭据引用、Skill、MCP、缓存等本地数据。
- `data/`：会话、历史记录、索引、工作区元数据和统计 telemetry。

卸载时可以选择是否删除已保存数据。

## 当前状态

QQ、微信、飞书和 Lark 机器人入口当前暂时关闭，后续版本会在连接稳定后重新开放。

DeepSeek-Orca 基于 Reasonix fork 改造。License: MIT.
