# 007 Windows Process Info

## 目标
实现 Windows 进程详情读取：名称、可执行路径、命令行。

## 负责文件
- `internal/process/windows.go`，必须有 `//go:build windows`。
- `internal/process/windows_test.go` 的查询解析测试。

使用 `OpenProcess`/`QueryFullProcessImageName` 获取 Executable；使用 Windows WMI `Win32_Process`（通过受控 `powershell.exe -NoProfile -NonInteractive` 调用）获取 Name、CommandLine、WorkingDirectory。PowerShell 输出使用 JSON 且限定单个 PID，解析失败返回包装错误；禁止拼接未经校验的用户输入。权限不足映射 `ErrAccessDenied`，不存在映射 `ErrProcessNotFound`。实现 `Exists`，仅打开查询句柄并及时关闭。

## 不负责
不做终止流程编排，不修改 command 包。

## 依赖
006。

## 验收与测试
Windows 上查询当前测试进程，Name/PID 至少正确；查询不存在 PID 返回 `ErrProcessNotFound`。测试命令：`go test ./internal/process`。

