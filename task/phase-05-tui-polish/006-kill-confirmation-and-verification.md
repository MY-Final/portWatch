# TUI-Polish-006: Kill Confirmation and Verification

## 目标

实现 TUI 专用的安全终止流程 helper：确认、终止、PID 验证、端口验证和结果消息。

## 负责文件

- `internal/tui/kill_flow.go`
- `internal/tui/kill_flow_test.go`

## 依赖

TUI-Polish-001、TUI-Polish-005。

## 必须实现

- 列表和详情都可进入确认状态
- Enter 确认，Esc 取消；y/n 可兼容但不能跳过确认
- 确认内容包含 PID、进程名、关联端口和警告
- 拒绝 PID 4、当前 PortWatch PID 和明确标记的关键进程
- Terminate 后调用 Exists，并通过 Scanner.Port 验证端口记录消失
- 成功、权限失败、验证失败和取消都返回可分类结果

## 不负责

- 不复制 CLI `Kill`/`Free` 实现
- 不修改 `internal/process` 平台实现
- 不直接调用 `os.Exit`

## 完成标准

- 未确认、取消、成功、权限失败、PID 仍存在、端口仍占用都有测试
- 测试替身记录 Terminate/Exists/Port 调用顺序
