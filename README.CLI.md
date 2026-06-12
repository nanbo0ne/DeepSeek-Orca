# DeepSeek-Orca CLI

DeepSeek-Orca CLI 是面向终端的 AI 编程 Agent。它把模型对话、项目上下文、文件工具、shell、权限控制、MCP、Skill、记忆、CodeGraph、会话恢复和自动化接口整合为一个命令行程序，适合程序员在本地项目、远程服务器、CI 辅助脚本和日常终端工作流中使用。

## 核心定位

CLI 版适合已经习惯终端和 Git 工作流的用户。它可以在当前目录理解项目，调用工具读取或修改文件，执行命令，审查 diff，保存会话，并通过配置和命令组合进入更可控的自动化流程。

CLI 端适合：

- 在终端中直接与 AI 编程 Agent 协作。
- 在远程机器或服务器上处理代码任务。
- 把一次性 AI 任务接入脚本。
- 对项目进行长期、多轮、可恢复的代码协作。
- 使用 MCP、Skill、CodeGraph 和项目记忆扩展 Agent 能力。
- 需要比桌面端更透明、更可组合、更容易自动化的开发者工作流。

## 交互模式

DeepSeek-Orca CLI 支持交互式和一次性两类主要工作方式。

交互式会话适合连续讨论、逐步修改和长期任务。它会打开 TUI，显示对话、工具调用、审批、状态栏、模型信息、tokens、缓存命中、费用和上下文状态。

一次性任务适合脚本和自动化。用户给出一个任务，DeepSeek-Orca 执行后输出结果并退出。

CLI 可以处理：

- 普通自然语言任务。
- 代码阅读与解释。
- 文件修改。
- 测试运行与修复。
- Git diff 审查。
- 日志分析。
- 管道输入。
- 脚本化批处理。

## 项目上下文

DeepSeek-Orca 会围绕当前工作目录构建项目上下文。

它会使用：

- 当前目录。
- 项目配置文件。
- 用户级配置。
- 环境变量。
- `.env`。
- AGENTS.md、DEEPSEEK_ORCA.md、CLAUDE.md 等指令文件。
- 记忆系统。
- 已启用的 MCP 和 Skill。
- CodeGraph 和 LSP 等代码智能工具。

这些上下文会合并到 Agent 的系统提示词和可用工具中，让它知道项目约定、可用能力和操作边界。

## 模型与 provider

CLI 支持 DeepSeek 和 OpenAI 兼容模型服务。

配置能力包括：

- 默认模型。
- 多 provider。
- provider 名称。
- base URL。
- 模型列表。
- API Key 环境变量。
- 余额查询 URL。
- 定价信息。
- 默认 effort。
- 子 Agent 模型和 effort。
- 独立 planner 模型。
- 网络代理。
- no_proxy。

模型引用通常采用 `provider/model` 形式。DeepSeek-Orca 会根据配置解析 provider、模型、密钥和调用参数。

## DeepSeek 能力

DeepSeek-Orca 包含 DeepSeek 相关预设和模型适配。

包括：

- DeepSeek 官方 API 接入。
- DeepSeek 模型预设。
- DeepSeek reasoner / chat / flash 类模型配置。
- 思考模式与 effort 映射。
- 余额查询支持。
- 中文提示词与中文界面文本。

对于支持思考模式的模型，DeepSeek-Orca 会尽量按模型服务要求传递思考相关参数；不支持的模型会使用默认行为。

## 工具系统

CLI 的核心能力来自工具调用。Agent 可以根据任务调用工具读取上下文、修改文件、运行命令和管理任务状态。

内置工具包括：

- 读取文件。
- 写入文件。
- 编辑文件。
- 多处编辑。
- 删除范围。
- 删除符号。
- 列目录。
- glob 搜索。
- grep 搜索。
- shell/bash/PowerShell 命令。
- 后台任务。
- notebook 编辑。
- todo 管理。
- 完成步骤记录。
- 文件预览。
- Web fetch。
- 工作区相关工具。

工具调用会受到权限、沙箱、工作区和 Plan 模式限制。

## 权限控制

CLI 提供权限边界，避免 Agent 未经确认执行高风险操作。

相关能力包括：

- 只读分析。
- 写操作审批。
- shell 命令审批。
- yolo/full access 模式。
- 工作区限制。
- 文件系统沙箱。
- 工具 allow/deny。
- Plan 模式下的写入阻止。
- 运行中取消。
- 重复写入保护。

权限系统适合区分探索、修改、批处理、迁移和高风险自动化任务。

## Plan 与双模型规划

DeepSeek-Orca 支持计划模式和双模型规划。

Plan 模式用于先生成方案，再进入执行。双模型模式可以使用 planner 模型负责规划，再把计划交给 executor 模型执行。

