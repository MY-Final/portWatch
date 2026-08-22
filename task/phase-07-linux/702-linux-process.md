# 702 Linux Process Manager

目标：在 `internal/process/linux.go` 通过 `/proc/<pid>/comm`, `exe`, `cmdline`, `cwd` 实现 Info/Exists/Terminate，权限错误映射统一 sentinel。

负责文件：`internal/process/linux.go`、Linux 测试。依赖：006。验收：临时子进程生命周期测试，字段缺失可返回空值但不 panic；不得修改其他平台文件。

