# 302 JSON Output

## 目标
在 `internal/command/json.go` 实现 `--json`，stdout 仅输出一份合法 JSON，诊断错误写 stderr；查询和 free 的 JSON 成功/失败结构分别固定。

## 文件边界
负责文件：`internal/command/json.go`、测试、最小根命令注册改动。不得修改 Windows 实现。

## 依赖
301、Phase 2 的 201。

## 验收与测试
`go test ./...`；用 `json.Decoder` 验证 stdout 无日志污染。
