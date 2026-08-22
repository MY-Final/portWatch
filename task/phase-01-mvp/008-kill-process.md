# 008 Kill Process

## 目标
在 Windows 进程管理器中实现安全、可验证的终止操作。

## 负责文件
- 继续修改 `internal/process/windows.go`（仅补充 `Terminate`）。
- `internal/process/terminate_windows_test.go`。

`Terminate` 先以 `PROCESS_TERMINATE` 打开句柄，调用 `TerminateProcess`，关闭句柄并等待最多 3 秒轮询 `Exists`；超时返回明确错误。拒绝 PID 0、当前系统关键 PID（至少 4）和无效 PID。将 ERROR_ACCESS_DENIED 映射为 `ErrAccessDenied`。不得强制终止多个 PID。

## 不负责
不询问用户，不扫描端口，不改 CLI。

## 依赖
006, 007。

## 验收与测试
使用测试子进程（由测试自身启动）验证 Terminate 后 `Exists=false`；权限不足和不存在 PID 有稳定错误分类。测试不得终止测试运行器自身。

