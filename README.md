# PortWatch

PortWatch 是一个面向开发者的端口诊断与进程管理 CLI。它可以查询 TCP
监听端口对应的 PID、进程名、命令行和可执行文件路径，也可以在确认后终止占用端口的进程。

当前版本：`v0.5.0`

## 当前支持

- 查询一个端口：`portwatch 8080`
- 查询端口范围：`portwatch 3000-4000`
- 查询指定端口集合：`portwatch --ports 3000,8080,8848`
- 列出全部 TCP `LISTENING` 端口：`portwatch`
- 释放端口：`portwatch free 8080`
- 按 PID 终止进程：`portwatch kill 1234`
- 按进程名搜索：`portwatch find node`
- JSON 输出：`portwatch --json 8080`
- JSON 搜索：`portwatch --json find node`
- 实时监听端口变化：`portwatch --interval 2s watch`
- JSON 监听事件：`portwatch --json watch`
- 只监听一个端口：`portwatch watch 8080`
- 按进程、PID 或状态筛选：`portwatch --process node`、`portwatch --pid 1234`、`portwatch --state LISTENING`
- 查看 PID 详情及其端口：`portwatch info 1234`
- JSON 查看 PID 详情：`portwatch --json info 1234`
- 交互式终端界面：`portwatch tui`、聚焦端口：`portwatch tui 8080`
- `--help`、`--version` 和 `--protocol tcp|udp|all`

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

# 查看 3000 到 4000 范围内的监听端口
portwatch 3000-4000

# 查看指定端口集合
portwatch --ports 3000,8080,8848

# 查看 UDP 绑定端口（Windows）
portwatch --protocol udp

# 查看 TCP 和 UDP 端口（Windows）
portwatch --protocol all

# 查看全部监听端口
portwatch

# 输出稳定 JSON，适合脚本处理
portwatch --json 8080

# 每 2 秒报告新增或移除的监听端口，按 Ctrl+C 退出
portwatch --interval 2s watch

# 只监控 8080 端口
portwatch watch 8080

# 以 JSON Lines 输出端口变化
portwatch --json watch

# 搜索名称包含 node 的进程
portwatch find node

# 以 JSON 搜索名称包含 node 的进程
portwatch --json find node

# 只查看 node 进程监听的端口
portwatch --process node

# 查看 PID 详情和它占用的端口
portwatch info 1234

# 以 JSON 输出 PID 详情
portwatch --json info 1234

# 启动交互式终端界面
portwatch tui

# 只关注一个端口
portwatch tui 8080

# 终止指定 PID（会要求确认）
portwatch kill 1234

# 释放 8080 端口（会要求确认并验证）
portwatch free 8080
```

### JSON 契约

JSON 响应顶层包含 `schema_version`。当前版本为 `"2"`，端口结果可能包含可选的
`service` 对象（例如 `name: "Vite"`）；脚本应先检查 schema 版本，再读取新增字段。
空结果使用空数组，例如：

```json
{"schema_version":"2","ports":[]}
```

`--json free` 和 `--json find <name>` 也使用同一个 schema 版本字段。
`--json watch` 使用 JSON Lines：每行是一个独立事件对象，包含 `event`、`observed_at`、端口、协议、PID 和进程名。

`--json info <pid>` 使用独立的进程详情 schema（当前为 `"1"`），返回
`process` 对象及其 `ports` 数组；它不会改变端口响应的 schema `"2"`。

### TUI 快速用法

TUI 默认只显示 TCP `LISTENING` 端口，适合先解决“端口被谁占用”。最短操作路径是：

```text
portwatch tui
  Up/Down 选择端口（方向键异常时，u 上移、j 下移）
  Enter   查看详情
  K       打开终止确认
  Enter   确认，Esc 取消
```

常用辅助操作：`/` 按端口、PID 或进程名过滤，`R` 刷新，`V` 选择
Listening/Connections/All，`?` 查看完整帮助，`Q` 退出。终止后 PortWatch
会验证 PID 已退出并检查端口是否释放。遇到 `Access denied` 时请用管理员权限
重新打开 PowerShell。

### 退出码

脚本可以依赖以下稳定退出码：

| 退出码 | 含义 |
| --- | --- |
| `0` | 查询成功、空结果、用户取消或 Ctrl+C |
| `2` | 参数、端口、PID 或筛选条件无效 |
| `3` | 扫描或进程信息读取失败 |
| `4` | 权限不足 |
| `5` | Kill 失败、关键 PID 拒绝或 Kill 后验证失败 |

所有全局选项应放在位置参数或子命令之前，例如使用
`portwatch --json 8080`，不要写成 `portwatch 8080 --json`。

## 平台说明

Windows 使用 IP Helper API 扫描 IPv4/IPv6 TCP 监听端口，并通过 Windows
进程 API 和 WMI 获取进程详情。Linux 使用 `/proc`，macOS 使用 `lsof`/`ps`；
这两个平台的实现已包含在代码中，但建议在目标系统上单独验证权限和命令可用性。

Windows CLI 支持 TCP 监听端口和 UDP 绑定端口，`--protocol` 可选 `tcp`、`udp` 或 `all`；Linux/macOS 当前只实现 TCP，使用 UDP 时会返回明确的不支持错误。TUI 可以通过 `portwatch tui` 或 `portwatch tui 8080` 显式启动，并默认进入 TCP `LISTENING` 视图；端口表格和端口 JSON 会提供可选的开发服务识别结果，无法确认时显示 `Unknown`。

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
V4 产品范围、退出码、安全边界和后续任务拆分见 [`prd/v4.md`](prd/v4.md)。
下一轮 TUI 的端口诊断工作流和交互边界见 [`prd/v6-guided-tui.md`](prd/v6-guided-tui.md)。
