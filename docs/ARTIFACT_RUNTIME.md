# Work 产物运行时

DeepSeek-Orca 2.1.0 在助手模式和 Orca 中提供内置的 DOCX、XLSX、PPTX 与 PDF 产物运行时。实现位于 Go 内核中，不会调用用户机器上的 Python、Microsoft Office 或 LibreOffice。

## 能力边界

- `artifact_create` 从结构化正文、工作表或幻灯片数据创建文件，并写入 `.orca-artifact.json` sidecar。
- `artifact_edit` 读取 sidecar 后执行追加文字、替换文字、设置单元格和增加幻灯片等结构化修改。
- `artifact_preview` 生成确定性的 PNG 布局预览，用于程序检查页面占用和内容密度。
- `artifact_validate` 重新打开输出文件，检查 ZIP/OOXML 关系、XML 可解析性、逻辑单元数量和文本块数量。
- 每次创建和修改都会在返回前重新解析并验证；验证失败不会报告成功。

## 字体与依赖

运行时仅使用 Go 标准库，不引入新的第三方二进制依赖。OOXML 文件声明通用拉丁与东亚字体回退，由打开文件的 Office 套件选择本机可用字体；PDF 使用标准 CJK CID 字体引用。预览是程序化布局图，不依赖系统字体，也不要求模型识别截图。因此本版没有需要随安装包分发的新增字体许可文件。

## 第三方文件

结构化编辑只面向带有 Orca sidecar 的产物。第三方复杂 Office 文件可能包含宏、SmartArt、嵌入对象、主题扩展或未知 OOXML 部件；在无法证明可以安全保留这些内容时，工具会明确拒绝修改，而不会声称实现了像素级或全保真 Office 编辑。原文件不会被覆盖。
