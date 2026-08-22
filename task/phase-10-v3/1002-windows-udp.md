# V3-1002 Windows UDP Scanner

## 目标

通过 Windows IP Helper API 支持 UDP 绑定端口，并让 CLI 接受：

```powershell
portwatch --protocol udp
portwatch --protocol all
portwatch --protocol udp 5353
```

## 负责文件

- `internal/port/scanner.go`
- `internal/port/windows.go`
- `internal/port/windows_test.go`
- `internal/command/flags.go`
- `internal/command/root.go`
- 对应文档和测试

不得改用 `netstat`、`tasklist` 或新增平台无关服务层。Linux/macOS 没有 UDP 实现时必须返回 `port.ErrUnsupported`。

## 数据约定

- `Protocol`: `UDP`
- `State`: `BOUND`
- UDP 没有 TCP 远端连接时，`RemoteAddr` 为空
- 必须支持 IPv4 和 IPv6
- PID 解析失败或进程信息无权限时保留端口记录，并只汇总一次错误提示

## 验收

```powershell
go test ./internal/port ./internal/command
go test ./...
go vet ./...
go build ./...
go run ./cmd/portwatch --json --protocol udp
```

测试至少覆盖 UDP IPv4/IPv6 表解析、截断数据、协议选择和 `all` 合并排序。
