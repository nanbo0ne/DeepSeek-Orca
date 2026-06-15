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

机器人功能当前暂时关闭；后续版本会重新开放 QQ / 微信连接。

## 常用命令

- `deepseek-orca setup`：创建或更新用户配置。
- `deepseek-orca chat`：启动交互式 TUI。
- `deepseek-orca run "task"`：执行一次性任务。
- `/init`：创建或更新项目指令。
- `/plan`：切换 Plan 模式。
- `/skill`：管理 Skill 工作流。
- `/mcp`：管理 MCP 服务。
- `/model`：切换模型。
- `/resume`：恢复历史会话。
- `/rewind`：回滚会话上下文或相关改动。
- `/compact`：手动执行 CONTEXT CHECKPOINT 交接式压缩。

## 功能说明

- 交互式 TUI 和一次性 `run` 模式。
- 从配置、Memory、项目指令、MCP、Skill 和 CodeGraph 加载项目上下文。
- 文件读取、搜索、编辑和创建。
- 运行 shell 命令，用于构建、测试和诊断。
- 对写文件和命令执行提供权限规则。
- Plan 模式支持先规划再执行。
- 会话保存、恢复、分支、总结和回滚。
- CONTEXT CHECKPOINT 上下文交接：自动达到阈值或手动 `/compact` 时，生成面向接续模型的交接摘要；阈值可通过 `[agent] compact_ratio` 调整。
- 支持 DeepSeek 和 OpenAI-compatible provider。
- 按模型能力处理 DeepSeek thinking mode。
- MCP 外部工具集成。
- 本地 Skill 工作流。
- Memory 长期偏好和项目事实。
- 兼容 UTF-8、UTF-16、GB18030 等常见文本编码；Windows 中文命令输出和文件名会尽量保持可读。
- 机器人入口当前暂时关闭。

## 桌面安装包

桌面端通过 Windows 安装器分发：

[DeepSeek-Orca-Setup-1.0.22-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v1.0.22/DeepSeek-Orca-Setup-1.0.22-windows-amd64.exe)
