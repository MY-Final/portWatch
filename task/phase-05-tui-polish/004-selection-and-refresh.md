# TUI-Polish-004: Stable Selection and Refresh

## 目标

提供按稳定行键保存和恢复选择的纯逻辑，解决刷新或排序后跳到错误端口的问题。

## 负责文件

- `internal/tui/selection.go`
- `internal/tui/selection_test.go`

## 依赖

TUI-Polish-001。

## 必须实现

- 根据行键查找原记录的新索引
- 原记录存在时保持选择
- 原记录消失时选择相邻合理位置
- 过滤后无可见记录时禁止 Enter/K 操作
- 列表排序固定为端口、协议、PID、本地地址、远端地址

## 不负责

- 不启动扫描、不创建 Bubble Tea Cmd
- 不修改 `app.go`

## 完成标准

- 覆盖记录插入、删除、排序变化、过滤无匹配和空列表测试
- 不依赖 slice index 作为稳定身份
