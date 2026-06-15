# DeepSeek-Orca V1.0

DeepSeek-Orca 是一个面向代码项目的 AI 编程 Agent，提供 Windows 桌面端和 CLI 两种入口。它基于 Reasonix fork 改造，保留 Agent、工具调用、MCP、Skill、Memory、权限控制、会话恢复、文件回滚等核心能力，并围绕 DeepSeek / OpenAI-compatible 模型重新整理了交互、统计和上下文管理体验。

## How to use

### 桌面端

下载 Windows 安装器：

[DeepSeek-Orca-Setup-1.0.21-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v1.0.21/DeepSeek-Orca-Setup-1.0.21-windows-amd64.exe)

双击安装器，阅读并同意许可协议，选择安装目录，然后启动 DeepSeek-Orca Desktop。首次使用时填写 DeepSeek API Key；如果你使用代理服务或其他兼容接口，也可以在设置中添加 OpenAI-compatible provider。

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

一次性任务：

```powershell
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe run "阅读这个项目并总结主要模块"
```

## 核心功能

- 桌面会话侧栏按置顶、项目、独立工作区分组，左侧就是主要切换入口。
- 新建对话时可选择项目文件夹，也可直接进入独立工作区。
- 推理、工具调用、审批、后台任务和 checkpoint 默认折叠，减少长对话滚动负担。
- 模型生成回复时不会强制把视角锁在底部；需要跟随最新输出时，点击对话区中的向下圆形按钮。
- Agent 运行时按 `Enter` 会把输入加入队列；按 `Ctrl+Enter` 会作为“引导”发送给当前运行中的 Agent。
- 支持 review、auto、yolo 三种权限模式；yolo 表示完全访问权限。
- Plan 模式独立于权限模式，可先规划再执行。
- 删除 AI 回复时可回滚该回复关联的文件改动，并通过 hash 检查避免覆盖用户后续修改。
- 支持 DeepSeek thinking mode、OpenAI-compatible provider、MCP、Skill、Memory、CodeGraph 和项目指令文件。
- 统计面板显示当前上下文窗口占用、会话 token、缓存命中、请求数、耗时和费用；这些信息会随会话持久化，重启后恢复。
- 自动和手动 `/compact` 都使用 CONTEXT CHECKPOINT 交接式压缩：保留当前进度、关键决策、用户偏好、约束、未完成事项和继续所需数据，供接续模型继续工作。
- 右侧上下文面板会在压缩、回滚和重新打开后重新计算，不把 provider 的计费 token 误当成当前上下文窗口占用。
- 中文文件名和 Windows 命令输出按 UTF-8/GB18030 等常见编码处理，流式工具输出也会保留 UTF-8 字符边界，减少乱码。

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

## 当前状态

QQ、微信、飞书和 Lark 机器人入口当前暂时关闭，后续版本会在连接稳定后重新开放。

DeepSeek-Orca 基于 Reasonix fork 改造。License: MIT.
