# 012 Unit Tests

## 目标
补齐跨包契约测试和 MVP 回归测试，不依赖开发者机器上的固定端口。

## 负责文件
- `internal/command/*_test.go`
- `internal/port/*_test.go`（仅可移植 fake/解析测试）
- `internal/process/*_test.go`（仅可移植 fake/错误映射测试）

使用 fake scanner/manager 测试 root 查询、free 确认和错误退出码；Windows 集成测试必须使用 `//go:build windows`，动态创建监听 socket/子进程并在 cleanup 中释放；禁止 kill 当前测试进程、禁止依赖 8080 一定空闲。

## 不负责
不改生产代码和公开契约；发现契约缺陷时在任务结果中报告给 013 Agent。

## 依赖
002 至 011 全部完成。

## 验收与测试
`go test ./...` 在 Windows 通过；非 Windows 至少能编译平台无关包并跳过 Windows 集成测试。

