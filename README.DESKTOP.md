# DeepSeek-Orca Desktop

DeepSeek-Orca Desktop 是一个 Windows 桌面 AI 编程 Agent。它把项目会话、文件改动、回滚、上下文统计、模型配置、MCP、Skill、Memory 和机器人渠道放在同一个桌面工作台里。

本项目基于 Reasonix fork 改造，保留原有 Agent 能力，并重构为 DeepSeek-Orca 的桌面体验。

## How to use

1. 下载 Windows 安装器：

   [DeepSeek-Orca-windows-amd64-installer.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v1.0.13/DeepSeek-Orca-windows-amd64-installer.exe)

2. 双击安装器，阅读并同意许可协议，选择安装目录。
3. 启动 DeepSeek-Orca Desktop。
4. 按引导填写 DeepSeek API Key，或在设置中配置 OpenAI-compatible provider。
5. 点击左侧新建会话。
6. 选择进入某个项目文件夹，或使用独立工作区。
7. 在底部输入任务，例如“解释这个项目结构”或“修复当前测试失败”。

当 Agent 正在运行时：

- 按 `Enter` 会把输入加入待发送队列。
- 按 `Ctrl+Enter` 会把输入作为“引导”发送给正在运行的 Agent。

## 功能说明

- **会话侧栏**：按置顶、项目、独立工作区分组展示会话。
- **项目绑定**：会话可以绑定一个项目文件夹，也可以不属于任何项目。
- **workflow 折叠**：推理、工具调用、审批和后台任务默认折叠，需要时可展开查看。
- **回滚**：删除 AI 回复时，可以回滚该回复关联的文件改动。
- **运行中排队**：模型还在思考时，新输入不会丢失，可以排队或直接引导。
- **上下文统计**：显示上下文窗口、token、缓存命中、请求数、耗时和费用。
- **权限模式**：支持 review、auto、yolo 等模式，其中 yolo 表示完全访问权限。
- **Plan 模式**：Plan 开关独立于权限模式。
- **MCP 与 Skill**：支持外部 MCP 工具和本地 Skill 工作流。
- **Memory**：保存偏好、项目事实和长期上下文。
- **CodeGraph**：在可用时帮助查询符号和代码结构。
- **机器人渠道**：支持 QQ、微信、飞书/Lark。

## QQ Bot 配置

QQ 使用官方 Bot API，需要 App ID 和 App Secret。

1. 打开设置。
2. 进入机器人渠道区域。
3. 选择 **QQ**。
4. 点击 **去申请**，打开 QQ Bot 官方后台。
5. 将官方后台里的 **App ID** 和 **App Secret** 填入 DeepSeek-Orca。
6. 测试阶段选择 **沙箱**，正式使用时选择 **正式**。
7. 点击 **保存并连接**。

保存后，如果要真正接收 QQ 消息，还需要启动 bot 网关：

```powershell
cd D:\AI-Reasonix
go run .\cmd\deepseek-orca bot start --channels qq
```

测试时可以临时允许所有用户访问；正式使用建议配置白名单。

## 微信机器人说明

微信 iLink 不需要手动复制 App Secret。它的凭据是扫码登录后保存在本机的登录 Token。若界面显示登录 Token 已保存但发消息无响应，请检查 bot 网关是否启动，以及 allowlist 是否允许当前微信用户。

## 安装后的文件

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
- `data/`：会话、历史、索引和工作区元数据。

卸载时可以选择是否删除已保存数据。如果不删除，会尽量保留配置、会话、记忆和缓存。

## 当前版本

桌面端当前版本：**v1.0.13**。

本版本修复 GitHub Actions 桌面端发布流程，使 release 发布可以安全重试，并保留 v1.0.12 加入的 QQ Bot 引导式连接配置。
