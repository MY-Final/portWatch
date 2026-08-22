# PortWatch Task Plan

PortWatch 的任务拆分目录。当前仓库只有产品需求文档，代码尚未初始化；所有实现 Agent 以本目录中的目标结构和接口契约为准。

## 执行规则

- 默认只执行 `phase-01-mvp/`，其他阶段均为规划，不应提前实现。
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
| Phase 1 | `phase-01-mvp/` | 当前执行 |
| Phase 2-9 | `phase-02-cli/` 至 `phase-09-release/` | 仅规划 |
| Backlog | `backlog/` | 低优先级候选 |

