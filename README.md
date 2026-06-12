# DeepSeek-Orca V1.0

DeepSeek-Orca 是一个基于 DeepSeek 工作流定制的 AI 编程 Agent。它提供桌面端和 CLI 两种入口：桌面端适合编程新手和日常项目维护，CLI 端适合程序员在终端、脚本和远程环境中使用。

本仓库当前以桌面端为主要交付物，同时保留 CLI、MCP、Skill、记忆、CodeGraph、权限控制、Plan 模式、会话恢复和工具调用等能力。

## 下载与安装

Windows 用户直接下载打包好的安装器：

```text
DeepSeek-Orca-Setup-<version>-windows-x64.exe
```

安装器会引导你同意许可协议并选择安装位置。默认安装到：

```text
C:\Users\<你的用户名>\AppData\Local\Programs\DeepSeek-Orca
```

安装后会包含桌面快捷方式、开始菜单入口和卸载入口。卸载时可以选择是否删除保存的数据；如果不勾选，配置、会话、记忆和缓存会尽量保留。

## 桌面端能力

- 多会话侧边栏：置顶会话、项目会话、独立工作区分组展示。
- 项目工作区：可以进入项目文件夹，也可以创建不绑定项目的独立对话。
- 文件读写与命令执行：Agent 能读取项目、修改文件、运行 shell，并按权限模式请求确认。
- 事务级回滚：AI 每次修改文件前会记录备份，点击“删除并回退”可删除最后一段 AI 回复并恢复本轮文件改动。
- workflow 折叠：思考、工具调用、权限审批、后台任务等默认折叠为一行，需要排查时可展开。
- Plan 模式独立开关：Plan 与权限模式分离，可和 review、auto、完全访问权限组合使用。
- 上下文与费用统计：显示 token、缓存命中、上下文占用和当前对话费用。DeepSeek V4 Flash/Pro 会按各自价格分别计算。
- MCP 与 Skill：支持外部 MCP 工具和可复用 Skill，适合扩展项目能力。
- CodeGraph：可选代码图谱能力，用于符号、调用关系和代码结构查询。
- 记忆与项目指令：读取 `DEEPSEEK_ORCA.md`、`AGENTS.md`、`CLAUDE.md` 等项目指令文件。
- 外观设置：支持主题、字体、字号、关闭窗口行为等桌面偏好。

## CLI 能力

CLI 入口为：

```bash
deepseek-orca
```

常用命令：

```bash
deepseek-orca setup
deepseek-orca chat
deepseek-orca run "修复这个项目里的测试失败"
deepseek-orca serve
deepseek-orca acp
```

CLI 支持交互式聊天、一次性任务、会话恢复、checkpoint/rewind、Plan 模式、权限规则、MCP、Skill、记忆、CodeGraph、Bot 网关和 HTTP/SSE 服务模式。

## 配置文件

项目级配置文件：

```text
./deepseek-orca.toml
```

用户级配置目录：

```text
~/.config/deepseek-orca/
```

Windows 上通常位于：

```text
C:\Users\<你的用户名>\AppData\Roaming\deepseek-orca
```

常见配置内容包括模型 provider、API key 环境变量、权限模式、MCP 插件、Skill 路径、CodeGraph、代理和桌面偏好。

## DeepSeek 费用计算

DeepSeek-Orca 会根据 usage 事件计算当前对话费用。DeepSeek V4 使用以下价格：

```text
deepseek-v4-flash:
缓存命中输入 0.02 元 / 1M tokens
缓存未命中输入 1 元 / 1M tokens
输出 2 元 / 1M tokens

deepseek-v4-pro:
缓存命中输入 0.025 元 / 1M tokens
缓存未命中输入 3 元 / 1M tokens
输出 6 元 / 1M tokens
```

计算公式：

```text
费用 = (缓存命中 tokens * 命中单价 + 缓存未命中 tokens * 未命中单价 + 输出 tokens * 输出单价) / 1,000,000
```

如果同一对话里切换 Flash 和 Pro，每一次模型调用都会按当时使用的模型价格计算并累加。

## 目录结构

```text
cmd/deepseek-orca/                    CLI 入口
desktop/                              Wails 桌面端
desktop/frontend/                     React 前端界面
desktop/build/windows/installer-go/   Windows 安装器
internal/agent/                       Agent 主循环
internal/provider/                    模型 provider
internal/config/                      配置加载与默认提示词
internal/tool/                        内置工具
internal/plugin/                      MCP 客户端
internal/skill/                       Skill 系统
internal/memory/                      记忆与指令文件
npm/deepseek-orca/                    npm 分发包装
docs/                                 文档与展示页
```

## 构建

CLI：

```bash
go build -o bin/deepseek-orca ./cmd/deepseek-orca
```

桌面端：

```bash
cd desktop
npm --prefix frontend install
wails build -webview2 embed
```

Windows 安装包：

```bash
cd desktop/build/windows/installer-go
go build -o ../../bin/DeepSeek-Orca-Setup-0.0.0-windows-amd64.exe .
```

## Fork 说明

DeepSeek-Orca V1.0 基于 `esengine/DeepSeek-Reasonix` 的 `main-v2` 分支改造，是一个 fork。当前版本在品牌、桌面界面、会话组织、回滚机制、中文提示词、DeepSeek V4 思考模式、费用统计和 Windows 安装体验上进行了定制。
