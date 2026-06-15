# DeepSeek-Orca V1.0.23

DeepSeek-Orca 是一个面向代码项目的 AI 编程 Agent，提供 Windows 桌面端和 CLI 两种入口。它基于 Reasonix fork 改造，保留 Agent、工具调用、MCP、Skill、Memory、权限控制、会话恢复、文件回滚、上下文压缩等核心能力，并围绕 DeepSeek / OpenAI-compatible 模型重新整理了桌面体验、统计面板和机器人入口。

## How to use

### Windows 桌面版

下载并运行安装器：

[DeepSeek-Orca-Setup-1.0.23-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v1.0.23/DeepSeek-Orca-Setup-1.0.23-windows-amd64.exe)

安装后打开 DeepSeek-Orca Desktop，在设置里填写 DeepSeek API Key，或添加 OpenAI-compatible provider。普通用户推荐先使用桌面版：它会显示会话列表、项目文件、上下文统计、回滚入口和机器人连接设置。

### CLI

```powershell
git clone https://github.com/nanbo0ne/DeepSeek-Orca.git
cd DeepSeek-Orca
go build -o bin/deepseek-orca.exe ./cmd/deepseek-orca

$env:DEEPSEEK_API_KEY="your DeepSeek API Key"
.\bin\deepseek-orca.exe setup

cd D:\your-project
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe chat
```

一次性任务：

```powershell
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe run "阅读这个项目并总结主要模块"
```

## 功能概览

- 桌面端左侧按置顶、项目、独立工作区组织会话，左侧会话列表是主要切换入口。
- 新建对话时可以进入项目文件夹，也可以创建独立工作区会话。
- 推理、工具调用、审批、后台任务和 checkpoint 默认折叠，长对话更容易阅读。
- 模型生成时不会强制锁定到底部；需要跟随最新输出时，点击对话区的向下圆形按钮。
- Agent 运行时按 `Enter` 会把输入加入队列；按 `Ctrl+Enter` 会把输入作为“引导”发送给正在运行的 Agent。
- 支持 review、auto、yolo 三种权限模式，其中 yolo 表示完全访问权限。
- Plan 模式独立于权限模式，可先规划再执行。
- 删除 AI 回复时可回滚该回复关联的文件改动，并通过 hash 检查避免覆盖用户后续修改。
- 支持 DeepSeek thinking mode、OpenAI-compatible provider、MCP、Skill、Memory、CodeGraph 和项目指令文件。
- 统计面板显示上下文窗口、会话 token、缓存命中、请求数、耗时和费用；统计信息会随会话持久化，重启后恢复。
- 自动和手动 `/compact` 都使用 CONTEXT CHECKPOINT 交接式压缩，保留当前进度、关键决策、约束、用户偏好、未完成事项和继续所需数据。
- 中文文件名和 Windows 命令输出按 UTF-8/UTF-16/GB18030 等常见编码处理，减少乱码。
- QQ 和微信机器人可在桌面设置中连接。手机端发送 `/start` 可选择最近 15 条会话，发送 `/new` 可创建新的独立工作区对话；手机端消息会同步到桌面对应会话，回复会以少量进度提示加最终完整回复呈现。

## 安装后目录

常见安装目录：

```text
C:\Users\<你的用户名>\AppData\Local\Programs\DeepSeek-Orca
```

主要文件：

- `deepseek-orca-desktop.exe`：桌面端主程序。
- `uninstall.exe`：安装器生成的卸载程序。
- `uninstall.bat`：备用卸载脚本。
- `node.exe`：随程序携带的 Node 运行时。
- `dist/`：桌面端前端资源。
- `.deepseek-orca/`：配置、凭据引用、Skill、MCP、缓存等本地数据。
- `data/`：会话、历史记录、索引、工作区元数据和统计 telemetry。

卸载时可以选择是否删除已保存数据。

DeepSeek-Orca 基于 Reasonix fork 改造，License: MIT.
