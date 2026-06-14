# DeepSeek-Orca V1.0

DeepSeek-Orca 是一个 AI 编程 Agent，基于 Reasonix fork 改造而来，提供 **桌面端** 和 **CLI** 两种入口。

- **桌面端**：适合用图形界面管理项目、会话、文件改动、回滚、模型、MCP、Skill、记忆和机器人渠道。
- **CLI**：适合程序员在终端、远程服务器、脚本和自动化流程中使用。

本项目保留 Reasonix 原有的 Agent、工具、MCP、Skill、Memory、权限控制和会话恢复能力，并统一重构为 DeepSeek-Orca。

## How to use

### 桌面端

下载打包好的 Windows 安装器：

[DeepSeek-Orca-Setup-1.0.13-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v1.0.13/DeepSeek-Orca-Setup-1.0.13-windows-amd64.exe)

也可以打开最新版 Release：

[GitHub Releases](https://github.com/nanbo0ne/DeepSeek-Orca/releases/latest)

安装包文件名格式：

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

## 主要功能

- 项目会话、置顶会话和独立工作区。
- Codex 风格的桌面侧栏与会话切换。
- workflow、工具调用、审批和推理过程默认折叠。
- AI 文件改动支持事务级回滚。
- Agent 运行中可排队消息，也可用 `Ctrl+Enter` 直接引导当前轮。
- Plan 模式与权限模式相互独立。
- 上下文、token、缓存命中、请求数量、耗时和费用统计。
- DeepSeek 与 OpenAI-compatible provider。
- 按 DeepSeek 官方接口约定处理 thinking mode。
- MCP、Skill、Memory、CodeGraph 和 slash command。
- QQ、微信、飞书/Lark 机器人渠道。
- QQ Bot 引导式配置：去官方后台申请 App ID / App Secret，选择沙箱或正式环境，然后保存并连接。
- Windows 一键安装包与卸载入口。

## 安装后的目录结构

常见安装目录：

```text
C:\Users\<你的用户名>\AppData\Local\Programs\DeepSeek-Orca
```

主要文件：

- `deepseek-orca-desktop.exe`：桌面端主程序。
- `uninstall.exe`：安装器生成的卸载程序。
- `uninstall.bat`：备用卸载脚本。
- `node.exe`：随程序携带的 Node 运行时。
- `dist/`：桌面端前端静态资源。
- `.deepseek-orca/`：本地配置、凭据引用、Skill、MCP、缓存等数据。
- `data/`：会话历史、工作区元数据、索引等用户数据。

## 当前版本

桌面端当前版本：**v1.0.13**。

本版本修复 GitHub Actions 桌面端发布流程：如果同名 GitHub Release 已存在，workflow 会覆盖上传资源，不会再因为 `release already exists` 而失败。同时保留 v1.0.12 加入的 QQ Bot 引导式配置。

## 致谢

DeepSeek-Orca 基于 Reasonix fork 改造。

## License

MIT License.
