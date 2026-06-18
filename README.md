# DeepSeek-Orca V2.0.14

DeepSeek-Orca 是基于 Reasonix fork 改造的 Windows 桌面端与 CLI AI 编程 Agent。它保留核心 Agent 循环、工具调用、MCP、Skill、记忆、权限控制、会话恢复、检查点、上下文压缩和回滚能力，并围绕 DeepSeek / OpenAI-compatible provider 做了桌面端体验增强。

## 下载

Windows 安装包：

[DeepSeek-Orca-Setup-2.0.14-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.14/DeepSeek-Orca-Setup-2.0.14-windows-amd64.exe)

## V2.0.14 新功能

### 底部状态栏审批状态更清晰

底部状态栏现在会稳定显示当前审批权限，并用不同颜色区分三种状态：`需要批准` 使用蓝色，`自动批准` 使用绿色，`完全访问权限` 使用红色。该修复避免了浅色背景下权限文字不可见的问题，也让用户不需要悬停 tooltip 才能确认当前权限模式。

### 模型切换更丝滑

切换模型时，按钮和底部状态栏会立即显示短模型名，例如 `deepseek-v4-pro` 或 `deepseek-v4-flash`，不再在后台控制器重建期间闪现 `provider/model` 路径。快速切换时以前端最后一次选择为准；真正发送消息前会等待后台切换完成。

### 自动化只用于持续任务

自动化功能被重新定义为明确的重复性、持续性、后台监控任务，例如：

- 每 30 分钟检查一次构建状态。
- 每天固定时间提醒整理日报。
- 持续监控日志并在失败时提醒。

普通一次性计时需求不会被当成自动化创建。`automation_create` 的 schema 已移除一次性 `delay_seconds` / `run_at` 创建路径，改为 `interval_seconds`、`daily_time`、`weekly_day + weekly_time`、`monitor` 等重复或持续规则。

### 自动化管理面板

左上角“机器人”按钮右侧新增“自动化”入口。点击后可以查看本地持久化自动化，支持：

- 查看运行中、等待中、已暂停、失败、已取消状态。
- 暂停 / 恢复自动化。
- 取消自动化。
- 清理已结束记录。
- 刷新状态。

自动化记录会保存在本地配置中，应用重启后仍可查看。Windows 通知不再 fallback 到居中的 `msg` 弹窗，避免任务完成时突然弹出阻塞式系统窗口。

### 右侧栏新增“侧边聊天”

右侧栏新增第四个页签“侧边聊天”。它是一个只读问答窗口，可以读取当前主对话内容，并重点留意最近 1-2 轮上下文。适合在不打断主任务的情况下询问：

- 刚才 AI 为什么这么做？
- 当前任务进展到哪一步？
- 这段工具输出是什么意思？
- 能不能解释上一轮回复里的某个决策？

侧边聊天有独立历史和清空按钮；清空只删除侧聊记录，不影响主对话。侧边聊天不会写入主对话 history，不参与主对话 token 统计、压缩、标题生成，也不会暴露写文件、shell、host 系统操作或自动化创建能力。

### 斜杠菜单中文化

中文环境下 `/skill`、`/hooks`、`/model`、`/effort`、`/mcp` 等斜杠菜单说明进一步中文化。命令本体仍保持英文，说明和 hint 尽量使用中文，减少同一菜单里中英文混杂的问题。

## V2 重点能力

### 增强模式

增强模式位于发送按钮附近。开启后当前会话会切换到独立的 Claude-like prompt/context 组装结构，适合复杂代码任务、长上下文工作、架构分析、代码审查和高质量回答场景。

增强模式的记忆注入方式与普通模式不同：核心系统提示词保持稳定，项目记忆通过 `<system-reminder>` 用户消息块动态注入。继续兼容 `DEEPSEEK_ORCA.md`、`AGENTS.md`、`CLAUDE.md` 及 local 变体，不新增 `deepseek.md`。

代价是可能消耗更多 token，并降低提示缓存命中率。长上下文中途切换模式不一定能发挥最佳效果。

### 偏好继承

新对话会继承最近一次实际发生对话的模型、思考强度、审批力度和增强模式。加号菜单里的询问、分步思考、计划模式、目标模式等临时协作选项不会继承，默认关闭。

### 询问与分步思考

询问工作流用于需求不清或风险较高的任务：先探索，再只问无法从代码和文档中发现的关键问题，锁定计划后再执行。

分步思考工作流用于复杂任务：探索上下文、构思方案、设计计划、分任务执行、任务级复查、最终复查。如果询问和分步思考同时开启，分步思考会跳过 brainstorm 环节，避免重复规划。

## 宿主工具库

DeepSeek-Orca 默认启用宿主工具库。模型可以在现有审批体系下使用：

- 原生宿主命令：通过 Windows `cmd.exe` / PowerShell 执行系统命令，避免 Git Bash 参数改写。
- 系统与进程：系统信息、进程列表、结束进程、启动应用、文本剪贴板。
- 通知与自动化：直接通知使用 `notify_user`；明确重复、持续、监控任务使用 `automation_create`。
- 联网搜索：不知道具体 URL 时使用 `web_search`，已有 URL 时使用 `web_fetch`。
- 运行时工具：Node / Python 轻量执行，用于脚本、数据处理、依赖探测和结构化输出。
- 文档工具：Word / PPT / Excel / PDF 的基础检查与文本提取。

当前不包含截图识图、OCR、坐标点击、键盘输入或视觉桌面控制。

## 构建

```powershell
git clone https://github.com/nanbo0ne/DeepSeek-Orca.git
cd DeepSeek-Orca
npm run build --prefix desktop/frontend
go test ./...
```

桌面端 Windows 安装包使用仓库脚本构建。

License: MIT.
