# DeepSeek-Orca

DeepSeek-Orca 是基于 Reasonix fork 改造的 AI 编程 Agent，提供 Windows 桌面端和 CLI。桌面端适合日常使用，CLI 适合终端和自动化工作流。

## 下载

[DeepSeek-Orca-Setup-1.0.23-windows-amd64.exe](https://github.com/nanbo0ne/DeepSeek-Orca/releases/download/desktop-v1.0.23/DeepSeek-Orca-Setup-1.0.23-windows-amd64.exe)

## 快速开始

1. 安装并打开 DeepSeek-Orca Desktop。
2. 在设置中填写 DeepSeek API Key，或配置 OpenAI-compatible provider。
3. 新建对话，选择项目文件夹或独立工作区。
4. 输入任务，例如“分析这个项目结构”或“修复测试失败”。

## 主要能力

- 项目会话、独立工作区和置顶会话管理。
- 文件读取、搜索、编辑、创建和回滚。
- 运行 shell 命令并辅助构建、测试、诊断。
- Plan、review、auto、yolo 权限模式组合。
- MCP、Skill、Memory、CodeGraph。
- DeepSeek thinking mode 与 OpenAI-compatible provider。
- CONTEXT CHECKPOINT 自动/手动上下文交接压缩。
- 会话 token、缓存命中、耗时、费用等统计持久化。
- QQ/微信机器人连接，支持 `/start` 选择最近会话和 `/new` 新建独立工作区会话。

## CLI

```powershell
git clone https://github.com/nanbo0ne/DeepSeek-Orca.git
cd DeepSeek-Orca
go build -o bin/deepseek-orca.exe ./cmd/deepseek-orca
$env:DEEPSEEK_API_KEY="your DeepSeek API Key"
.\bin\deepseek-orca.exe setup
.\bin\deepseek-orca.exe chat
```

License: MIT.
