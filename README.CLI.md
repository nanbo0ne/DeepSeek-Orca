# DeepSeek-Orca CLI

DeepSeek-Orca CLI 是面向终端工作流的 AI 编程 Agent。它适合熟悉命令行、Git、远程服务器和自动化脚本的开发者使用。

本项目基于 Reasonix fork 改造而来。CLI 保留核心 Agent、工具、MCP、Skill、Memory、权限控制、回滚和会话恢复能力，并统一命名为 DeepSeek-Orca。

## How to use

从源码构建：

```powershell
git clone https://github.com/nanbo0ne/DeepSeek-Orca.git
cd DeepSeek-Orca
go build -o bin/deepseek-orca.exe ./cmd/deepseek-orca
```

配置 API Key：

```powershell
$env:DEEPSEEK_API_KEY="你的 DeepSeek API Key"
.\bin\deepseek-orca.exe setup
```

在项目目录启动交互式会话：

```powershell
cd D:\your-project
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe chat
```

执行一次性任务：

```powershell
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe run "阅读这个项目并总结主要模块"
```

启动 IM 机器人网关：

```powershell
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe bot start --channels qq,weixin
```

检查机器人配置：

```powershell
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe bot doctor
```

## 常用命令

- `deepseek-orca setup`：创建或更新用户配置。
- `deepseek-orca chat`：启动交互式 TUI。
- `deepseek-orca run "task"`：执行一次性任务。
- `deepseek-orca bot start --channels qq,weixin`：启动 QQ / 微信机器人网关。
- `deepseek-orca bot doctor`：检查机器人配置。
- `deepseek-orca bot weixin-login`：保存微信 iLink / OpenClaw 登录 token。

微信 iLink / OpenClaw 返回的数字型 `message_id`、`from_user_id`、`chat_id` 和 `context_token` 会自动按字符串兼容处理。

会话中的常用 slash command：

- `/init`：创建或更新项目指令。
- `/plan`：切换 Plan 模式。
- `/skill`：管理 Skill 工作流。
- `/mcp`：管理 MCP 服务。
- `/model`：切换模型。
- `/resume`：恢复历史会话。
- `/rewind`：回滚会话上下文或相关改动。

## 机器人说明

CLI 机器人默认启用，并默认允许所有用户消息进入机器人；不再因为缺少白名单而拒绝启动。

QQ 使用官方 Bot API：

```toml
[bot.qq]
enabled = true
app_id = "your_app_id"
app_secret_env = "QQ_BOT_APP_SECRET"
environment = "production"
```

设置密钥：

```powershell
$env:QQ_BOT_APP_SECRET="your_app_secret"
```

微信使用扫码登录：

```powershell
deepseek-orca bot weixin-login
```

然后启动：

```powershell
deepseek-orca bot start --channels qq,weixin
```

飞书/Lark 暂未开放，不建议在当前版本中启用。

## 功能说明

- 交互式 TUI 和一次性 `run` 模式。
- 从配置、Memory、项目指令、MCP、Skill 和 CodeGraph 加载项目上下文。
- 文件读取、搜索、编辑和创建。
- 运行 shell 命令，用于构建、测试和诊断。
- 对写文件和命令执行提供权限规则。
- Plan 模式支持先规划再执行。
- 会话保存、恢复、分支、总结和回滚。
- 自动上下文压缩，压缩阈值可通过 `[agent] compact_ratio` 调整。
- 支持 DeepSeek 和 OpenAI-compatible provider。
- 按模型能力处理 DeepSeek thinking mode。
- MCP 外部工具集成。
- 本地 Skill 工作流。
- Memory 长期偏好和项目事实。
- QQ / 微信机器人网关。

## 桌面安装包

桌面端通过 Windows 安装器分发：

[DeepSeek-Orca Desktop 最新 Release](https://github.com/nanbo0ne/DeepSeek-Orca/releases/latest)
