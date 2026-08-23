# 1102 Config Defaults Wiring

## 目标

把配置接入命令默认值。优先级固定为：**CLI flag > 配置文件 > 内置默认**，未配置时全量行为与 v0.8.0 逐字节一致。

## 负责文件

- `cmd/portwatch/main.go`（加载配置并传入 command）
- `internal/command/flags.go` / `root.go`
- `internal/command/*_test.go`（新增用例）
- `README.md`（新增「配置文件」一节）

## 行为要求

- `interval`：作用于 `watch` 与 `wait`。显式 `--interval` 覆盖配置；wait 的优先级变为 flag > 配置 > 200ms 内置默认（复用 IntervalSet 检测）
- `process`：作为默认进程名过滤（等价隐式 `--process`）。显式 `--process` 覆盖；显式 `--process ""` 也视为显式设置并清空过滤，不回落配置
- 配置的默认过滤参与现有规则：对不允许过滤的命令（kill/info/uninstall 等）按现有 ParseError 报错——因此配置 `process` 后裸跑这些命令会得到退出码 2，此为有意行为，README 必须写明
- 配置加载失败（语法/字段错误）打印一行 stderr 警告（含定位信息）后按无配置继续，退出码不受影响
- `--json`、TUI、wait 事件格式、退出码表不变；不引入新的 JSON 字段
- 测试覆盖优先级矩阵：flag 覆盖 config、config 覆盖默认、无配置等价现状、坏配置警告不中断

## 验收

```powershell
go test ./...
go vet ./...
go build ./...
$env:PORTWATCH_CONFIG = "test fixture path"; go run ./cmd/portwatch --json
Remove-Item Env:PORTWATCH_CONFIG
```

README「配置文件」一节含：路径表、两个键的语义、优先级说明、坏配置的宽松行为。
