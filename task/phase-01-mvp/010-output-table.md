# 010 Output Table

## 目标
提供确定性、无颜色依赖的文本输出，适合终端和测试。

## 负责文件
- `internal/command/output.go`

实现 `RenderPorts(io.Writer, []model.PortInfo)` 和 `RenderProcess(io.Writer, model.ProcessInfo, model.PortInfo)`。表头固定为 `PORT PROTOCOL STATE PID PROCESS NAME COMMAND EXECUTABLE PATH`；缺失字段显示 `-`；不截断命令行；按 Port、PID 排序；通过 `text/tabwriter` 对齐。输出函数不读系统、不改数据、不写 stderr。

## 不负责
不实现 JSON、不处理终端颜色、不解析参数。

## 依赖
002, 003。

## 验收与测试
给定固定模型生成 golden 字符串测试；验证空列表和包含换行命令行时不会 panic。运行 `go test ./internal/command`。

