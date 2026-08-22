# TUI-Polish-003: Header and Table Layout

## 目标

实现终端原生的 Header、模式状态栏、端口表和更新时间显示，不负责接入主 Model。

## 负责文件

- `internal/tui/view_header.go`
- `internal/tui/view_table.go`
- `internal/tui/view_header_test.go`

## 依赖

TUI-Polish-001。

## 必须实现

- Header 显示 `PortWatch` 和 `Port & Process Manager`
- 状态栏显示模式、协议、记录数和更新时间
- 列表至少显示 PORT、PROTOCOL、PID、PROCESS；宽度允许时显示 STATE
- 选中行使用 `>` 等纯文本标记
- Unknown 进程名不能渲染成孤立的 `-`
- 所有 helper 接收明确数据，不读取全局变量或终端尺寸

## 不负责

- 不修改 `app.go`、`actions.go`
- 不实现模式切换、详情、Kill 或过滤
- 不加入颜色主题、渐变或装饰性 Logo

## 完成标准

- 对空列表、Unknown、长进程名和多种宽度有快照/字符串测试
- `go test ./internal/tui`、`gofmt`、`git diff --check` 通过
