# 202 Kill and Find Commands

## 目标
新增 `portwatch kill <pid>` 和 `portwatch find <name>`；kill 必须复用进程详情、y/N 确认和终止后 Exists 验证，find 只在当前扫描结果中按进程名大小写不敏感过滤并聚合端口。

## 文件边界
负责文件：`internal/command/kill.go`、`internal/command/find.go`、各自测试、根命令注册文件。不得修改平台 API。

## 依赖
201。

## 验收与测试
fake 依赖覆盖拒绝/确认/权限错误，多 PID 输出稳定排序，`go test ./...` 通过。
