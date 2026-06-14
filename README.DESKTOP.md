# DeepSeek-Orca Desktop

DeepSeek-Orca Desktop 是面向 Windows 的 AI 编程 Agent 工作台。它把项目会话、文件改动、回滚、上下文统计、模型配置、MCP、Skill、Memory 和机器人渠道放在同一个桌面界面里。

本项目基于 Reasonix fork 改造，保留原有 Agent 能力，并重构为 DeepSeek-Orca 的桌面体验。

## How to use

1. 下载 Windows 安装器：

   [DeepSeek-Orca-windows-amd64-installer.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/latest)

2. 双击安装器，阅读并同意许可协议，选择安装目录。
3. 启动 DeepSeek-Orca Desktop。
4. 按引导填写 DeepSeek API Key，或在设置中配置 OpenAI-compatible provider。
5. 点击左侧新建对话。
6. 选择进入项目文件夹，或使用独立工作区。
7. 在底部输入任务，例如“解释这个项目结构”或“修复当前测试失败”。

Agent 正在运行时：

- 按 `Enter` 会把输入加入待发送队列。
- 按 `Ctrl+Enter` 会把输入作为“引导”发送给正在运行的 Agent。

## 功能说明

- 会话侧栏：按置顶、项目、独立工作区分组展示会话。
- 项目绑定：会话可以绑定项目文件夹，也可以不属于任何项目。
- 固定外观：桌面端使用统一的浅色 DeepSeek-Orca 风格；设置中只保留字号与字体。
- workflow 折叠：推理、工具调用、审批和后台任务默认折叠，需要时可展开查看。
- 回滚：删除 AI 回复时，可以回滚该回复关联的文件改动。
- 运行中排队：模型还在思考时，新输入不会丢失，可以排队或直接引导。
- 上下文统计：显示上下文窗口、token、缓存命中、请求数、耗时和费用。
- 自动压缩：当对话接近模型上下文窗口阈值时，较早内容会被摘要压缩；阈值可在设置的模型运行参数中调整，压缩后右侧上下文窗口会重新计算。
- 权限模式：支持 review、auto、yolo，其中 yolo 表示完全访问权限。
- Plan 模式：Plan 开关独立于权限模式。
- MCP 与 Skill：支持外部 MCP 工具和本地 Skill 工作流。
- Memory：保存偏好、项目事实和长期上下文。
- CodeGraph：在可用时帮助查询符号和代码结构。
- 机器人：支持 QQ 与微信；飞书/Lark 暂未开放。

## 机器人 How to use

### QQ

1. 打开设置，进入“机器人”。
2. 选择 QQ。
3. 点击“去申请”，打开 QQ 机器人官方后台。
4. 申请并复制 App ID / App Secret。
5. 回到 DeepSeek-Orca，填写 App ID / App Secret。
6. 默认使用“正式”环境；只有调试沙箱应用时才切到“沙箱”。
7. 点击“保存并连接”。DeepSeek-Orca 会先验证 QQ token，再自动在后台接收消息。

### 微信

1. 打开设置，进入“机器人”。
2. 选择微信。
3. 使用微信扫码并在手机上确认。
4. 连接成功后，DeepSeek-Orca 会保存本地 token，并验证微信 `getupdates` 是否可用。
5. 之后只要桌面端运行，微信机器人会自动在后台接收消息。

微信通道兼容 OpenClaw / iLink 返回的字符串或数字型消息 ID、用户 ID 与上下文 token。

QQ 和微信都不需要手动启动机器人网关，也不需要配置白名单。所有进入渠道的用户默认可用。

在机器人里发送 `/start` 会返回最近 15 条对话；输入数字即可进入对应对话。

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
- `data/`：会话、历史记录、索引和工作区元数据。

卸载时可以选择是否删除已保存数据。
