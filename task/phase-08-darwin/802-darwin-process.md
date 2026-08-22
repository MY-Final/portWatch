# 802 Darwin Process Manager

目标：在 `internal/process/darwin.go` 通过 `ps`/`proc_pidpath` 获取进程详情，`kill -TERM` 后验证 Exists；所有外部命令参数固定且不经 shell。

负责文件：`internal/process/darwin.go`、Darwin tests。依赖：006。验收：解析 fixture、权限和不存在 PID 分类；不得修改 CLI 或其他平台文件。

