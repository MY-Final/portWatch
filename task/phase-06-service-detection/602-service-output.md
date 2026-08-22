# 602 Service Output

目标：把可选 Detector 接入表格、JSON 和 TUI 的展示 DTO；检测错误不能阻断端口查询。

负责文件：`internal/command/output_service.go`、JSON DTO 扩展及测试。依赖：601、Phase 3 的 302。验收：无 detector 时输出 Unknown，已有 MVP 文本和 JSON 测试仍通过。

