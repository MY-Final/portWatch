# 201 CLI Flags and Help

## 目标
在 `internal/command/flags.go` 实现标准库 `flag` 解析，支持 `--version`、`--help`、`--protocol tcp` 和明确的未知参数错误；更新 `cmd/portwatch/main.go` 的版本注入。

## 文件边界
负责文件：`internal/command/flags.go`、`internal/command/root.go`、`cmd/portwatch/main.go`、对应测试。不得修改 `pkg/model` 或平台文件。

## 依赖
Phase 1 的 013。

## 验收与测试
`go test ./...`；帮助文本可复制执行且错误退出码为 2。
