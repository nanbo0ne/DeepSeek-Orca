# DeepSeek-Orca Desktop

DeepSeek-Orca Desktop 是面向 Windows 的 AI 编程 Agent 工作台。它把项目会话、文件改动、回滚、上下文统计、模型配置、MCP、Skill 和 Memory 放在一个桌面界面里。

本项目基于 Reasonix fork 改造，保留原有 Agent 能力，并重构为 DeepSeek-Orca 的桌面体验。

## How to use

1. 下载 Windows 安装器：

   [DeepSeek-Orca-Setup-1.0.22-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v1.0.22/DeepSeek-Orca-Setup-1.0.22-windows-amd64.exe)

2. 双击安装器，阅读并同意许可协议，选择安装目录。
3. 启动 DeepSeek-Orca Desktop。
4. 按引导填写 DeepSeek API Key，或在设置中配置 OpenAI-compatible provider。
5. 点击左侧新建对话。
6. 选择进入项目文件夹，或使用独立工作区。
7. 在底部输入任务，例如“解释这个项目结构”或“修复当前测试失败”。

Agent 正在运行时：

- 按 `Enter` 会把输入加入待发送队列。
- 按 `Ctrl+Enter` 会把输入作为“引导”发送给正在运行的 Agent。
- 模型流式输出时不会强制锁住滚动；点击对话区的向下圆形按钮才会重新跟随最新输出。

## 功能说明

- 会话侧栏：按置顶、项目、独立工作区分组展示会话。
- 项目绑定：会话可以绑定项目文件夹，也可以不属于任何项目。
- workflow 折叠：推理、工具调用、审批、后台任务和 checkpoint 默认折叠，需要时可展开查看。
- 回滚：删除 AI 回复时，可以回滚该回复关联的文件改动；如果文件已经被用户后续修改，会阻止自动回滚。
- 运行中排队：模型还在思考时，新输入不会丢失，可以排队或直接引导。
- 自由滚动：模型生成回复时用户可以继续翻看上文，需要回到底部时再点击“跟随最新输出”按钮。
- 上下文统计：显示当前上下文窗口、token、缓存命中、请求数、耗时和费用；统计信息会持久化，重启后恢复。
- 统计修正：右上角上下文百分比基于当前会话消息重新估算，不再使用 provider 的计费 prompt tokens，避免思考过程中突然暴涨再恢复。
- CONTEXT CHECKPOINT：当对话接近阈值时生成面向接续模型的交接摘要；手动 `/compact` 也执行同样的交接式压缩。
- 编码兼容：读取文件和工具输出时兼容 UTF-8、UTF-16、GB18030；bash 流式输出会避免把中文文件名拆成乱码。
- 权限模式：支持 review、auto、yolo，其中 yolo 表示完全访问权限。
- Plan 模式：Plan 开关独立于权限模式。
- MCP 与 Skill：支持外部 MCP 工具和本地 Skill 工作流。
- Memory：保存偏好、项目事实和长期上下文。
- CodeGraph：在可用时帮助查询符号和代码结构。
- 机器人：入口当前暂时关闭，后续版本会重新开放 QQ / 微信连接。

## 安装后文件

常见安装目录：

```text
C:\Users\<你的用户名>\AppData\Local\Programs\DeepSeek-Orca
```

主要文件：

- `deepseek-orca-desktop.exe`：桌面端主程序。
- `uninstall.exe`：安装器生成的卸载程序。
- `uninstall.bat`：备用卸载脚本。
- `node.exe`：随程序携带的 Node 运行时。
- `dist/`：桌面端前端资源。
- `.deepseek-orca/`：配置、凭据引用、Skill、MCP、缓存等本地数据。
- `data/`：会话、历史记录、索引、工作区元数据和统计 telemetry。

卸载时可以选择是否删除已保存数据。
