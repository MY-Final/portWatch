# 402 Watch Command

目标：在 `internal/command/watch.go` 注册 `portwatch watch [--interval duration]`，将引擎事件渲染为时间戳和 `+/-` 行，Ctrl-C 正常退出。

负责文件：`internal/command/watch.go`、测试、根命令注册。依赖：401。验收：interval 非正数返回参数错误，context 取消不泄漏 goroutine，`go test ./...`。

