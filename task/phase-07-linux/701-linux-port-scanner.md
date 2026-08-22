# 701 Linux Port Scanner

目标：在 `internal/port/linux.go` 实现 TCP LISTEN 扫描，优先解析 `/proc/net/tcp`、`/proc/net/tcp6` 并通过 `/proc/<pid>/fd` 映射 socket inode；权限不足返回可分类错误。

负责文件：`internal/port/linux.go`、Linux 专属测试/fixtures。依赖：004。验收：fixture 解析测试、真实临时 listener 集成测试；不得修改 Windows 或 command 文件。

