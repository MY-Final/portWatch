# 601 Service Detector

目标：在 `internal/service/detector.go` 定义 Detector 接口和规则实现，根据进程名、命令行、工作目录匹配 Node/Vite、Spring Boot、Uvicorn；返回 Name、Type、Confidence。

负责文件：`internal/service/detector.go`、规则测试。依赖：Phase 1 的 ProcessInfo/PortInfo。验收：大小写和参数顺序测试，未知样本为 `Unknown`/0；不调用系统 API。

