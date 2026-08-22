# V3-1001 Port Range Query

## 目标

实现 `portwatch START-END`、`portwatch --json START-END` 和 `--ports p1,p2`，查询指定 TCP 监听端口。

## 负责文件

- `internal/command/root.go`
- `internal/command/root_test.go`
- `README.md`

如需修改其他文件，必须先确认现有输出模型无法复用；平台扫描器和进程管理器不在本任务范围内。

## 行为要求

- `START`、`END` 必须是 1 到 65535 的十进制整数
- `START <= END`
- 范围结果复用现有表格、JSON 和服务识别输出
- `--ports` 使用逗号分隔，重复端口自动去重并排序
- `--ports` 不能和位置参数组合使用
- 结果按现有端口/PID排序
- 无匹配结果仍返回成功；JSON 输出 `ports: []`
- 非法范围返回参数错误和非零退出码
- 单端口、`free`、`find`、`watch` 和 `tui` 行为不得改变
- `--protocol tcp|udp|all` 在支持协议的扫描器上保持有效

## 验收

```powershell
go test ./internal/command
go test ./...
go vet ./...
go build ./...
go run ./cmd/portwatch --json 3000-4000
```

测试至少覆盖：正常范围、反向范围、越界范围、缺少一端、JSON排序和空结果。
