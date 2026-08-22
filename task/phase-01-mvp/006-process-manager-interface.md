# 006 Process Manager Interface

## 目标
定义进程查询与终止的最小抽象。

## 负责文件
- `internal/process/manager.go`

定义 `Manager`：`Info(ctx, pid int) (model.ProcessInfo, error)`、`Exists(ctx, pid int) (bool, error)`、`Terminate(ctx, pid int) error`。提供 `ErrProcessNotFound`, `ErrAccessDenied`, `ErrNotSupported`；所有平台错误使用 `errors.Is` 可判断。定义 `ErrInvalidPID`，PID 必须大于 0。

## 不负责
不实现 Windows API，不询问用户，不输出文本。

## 依赖
002。

## 验收与测试
fake manager 能编译并覆盖错误判断；`go test ./internal/process` 通过。

