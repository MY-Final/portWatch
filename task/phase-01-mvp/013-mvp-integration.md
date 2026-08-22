# 013 MVP Integration

## 目标
将已完成的模型、扫描器、进程管理器和命令编排接入可运行二进制，并写基础 README。

## 负责文件
- `cmd/portwatch/main.go`（接线和退出码）
- `internal/command/root.go`（仅接线所需的小幅修改）
- `README.md`

在 Windows build 下实例化 `windows` scanner/manager，注入 `command.Dependencies`；实现 `portwatch <port>` 查询并渲染端口及进程详情，实现 `portwatch free <port>`；无参数列出全部监听端口。进程信息读取失败时保留端口记录并显示错误，不 panic。README 只记录安装、Windows 要求、三个 MVP 命令、管理员权限提示和当前明确不支持的平台/功能。

## 不负责
不新增功能、不引入框架、不修改平台实现或模型字段。

## 依赖
009, 010, 011, 012。

## 验收与测试
`go test ./...`；`go build -o portwatch.exe ./cmd/portwatch`；运行 `portwatch.exe 1`、`portwatch.exe 8080` 和 `portwatch.exe free 8080`，确认错误和取消路径有退出码且不会 panic。提交前检查 `git diff --stat` 仅包含本任务允许文件及前序任务产物。

