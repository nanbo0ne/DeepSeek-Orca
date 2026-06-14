# DeepSeek-Orca V1.0

DeepSeek-Orca 是一个 AI 编程 Agent，基于 Reasonix fork 改造而来，提供桌面端和 CLI 两种入口。它保留 Reasonix 的 Agent、工具、MCP、Skill、Memory、权限控制和会话恢复能力，并针对 DeepSeek / OpenAI-compatible 工作流做了桌面体验重构。

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
- CONTEXT CHECKPOINT 上下文交接：自动达到阈值或手动 `/compact` 时，生成面向接续模型的交接摘要；压缩后右侧上下文窗口会重新计算。
- DeepSeek 与 OpenAI-compatible provider。
- 按 DeepSeek 官方接口约定处理 thinking mode。
- MCP、Skill、Memory、CodeGraph 和 slash command。
- 机器人入口当前暂时关闭，后续版本会重新开放 QQ / 微信连接。
- Windows 一键安装包与卸载入口。

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
