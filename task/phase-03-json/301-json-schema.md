# 301 JSON Schema

## 目标
在 `pkg/model/json.go` 定义版本化 JSON DTO，字段为 `port, protocol, local_addr, remote_addr, state, pid, process_name, executable, command`，缺失值输出空字符串，schema 版本写入顶层 `schema_version`。

## 文件边界
负责文件：`pkg/model/json.go`、模型测试。不得改变 PortInfo/ProcessInfo 字段。

## 依赖
Phase 1 的 002。

## 验收与测试
固定 fixture 的 `encoding/json` 输出稳定，字段无额外平台句柄。
