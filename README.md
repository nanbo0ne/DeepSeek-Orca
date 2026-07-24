# DeepSeek-Orca

[中文](README.zh-CN.md) | [English](README.DESKTOP.md) | [桌面端更新记录](DESKTOP_CHANGELOG.md)

DeepSeek-Orca 是一个面向桌面工作流的本地 AI 助手与工程 Agent。它把多模型接入、项目工作区、工具调用、MCP、Skill、CodeGraph、长期记忆、计划执行和自动化集中在一个应用中，并保留可审查、可暂停、可回滚的交互方式。

## 下载

- [Windows 安装包](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.25/DeepSeek-Orca-Setup-2.0.25-windows-amd64.exe)
- [Windows 便携版](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.25/DeepSeek-Orca-windows-amd64.zip)
- [macOS 通用 DMG](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.25/DeepSeek-Orca-darwin-universal.dmg)
- [Linux amd64 DEB](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v2.0.25/DeepSeek-Orca-linux-amd64.deb)
- [全部发布文件](https://github.com/nanbo0ne/DeepSeek-Orca/releases/tag/desktop-v2.0.25)

首次启动后，在“设置 > 模型”中配置 API Key、服务地址和模型。多模态识图默认关闭，只有主动开启后图片才会上传给当前模型服务商。

## 三种模式

- **助手模式**：以 `Orca` 身份处理日常问答、轻量任务和陪伴式交流。它可以静默整理助手画像记忆，并在相关话题中联想用户偏好。
- **普通模式**：以 `DeepSeek-Orca` 身份处理查询、写作、分析和常规开发任务，兼顾速度与工具能力。
- **增强模式**：面向复杂代码、架构分析、长任务、代码审查和高质量推理，采用更主动的工程 Agent 工作方式。

模式保存在会话中。切换模式会重建该会话的控制器，但不会丢失历史、模型、思考强度、审批方式、计划和目标状态。

## 模型与 Provider

DeepSeek-Orca 支持 DeepSeek、OpenAI-compatible 服务和 Anthropic 协议服务。可以为主执行模型和规划模型分别配置 Provider、模型名称、上下文窗口、思考强度及价格信息。

- 模型可按会话切换，新会话继承最近使用的偏好。
- 支持单模型执行和双模型“规划器 + 执行器”结构。
- 余额查询异步加载，不会阻塞会话打开和历史恢复。
- 多模态识图支持聊天附件和工作区图片引用；模型本身仍需支持图片输入。

## 工作区与会话

- **项目工作区**绑定真实目录，适合持续开发、Git 操作和项目级记忆。
- **独立工作区**为每个话题建立隔离目录，适合临时任务和不应污染项目的对话。
- 会话支持固定、重命名、历史预览、恢复、分支、回收站和检查点回滚。
- 新会话按创建时间显示在所属工作区顶部。
- 上下文过长时会自动压缩；长期对话检索可以按需查找更早的本地对话片段。

## 多模态识图

在“设置 > 模型”中开启后，可以粘贴、拖入或通过 `@` 引用 PNG、JPEG、WebP 和 GIF 图片。

- 每轮最多 8 张，总大小不超过 20 MB，单张不超过 10 MB。
- 会话只保存本地附件路径、名称和 MIME，不保存 base64。
- OpenAI-compatible 和 Anthropic 请求会使用各自的图片消息格式。
- 不支持视觉的模型会返回明确错误，不会静默删除图片后重试。
- 当前版本不包含截图识别、鼠标键盘控制或完整 computer-use。

## 工具库

“工具库”面板可以按组控制扩展工具是否注册并暴露给模型：

- 联网搜索与网页读取
- 系统信息、原生命令、进程、应用、剪贴板和通知
- Node REPL 与 Python REPL
- Word、PowerPoint、Excel 和 PDF 检查/提取
- 本地会话线程管理
- 长期对话搜索与分段读取
- 更积极调用工具的提示策略

关闭工具组后，模型看不到该组专用工具，对应路由提示也会移除。`bash` 继续作为构建、测试、Git 和包管理等开发任务的通用兜底。

## MCP、Skill 与 CodeGraph

- **MCP**：管理本地或远程 MCP Server、连接状态、授权、重试和工具启用状态。
- **Agent Skill**：按需加载工作流说明和领域能力，避免把所有规则永久塞入系统提示词。
- **CodeGraph**：使用真实的 `mcp__codegraph__...` 工具进行符号搜索、调用关系和架构上下文分析。
- **斜杠菜单**：提供内置命令、Skill 和 MCP Prompt 的统一入口。

## 记忆

记忆按用途分区：

- 助手模式只读写 `assistant` 画像记忆。
- 普通和增强模式读取助手记忆与共享 Agent 记忆，写入 `shared-agent`。
- 旧未分类记忆按共享 Agent 记忆兼容处理。

助手模式可以在切换或关闭会话后静默整理新增内容。生成失败不会阻塞退出或主对话，最多重试 5 次后忽略该批次；用户可以关闭自动记忆、关闭主动联想、删除单条记忆或清空助手记忆。

## 计划、Todo 与 Goal

- 计划模式先生成完整计划卡，用户批准后再执行，也可以要求整体重写计划。
- Todo 用于跟踪复杂多步骤任务的当前项和完成状态。
- Goal 模式可自动续跑目标；模型发出内部完成/阻塞标记后会立即停止，不把控制标记显示在正文中。
- 询问、计划、Todo 和待发送内容使用统一底部布局，不会遮挡“跳到最新”按钮。

## 过程展示

在“设置 > 常规”中选择：

- **精简**：思考和工具过程以单行白色圆角栏显示，默认折叠。
- **标准**：运行中按真实时间顺序显示过程卡，完成后原位折叠。
- **详细**：按真实时间顺序显示并默认展开过程详情。

正文、思考、工具、图片读取和压缩过程严格按实际事件顺序交错排列。回合真正收到 `TurnDone` 前不会显示复制、分支、总结和回溯等完成操作。用户阅读历史时，新增工具输出会保持当前可见锚点，不会强行跳回工具区或底部。

## 自动化、机器人与侧边聊天

- 自动化用于明确的周期任务、持续监控和提醒，可查看、暂停、恢复或取消。
- 机器人连接支持为渠道选择模型和助手/普通/增强模式。
- 侧边聊天位于右侧 Dock，可参考主会话内容但不写入主历史，也不参与主会话 token、压缩和标题生成。
- 后台 bash 和 subagent 会持续显示运行状态；当前回答依赖其结果时，模型必须先调用 `wait`。

## 权限与隐私

- 支持“需要批准”“自动批准”“完全访问”三档工具审批。
- 主机命令、文件写入和外部操作仍受工作区、沙箱及权限策略约束。
- 会话、记忆、索引、工作区元数据和设置默认保存在本地。
- API 请求会发送给用户配置的模型服务商；开启识图后，图片也会发送给该服务商。
- 使用完全访问权限前，请确认任务边界、工作目录和命令风险。

## 常见问题

**模型无法识图**：确认已开启多模态识图，并确认当前模型及服务商兼容图片消息格式。

**工具不存在**：检查工具库分组、MCP 连接、Skill 状态和当前模式；关闭的工具不会出现在模型 schema 中。

**后台任务已经结束但模型没有继续**：后台任务不会自动创建新用户轮次。依赖结果的当前任务应使用 `wait`，独立任务结果会在下一次用户消息时注入。

**大项目 Git 视图较慢**：只有打开“文件/改动”页签时才请求相关数据；仍可减少未跟踪文件或缩小工作区范围。

**配置或会话异常**：先在设置中检查 Provider、模型和工作区路径，再查看本地日志。删除数据前建议备份用户数据目录。

## 开发与构建

需要 Go、Node.js、npm、Wails v2；Windows 安装包还需要 NSIS。

```powershell
git clone https://github.com/nanbo0ne/DeepSeek-Orca.git
cd DeepSeek-Orca
npm install --prefix desktop/frontend
npm run test:all --prefix desktop/frontend
npm run build --prefix desktop/frontend
go test ./...
cd desktop
wails build
```

桌面配置位于 `desktop/wails.json`，发布工作流位于 `.github/workflows/release-desktop.yml`。

## 限制与许可

- 模型能力、上下文窗口、工具调用和视觉支持取决于所选 Provider。
- computer-use、屏幕截图理解和坐标级桌面控制尚未内置。
- 自动记忆是辅助能力，不应保存密码、密钥、支付信息或高敏感个人画像。

License: MIT.
