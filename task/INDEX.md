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
| TUI-Polish-001 至 010 | superseded by V6 | [`phase-05-tui-polish/README.md`](phase-05-tui-polish/README.md) |

产品定义见 [`../prd/v5-tui-polish.md`](../prd/v5-tui-polish.md)。该提案已归档，后续以 V6 Guided TUI 产品定义为准。

## V6 Guided TUI Implementation

V6 将 TUI 从“端口表”收敛为端口诊断工作流：默认 Listening、可选端口聚焦、独立详情和确认页、Kill 后验证，以及通过 View 菜单进入 Connections/All。

产品定义见 [`../prd/v6-guided-tui.md`](../prd/v6-guided-tui.md)。当前实现已覆盖命令入口、核心页面状态、Windows Scope 扫描、测试和文档；后续只添加 V6 PRD 明确的验收修正，不继续恢复 V5 首屏快捷键堆叠。

## V6 之后实现日志（v0.6.0 – v0.8.0）

V6 交付后进入人机协作迭代：以下工作均由用户与本仓库 Agent 会话完成，无预拆分任务卡，按版本与主题补记于此，细节以各 commit message 与 README 为准。

### v0.6.0 — Windows 进程信息直读 PEB

| 主题 | 内容 | Commit |
| --- | --- | --- |
| 进程参数直读 | `NtQueryInformationProcess` + `ReadProcessMemory` 读 PEB 的命令行/工作目录，替代 PowerShell/WMI 外部进程，百端口级列表从 ~10s 降到 ~50ms | `442d099` |
| 工作目录修复 | Win32_Process 本无 WorkingDirectory，PEB `CurrentDirectory` 补齐真实值 | `442d099` |
| 错误分类去本地化 | NTSTATUS/Win32 错误码映射替代英文消息匹配 | `442d099` |
| 文档 | 用户操作手册（后随仓库卫生改为 `docs/operation-manual.md`） | `275efbe` |

### v0.7.0 — pw 别名与维护批次

| 主题 | 内容 | Commit |
| --- | --- | --- |
| pw 短别名 | argv[0] 识别二进制名（BusyBox 模式），`pw.exe` 的帮助与错误前缀跟随显示 | `6e4669f` |
| 维护批次 | TUI 版本注入、错误报告按 PID 升序确定化、root.go 去重、进程信息有界并发（≤8）、GNU 风格参数交错、平台能力矩阵文档、仓库卫生 | `259a17d` |
| processinfo 抽包 | 有界并发查询移入 `internal/processinfo`，TUI refresh 复用（消除 command↔tui 循环依赖） | `bd2839d` |
| Linux /proc 单遍扫描 | inode→PID 映射一次构建，两表共用；顺带修复 parseProcAddress 解码长度判断（原实现在真实 /proc 上必错） | `83957ef` |

### v0.7.x — 卸载与分发

| 主题 | 内容 | Commit |
| --- | --- | --- |
| uninstall 子命令 | y/yes 确认 + `--yes`；Windows 改名 + 分离进程延迟删除；保守用户 PATH 清理（默认目录且清空才动）；退出码 0/2/3/4 | `46c0c44` |
| 自删除修正 | 引号经 exec.Command 被 MSVCRT 转义导致 cmd 语法错误 → 改为生成自删除批处理；过渡产物不再阻断 PATH 清理 | `72ddaa2` |
| 安装脚本 | `install.ps1` / `install.sh`：Release 下载 + SHA256 校验 + 双别名安装 + PATH/rc 配置 + 卸载模式 | `e27719c`, `9514f21`, `e80ac5c` |
| 模块路径 | `github.com/portwatch/portwatch` → `github.com/MY-Final/portWatch`，全量 import 同步 | `a48b793` |
| README 重设计 | pw 优先、徽章/表格/折叠块、真实输出示例 | `4f1d03f` |
| CI | 平台无关化测试修正（驱动 ubuntu runner 抓出 fixture 缺陷）、actions 升 v7 | `5aa6adc`, `99bd3b0` |
| 帮助跟随程序名 | usage 行使用 BinaryName | `2976122` |
| 别名一并卸载 | 卸载自身时清理同目录另一别名 | `98013c3` |

### v0.8.0 — wait / 父进程链 / 平台对齐（分支 `feat/wait-parent-chain`）

| 主题 | 内容 | Commit |
| --- | --- | --- |
| wait 子命令 | `pw wait <port>` 阻塞等待空闲/占用（`--expect`、`--timeout` 退出码 124、`--json` 单事件） | `89bd510` |
| 父进程链 | `ProcessInfo.ParentPID`（WithParent 修饰）；Windows PBI InheritedFromUniqueProcessID / Linux status PPid / macOS ps；`processinfo.Ancestors` ≤8 跳含环保护；info 文本 + TUI 详情 + JSON 加法字段 | `f6b38cf` |
| 平台对齐 | Linux UDP（/proc/net/udp）+ 全状态连接视图（ListScope + 状态码表）+ cwd；macOS UDP（lsof）+ cwd；协议不支持文案去 Windows 专属表述 | `c53e090`, `e644566` |
| 发布 | goreleaser 发布目标钉死 GitHub；`v0.7.0`/`v0.8.0` 双 Release 发布（含双远端 tag 推送） | `e0ce5d6` + tags |

### Backlog 状态更新

- `backlog/001-udp.md`：已由 v0.8.0 完成全部三平台实现，卡片已标记 done。
- `backlog/002-config.md`、`backlog/003-security-review.md`：仍为候选，未调度。

## Phase 12: Security Review（已排期）

来源 `backlog/003-security-review.md`，范围按 PEB/卸载/安装脚本现状重写。

backlog 002（配置文件）已于 2026-08-23 评估后 declined，结论记录在卡片内；曾短暂排期的
Phase 11 任务卡已删除。

| ID | 状态 | 文件 | 结果 |
| --- | --- | --- | --- |
| 1201 | **done** | [`phase-12-security/1201-security-audit.md`](phase-12-security/1201-security-audit.md) | [`../docs/security-review.md`](../docs/security-review.md)：2 中 + 1 低已修复（PID 重用身份复核、自删除敌意路径、Exists 最小权限），3 信息级文档化，无高危 |

## V6 之后实现日志（v0.9.0 –）

### v0.9.0 — CLI 表格截断 / 发布自动化 / 版本单源 / TUI 自动刷新

| 主题 | 内容 | Commit |
| --- | --- | --- |
| CLI 表格截断 | COMMAND / EXECUTABLE PATH 列 80 字符截断（77+`...`），仅表格渲染层，JSON 保持完整；修复超长命令行把 tabwriter 撑爆导致终端刷空白/黑屏 | `703ab7e` |
| 发布自动化 | tag 推送触发 goreleaser Release 工作流（已用一次性 tag 实测通过） | `4d7c2b5` |
| 版本单源 | flags.go 不再持有版本号副本，仅 main.go 字面量（ldflags 契约）；main_test 守护两条契约 | `121a535` |
| TUI 自动刷新 | `T` 键开关（默认关），2s 轮询复用 refresh 通道，in-flight 防重叠；portsLoadedMsg 不再重置页面，后台刷新不打断详情/确认页 | 本次 |
