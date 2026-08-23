# 001 Project Init

## 目标
初始化可编译的 Go CLI 工程，不实现端口或进程功能。

## 负责文件
- 新建 `go.mod`，module path 固定为 `github.com/MY-Final/portWatch`，Go 版本固定为 `1.22`，并声明 `golang.org/x/sys`（供后续 Windows API 使用）。
- 新建 `.gitignore`，忽略构建产物、IDE 文件和临时覆盖率文件。
- 创建空目录 `cmd/portwatch`, `internal/command`, `internal/port`, `internal/process`, `pkg/model`；不在这些目录实现业务逻辑。

## 不负责
不修改 PRD，不添加 CLI 框架，不添加平台代码。

## 依赖
无。

## 验收与测试
在 `portWatch` 根目录运行 `go mod tidy` 和 `go test ./...`，两者成功；`go list ./...` 能列出工程包。
