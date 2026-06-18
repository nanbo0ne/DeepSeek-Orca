# DeepSeek-Orca V2.0.11

DeepSeek-Orca 是基于 Reasonix fork 改造的 Windows 桌面端与 CLI AI 编程 Agent。它保留了原有的核心 Agent 循环、工具调用、MCP、技能、记忆、权限控制、会话恢复、检查点、上下文压缩和回滚能力，并围绕 DeepSeek 与 OpenAI-compatible provider 做了桌面端体验和 V2 工作流增强。

## 下载

Windows 安装包：

[DeepSeek-Orca-Setup-2.0.11-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.11/DeepSeek-Orca-Setup-2.0.11-windows-amd64.exe)

## V2.0.11 宿主工具库

V2.0.11 默认启用新的宿主工具库，不再额外放置“工具库拓展”开关。模型现在可以在现有审批体系下使用这些能力：

- 原生宿主命令：通过 Windows `cmd.exe` / PowerShell 执行系统命令，避免 Git Bash 参数改写，适合 `shutdown`、`taskkill` 等 Windows 命令。
- 系统与进程：读取系统信息、列出进程、结束进程、启动应用、读写文本剪贴板。
- 提醒与自动化：发送系统通知，创建本进程内的定时提醒或定时宿主命令任务。自动化任务按无人值守方式运行，不会卡在提问上。
- 联网搜索：新增 `web_search`，不知道具体 URL 时可以先搜索，再用 `web_fetch` 读取目标页面。
- 运行时工具：新增轻量 Node / Python 执行工具，适合脚本、数据处理、依赖探测和结构化输出。
- 文档工具：新增 Word / PPT / Excel / PDF 的基础检查与文本提取入口；复杂编辑可继续用 Python 运行时处理。

本版暂不包含截图识图、OCR、坐标点击、键盘输入或视觉桌面控制。审批语义同步调整为：`ask` 正常询问，`auto` 自动审批工具权限请求，`yolo` 是真正的工具全权限模式，会允许任何工具操作。

同时修复输入框底部参数行被长文本框遮挡的问题，当前审批、模型、思考强度等状态会稳定露出。

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

## V2.0.10 补丁

V2.0.10 重点修复自动标题和 shell 报错可读性：

- 增强模式下的 `<system-reminder>`、`<workflow-reminder>`、`<memory-update>` 等注入块不再污染会话标题。
- 首次真实用户消息与首次 AI 回复完成后，会用独立轻量对话自动生成会话标题，不写入主对话历史，不影响主上下文统计。
- 左侧会话右键菜单新增“AI 生成标题”，可手动重新生成当前会话标题。
- shell / bash 工具非零退出时会显示退出码、shell 类型、原始命令和简短原因，不再只显示 `exit status 1`。
- Windows Git Bash 下的 `shutdown /s /t ... /c ...` 会自动通过 `cmd.exe /c` 兼容执行；失败时给出参数解析或权限不足提示。
- 修复根目录测试中的 DeepSeek 默认 reasoning effort、Windows symlink 权限测试、serve 默认语言断言和中文 forget 输出兼容问题。

## V2.0.9 补丁

V2.0.9 在 V2.0.6 的基础上继续修补两个主界面交互问题：

- 下箭头按钮改为对话区域底部居中的白色圆形按钮，箭头使用蓝色；只要当前视图没有贴住最下方就稳定显示，不再只在滚动时闪烁一下。
- 用户消息气泡的复制按钮改为真正的悬浮层，不再在消息正文前方挤出一行空白。

同时继续包含 V2.0.6 的稳定性修复：

- 下箭头改成真正的贴底跟随按钮。它会在不贴底时出现，点击后持续跟随最后一行，只有用户主动上滚才停止。
- 左侧项目树和会话排序更稳定。用户选中的工作区会排在独立工作区前面，独立工作区固定在最底部；每个工作区里的会话按最近活动时间倒序排列。
- 运行中可切换模型、思考强度、审批、协作开关和增强模式。当前轮保持原状，但下一次发送会使用新设置。
- 增强模式切换不再把右上角上下文统计刷成异常高值。统计只在 provider 返回真实 usage 后更新。
- 运行中的 task / 工具卡会显示流动光效，能一眼看出任务还在继续，不会显得像卡住。
- 用户消息也增加了复制按钮，鼠标移到气泡上才出现，和 AI 消息的复制体验一致。
- 新对话继承最近一次实际使用的模型、思考强度、审批和增强模式；询问、分步思考、计划、目标这些临时工作流默认关闭。

V2.0.9 继续保留上下文统计修复和 reasoning 回传收窄：右上角上下文统计使用 provider 的真实 `PromptTokens`，DeepSeek 的 `reasoning_content` 只在需要的 tool-call 轮次回传，避免不必要的 token、缓存失效和压缩异常。

同时，DeepSeek 的 `reasoning_content` 回传范围被收窄：普通 assistant reasoning 只保留在本地显示和历史中，不再进入后续请求；只有带 `tool_calls` 的 DeepSeek assistant 轮次会按需回传。这可以降低不必要的 prompt token、缓存失效和压缩异常风险。

## 桌面端能力

- 左侧栏支持项目会话、独立工作区会话、置顶、历史、回收站和话题管理。
- 输入框中可直接切换模型、思考力度、审批力度、增强模式，以及询问、分步思考、计划、目标等协作方式。
- 支持文件、图片、工作区引用、命令引用和历史会话引用。
- 支持 ask、auto、yolo/完全访问等工具审批力度。
- 支持 checkpoint、会话回滚、文件改动回滚，并通过 hash 检查避免覆盖用户后续修改。
- 右侧上下文面板显示上下文窗口、token消耗、缓存命中、请求数、耗时和费用。
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
