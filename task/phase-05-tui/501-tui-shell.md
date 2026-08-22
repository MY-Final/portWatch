# 501 TUI Shell

目标：在 `internal/tui/app.go` 使用 Bubble Tea 建立端口表、加载/错误/空状态和退出模型；依赖注入 Scanner/Manager。

负责文件：`internal/tui/app.go`、`go.mod` 中 TUI 依赖、TUI 单测。依赖：Phase 1 的 013。验收：模型测试覆盖 q、刷新、加载失败；不修改 command 或 platform 包。

