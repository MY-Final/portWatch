# 009 Free Port Command

## 目标
实现 `portwatch free <port>` 的完整交互编排。

## 负责文件
- `internal/command/free.go`

实现 `Free(ctx, scanner port.Scanner, manager process.Manager, port int, in io.Reader, out io.Writer) error`：扫描端口；无记录返回“端口已可用”状态；对每条记录查询进程信息并展示；读取一行确认，仅精确接受大小写不敏感的 `y`/`yes`，其他输入取消；确认后只终止对应 PID；等待并重新扫描，只有端口无记录才报告成功，否则返回验证错误。EOF、扫描失败、进程详情失败、终止失败都要保留可判断 cause。

## 不负责
不解析根命令参数，不实现表格通用渲染，不直接调用 Windows API。

## 依赖
003, 004, 006, 007, 008。

## 验收与测试
用 fake scanner/manager 覆盖：空闲、拒绝、确认成功、终止失败、验证仍占用、多个 PID；运行 `go test ./internal/command`。

