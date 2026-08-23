# PortWatch MVP Architecture

## Current repository

`portWatch/` currently contains only `prd/v1.md`; there is no Go module, source package, test suite, or CI configuration. The task plan therefore defines the initial package boundaries instead of adapting an existing implementation.

## Runtime flow

```text
cmd/portwatch/main.go
  -> internal/command.Run
      -> port.Scanner (TCP LISTENING records)
      -> process.Manager (details by PID)
      -> internal/command output/error helpers
```

`free` adds a confirmation gate, calls `Manager.Terminate` for the selected PID, then invokes `Scanner.Port` again. A successful kill is never reported without the second scan showing no matching record.

## Dependency direction

`pkg/model` has no internal dependencies. `internal/port` and `internal/process` depend only on model plus standard library/platform packages. `internal/command` depends on both interfaces and model. `cmd` depends only on command. Platform files implement interfaces and never import command.

## Platform boundary

The interface files are platform-neutral. Implementations are selected at compile time with `//go:build windows`, `linux`, or `darwin`. MVP only ships the Windows implementation. Unsupported platforms must return `ErrUnsupported` from a small tagged stub added by the integration task if needed for cross-compilation; they must not contain runtime OS-name switches.

## Error and safety contract

Library packages return wrapped sentinel errors and never print, panic, or exit. The command package maps errors to stable exit codes. Termination always targets one validated PID, requires explicit interactive confirmation in the command layer, and is followed by process/port verification.

## Parallel ownership

After project initialization and models, scanner interface, process interface, CLI parsing, and error mapping can proceed independently. Windows scanner and process implementation are independent files. The only intentional same-file sequence is `007 -> 008` for `internal/process/windows.go`, followed by `013` for final `main.go` wiring. This is reflected in `INDEX.md`.


## Post-MVP evolution（v0.6.0 起）

初始边界的依赖方向保持不变；在此之上新增的组件：

- `internal/processinfo`：`command` 与 `tui` 共用的有界并发进程信息查询与父链遍历，仅依赖 model，位于两包之下避免循环依赖；
- `wait` / `uninstall`：均在 `internal/command` 内编排；wait 轮询 `Scanner.Port`，uninstall 通过注入钩子隔离 `os.Executable` 与用户 PATH 操作；
- 分发层（仓库根 `install.ps1` / `install.sh` + goreleaser）在 Go 模块之外，安装脚本按资产后缀匹配 Release，不感知具体命名。

平台实现的实际能力以 README 能力矩阵为现行事实来源；本文件的历史段落保留 MVP 时期的规划语境。
