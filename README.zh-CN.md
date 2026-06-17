# DeepSeek-Orca V2.0.0

DeepSeek-Orca 是基于 Reasonix fork 改造的 AI 编程 Agent，提供 Windows 桌面端和 CLI。V2 重点升级桌面端的上下文组装、增强模式、询问工作流、分步思考，以及新对话偏好继承。

## 下载

Windows 安装包：

[DeepSeek-Orca-Setup-2.0.0-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.0/DeepSeek-Orca-Setup-2.0.0-windows-amd64.exe)

安装后打开 DeepSeek-Orca Desktop，在设置里填写 DeepSeek API Key，或添加 OpenAI-compatible provider。

## V2 新功能

### 增强模式

输入框右下角新增“增强模式”按钮。开启后，会切换到一套独立的 Claude-like prompt/context 组装结构：

- 稳定系统提示词与动态项目记忆分离，减少动态内容对缓存前缀的影响。
- `DEEPSEEK_ORCA.md`、`AGENTS.md`、`CLAUDE.md` 等记忆文件以 `<system-reminder>` 注入用户消息，而不是折进 system prompt。
- 增强模式会在每轮重新读取记忆，让会话中修改的项目规则更快生效。
- 继续兼容现有记忆文件名，不新增 `deepseek.md`。

增强模式适合复杂代码任务、长上下文重构、架构分析、审查和需要更高回答质量的场景。代价是可能消耗更多 tokens，并降低提示缓存命中率。若当前对话上下文超过 50,000 tokens，中途切换时会弹出提醒：切换可能导致缓存无法命中，且模型表现可能无法发挥完全水平。

### 新对话偏好继承

新对话会继承“最近一次实际发生的对话”的模型、思考力度、审批力度和增强模式。这样你不用每次新建对话都重新选择常用模型和权限力度。

加号二级菜单里的临时协作选项不会继承：询问、分步思考、计划模式、目标模式等在新对话中默认关闭，避免上一个复杂任务的工作流状态影响下一个任务。

### 询问工作流

加号菜单新增“询问”开关，参考 `grill-me-codex` 做成完整工作流：

- 先只读探索仓库，能从代码和文档发现的信息不问用户。
- 只询问真正属于用户决策的问题，并且一次只问一个关键问题。
- 执行前锁定计划。
- 使用现有 planner/subagent 能力进行内部对抗式审查。
- 高风险修改前等待用户最终确认。

它适合需求还不完全清楚、产品取舍较多、影响范围较大的任务。

### 分步思考

加号菜单新增“分步思考”开关，参考 `superpowers` 做成阶段式复杂任务流程：

- 探索上下文。
- 构思可行方案。
- 选择并设计实现。
- 写出实施计划。
- 分任务执行。
- 任务级复查和最终复查。

如果“询问”和“分步思考”同时开启，分步思考会跳过 brainstorm/构思环节，避免重复规划。

## 桌面端能力

- 左侧按置顶、项目、独立工作区组织会话。
- 新建会话可绑定项目文件夹，也可创建独立工作区。
- 输入框内可切换模型、思考力度、审批力度、增强模式，以及询问/分步思考/计划/目标等协作方式。
- 支持文件、图片、工作区引用、命令引用和历史会话引用。
- 支持 ask、auto、yolo/完全访问三种工具审批力度。
- 支持 checkpoint、会话回滚、文件改动回滚，并通过 hash 检查避免覆盖用户后续修改。
- 右侧上下文面板显示上下文窗口、tokens、缓存命中、请求数、耗时和费用。
- token、花费、缓存等统计会随会话持久化，重启后恢复。
- 自动和手动 `/compact` 都使用 CONTEXT CHECKPOINT 交接式压缩。
- 支持 MCP、Skill、Memory、CodeGraph、slash commands。
- 支持 QQ/微信机器人连接，手机端可用 `/start` 选择最近会话，或 `/new` 新建独立工作区会话。

## CLI

从源码构建：

```powershell
git clone https://github.com/nanbo0ne/DeepSeek-Orca.git
cd DeepSeek-Orca
go build -o bin/deepseek-orca.exe ./cmd/deepseek-orca
```

配置 API Key：

```powershell
$env:DEEPSEEK_API_KEY="your DeepSeek API Key"
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

License: MIT.
