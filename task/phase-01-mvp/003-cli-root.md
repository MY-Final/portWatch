# 003 CLI Root

## 目标
提供无框架 CLI 入口和参数解析骨架，保留后续命令接线点。

## 负责文件
- `internal/command/root.go`
- `cmd/portwatch/main.go`（仅建立 `main -> command.Run` 调用）

`command.Run(ctx, args, deps, stdout, stderr) int` 返回退出码，不调用 `os.Exit`。MVP 参数规则：`portwatch <port>` 查询单端口；无参数列出全部（扫描逻辑由后续任务接入）；`free` 和帮助/未知参数先返回结构化错误。依赖对象通过一个明确的 `Dependencies` struct 注入，禁止包级全局可变状态。

## 不负责
不实现扫描、进程读取、终止、表格渲染。

## 依赖
001, 002。

## 验收与测试
对空参数、数字端口、非数字参数、未知命令写解析测试；`go test ./internal/command` 通过，`go run ./cmd/portwatch --help` 不 panic。

