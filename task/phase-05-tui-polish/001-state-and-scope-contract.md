# TUI-Polish-001: State and Scope Contract

## 目标

冻结 TUI Polish 所需的视图状态、扫描范围和进程信息失败分类，给后续任务提供可编译的最小契约。

## 负责文件

- `internal/port/scope.go`
- `internal/port/scope_test.go`
- `internal/tui/state.go`
- `internal/tui/state_test.go`

## 依赖

无。

## 必须实现

- 定义 `port.Scope`：`Listening`、`Connections`、`All`
- 定义可选接口 `port.ScopedScanner`，不修改现有 `port.Scanner`
- 定义 TUI 的页面模式：列表、详情、过滤、Kill 确认
- 定义稳定行键：协议、端口、PID、本地地址、远端地址
- 定义 lookup 状态：成功、权限不足、进程已退出、未知失败
- 给出 `Listening` 默认值和字符串显示值

## 不负责

- 不修改 `internal/tui/app.go` 或 `actions.go`
- 不实现 Windows 连接扫描
- 不实现任何渲染或键盘处理

## 完成标准

- Scope 和状态值可被其他包引用
- 不支持的 Scope 可通过 `errors.Is` 或明确 sentinel 判断
- 单元测试覆盖默认值、字符串值和稳定行键相等性
- `gofmt -w`、`go test ./internal/port ./internal/tui`、`git diff --check` 通过
