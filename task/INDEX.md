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
| 501-502 | pending | - |
| 601-602 | done | current service detector commit |
| 701-702 | done | current Linux implementation commit |
| 801-802 | done | current Darwin implementation commit |
| 901-902 | pending | - |
