# Phase 05: TUI Polish

本阶段依据 [`../../../prd/v5-tui-polish.md`](../../../prd/v5-tui-polish.md)，目标是把现有 Bubble Tea TUI 变成面向开发者的端口与进程管理界面。

## 执行规则

- 每个 Worker 只修改自己任务列出的文件
- 不修改 `task/INDEX.md`，由主控 Agent 最后更新
- 不更换 Bubble Tea，不重写 CLI、Process Manager 或默认 `Scanner.List()` 合同
- 每个任务完成前执行任务卡中的测试、`gofmt` 和 `git diff --check`
- `010` 之前不得直接重构 `internal/tui/app.go` 的统一 Update/View 主流程

## 依赖图

```text
001 -> (002, 003, 004, 005, 006, 007, 008)
(002, 003, 004, 005, 006, 007, 008) -> 009 -> 010
```

`003`、`004`、`007`、`008` 可并行；`002` 只修改平台扫描文件，不与 TUI Worker 争抢文件。
