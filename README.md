<div align="center">

# PortWatch

**面向开发者的跨平台端口诊断与进程管理 CLI**

[![version](https://img.shields.io/badge/version-v0.8.0-blue)](https://github.com/MY-Final/portWatch/releases)
[![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-607D8B)](#-平台支持)
[![dist](https://img.shields.io/badge/dist-single%20binary-2E7D32)](https://github.com/MY-Final/portWatch/releases)

```powershell
irm https://raw.githubusercontent.com/MY-Final/portWatch/main/install.ps1 | iex
```

</div>

> [!NOTE]
> **日常用 `pw` 就够了。** `pw` 是 `portwatch` 的等价短别名（同一份程序，按调用名显示帮助与错误前缀），
> 安装脚本会同时装好 `portwatch` 和 `pw`。本文示例统一写 `pw`，换成 `portwatch` 行为完全一致。
> （`pw` 仅在 FreeBSD 上与系统命令撞名，Windows/Linux/macOS 无冲突。）

## ✨ 特性

| | |
| --- | --- |
| 🔍 **查端口** | 单个 / 范围 / 集合查询，按进程名、PID、状态筛选；三平台均支持 UDP（`--protocol udp\|all`） |
| 👀 **找进程** | 端口 → PID → 命令行 / 可执行文件 / 工作目录；`info` 展示父进程链；`find` 按进程名反查端口 |
| 🔥 **释放端口** | `free` / `kill` 先展示进程详情再确认，终止后自动验证端口已释放 |
| 📡 **实时监控** | `watch` 增量报告端口上下线；`wait` 阻塞等待端口空闲/被占用（`--timeout` 限时） |
| 🖥️ **交互界面** | `tui` 一屏搞定：键盘导航、过滤、详情、终止确认 |
| 🤖 **脚本友好** | 稳定 JSON 契约（`schema_version`）、JSON Lines 事件流、可复现退出码、GNU 风格参数交错 |
| ⚡ **快** | Windows 直读进程 PEB，不拉起 PowerShell/WMI，上百端口的进程信息毫秒级返回 |
| 🧹 **干净卸载** | `pw uninstall` 自清理，保守处理用户 PATH |

## 📦 安装

### Windows（PowerShell）

```powershell
irm https://raw.githubusercontent.com/MY-Final/portWatch/main/install.ps1 | iex
```

从 Release 下载对应架构的 zip，SHA256 校验后装入 `%USERPROFILE%\bin`，按需补充用户 PATH。
管道方式无法向脚本传参（`iex` 只接收整段脚本文本），需要带参数时改用：

```powershell
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/MY-Final/portWatch/main/install.ps1))) -Version v0.8.0
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/MY-Final/portWatch/main/install.ps1))) -Uninstall
```

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/MY-Final/portWatch/main/install.sh | bash
# 固定版本：
curl -fsSL https://raw.githubusercontent.com/MY-Final/portWatch/main/install.sh | PORTWATCH_VERSION=v0.8.0 bash
```

装入 `~/.local/bin`；目录不在 PATH 时自动在 `~/.bashrc`（zsh 为 `~/.zshrc`）追加一行**带标记**的
`export PATH`（卸载时只删这一行，不动你自己写的内容），当前终端 `source` 一下或重开即可使用；
其他 shell 会打印手动添加提示。

### 从源码构建

要求 Go 1.22+，在项目根目录执行：

<details>
<summary>展开构建与 PATH 配置命令</summary>

```powershell
$bin = Join-Path $env:USERPROFILE "bin"
New-Item -ItemType Directory -Force $bin | Out-Null
go build -o (Join-Path $bin "portwatch.exe") ./cmd/portwatch
go build -o (Join-Path $bin "pw.exe") ./cmd/portwatch

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (($userPath -split ';') -notcontains $bin) {
    [Environment]::SetEnvironmentVariable(
        "Path", (($userPath.TrimEnd(';') + ';' + $bin).Trim(';')), "User"
    )
}
```

重新打开 PowerShell 后 `pw --version` 验证。

</details>

## 🚀 快速开始

```text
$ pw 5173
PORT  PROTOCOL STATE     PID   PROCESS NAME COMMAND                                            EXECUTABLE PATH      SERVICE
5173  TCP      LISTENING 23184 node.exe    node D:\proj\web\node_modules\vite\bin\vite.js   node.exe             Vite

$ pw find node
PID   PROCESS      PORTS
23184 node.exe     5173

$ pw info 23184
FIELD                VALUE
PID                  23184
PROCESS NAME         node.exe
PARENT CHAIN         node.exe (23184) ← npm.exe (21000) ← Code.exe (5678)
EXECUTABLE PATH      C:\Program Files\nodejs\node.exe
COMMAND              node D:\proj\web\node_modules\vite\bin\vite.js
WORKING DIRECTORY    D:\proj\web
USER                 -
PORTS                5173

$ pw free 5173
PORT  PROTOCOL STATE     PID   PROCESS NAME ...
5173  TCP      LISTENING 23184 node.exe    ...
Kill 1 process(es) listening on port 5173? [y/N] y
Verifying port release...
Port 5173 is now available.

$ pw --expect occupied --timeout 30s wait 5173
Port 5173 is occupied.
```

查询或终止系统进程时出现 `access denied`，用「以管理员身份运行」的终端重试即可。

## 📖 命令总览

| 命令 | 说明 |
| --- | --- |
| `pw` | 列出全部 TCP `LISTENING` 端口 |
| `pw <port>` | 查询一个端口（如 `pw 8080`） |
| `pw <start>-<end>` | 查询端口范围（如 `pw 3000-4000`） |
| `pw --ports 3000,8080,8848` | 查询指定端口集合 |
| `pw free <port>` | 确认后终止占用端口的进程并验证释放 |
| `pw kill <pid>` | 确认后按 PID 终止进程并验证 |
| `pw find <name>` | 按进程名搜索及其监听端口 |
| `pw info <pid>` | 查看 PID 详情（命令行 / 路径 / 工作目录 / 父进程链）及其端口 |
| `pw watch [port]` | 实时报告监听端口变化，Ctrl+C 退出 |
| `pw wait <port>` | 阻塞等待端口空闲（`--expect occupied` 等被占用，`--timeout` 限时可选） |
| `pw tui [port]` | 交互式终端界面，可聚焦单个端口 |
| `pw uninstall` | 卸载自身（`--yes` 跳过确认） |
| `pw --help` / `pw --version` | 帮助与版本 |

### 全局选项

| 选项 | 说明 |
| --- | --- |
| `--json` | 输出稳定 JSON（端口 / 搜索 / 详情 / watch 事件） |
| `--yes` | 跳过确认提示（`free` / `kill` / `uninstall`） |
| `--protocol tcp\|udp\|all` | 协议选择（三平台均支持） |
| `--process <name>` | 按进程名筛选 |
| `--pid <p1,p2>` | 按 PID 集合筛选 |
| `--state <state>` | 按端口状态筛选（如 `LISTENING`） |
| `--interval <duration>` | `watch` 轮询周期（默认 `1s`；`wait` 未指定时默认 `200ms`） |
| `--expect free\|occupied` | `wait` 等待的目标状态（默认 `free`） |
| `--timeout <duration>` | `wait` 的等待上限，超时退出码 `124`（默认不限时） |

选项与位置参数可任意交错（GNU 风格）：`pw --json 8080`、`pw 8080 --json`、`pw find node --json`
均合法；`--flag value` 与 `--flag=value` 两种写法都支持。

## 🧭 更多示例

```powershell
# 全部监听端口
pw

# UDP 绑定端口
pw --protocol udp

# 只看 node 进程监听的端口
pw --process node

# 按 PID / 状态筛选
pw --pid 1234
pw --state LISTENING

# PID 详情
pw info 1234

# 每 2 秒报告端口变化
pw --interval 2s watch

# 只监控 8080
pw watch 8080

# 阻塞等待 8080 空闲（比如等旧进程退出），Ctrl+C 退出码 0
pw wait 8080

# 等服务起来（端口被占用即返回），最多等 30 秒
pw --expect occupied --timeout 30s wait 8080
```

### JSON 输出

`--json` 的端口响应顶层带 `schema_version`（当前 `"2"`），空结果用空数组，字段只增不改：

```json
{"schema_version":"2","ports":[]}
```

- 端口记录含可选 `service` 对象（如 `{"name":"Vite","type":"Node.js","confidence":95}`），
  脚本应先检查 schema 版本再读取新增字段；
- `--json find <name>`、`--json free <port>` 复用同一版本字段；
- `--json info <pid>` 使用独立的进程详情 schema（当前 `"1"`）；v0.8.0 起进程对象含纯加法字段
  `parent_pid` 与 `ancestors`（`[{"pid":21000,"name":"npm.exe"},...]`，最多 8 级）；
- `--json watch` 输出 JSON Lines：每行一个独立事件对象
  （`event`、`observed_at`、端口、协议、PID、进程名）；
- `--json wait <port>` 结束时输出单个事件对象（无 schema 版本）：

```json
{"event":"port_free","observed_at":"2026-08-23T15:30:00+08:00","port":5173,"protocol":"TCP"}
```

### 退出码

| 退出码 | 含义 |
| --- | --- |
| `0` | 查询成功、空结果、用户取消、Ctrl+C 或卸载成功 |
| `2` | 参数、端口、PID 或筛选条件无效 |
| `3` | 扫描或进程信息读取失败，或卸载等其他执行失败 |
| `4` | 权限不足（含卸载时二进制被占用） |
| `5` | Kill 失败、关键 PID 拒绝或 Kill 后验证失败 |
| `124` | `wait` 等待超时（GNU timeout 惯例） |

## 🖥️ TUI

`pw tui`（或 `pw tui 8080` 聚焦单个端口）默认进入 TCP `LISTENING` 视图：

| 按键 | 操作 |
| --- | --- |
| `↑` `↓`（备用 `u` `j`） | 选择端口 |
| `Enter` | 查看进程详情 |
| `K` | 打开终止确认（`Enter` 确认，`Esc` 取消） |
| `/` | 按端口、PID 或进程名过滤 |
| `R` | 刷新 |
| `V` | 切换 Listening / Connections / All 视图 |
| `?` | 完整帮助 |
| `Esc` / `Q` | 返回 / 退出 |

终止后 PortWatch 会验证 PID 已退出且端口已释放。详情页会展示该进程的父进程链
（与 `pw info` 同款 `node.exe (23184) ← npm.exe (21000)` 一行）；Linux 上 `V` 切换
Connections/All 视图可看全部 TCP 状态，macOS 暂仅 LISTENING 视图。

## 📊 平台支持

| 能力 | Windows | Linux | macOS |
| --- | --- | --- | --- |
| TCP 监听端口 | ✓ | ✓ | ✓（依赖 `lsof`） |
| TCP 连接（非 LISTENING） | ✓ | ✓（TUI Connections/All 视图） | ✗ |
| UDP 绑定端口 | ✓ | ✓（`/proc/net/udp`） | ✓（依赖 `lsof`） |
| 进程命令行 | ✓（直读 PEB） | ✓（`/proc/<pid>/cmdline`） | ✓（依赖 `ps`） |
| 进程工作目录 | ✓（直读 PEB） | ✓（`/proc/<pid>/cwd`） | ✓（依赖 `lsof`） |

- **Windows**：IP Helper API 扫描 IPv4/IPv6，进程信息直接读取 PEB，无 PowerShell/WMI 依赖；
- **Linux**：`/proc`（单次遍历构建 inode→PID 映射）；
- **macOS**：`lsof` + `ps`。

Linux 的 TUI Connections/All 视图读取 `/proc/net/tcp{,6}` 的全部状态并与 Windows 同名；macOS 的 TUI 非 LISTENING 视图暂不支持，会返回明确错误。

## 🧹 卸载

| 方式 | 命令 |
| --- | --- |
| 命令行（推荐） | `pw uninstall`（或 `portwatch uninstall`；同目录的另一个别名一并删除） |
| 安装脚本 | `.\install.ps1 -Uninstall`（Windows）/ `./install.sh --uninstall`（Linux/macOS），两个别名一并清理 |
| 手动 | 删除 `%USERPROFILE%\bin\portwatch.exe`、`pw.exe`（或 `~/.local/bin/` 下同名文件），目录清空后按需从用户 PATH 移除 |

<details>
<summary>卸载行为细节</summary>

- 确认交互与 `free`/`kill` 一致：显示将要删除的路径（含同目录的另一个别名），`y`/`yes` 确认，回车取消（退出码 0），`--yes` 跳过；
- 同目录下的另一个别名（`portwatch` ↔ `pw`）随自身一并删除，一条命令卸载干净；正在运行而删不掉的别名会保留并由残留提示说明；
- Windows 下运行中的 exe 无法直接删除：先改名为 `<name>.uninstalling.exe`，由分离进程在主进程退出后延迟删除；
- PATH 清理是保守的：仅当二进制位于默认安装目录（`%USERPROFILE%\bin` / `~/.local/bin`）且删除后为空时，
  才从用户级 PATH 移除该目录，绝不触碰系统级 PATH；unix 无统一用户级 PATH 存储，只提示需从 shell 配置移除的行；
- `install.sh` 卸载时会移除安装时写入 rc 文件的那行带标记的 `export PATH`；
- 脚本卸载在未安装时输出 `already uninstalled` 并以退出码 0 结束。

</details>

## 🔧 开发

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

发布配置见 [`.goreleaser.yaml`](.goreleaser.yaml)，CI 见 [`.github/workflows/`](.github/workflows/)；
开发任务与产品规划见 [`task/`](task/) 与 [`prd/`](prd/)。

## 许可

暂未声明开源协议；如需引入第三方使用，请先联系作者。
