# 401 Watch Engine

目标：在 `internal/watch/engine.go` 实现按 context 停止的轮询引擎；比较 `(protocol,port,pid)` 集合，产生 Added/Removed 事件，首次快照只产生 Added。

负责文件：`internal/watch/engine.go`、测试。依赖：Phase 1 的 `port.Scanner`。验收：fake scanner 与 fake clock 测试去重、排序、取消；不得修改 CLI。

