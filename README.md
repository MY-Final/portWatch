# PortWatch

PortWatch 是一个面向开发者的端口诊断与进程管理 CLI。它可以查询 TCP
监听端口对应的 PID、进程名、命令行和可执行文件路径，也可以在确认后终止占用端口的进程。

## 当前支持

- 查询一个端口：`portwatch 8080`
- 列出全部 TCP `LISTENING` 端口：`portwatch`
- 释放端口：`portwatch free 8080`
- 按 PID 终止进程：`portwatch kill 1234`
- 按进程名搜索：`portwatch find node`
- JSON 输出：`portwatch --json 8080`
- JSON 搜索：`portwatch --json find node`
- 实时监听端口变化：`portwatch --interval 2s watch`
- 只监听一个端口：`portwatch watch 8080`
- 交互式终端界面：`portwatch tui`
- `--help`、`--version` 和 `--protocol tcp`

`free` 会先显示进程详情并要求输入 `y` 或 `yes`，终止后重新扫描端口；直接回车或输入其他内容都会取消操作。

## Windows 安装

要求 Windows 和 Go 1.22 或更新版本。在项目根目录执行：

```powershell
$bin = Join-Path $env:USERPROFILE "bin"
New-Item -ItemType Directory -Force $bin | Out-Null
go build -o (Join-Path $bin "portwatch.exe") ./cmd/portwatch

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (($userPath -split ';') -notcontains $bin) {
    [Environment]::SetEnvironmentVariable(
        "Path",
        (($userPath.TrimEnd(';') + ';' + $bin).Trim(';')),
        "User"
    )
}
```

关闭并重新打开 PowerShell，然后验证：

```powershell
Get-Command portwatch
portwatch --version
portwatch 8080
```

如果查询或终止进程时出现 `access denied`，请使用“以管理员身份运行”的 PowerShell。

## 使用示例

```powershell
# 查看 8080 端口
portwatch 8080

# 查看全部监听端口
portwatch

# 输出稳定 JSON，适合脚本处理
portwatch --json 8080

# 每 2 秒报告新增或移除的监听端口，按 Ctrl+C 退出
portwatch --interval 2s watch

# 只监控 8080 端口
portwatch watch 8080

# 搜索名称包含 node 的进程
portwatch find node

# 以 JSON 搜索名称包含 node 的进程
portwatch --json find node

# 启动交互式终端界面
portwatch tui

# 终止指定 PID（会要求确认）
portwatch kill 1234

# 释放 8080 端口（会要求确认并验证）
portwatch free 8080
```

所有全局选项应放在位置参数或子命令之前，例如使用
`portwatch --json 8080`，不要写成 `portwatch 8080 --json`。

## 平台说明

Windows 使用 IP Helper API 扫描 IPv4/IPv6 TCP 监听端口，并通过 Windows
进程 API 和 WMI 获取进程详情。Linux 使用 `/proc`，macOS 使用 `lsof`/`ps`；
这两个平台的实现已包含在代码中，但建议在目标系统上单独验证权限和命令可用性。

当前 CLI 只扫描 TCP 监听端口，`--protocol` 暂时只接受 `tcp`。TUI 可以通过 `portwatch tui` 显式启动；端口表格和端口 JSON 会提供可选的开发服务识别结果，无法确认时显示 `Unknown`。

## 开发

```powershell
go test ./...
go vet ./...
go build ./...
```

交叉编译示例：

```powershell
$env:GOOS = "windows"; $env:GOARCH = "amd64"; go build ./cmd/portwatch
$env:GOOS = "linux";   $env:GOARCH = "amd64"; go build ./cmd/portwatch
$env:GOOS = "darwin";  $env:GOARCH = "arm64"; go build ./cmd/portwatch
```

发布配置位于 `.goreleaser.yaml`，CI 配置位于 `.github/workflows/`。

## 任务规划

开发任务和后续阶段见 [`task/README.md`](task/README.md) 与 [`task/INDEX.md`](task/INDEX.md)。
