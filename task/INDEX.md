# Task Index

整体边界与数据流见 [`ARCHITECTURE.md`](ARCHITECTURE.md)；执行规则见 [`README.md`](README.md)。

## Phase 1: MVP

执行顺序和并行关系：

```text
001 -> 002 -> (003, 004, 006) -> (005, 007) ->
(008, 010, 011) -> 009 -> 012 -> 013
```

括号内任务可并行，但只有在其依赖完成后开始。`013` 是唯一负责最终 CLI 接线和验收的任务。

| ID | 状态 | 文件 | 主要文件所有权 | 依赖 | Worker | Commit | 测试 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 001 | **done** | `001-project-init.md` | `go.mod`, `.gitignore`, 目录骨架 | 无 | Luna-001 | `e6c150c` (+ `96b0615`) | `go test ./...`; `go build ./...`; `go list ./...` |
| 002 | **done** | `002-domain-model.md` | `pkg/model/*.go` | 001 | Luna-002 | `69f99a0` | `go test ./pkg/model`; `go test ./...`; `go build ./...` |
| 003 | **done** | `003-cli-root.md` | `internal/command/root.go`, `cmd/portwatch/main.go` 初稿 | 001, 002 | Luna-003 | `5f3e697` | `go test ./...`; `go build ./...`; `git diff --check` |
| 004 | **done** | `004-port-scanner-interface.md` | `internal/port/scanner.go` | 002 | Luna-004 | `6f9b88b` | `go test ./...`; `go build ./...`; `git diff --check` |
| 005 | **done** | `005-windows-port-scanner.md` | `internal/port/windows.go` | 004 | Luna-005 | `cdcc92b` | `go test ./...`; `go build ./...`; `git diff --check` |
| 006 | **done** | `006-process-manager-interface.md` | `internal/process/manager.go` | 002 | Luna-006 | `12cda8d` | `go test ./...`; `go build ./...`; `git diff --check` |
| 007 | **done** | `007-windows-process-info.md` | `internal/process/windows.go` | 006 | Luna-007 | `930bc88` | `go test -mod=mod ./...`; `go build -mod=mod ./...`; `git diff --check` |
| 008 | **done** | `008-kill-process.md` | `internal/process/windows.go` tests/termination additions | 006, 007 | Luna-008 | `1334bee` | `go test -v ./internal/process`; `go test ./...`; `go build ./...`; `git diff --check` |
| 009 | **done** | `009-free-port-command.md` | `internal/command/free.go` | 003, 004, 006, 007, 008 | Main Agent | `20956bc` | `go test ./...`; `go build ./...`; `git diff --check` |
| 010 | **done** | `010-output-table.md` | `internal/command/output.go` | 002, 003 | Luna-010 | `fa50537` | `go test ./...`; `go build ./...`; `git diff --check` |
| 011 | **done** | `011-error-handling.md` | `internal/command/errors.go` | 003, 004, 006 | Luna-011 | `2ee32d6` | `go test ./...`; `go build ./...`; `git diff --check` |
| 012 | **done** | `012-unit-tests.md` | `*_test.go` beside owned packages | 002-011 | Main Agent | `40a33c3` | `go test ./...`; `go build ./...`; `git diff --check` |
| 013 | **done** | `013-mvp-integration.md` | final wiring in `cmd/portwatch/main.go`, README | 009-012 | Main Agent | `eafa2fe` | `go test ./...`; `go vet ./...`; `go build -o portwatch.exe ./cmd/portwatch`; Windows runtime checks |

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

## Future Task Status

| Task | 状态 | Commit |
| --- | --- | --- |
| 201-202 | done | current Phase 2 commit |
| 301-302 | done | current Phase 3 commit |
| 401-402 | done | current Phase 4 commit |
| 501-502 | done | `af2c46c` |
| 601-602 | done | `af2c46c` |
| 701-702 | done | `af2c46c` |
| 801-802 | done | `f7d7902` |
| 901-902 | done | `68dcee1` |

## V3 Active Planning

| Task | 状态 | 文档 |
| --- | --- | --- |
| V3-1001 | done | [`phase-10-v3/1001-port-range.md`](phase-10-v3/1001-port-range.md) |
| V3-1002 | done | [`phase-10-v3/1002-windows-udp.md`](phase-10-v3/1002-windows-udp.md) |
| V3-1003 | done | [`phase-10-v3/1003-json-watch.md`](phase-10-v3/1003-json-watch.md) |

## V4 Product Planning

V4 的产品范围、JSON 合同、退出码、安全边界和建议任务拆分见
[`../prd/v4.md`](../prd/v4.md)。V4 任务原计划按依赖关系调度，当前实现已完成并保留该拆分供后续维护。

## V4 Implementation Status

| Task | 状态 | 实现位置 |
| --- | --- | --- |
| V4-01 至 V4-03 | done | `internal/command/filter.go`, `flags.go`, `root.go`, `watch.go` |
| V4-04 至 V4-06 | done | `pkg/model/json.go`, `internal/command/info.go` |
| V4-07 至 V4-08 | done | `internal/command/errors.go`, `kill.go`, `free.go` |
| V4-09 至 V4-10 | done | CI/build 验证、`README.md`、`prd/v4.md` |

## TUI Polish Proposal

| 范围 | 状态 | 文档 |
| --- | --- | --- |
| TUI-Polish-001 至 010 | proposed | [`phase-05-tui-polish/README.md`](phase-05-tui-polish/README.md) |

产品定义见 [`../prd/v5-tui-polish.md`](../prd/v5-tui-polish.md)。本阶段尚未实现，必须先按任务依赖进行拆分和 review。
