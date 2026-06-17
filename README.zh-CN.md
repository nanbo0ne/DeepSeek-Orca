# DeepSeek-Orca V2.0.5

DeepSeek-Orca 是基于 Reasonix fork 改造的 Windows 桌面端与 CLI AI 编程 Agent。它保留了原有的核心 Agent 循环、工具调用、MCP、技能、记忆、权限控制、会话恢复、检查点、上下文压缩和回滚能力，并围绕 DeepSeek 与 OpenAI-compatible provider 做了桌面端体验和 V2 工作流增强。

## 下载

Windows 安装包：

[DeepSeek-Orca-Setup-2.0.5-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.5/DeepSeek-Orca-Setup-2.0.5-windows-amd64.exe)

安装后打开 DeepSeek-Orca Desktop，在设置中填写 DeepSeek API Key，或添加兼容 OpenAI 接口的 provider。

## V2 重点功能

### 增强模式

V2 在发送按钮旁新增“增强模式”。开启后，当前会话会切换到一套独立的 Claude-like prompt/context 组装结构，适合复杂代码任务和长上下文工作：

- 稳定的核心系统提示词与动态项目记忆分离，减少动态内容对缓存前缀的影响。
- 项目记忆通过 `<system-reminder>` 用户消息块注入，而不是折进 system prompt。
- 增强模式会在每轮重新读取记忆文件，让会话中途修改的项目规则更快生效。
- 继续兼容现有记忆文件链：`DEEPSEEK_ORCA.md`、`AGENTS.md`、`CLAUDE.md` 以及对应的 local 变体，不新增 `deepseek.md`。

增强模式更适合复杂重构、架构分析、代码审查、长期上下文任务和需要更高回答质量的场景。代价是可能消耗更多 tokens，并降低提示缓存命中率。若会话已经进入较长上下文，中途切换模式也可能导致缓存无法命中，模型表现未必能发挥到最佳状态。

### 新对话偏好继承

新对话会继承最近一次实际发生对话的模型、思考力度、审批力度和增强模式。这样常用模型与权限策略只需要设置一次，后续新建对话可以直接沿用。

加号二级菜单中的临时协作选项不会继承。询问、分步思考、计划模式、目标模式等在新对话中默认关闭，避免上一个复杂任务的工作流状态影响下一个任务。

### 询问工作流

加号菜单中的“询问”开关用于更谨慎的需求澄清流程，参考 `grill-me-codex` 的思路实现：

- 先只读探索仓库，能从代码和文档发现的信息不再重复询问用户。
- 只询问真正需要用户决策的问题，并且一次只问一个关键问题。
- 执行前锁定计划。
- 需要时使用现有 planner/subagent 能力做内部对抗式审查。
- 高风险修改前等待用户最终确认。

这个模式适合需求边界不清、产品取舍较多、影响范围较大的任务。

### 分步思考

加号菜单中的“分步思考”开关用于阶段式推进复杂任务，参考 `superpowers` 的流程：

- 探索上下文。
- 构思可行方案。
- 选择并设计实现。
- 写出实施计划。
- 分任务执行。
- 做任务级复查和最终复查。

如果“询问”和“分步思考”同时开启，分步思考会跳过 brainstorm/方案构思环节，避免重复规划。

## V2.0.5 修复

V2.0.5 修复了大项目分析时上下文窗口统计异常陡增的问题。现在右上角上下文统计恢复为 provider 返回的真实 `PromptTokens` 口径，普通模式和增强模式一致，避免把本地 reasoning 显示内容、工具输出或历史估算误算进当前上下文窗口。

同时，DeepSeek 的 `reasoning_content` 回传范围被收窄：普通 assistant reasoning 只保留在本地显示和历史中，不再进入后续请求；只有带 `tool_calls` 的 DeepSeek assistant 轮次会按需回传。这可以降低不必要的 prompt token、缓存失效和压缩异常风险。

## 桌面端能力

- 左侧栏支持项目会话、独立工作区会话、置顶、历史、回收站和话题管理。
- 输入框中可直接切换模型、思考力度、审批力度、增强模式，以及询问、分步思考、计划、目标等协作方式。
- 支持文件、图片、工作区引用、命令引用和历史会话引用。
- 支持 ask、auto、yolo/完全访问等工具审批力度。
- 支持 checkpoint、会话回滚、文件改动回滚，并通过 hash 检查避免覆盖用户后续修改。
- 右侧上下文面板显示上下文窗口、token 消耗、缓存命中、请求数、耗时和费用。
- token、花费、缓存等统计会随会话持久化，应用重启后可恢复。
- 支持自动和手动 `/compact`，使用 CONTEXT CHECKPOINT 交接式压缩。
- 支持 MCP、本地技能、记忆文件、CodeGraph 和 slash commands。
- 支持 QQ/微信机器人连接，可在手机端使用 `/start` 选择最近会话，或使用 `/new` 新建独立工作区会话。

## CLI

从源码构建：

```powershell
git clone https://github.com/nanbo0ne/DeepSeek-Orca.git
cd DeepSeek-Orca
go build -o bin/deepseek-orca.exe ./cmd/deepseek-orca
```

配置 API Key 并启动会话：

```powershell
$env:DEEPSEEK_API_KEY="your DeepSeek API Key"
.\bin\deepseek-orca.exe setup

cd D:\your-project
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe chat
```

执行一次性任务：

```powershell
D:\path\to\DeepSeek-Orca\bin\deepseek-orca.exe run "阅读这个项目并总结主要模块"
```

## 安装目录

Windows 默认安装位置通常为：

```text
C:\Users\<your-user>\AppData\Local\Programs\DeepSeek-Orca
```

主要文件：

- `deepseek-orca-desktop.exe`：桌面应用主程序。
- `uninstall.exe`：安装器生成的卸载程序。
- `uninstall.bat`：备用卸载脚本。
- `node.exe`：随应用携带的 Node runtime。
- `dist/`：桌面端前端资源。
- `.deepseek-orca/`：配置、凭据引用、技能、MCP、缓存和本地数据。
- `data/`：会话、历史索引、工作区元数据和统计信息。

DeepSeek-Orca 基于 Reasonix fork。License: MIT.
