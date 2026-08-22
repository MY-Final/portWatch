# TUI-Polish-010: Integration and Acceptance

## 目标

完成 TUI 主 Model、Bubble Tea Update/View、扫描 Cmd 和动作 Cmd 的最终接线，并执行全量验收。

## 负责文件

- `internal/tui/app.go`
- `internal/tui/actions.go`
- `internal/tui/app_test.go`
- `README.md`
- `prd/v5-tui-polish.md`（只更新验收状态）

## 依赖

TUI-Polish-009。

## 必须实现

- 默认 Listening，L/C/A 切换并处理 unsupported
- 列表、详情、过滤、Kill confirm 四种状态互斥且可返回
- 刷新使用稳定行键恢复选择
- 进程 lookup 失败显示 Unknown 和可解释状态
- Kill 后验证并刷新，不确认不调用 Terminate
- 更新 README 的真实快捷键和手测步骤

## 不负责

- 不改变 CLI 默认输出、JSON schema 或平台扫描 API 的既有行为
- 不增加自动刷新、UDP TUI、批量 Kill 或 Service Detection

## 完成标准

- 自动化覆盖完整主路径和取消路径
- Windows 真实终端手测四种尺寸及 Listening/Connections/All
- `go test ./...`、`go vet ./...`、`go build ./...`、`git diff --check` 通过
- 只在主控 Agent review 后合并
