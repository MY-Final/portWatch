# 902 Release Packaging

目标：新增 `.goreleaser.yaml` 和发布文档，产出 windows amd64/arm64、linux amd64/arm64、darwin amd64/arm64 二进制及校验和。

负责文件：`.goreleaser.yaml`、`docs/release.md`。依赖：901。验收：`goreleaser check` 成功，版本由 `-ldflags` 注入且不把管理员权限写死在构建脚本中。

