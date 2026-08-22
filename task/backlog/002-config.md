# Backlog 002 Configuration

## 目标
评估并在需求成立时增加用户配置文件（默认端口过滤、刷新间隔）。

## 文件边界
预期为 `internal/config/config.go`、平台路径适配文件和配置测试；不得建立配置中心、数据库或修改平台扫描 API。

## 依赖
Phase 2 CLI 与 Phase 4 watch 的需求反馈。

## 验收与测试
未配置时行为与默认值完全一致；非法配置有行号/字段错误；Linux 使用 XDG，Windows 使用 AppData；`go test ./...` 通过。