适合场景包括：

- 大范围重构。
- 跨模块修改。
- 数据结构迁移。
- 不确定 bug。
- 需要先评估影响面的任务。
- 需要明确步骤、风险和验证方法的任务。

## 会话保存、恢复与分叉

CLI 会把会话保存为 JSONL 文件，便于恢复和审计。

支持能力包括：

- 新建会话。
- 恢复历史会话。
- 继续当前会话。
- 会话列表。
- 会话预览。
- 分支和 fork。
- 会话树。
- 标题和元数据。
- 会话压缩。
- 会话迁移。
- 删除与回收。

保存格式是按消息逐条写入的 JSONL，适合排查和后续兼容。

## 回滚与 checkpoint

DeepSeek-Orca CLI 支持 checkpoint 与 rewind。

能力包括：

- 按用户轮次记录 checkpoint。
- 恢复某轮之后的代码改动。
- 截断对话到某个轮次。
- 同时回滚代码和对话。
- 从某轮分叉出新会话。
- 生成某段会话总结。
- 保护后续用户改动，避免无提示覆盖。

回滚用于处理 AI 修改方向错误、误操作、实验性重构失败或需要回到早期讨论状态的场景。

## 上下文压缩

长会话会逐渐消耗模型上下文。DeepSeek-Orca 支持自动压缩旧上下文。

压缩会保留：

- 用户目标。
- 已做决策。
- 关键文件和代码事实。
- 已运行命令与结果。
- 遇到的问题和修复。
- 待办与下一步。

压缩后的摘要会继续作为后续 Agent 的上下文，减少重新阅读和重复推理。

## Token、缓存和费用统计

CLI 会从模型返回的 usage 中提取运行指标。

包括：

- prompt tokens。
- completion tokens。
- reasoning tokens。
- total tokens。
- cache hit tokens。
- cache miss tokens。
- 当前缓存命中率。
- 平均缓存命中率。
- 会话 tokens。
- 费用估算。
- 上下文窗口占用。
- 压缩阈值。

这些指标会显示在 TUI 状态栏或事件流中，用于判断上下文健康、缓存效果和本次任务成本。

## MCP 扩展

DeepSeek-Orca CLI 支持 MCP 服务器。

能力包括：

- 从配置加载 MCP。
- 导入 `.mcp.json`。
- 添加 stdio / HTTP / SSE MCP。
- 启用和禁用 MCP。
- 查看连接状态。
- 查看工具、资源和 prompts。
- 管理认证。
- 失败诊断。
- 重试连接。
- 清理无效配置。
- 安装和卸载 MCP 能力。

MCP 可以扩展数据库、浏览器、文档、内部系统、代码服务和团队工具。

## Skill 系统

Skill 是可复用的任务 playbook。DeepSeek-Orca 会扫描内置、项目、自定义和用户级 Skill。

Skill 支持：

- `SKILL.md` 目录布局。
- 扁平 Markdown 兼容。
- frontmatter。
- references。
- scripts。
- assets。
- 内联 Skill。
- 子 Agent Skill。
- Skill 启用/停用。
- Skill 搜索、查看和删除。
- Skill 来源管理。

内置 Skill 包括：

- init：生成或刷新项目指令文件。
- explore：探索代码库并返回聚焦结论。
- research：结合代码和网页资料调研。
- install-capability：安装或卸载 MCP 与 Skill。
- review：代码审查。
- security-review：安全审查。
- test：运行测试并修复失败。

## 记忆系统

DeepSeek-Orca 支持项目和用户级记忆。

能力包括：

- 保存偏好。
- 保存项目事实。
- 保存团队约定。
- 读取记忆文档。
- 快速追加记忆。
- 删除过期记忆。
- 在会话启动时合并相关记忆。

记忆适合保存长期有效的信息，例如构建命令、代码风格、目录约定、测试策略和用户偏好。

## CodeGraph

CodeGraph 是内置代码智能引擎，以 MCP 工具形式提供符号级代码理解。

能力包括：

- 符号搜索。
- 入口点和相关符号聚合。
- 调用方查询。
- 被调用方查询。
- 改动影响分析。
- 调用路径追踪。
- 文件树和符号统计。

CodeGraph 适合回答架构、调用链、影响面和“某个功能如何工作”这类问题，比纯 grep 更适合符号级分析。

## LSP 工具

DeepSeek-Orca 支持可选 LSP 工具。

能力包括：

- definition。
- references。
- hover。
- diagnostics。

LSP 服务器不强制随程序安装。启用后会从 PATH 或配置路径解析，对应语言服务器可用时才生效。

## Slash Command

CLI 支持 slash command 和项目级命令。

命令来源包括：

- 内置命令。
- `.deepseek-orca/commands/`。
- 项目配置。
- MCP prompts。
- Skill。

