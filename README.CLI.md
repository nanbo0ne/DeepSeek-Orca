# DeepSeek-Orca CLI

DeepSeek-Orca CLI 是面向终端工作流的 AI 编程 Agent。它把模型对话、项目上下文、文件工具、shell、权限控制、MCP、Skill、记忆、CodeGraph、会话恢复和自动化接口整合为一个命令行程序，适合程序员在本地项目、远程服务器、CI 辅助脚本和日常终端工作流中使用。

本项目基于 Reasonix fork 改造而来，CLI 保留原有核心能力，并统一命名为 DeepSeek-Orca。

## How to use

从源码构建 CLI：

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

在项目根目录启动交互式会话：

```powershell
cd D:\your-project
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe chat
```

执行一次性任务：

```powershell
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe run "阅读这个项目并总结主要模块"
```

常用 slash command：

- `/init`：生成或更新项目记忆。
- `/plan`：切换 Plan 模式。
- `/skill`：管理 Skill。
- `/mcp`：管理 MCP 服务。
- `/model`：切换模型。
- `/resume`：恢复历史会话。
- `/rewind`：回滚会话上下文或代码改动。

## Features

- 交互式 TUI：显示对话、工具调用、审批、状态栏、tokens、缓存命中、费用和上下文状态。
- 一次性任务：适合脚本、管道和自动化。
- 项目上下文：读取当前目录、配置、环境变量、项目指令文件、记忆、MCP、Skill 和 CodeGraph。
- 文件工具：读取、搜索、编辑、创建文件，并记录可恢复的改动边界。
- Shell 工具：运行项目命令、测试、构建和诊断脚本。
- 权限控制：写文件、执行命令等操作可按规则询问、允许或阻止。
- Plan 模式：先生成计划，再在用户确认后执行高风险改动。
- MCP：通过外部 MCP 服务扩展工具能力。
- Skill：复用本地 Skill 指令和工作流。
- Memory：保存跨会话偏好、项目事实和长期上下文。
- CodeGraph：用于符号、调用关系和代码结构查询。
- 会话管理：保存、恢复、分叉、总结和回滚会话。
- Provider 配置：支持 DeepSeek 预设，也支持 OpenAI-compatible endpoint。
- DeepSeek thinking mode：按官方接口约定处理支持推理的模型与思考强度。

## Notes

CLI 更适合熟悉终端和 Git 工作流的用户。若你希望通过图形界面管理会话、项目、上下文和回滚，请使用 DeepSeek-Orca Desktop。
