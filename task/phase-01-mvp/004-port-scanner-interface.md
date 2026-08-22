# 004 Port Scanner Interface

## 目标
定义端口扫描抽象，使 CLI 不依赖 Windows 实现。

## 负责文件
- `internal/port/scanner.go`

定义 `Scanner` 接口：`List(ctx context.Context) ([]model.PortInfo, error)` 与 `Port(ctx context.Context, number int) ([]model.PortInfo, error)`。`Port` 返回该端口的全部监听记录，找不到时返回空切片而非错误；接口实现不得修改返回模型。提供 `ErrUnsupported` 和 `ErrInvalidPort`，错误可用 `errors.Is` 判断。

## 不负责
不添加 build tag 文件，不执行系统命令。

## 依赖
002。

## 验收与测试
编译一个只依赖接口的 fake scanner；运行 `go test ./internal/port`。

