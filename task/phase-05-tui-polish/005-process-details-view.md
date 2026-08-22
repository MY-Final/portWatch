# TUI-Polish-005: Process Details View

## 目标

把当前追加文本的详情改造成独立、可返回的详情状态和渲染 helper。

## 负责文件

- `internal/tui/view_details.go`
- `internal/tui/view_details_test.go`

## 依赖

TUI-Polish-001。

## 必须实现

- 展示 Port、Protocol、State、Local/Remote Address、PID
- 展示 Process Name、Executable、Command、Working Directory
- 缺失值显示 `Unknown` 或 `Unavailable`
- 详情支持明确的返回状态，不与列表文本混排
- 提供详情状态下的 `K` 入口信号，但不在本任务执行终止

## 不负责

- 不修改 `actions.go`
- 不调用平台 API，不猜测字段内容

## 完成标准

- 长命令、空字段、Unknown lookup 状态有测试
- 详情渲染在 80 列宽下不发生文字覆盖
