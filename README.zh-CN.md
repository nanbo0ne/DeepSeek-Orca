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

## 核心能力

- 助手、普通、增强三种提示词模式，分别面向个性化助理、通用任务和复杂工程工作。
- DeepSeek、OpenAI-compatible 与 Anthropic Provider，多模型和双模型规划/执行。
- 项目工作区、独立工作区、会话恢复、分支、检查点、回滚和回收站。
- 可选多模态识图，支持粘贴、拖入和 `@` 引用工作区图片。
- 工具库、MCP、Agent Skill、CodeGraph、联网搜索、REPL 和文档工具。
- 助手画像记忆、共享 Agent 记忆、上下文压缩和长期对话检索。
- Plan、Todo、Goal、自动化、机器人连接和只读侧边聊天。
- 精简、标准、详细三档过程展示，正文与工具按真实时间顺序排列。
- 工具审批、工作区限制、沙箱和本地数据管理。

## 三种模式

- **助手模式**：以 `Orca` 身份处理日常问答、轻量任务和陪伴式交流。它可以静默整理助手画像记忆，并在相关话题中联想用户偏好。
- **普通模式**：以 `DeepSeek-Orca` 身份处理查询、写作、分析和常规开发任务，兼顾速度与工具能力。
- **增强模式**：面向复杂代码、架构分析、长任务、代码审查和高质量推理，采用更主动的工程 Agent 工作方式。

模式保存在会话中。切换模式会重建该会话的控制器，但不会丢失历史、模型、思考强度、审批方式、计划和目标状态。

## 工作区、多模态与工具

项目工作区绑定真实目录；独立工作区为每个话题建立隔离目录。会话支持固定、重命名、历史预览、恢复、分支和检查点回滚。上下文过长时会自动压缩，长期对话工具可以搜索并分段读取更早的本地内容。

开启多模态识图后，每轮最多发送 8 张 PNG、JPEG、WebP 或 GIF，总大小不超过 20 MB。会话只保存本地路径和 MIME，不保存 base64。不支持视觉的模型会返回明确错误；当前版本不包含截图、鼠标键盘控制或完整 computer-use。

工具库按组管理联网搜索、主机系统、Node/Python REPL、文档、线程和长期对话检索。关闭工具组后，该组工具和路由提示都会从模型上下文移除。`bash` 保留为构建、测试、Git 和包管理的通用兜底。

## MCP、Skill、CodeGraph 与记忆

MCP 面板管理 Server、授权、连接和工具状态；Agent Skill 按需加载工作流；CodeGraph 使用真实 `mcp__codegraph__...` 工具分析符号和调用关系。

助手模式只读写助手画像记忆。普通和增强模式读取全部记忆，写入共享 Agent 记忆。助手自动记忆在离开会话后静默运行，失败最多重试 5 次且不会阻塞退出。设置中可以关闭自动记忆或主动联想，也可以删除和清空助手记忆。

## 计划、过程与后台任务

计划模式先展示可批准或重写的完整计划；Todo 跟踪多步任务；Goal 可以自动续跑并在内部完成标记后停止。询问、计划、Todo 和待发送内容使用统一底部布局。

- **精简**：单行白色圆角过程栏，默认折叠。
- **标准**：运行中按顺序显示过程卡，完成后原位折叠。
- **详细**：按顺序显示并默认展开详情。

正文、思考、工具和压缩过程严格按事件顺序交错。只有收到 `TurnDone` 才显示完成操作。后台 bash/subagent 不会自动开启新模型轮次；当前回答依赖结果时必须先调用 `wait`。

## 自动化、机器人与侧边聊天

自动化用于周期任务、持续监控和提醒；机器人连接可选择模型及三种模式；侧边聊天可以参考主会话，但不写入主历史，也不参与主 token、压缩或标题生成。

## 权限、隐私与排障

工具审批包含需要批准、自动批准和完全访问。会话、记忆、索引和设置默认保存在本地；API 请求及已开启的图片输入会发送给用户配置的服务商。

- 无法识图：检查识图开关及当前模型能力。
- 工具不存在：检查工具库、MCP、Skill 和模式。
- 后台完成后不继续：后台完成只更新状态；依赖结果的任务应使用 `wait`。
- 大项目较慢：只在需要时打开文件/Git 页签，并减少无关未跟踪文件。

## 开发构建

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

License: MIT.
