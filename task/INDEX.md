# Task Index

整体边界与数据流见 [`ARCHITECTURE.md`](ARCHITECTURE.md)；执行规则见 [`README.md`](README.md)。

## Phase 1: MVP

执行顺序和并行关系：

```text
001 -> 002 -> (003, 004, 006) -> (005, 007) ->
(008, 010, 011) -> 009 -> 012 -> 013
```

括号内任务可并行，但只有在其依赖完成后开始。`013` 是唯一负责最终 CLI 接线和验收的任务。

| ID | 文件 | 主要文件所有权 | 依赖 |
| --- | --- | --- | --- |
| 001 | `001-project-init.md` | `go.mod`, `.gitignore`, 目录骨架 | 无 |
| 002 | `002-domain-model.md` | `pkg/model/*.go` | 001 |
| 003 | `003-cli-root.md` | `internal/command/root.go`, `cmd/portwatch/main.go` 初稿 | 001, 002 |
| 004 | `004-port-scanner-interface.md` | `internal/port/scanner.go` | 002 |
| 005 | `005-windows-port-scanner.md` | `internal/port/windows.go` | 004 |
| 006 | `006-process-manager-interface.md` | `internal/process/manager.go` | 002 |
| 007 | `007-windows-process-info.md` | `internal/process/windows.go` | 006 |
| 008 | `008-kill-process.md` | `internal/process/windows.go` tests/termination additions | 006, 007 |
| 009 | `009-free-port-command.md` | `internal/command/free.go` | 003, 004, 006, 007, 008 |
| 010 | `010-output-table.md` | `internal/command/output.go` | 002, 003 |
| 011 | `011-error-handling.md` | `internal/command/errors.go` | 003, 004, 006 |
| 012 | `012-unit-tests.md` | `*_test.go` beside owned packages | 002-011 |
| 013 | `013-mvp-integration.md` | final wiring in `cmd/portwatch/main.go`, README | 009-012 |

## Future phases

Each future card is intentionally isolated in its own files and must not be scheduled with Phase 1 unless its prerequisites are explicitly met.

| Phase | Scope |
| --- | --- |
| 2 | CLI flags, `kill`, `find`, port ranges, version/help |
| 3 | Stable JSON schema and JSON output |
| 4 | Polling watch command and change detection |
| 5 | TUI after CLI contracts stabilize |
| 6 | Development-service detection |
| 7 | Linux `/proc`/`ss` implementation |
| 8 | macOS `libproc`/`netstat` implementation |
| 9 | CI, release binaries, packaging and cross-platform verification |
