# TUI-Polish-008: Keybinding and Status Help

## 目标

统一快捷键、模式提示和操作状态文本，让用户无需猜测 TUI 如何使用。

## 负责文件

- `internal/tui/help.go`
- `internal/tui/help_test.go`

## 依赖

TUI-Polish-001。

## 必须实现

- 帮助栏包含 Navigate、Details、Kill、Filter、Listening、Connections、All、Refresh、Quit
- 当前模式和过滤状态有明确文本
- 刷新中、成功、取消、终止成功、终止失败有统一状态消息
- 颜色不是唯一表达方式
- `A` 只表示 All；自动刷新不占用本阶段快捷键

## 不负责

- 不修改平台代码或主 Update/View
- 不实现自动刷新

## 完成标准

- 所有核心 key 的显示文本有测试
- 文案在窄终端可截断但不重叠
