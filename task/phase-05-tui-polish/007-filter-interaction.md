# TUI-Polish-007: Filter Interaction

## 目标

实现实时过滤输入和可见行计算，覆盖 Port、PID、Process Name 以及已存在的 Service 字段。

## 负责文件

- `internal/tui/filter.go`
- `internal/tui/filter_test.go`

## 依赖

TUI-Polish-001。

## 必须实现

- `/` 进入输入，字符输入实时改变可见记录
- 匹配 Port、PID、Process Name；Service 有数据时匹配 Service
- 大小写不敏感
- Enter 保留过滤并回到导航，Esc 清空并退出
- 无匹配时禁止详情和 Kill
- 过滤改变后选择归一化到第一个可见记录

## 不负责

- 不修改 `app.go` 主 Update/View
- 不执行扫描或进程查询

## 完成标准

- 覆盖输入、退格、Enter、Esc、数字和大小写匹配
- 空结果不能触发操作的测试通过