Slash command 适合沉淀团队常用流程，例如 review、生成 release note、检查配置、运行测试方案或执行项目特定操作。

## TUI 界面

交互式 CLI 使用 TUI 展示会话。

能力包括：

- 流式回复。
- 工具调用卡片。
- 折叠和展开工作流。
- approval 请求。
- ask 请求。
- 背景任务状态。
- 状态栏。
- 模型切换。
- effort 切换。
- 主题。
- 复制内容。
- 滚动历史。
- resume picker。
- MCP 管理器。
- Skill picker。
- memory view。
- model view。
- hooks view。

## 附件与引用

CLI 支持把外部内容带入会话。

能力包括：

- 文件引用。
- 目录引用。
- 图片/媒体引用。
- 粘贴图片。
- 管道输入。
- workspace 文件补全。
- 引用解析。

引用能力用于把局部上下文精确交给 Agent，减少无关文件进入上下文。

## Bot 网关

DeepSeek-Orca 包含多渠道 Bot 网关配置。

支持能力包括：

- QQ Bot。
- 飞书/Lark Bot。
- 微信 iLink Bot。
- webhook。
- websocket。
- allowlist。
- debounce 消息合并。
- token 环境变量。
- bot 专用模型。

Bot 网关用于把 DeepSeek-Orca Agent 接入 IM 或团队消息系统。

## ACP / 服务接口

DeepSeek-Orca 包含 ACP 和 serve 相关能力，用于把 Agent 作为服务或协议端点使用。

能力包括：

- ACP 协议分发。
- 会话控制。
- 事件广播。
- 附件处理。
- slash command 控制。
- 权限请求。
- replay pending。
- 分支与回滚控制。

这些能力适合被桌面端、外部控制器或自动化系统复用。

## 配置系统

DeepSeek-Orca 配置采用 TOML。

配置范围包括：

- 默认模型。
- provider。
- 网络代理。
- 通知。
- UI。
- 桌面端偏好。
- 权限。
- 沙箱。
- MCP。
- Skill。
- 记忆。
- LSP。
- CodeGraph。
- Bot。
- hooks。
- statusline。
- auto plan。
- output style。

配置解析支持项目级和用户级配置，密钥优先通过环境变量读取。

## 网络与代理

DeepSeek-Orca 支持网络代理配置。

能力包括：

- 自动代理。
- 环境变量代理。
- 自定义代理。
- no_proxy。
- 对国内/特定 provider 的直连策略。
- provider 级请求复用。

这用于处理公司代理、跨网络访问、模型服务和 MCP 访问。

## 通知与 hooks

CLI 支持通知和 hooks。

通知可用于：

- turn 完成。
- 需要 approval。
- 需要用户回答。

Hooks 可用于在用户提交、任务停止等节点执行外部逻辑，适合团队规范检查、审计和集成自动化。

## 审查与安全

内置审查能力包括普通代码审查和安全审查。

普通 review 关注：

- 正确性 bug。
- 边界条件。
- 行为回归。
- 缺失测试。
- 有实际影响的一致性问题。

安全 review 关注：

- 注入。
- 路径穿越。
- 认证授权缺失。
- secret 泄漏。
- 不安全反序列化。
- XSS、SSRF、开放重定向。
- cookie 和限流问题。

## 测试修复

内置 test Skill 用于运行测试、诊断失败、修改代码或测试并重新运行。

它会区分：

- 生产代码 bug。
- 测试本身错误。
- 环境问题。
- 缺依赖。
- 工具链问题。

它不会为了变绿而跳过、删除或禁用失败测试。

## 适合的工作流

推荐流程：

- 先保持 Git 工作区可审查。
- 使用 CLI 让 Agent 阅读项目结构。
- 复杂任务先要求计划。
- 逐步执行，每步后运行测试。
- 使用 diff 审查 AI 改动。
- 需要时使用 rewind 回滚。
- 将稳定流程沉淀为 slash command 或 Skill。

## 目录与文件能力

常见项目文件包括：

- `cmd/deepseek-orca/`：CLI 入口。
- `internal/agent/`：Agent 运行循环、上下文、压缩、任务和会话。
- `internal/control/`：控制器、权限、输入、回滚、分支和会话管理。
- `internal/tool/`：内置工具。
- `internal/config/`：配置加载、渲染和迁移。
- `internal/skill/`：Skill 发现、索引和调用。
- `internal/memory/`：记忆文档和长期事实。
- `internal/codegraph/`：CodeGraph 集成。
- `internal/plugin/`：MCP 插件系统。
- `internal/cli/`：命令行与 TUI。
- `internal/serve/`：服务接口。
- `internal/bot/`：IM Bot 网关。
- `npm/deepseek-orca/`：npm 分发包装。
