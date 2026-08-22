# TUI-Polish-002: Windows Connection Scan

## 目标

为 Windows 实现 `ScopedScanner`，让 TUI 能真实区分 Listening、Connections 和 All；保留现有 CLI 默认监听行为。

## 负责文件

- `internal/port/windows.go`
- `internal/port/windows_connections_test.go`

## 依赖

TUI-Polish-001。

## 必须实现

- `Listening` 复用现有 TCP LISTENING 结果
- `Connections` 返回 TCP 非 LISTENING 状态，并保留真实 State/PID/地址
- `All` 返回监听和连接记录的稳定合并结果
- 继续使用 Windows IP Helper API，不调用 `netstat` 子进程
- `List()` 和 `Port()` 的既有语义、排序和测试不得回归

## 不负责

- 不修改 `internal/tui`
- 不增加 UDP TUI 模式
- 不改变 CLI 的 `--protocol` 或 JSON 合同

## 完成标准

- Windows build tag 下解析 LISTENING、ESTABLISHED、TIME_WAIT 等状态
- 截断表、非法行和 API 错误有测试
- 非 Windows 编译不受影响
- `go test ./internal/port`、Windows 交叉编译、`git diff --check` 通过
