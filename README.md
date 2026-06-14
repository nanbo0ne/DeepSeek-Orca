# DeepSeek-Orca V1.0

DeepSeek-Orca 是一个 AI 编程 Agent，提供桌面端和 CLI 两种入口。桌面端适合通过图形界面管理项目、会话、上下文、文件改动和回滚；CLI 适合程序员在终端、远程环境、脚本和自动化流程中使用。

本项目基于 Reasonix fork 改造而来，并统一重命名为 DeepSeek-Orca。

## How to use

### Desktop

Windows 用户直接下载打包好的安装器：

[打开最新版 GitHub Release](https://github.com/nanbo0ne/DeepSeek-Orca/releases/latest)

下载文件名类似：

```text
DeepSeek-Orca-Setup-<version>-windows-amd64.exe
```

双击安装器，阅读并同意许可协议，选择安装目录，然后启动 DeepSeek-Orca Desktop。首次启动时填写 DeepSeek API Key，或在设置中配置 OpenAI-compatible provider。

桌面端详细说明见 [README.DESKTOP.md](./README.DESKTOP.md)。

### CLI

从源码构建：

```powershell
git clone https://github.com/nanbo0ne/DeepSeek-Orca.git
cd DeepSeek-Orca
go build -o bin/deepseek-orca.exe ./cmd/deepseek-orca
```

配置并启动：

```powershell
$env:DEEPSEEK_API_KEY="你的 DeepSeek API Key"
.\bin\deepseek-orca.exe setup
cd D:\your-project
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe chat
```

CLI 详细说明见 [README.CLI.md](./README.CLI.md)。

## Desktop Features

- Codex 风格左侧会话栏：置顶、项目分组、独立工作区分组展示。
- 新建会话时可选择进入项目或独立工作区。
- 运行中按 Enter 会把输入加入待发送队列，Ctrl+Enter 会直接引导当前 Agent。
- 正在运行的会话会在左侧显示转圈状态。
- Todo 全部完成后自动移除待办框。
- workflow 消息默认折叠，减少对话滚动长度。
- 支持事务级回滚，可删除 AI 回复并恢复该轮文件改动。
- 右侧上下文面板显示当前窗口占用、会话 tokens、缓存命中、费用和读取/变更文件。
- 支持 DeepSeek、OpenAI-compatible provider、MCP、Skill、Memory、CodeGraph、权限控制和 Plan 模式。

## CLI Features

- 交互式 TUI 和一次性 `run` 任务。
- 读取项目上下文、项目指令文件、记忆、MCP、Skill 和 CodeGraph。
- 支持文件读写、shell 命令、Git diff 审查、测试运行、会话恢复和回滚。
- 支持 Plan 模式、权限规则、provider 配置和 DeepSeek thinking mode。

## Installed Desktop Structure

默认安装目录通常是：

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

## License

MIT License. DeepSeek-Orca is based on a Reasonix fork.
