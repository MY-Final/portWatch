# 801 Darwin Port Scanner

目标：在 `internal/port/darwin.go` 使用 `lsof -nP -iTCP -sTCP:LISTEN` 的受控参数调用实现扫描和 PID 解析；解析器与命令执行器分离。

负责文件：`internal/port/darwin.go`、Darwin fixtures/tests。依赖：004。验收：固定 lsof 输出解析、命令失败分类、真实 listener 可选集成测试。

