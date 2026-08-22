# TUI-Polish-009: Responsive Layout and Review

## 目标

将前置 helper 组合成可在常见终端尺寸稳定显示的布局，并修复视觉层级问题；这是第一次允许修改主 View 组合代码。

## 负责文件

- `internal/tui/layout.go`
- `internal/tui/layout_test.go`
- 如确有必要，`internal/tui/app.go` 的 View 组合部分

## 依赖

TUI-Polish-002 至 008。

## 必须实现

- 验证 80x24、100x30、120x40、160x50
- 宽度不足时保留 Port、PID、Process，隐藏/截断低优先级字段
- Header、状态栏、表格、过滤行、帮助栏不互相覆盖
- 详情页长字段换行，不导致布局跳动
- 保留旧键盘行为的兼容测试

## 不负责

- 不新增功能、不修改 CLI 或 JSON
- 不引入新的 TUI 框架或重量级依赖

## 完成标准

- 纯字符串/模型测试覆盖关键宽度
- `go test ./internal/tui` 和 `git diff --check` 通过
