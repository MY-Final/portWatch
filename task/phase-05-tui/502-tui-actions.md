# 502 TUI Actions

目标：在 `internal/tui/actions.go` 增加 Enter 详情、`k` 确认终止、`r` 刷新和 `/` 搜索，复用 Manager/Scanner 接口并显示权限错误。

负责文件：`internal/tui/actions.go`、测试。依赖：501。验收：所有动作有模型测试，终止前无确认不得调用 Terminate；不得复制 CLI 的终止逻辑。

