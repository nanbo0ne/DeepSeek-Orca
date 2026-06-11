# DeepCode Desktop V1.0

DeepCode Desktop 是 DeepCode 的桌面版本，适合希望通过图形界面使用 AI 编程助手的用户。它保留了 Agent 读写文件、调用工具、管理上下文和执行任务的能力，同时用会话列表、设置面板、折叠工作流和回退按钮降低使用门槛。

## 适合的使用方式

- 打开一个项目，让 AI 帮你阅读、解释、修改代码。
- 管理多个项目里的 AI 对话。
- 上传文件或图片作为上下文。
- 让 AI 修改文件，并在结果不合适时回退。
- 使用 Plan 模式先确认方案，再执行改动。

如果你主要在终端里工作，或者想把 DeepCode 接入脚本，请使用 CLI 仓库。

## 主要功能

- DeepSeek 风格的蓝白界面和 DeepCode O 形标志。
- 左侧会话列表按置顶、项目、独立工作区分组。
- 新建会话时可选择是否进入项目。
- AI 思考、工具调用、工具结果默认折叠。
- Plan 模式独立开关。
- Review / Auto / YOLO 权限模式。
- AI 文件改动事务回退。
- 设置中可选择关闭窗口时后台运行或结束程序。
- Windows 安装器支持选择安装目录。
- 卸载器支持选择是否删除保存数据。

## 安装

Windows 安装包位于：

```text
desktop/build/bin/DeepCode-Setup-0.0.0-windows-amd64.exe
```

双击后按安装向导操作即可。安装过程中可以选择安装目录。

默认安装目录通常是：

```text
C:\Users\<你的用户名>\AppData\Local\Programs\DeepCode
```

安装后常见文件：

```text
deepcode-desktop.exe   主程序
node.exe               附带的 Node 运行时
uninstall.exe          标准卸载程序
uninstall.bat          兜底卸载脚本
```

## 第一次使用

1. 打开 DeepCode。
2. 进入设置，配置 DeepSeek 或其他兼容模型。
3. 新建会话。
4. 选择项目文件夹，或进入独立工作区。
5. 先让 AI 阅读项目，再让它修改文件。

比较稳妥的第一句话：

```text
请先阅读这个项目的结构，说明主要模块和启动入口。不要修改文件。
```

## 模型配置

在设置里的模型页面添加模型服务。使用 DeepSeek 时通常需要：

```text
API Key
服务地址
模型名称
```

建议把 API Key 当作敏感信息处理，不要截图或提交到仓库。

如果模型无法使用，优先检查：

- API Key 是否正确。
- 网络是否能访问模型服务。
- 模型名称是否存在。
- 当前是否还有任务在运行。

## 主界面

### 左侧会话栏

会话按三类展示：

1. 置顶会话。
2. 项目会话。
3. 独立工作区会话。

置顶只改变显示位置，不改变会话属于哪个项目。

### 中间对话区

这里显示用户消息、AI 回复和工作流记录。工具调用、思考过程和后台任务默认折叠，点击后可以展开查看。

### 底部输入框

左下角 `+` 按钮包含附件和模式入口：

- 上传文件。
- 上传图片。
- 引用文件或目录。
- 使用 slash command。
- 打开或关闭 Plan 模式。

右下角是发送按钮。AI 运行时会在同一位置显示运行状态和停止按钮。

### 右侧上下文

右侧用于查看项目文件、上下文占用和相关信息。上下文占用以横条展示，便于快速判断当前对话还能放入多少信息。

## 权限模式

权限模式控制 AI 能否自动调用工具：

- Review：更谨慎，适合不确定任务。
- Auto：常规自动模式。
- YOLO：更激进，只建议在明确知道风险时使用。

建议默认使用 Review 或 Auto。涉及删除、批量替换、重构时，先让 AI 给计划。

## Plan 模式

Plan 模式用于复杂任务。开启后，AI 更倾向于先说明步骤，再进入执行。

适合开启 Plan 的任务：

- 跨多个文件的修改。
- UI 重构。
- 配置或数据结构调整。
- 修 bug 但原因还不明确。

简单问答、解释代码、小范围改文案通常不需要开启。

## 回退机制

AI 修改文件前会保存备份。一个 AI 回复中的所有文件改动会记录为同一个事务。

如果最后一次结果不满意，可以使用“删除并回退”：

1. 删除最后一段 AI 回复。
2. 删除关联工作流展示。
3. 恢复这一轮 AI 修改过的文件。

如果文件在 AI 修改后又被手动改过，DeepCode 会提示冲突，避免覆盖新的改动。

## 设置建议

### 关闭窗口时

可选：

- 在后台运行：关闭窗口后程序继续留在后台。
- 结束程序：关闭窗口时退出应用。

不希望常驻后台时，选择“结束程序”。

### 思考过程显示

默认折叠更适合日常使用。需要排查问题时再展开。

### 字体

设置中可选择系统字体、微软雅黑、苹方、Noto Sans 和 Comic Sans。

## 从源码构建

需要准备：

- Go
- Node.js / npm
- Wails
- NSIS，用于 Windows 安装包

前端构建：

```powershell
cd desktop/frontend
npm run build
```

桌面端打包：

```powershell
cd desktop
wails build -nsis -webview2 embed
```

输出：

```text
desktop/build/bin/deepcode-desktop.exe
desktop/build/bin/DeepCode-Setup-0.0.0-windows-amd64.exe
```

## 仓库结构

```text
desktop/                         Wails 桌面端
desktop/frontend/                React 前端
desktop/frontend/src/components  UI 组件
desktop/frontend/src/locales     界面文案
desktop/build/windows/installer  Windows 安装器脚本
internal/                        桌面端依赖的 Agent 核心
docs/                            文档
```

## 使用建议

- 大改动前先提交 Git。
- 不确定时使用 Review 权限。
- 复杂任务先开 Plan。
- AI 修改后看 diff，再运行测试。
- 不要把 API Key 上传到 GitHub。

## 项目来源

DeepCode Desktop V1.0 基于 `esengine/DeepSeek-Reasonix` 的 `main-v2` 分支改造，是 Reasonix 的 fork。此版本在桌面 UI、会话组织、回退机制、提示词中文化、DeepSeek 风格配色、图标和 Windows 安装/卸载体验等方面进行了定制。
