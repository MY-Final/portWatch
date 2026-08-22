# 005 Windows Port Scanner

## 目标
实现 Windows TCP LISTENING 扫描，将本机监听端口映射到 PID。

## 负责文件
- `internal/port/windows.go`，文件头必须有 `//go:build windows`。
- `internal/port/windows_test.go` 只测试可注入解析逻辑，不要求固定机器端口。

使用 `golang.org/x/sys/windows` 调用 `GetExtendedTcpTable`（`TCP_TABLE_OWNER_PID_ALL`），读取 IPv4/IPv6 owner PID 表，仅保留 `MIB_TCP_STATE_LISTEN`。填充 `PortInfo` 的 Port、Protocol=`TCP`、LocalAddr、RemoteAddr、State=`LISTENING`、PID；ProcessName 留空，由进程层补充。按端口升序、地址稳定排序；API 错误包装为可识别 error。支持取消 context。

## 不负责
不读取进程名，不终止进程，不改 CLI 文件。

## 依赖
004。

## 验收与测试
Windows 上运行 `go test ./internal/port`；另运行一个临时监听 TCP socket，通过 `List` 能找到其端口和当前 PID。非 Windows 环境必须因 build tag 跳过该测试。

