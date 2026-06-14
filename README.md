# DeepSeek-Orca V1.0

DeepSeek-Orca 是一个 AI 编程 Agent，基于 Reasonix fork 改造而来，提供桌面端和 CLI 两种入口。它保留 Reasonix 的 Agent、工具、MCP、Skill、Memory、权限控制、会话恢复和机器人能力，并针对 DeepSeek / OpenAI-compatible 工作流做了桌面体验重构。

## How to use

### 桌面端

下载打包好的 Windows 安装器：

[DeepSeek-Orca-windows-amd64-installer.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/latest)

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
- 统一浅色 DeepSeek-Orca 外观，设置中只保留字号与字体。
- workflow、工具调用、审批和推理过程默认折叠。
- AI 文件改动支持事务级回滚。
- Agent 运行中可排队消息，也可用 `Ctrl+Enter` 直接引导当前轮。
- Plan 模式与权限模式相互独立。
- 上下文、token、缓存命中、请求数量、耗时和费用统计。
- DeepSeek 与 OpenAI-compatible provider。
- 按 DeepSeek 官方接口约定处理 thinking mode。
- MCP、Skill、Memory、CodeGraph 和 slash command。
- QQ 与微信机器人渠道；飞书/Lark 暂未开放。
- Windows 一键安装包与卸载入口。

## 机器人说明

桌面端中，QQ 和微信连接成功后会自动在后台运行机器人网关，不需要手动启动，也不需要设置白名单。所有能给机器人发消息的用户默认都可以使用机器人。

QQ 使用官方 Bot API，需要先在 QQ 机器人官方后台申请 App ID / App Secret，然后在桌面设置页填写并验证。

微信使用 ClawBot / OpenClaw Weixin 风格的扫码登录流程。扫码确认后，DeepSeek-Orca 会保存本地 token，并自动验证 `getupdates` 是否可用。

飞书和 Lark 暂时置灰，不作为当前版本可用渠道。

## 安装后目录结构

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
- `data/`：会话、历史记录、索引和工作区元数据。

## 致谢

DeepSeek-Orca 基于 Reasonix fork 改造。

## License

MIT License.
