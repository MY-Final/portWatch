# PortWatch Task Plan

PortWatch 的任务拆分目录。MVP 至 V6 的各阶段均已实现；V6 之后（v0.6.0–v0.8.0）的工作按实现日志补记于
[`INDEX.md`](INDEX.md) 末节，后续实现 Agent 仍以本目录中的目标结构和接口契约为准。

## 执行规则

- 未明确调度的阶段默认只作为规划，不应自行提前实现。
- 调度器必须先满足任务依赖；没有依赖的任务可以并行。
- 一个 Agent 只修改自己任务列出的文件。需要修改同一文件的任务按依赖串行。
- 平台实现使用 Go build tags；Windows MVP 不应引入 Linux/macOS 代码。
- 使用标准库优先，MVP 仅允许 `golang.org/x/sys/windows` 作为 Windows API 依赖；不引入 CLI/TUI 框架。
- 所有用户可见错误返回到 CLI，由 CLI 负责退出码和文字；库层不得 `panic` 或直接 `os.Exit`。
- 每个任务完成后运行其卡片中的测试，并记录 Windows 版本和 Go 版本。

## MVP 目标

支持 Windows TCP `LISTENING`：

```text
portwatch 8080
portwatch free 8080
```

输出端口、协议、状态、PID、进程名、命令行和可执行路径；`free` 必须显示信息、要求明确确认、终止进程并重新扫描确认端口释放。

## 目标代码结构

```text
cmd/portwatch/main.go
pkg/model/port.go
pkg/model/process.go
internal/port/scanner.go
internal/port/windows.go
internal/process/manager.go
internal/process/windows.go
internal/command/root.go
internal/command/output.go
internal/command/free.go
internal/command/errors.go
```

`cmd/portwatch/main.go` 只负责调用命令入口；业务编排在 `internal/command`，平台能力在 `internal/port` 和 `internal/process`。

## 阶段

| 阶段 | 目录 | 状态 |
| --- | --- | --- |
| Phase 1-10 | `phase-01-mvp/` 至 `phase-10-v3/` | 已完成 |
| V4 / V5 / V6 | 见 [`../prd/`](../prd/) 与本目录各表 | 已完成（V5 被 V6 取代） |
| v0.6.0–v0.8.0 | 无预拆分卡片，见 [`INDEX.md`](INDEX.md) 实现日志 | 已完成 |
| Backlog | `backlog/` | 001-udp 已完成；002/003 候选 |

TUI 产品重构见 [`../prd/v5-tui-polish.md`](../prd/v5-tui-polish.md)，对应任务位于
`phase-05-tui-polish/`。该阶段默认先完成任务评审，再按 README 中的依赖图调度。
