# PortWatch 安全审计报告

- 日期：2026-08-23
- 范围：v0.8.0（`e644566`）代码基线
- 方法：按 `task/phase-12-security/1201-security-audit.md` 的六个审计面逐项核对；
  每个发现要求代码位置或复现证据，仅有复现证据的问题才修改生产代码。
- 结论：**发现 2 项中级问题（均已修复）、1 项低级问题（已修复）、3 项信息级（文档化）**，
  其余面核实为已防护。无高危发现。

## 裁决汇总

| ID | 审计面 | 发现 | 等级 | 状态 |
| --- | --- | --- | --- | --- |
| S1-1 | PEB | Exists 请求了超出所需的查询权限 | 低 | 已修复 |
| S1-2 | PEB | 恶意 UNICODE_STRING 长度被截断而非拒绝 | 信息 | 接受（有界，无害） |
| S1-3 | PEB | 三架构偏移表与公共 ABI 布局一致 | — | 已防护 |
| S2-1 | PID 重用 | kill/free 确认窗口期的 PID 重用可致误杀无关进程 | 中 | 已修复 |
| S3-1 | 自删除 | 安装目录含 `& ( ) ^ %` 时 cmd /c 引号失守，延迟删除静默失效 | 中 | 已修复 |
| S3-2 | 自删除 | 批处理内容中 `%VAR%` 展开，路径含变量名字面量时删除目标被改写 | 中 | 已修复 |
| S4-1 | 安装脚本 | 双平台校验逻辑对篡改资产正确拒绝（实测） | — | 已防护 |
| S4-2 | 安装脚本 | 信任模型边界：同源 checksum 不防仓库沦陷；`irm \| iex` 固有风险 | 信息 | 文档化 |
| S4-3 | 安装脚本 | rc 标记行为固定字面量，无目录名插值 | — | 已防护 |
| S5 | 外部命令 | 除自删除批处理外全部参数化传递，无 shell 拼接 | — | 已防护 |
| S6 | wait/watch | 取消传播与资源释放正确；`--timeout` 缺省无限等待为文档化设计 | — | 已防护 |

## 1. Windows 进程内存读取（PEB）

**S1-1（低，已修复）**：`Exists` 用 `PROCESS_QUERY_INFORMATION` 打开进程，
而 `GetExitCodeProcess` 仅需 `PROCESS_QUERY_LIMITED_INFORMATION`。已降级为最小权限
（`internal/process/windows.go`）。Info 路径的
`QUERY_LIMITED|VM_READ` 与 Terminate 路径的 `TERMINATE|SYNCHRONIZE` 复核为各自操作的最小集。

**S1-2（信息，接受）**：`readRemoteUnicodeString` 对超过 64KB 的 `Length` 采用**截断**而非报错。
复核结论：分配有界（无法用伪造长度触发大分配）、读回长度逐一校验、`UTF16ToString` 在 NUL 处截断，
截断只会导致字段变短，无内存安全问题。保持现状。

**S1-3（已防护）**：`windows_peb_offsets_64.go`（x64/arm64）与 `_32.go`（x86）的
PEB→ProcessParameters、CURDIR、CommandLine、UNICODE_STRING.Buffer 偏移与公共 ABI 布局
（phnt/ReactOS 定义）逐项一致；32 位进程读 64 位目标时按设计返回明确错误而非错数据。

## 2. PID 重用 TOCTOU（kill / free）

**S2-1（中，已修复）**：原流程为「扫描 → Info 展示 → 用户确认 → Terminate」，
确认期间人机交互可达数秒；Windows PID 复用激进，原进程退出后 PID 可被无关进程接管，
届时终止的是接管者。`free` 的事后端口复查只能发现"端口仍被占用"，不能发现"杀错了进程"。

修复：`kill` 与 `free` 在确认后、终止前重新调用 `Manager.Info` 并比对
`Name` + `Executable`，不一致即以 `ErrKillFailed`（退出码 5）中止，
错误信息明确写出身份变化（`internal/command/kill.go`、`free.go`）。
回归测试 `TestKillAbortsWhenIdentityChanges`、`TestFreeAbortsWhenIdentityChanges`
以「两次 Info 返回不同身份」的 fake 复现窗口并断言不发生 Terminate。

**残余风险（接受）**：复核与终止之间仍有微秒级窗口，且同名进程重用可绕过比对。
彻底关闭需要进程创建时间令牌（`GetProcessTimes`），涉及数据模型加字段，留作后续硬化项，
非本轮范围。

