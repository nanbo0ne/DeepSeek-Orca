# DeepSeek-Orca V1.0 中文说明

DeepSeek-Orca 是基于 Reasonix fork 改造的 AI 编程 Agent，提供桌面端和 CLI 两种入口。

- **桌面端**：适合通过图形界面管理项目、会话、文件改动、回滚、模型、MCP、Skill、记忆和机器人渠道。
- **CLI**：适合程序员在终端、远程服务器、脚本、自动化流程和 CI 辅助场景中使用。

## 如何使用

### 桌面端

下载 Windows 安装包：

[DeepSeek-Orca-Setup-1.0.18-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v1.0.18/DeepSeek-Orca-Setup-1.0.18-windows-amd64.exe)

或打开最新版 Release：

[GitHub Releases](https://github.com/nanbo0ne/DeepSeek-Orca/releases/latest)

安装后首次启动时，填写 DeepSeek API Key，或在设置中配置其他 OpenAI-compatible provider。

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

## 主要特性

- 项目会话、置顶会话和独立工作区。
- Codex 风格的桌面侧栏和会话切换。
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

## 当前版本

桌面端当前版本：**v1.0.13**。

本版本修复桌面端 GitHub Release workflow：当同名 Release 已存在时，Actions 会覆盖上传资源而不是失败。同时保留 v1.0.12 加入的 QQ Bot 引导式配置。

## 致谢

DeepSeek-Orca 基于 Reasonix fork 改造。

## License

MIT License.
