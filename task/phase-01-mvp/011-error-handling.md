# 011 Error Handling

## 目标
统一底层错误到用户可读信息和退出码的映射。

## 负责文件
- `internal/command/errors.go`

定义 `ExitCode(error) int`（成功 0、参数错误 2、权限错误 3、系统/扫描错误 1、用户取消 4）与 `PrintError(io.Writer, error)`。使用 `errors.Is` 识别 `port`/`process` 包错误；输出不得包含堆栈和内部句柄，权限错误提示“请以管理员身份运行”。未知错误保留简短上下文。

## 不负责
不吞掉错误，不改底层 API，不实现重试。

## 依赖
003, 004, 006。

## 验收与测试
覆盖每种 sentinel error、包装 error 和未知 error；确保 stderr 有内容且退出码稳定。运行 `go test ./internal/command`。