## 3. 卸载自删除链（Windows）

**S3-1（中，已修复，实测复现）**：延迟删除通过 `cmd /c <脚本路径>` 执行。
对抗性目录名实测（六个敌意目录、真实二进制自卸载）：`amp&ersand`、`(parens)`、`caret^up`、
`per%TEMP%cent` 四类目录下**延迟删除静默失效**——cmd /c 对含特殊字符的脚本路径会剥引号，
命令在 `&` 处断开，批处理从未执行（主程序仍报成功）。

修复：脚本改为创建在系统临时目录（路径可控、无特殊字符），批处理**内容**中引用原始路径——
批处理内双引号中的 `& ( ) ^` 为字面量。复测六个敌意目录全部清理成功且临时目录无脚本残留。

**S3-2（中，已修复）**：批处理内容中 `%` 会展开环境变量（含双引号内），路径含 `%TEMP%`
字面量时 `del` 目标被改写。修复：`quoteWindowsPath` 现将 `%` 转义为 `%%`（批处理字面量），
自删除用的 `%~f0` 不经过该函数、不受影响。`per%TEMP%cent`、`percent%sign` 目录实测通过。

引号注入面结构性不存在：Windows 文件名不允许 `"`。注册表 PATH 写入保持原始值与
值类型（`DoNotExpandEnvironmentNames`），`%JAVA_HOME%` 类条目不受影响（复核确认）。

## 4. 安装脚本信任链

**S4-1（已防护，实测）**：对 v0.8.0 真实资产做篡改实验——翻转 zip 中部一个字节后，
`install.ps1` 的 `Get-FileHash` 比对与 `install.sh` 的 `grep+awk` 比对均正确检出
mismatch 并终止（两侧 expected/actual 实测输出留档于审计过程记录）。

**S4-2（信息，文档化）**：信任模型为 **TLS + 与资产同源的 checksums.txt**——
防传输损坏与镜像篡改，**不防 GitHub 仓库/Release 本身沦陷**（checksum 与资产可被一并替换）。
`irm | iex` 管道执行是远程代码执行的固有暴露，与官方安装脚本生态一致。
升级路径（未实施）：release 签名（cosign/minisign）+ 在 README 公布签名验证命令。

**S4-3（已防护）**：`install.sh` 写入 rc 的标记行为固定字面量
`export PATH="$HOME/.local/bin:$PATH" # portwatch-install`，不含任何目录名插值，
无通过目录名注入 rc 内容的通路。

## 5. 外部命令调用清单

仓库内 `exec.Command*` 全量清单（v0.8.0 基线）：

| 位置 | 命令 | 参数来源 |
| --- | --- | --- |
| `internal/port/darwin.go` | `lsof` | 固定参数表 |
| `internal/process/darwin.go` | `ps` ×3、`lsof` | `strconv.Itoa(pid)` |
| `internal/process/darwin.go` / `linux.go` | `kill -0` / `kill -TERM` | `strconv`/`fmt.Sprint(pid)`，pid 经正整数校验 |
| `internal/command/uninstall_windows.go` | `cmd /c <脚本>` | 唯一 shell 字符串面，见第 3 节处置 |

PID 均经 `ValidatePID`（正整数）后再转字符串，无用户可控字符串进入任何命令行；
除自删除批处理（已单独处置）外不存在 shell 拼接。

## 6. wait / watch 长驻路径

- 取消传播：`wait` 与 watch 引擎的休眠均为 `timer + select{ctx.Done(), timer.C}` 且
  `defer timer.Stop()`，Ctrl+C 即时返回、退出码 0（有测试覆盖）；
- goroutine 生命周期：`processinfo.Resolve` 的 worker 上限 8、随 channel 关闭退出，
  TUI refresh 使用模型 context；
- `--timeout` 缺省为无限等待属文档化设计（README 全局选项表），单端口 200ms 轮询的
  资源占用可忽略。

## 修复对应的提交

| 修复 | 内容 |
| --- | --- |
| S1-1 | `internal/process/windows.go` Exists 权限降级 |
| S2-1 | `internal/command/kill.go`、`free.go` 身份复核 + kill/free 测试 |
| S3-1 + S3-2 | `internal/command/uninstall_windows.go` 脚本移至临时目录 + `%` 转义 |
