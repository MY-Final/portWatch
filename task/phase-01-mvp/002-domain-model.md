# 002 Domain Model

## 目标
定义平台无关的数据结构，作为扫描器、进程管理器和输出层的唯一数据契约。

## 负责文件
- `pkg/model/port.go`
- `pkg/model/process.go`

定义：

```go
type PortInfo struct {
    Port int; Protocol string; LocalAddr string; RemoteAddr string
    State string; PID int; ProcessName string
}
type ProcessInfo struct {
    PID int; Name string; Executable string; Command string
    WorkingDir string; User string
}
```

字段使用 JSON tag（小写 snake_case），不加入平台句柄、数据库 ID 或业务方法。端口号必须是 1-65535，PID 非负；模型构造校验返回普通 error。

## 不负责
不读取 Windows API，不实现格式化，不改变接口签名。

## 依赖
001。

## 验收与测试
为非法端口/PID 添加表驱动单元测试；运行 `go test ./pkg/model`。

