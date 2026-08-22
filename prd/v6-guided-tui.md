# PortWatch V6 PRD：端口诊断工作流

## 1. 文档状态

- 状态：提案，作为下一轮 TUI 产品和实现的唯一入口
- 目标版本：`v0.5.0`
- 目标平台：Windows 优先；已有 Linux/macOS 能力不得回归
- 影响范围：`portwatch tui`、TUI 直接依赖的扫描和进程查询

这不是一次视觉换肤，也不是把所有功能塞进一个终端页面。V6 解决的是当前产品问题：用户能看到一张端口表，却不知道从哪里开始、为什么要切换模式、如何安全完成“端口 -> 进程 -> 释放”的任务。

## 2. 产品定位

PortWatch TUI 是一个“端口诊断工作台”，服务开发者的一个高频问题：

> “我的开发端口被谁占用了？我能安全地释放它吗？”

首屏应该让用户在几秒内完成定位，而不是要求用户先理解网络连接、状态枚举和快捷键体系。

设计原则：

1. 端口优先：默认只显示 TCP `LISTENING`。
2. 逐步展开：列表、详情、危险操作是三个清晰层级。
3. 可解释：每个失败都说明原因和下一步，而不是显示 `-` 或吞错。
4. 可逆安全：Kill 永远先确认，完成后验证 PID 和端口。
5. 终端原生：高密度、键盘优先、无装饰性 Dashboard。

## 3. 当前产品问题

基于真实使用反馈和现有实现，V6 要修复以下问题：

- `portwatch tui` 的入口和使用方式不够明确；如果旧二进制把 `tui` 当端口解析，必须给出版本/升级提示。
- 首屏同时暴露 Listening、Connections、All、Filter、Kill 等概念，首次使用者不知道哪个与“端口被占用”有关。
- 表格缺少明显的“当前选中项”和当前任务状态。
- Enter 详情、Esc 返回、K 确认之间的页面层级不稳定，容易误以为详情只是列表末尾的一段文本。
- 进程信息失败显示 `-`，用户无法判断是权限、进程退出还是系统进程。
- Refresh 后按索引恢复选择，端口变化时可能跳到另一条记录。
- 帮助栏罗列按键，但没有告诉用户一条最短可用路径。

## 4. 目标用户和核心任务

### 4.1 主要用户

- 在 Windows 上运行 Node、Java、Go、Python、Docker Desktop 等开发服务的开发者。
- 需要快速处理端口冲突，不希望记住 `netstat`、`tasklist`、`taskkill` 参数的人。

### 4.2 V6 必须完成的任务

```text
启动 TUI
  -> 默认看到监听端口
  -> 通过端口/PID/进程名过滤
  -> 选中一行并查看详情
  -> 明确确认后终止进程
  -> 验证 PID 已退出且端口已释放
```

下列任务属于辅助能力，不得抢占首屏：查看普通连接、查看全部记录、刷新和退出。

## 5. 命令合同

### 5.1 启动方式

```powershell
portwatch tui
portwatch tui 8080
```

- 无参数：进入 Listening 列表。
- 带端口：进入 Listening 列表，并只展示该端口；端口不存在时显示空结果和可理解提示。
- `portwatch tui` 必须被解析为 TUI 子命令，不能再被当作端口参数。
- TUI 启动时显示 PortWatch 版本；发现不匹配的旧二进制时，错误提示应包含重新构建/更新建议。

### 5.2 CLI 与 TUI 边界

- TUI 复用现有 `port.Scanner`、`process.Manager` 和端口/进程模型。
- TUI 不改变 `portwatch 8080`、`free`、`kill`、JSON 和 watch 的输出合同。
- TUI 中的 Kill 与 CLI 一样必须确认；不提供无确认快捷键。

## 6. 首屏信息架构

首屏固定为四块，且只展示完成核心任务所需的信息：

```text
PortWatch  ·  Port & Process Manager                         v0.5.0

Listening ports                         18 results · Updated 2s ago
Filter: -

  PORT   PID      PROCESS                 STATE
> 8080   18232    java.exe                LISTENING
  3000   12884    node.exe                LISTENING

↑↓ Select   Enter Details   / Filter   K Kill   R Refresh   ? Help   Q Quit
```

要求：

- `PortWatch`、当前视图、结果数量和更新时间必须可见。
- 选中项使用 `>` 或 `▶` 等文本标记，颜色只能辅助。
- 80 列宽时保留 `PORT`、`PID`、`PROCESS`；`STATE` 和 `PROTOCOL` 可压缩但不能覆盖。
- 没有结果时显示下一步，例如 `No listening ports found` 或 `No match for "8080"`。
- 底部只保留最常用操作；高级能力放进帮助/视图菜单。

## 7. 视图与导航

### 7.1 默认视图

默认视图是 `Listening`，只包含真实状态为 `LISTENING` 的 TCP 记录。不能通过端口号范围、PID 大小或字符串 `TCP` 猜测监听状态。

### 7.2 高级视图

普通连接和全部记录仍需保留，但不放在首屏快捷键主路径中：

- `v` 打开轻量 View 菜单：`Listening`、`Connections`、`All`。
- `l/c/a` 在菜单中可作为快捷键。
- 未实现该 Scope 的平台显示 `Unavailable on this platform`，不得伪造结果或 panic。
- `A` 不得再表示自动刷新；自动刷新延期到后续版本。

### 7.3 页面层级

```text
List -> Details -> Confirm Kill -> Result -> List
```

- `Enter`：进入独立详情页。
- `Esc`：返回上一级；过滤输入中先退出输入，再清除过滤。
- `?`：打开帮助页/帮助覆盖层；Esc 返回原页面。
- 列表中无可见记录时，Enter/K 必须无操作。

## 8. 过滤和搜索

- `/` 进入输入；输入实时生效，不区分大小写。
- 匹配字段：端口、PID、进程名；未来有 Service 时再加入 Service。
- `Enter` 保留过滤并回到列表，`Esc` 清空过滤并回到列表。
- 过滤结果为空时保留输入内容，并明确禁止 Enter/K 操作。
- 过滤后刷新不得丢失条件；原选中记录消失时选择最近的可见记录。

## 9. 详情页

详情页只展示当前选中记录及其进程信息，不与列表混排：

- Port、Protocol、State
- Local Address、Remote Address
- PID、Process Name
- Executable Path、Command Line、Working Directory（可获取时）

读取不到的字段显示 `Unknown` 或 `Unavailable`。进程信息失败必须分类：

| 情况 | 列表 | 详情 |
| --- | --- | --- |
| 权限不足 | `Unknown` | `Process information unavailable. Access denied.` |
| 进程已退出 | `Unknown` | `Process information unavailable. Process exited.` |
| 无法分类 | `Unknown` | `Process information unavailable.` |

不能根据 PID 数字或空名称猜测 `System`。

## 10. Kill 和验证

列表或详情按 `K` 后进入确认页，必须显示 PID、进程名、端口和警告：

```text
Terminate process?
PID      18232
Process  java.exe
Port     8080

This will terminate the process.
Enter Confirm   Esc Cancel
```

完成 Kill 后：

1. 验证 PID 已退出。
2. 重新扫描并验证目标端口不再由该 PID 占用。
3. 成功显示 `Process terminated` 和 `Port 8080 is available`。
4. 失败保留上下文，显示原始错误分类；权限错误提示管理员权限。
5. 禁止终止 PID 4、当前 PortWatch 进程和平台明确标记的关键进程。

## 11. 刷新和状态

- `R` 只触发一次刷新；刷新期间保留旧列表并显示 `Refreshing...`。
- 刷新使用协议、端口、PID、本地地址、远端地址组成的稳定键恢复选择。
- 状态栏只显示当前视图、结果数量、更新时间和最近一次操作结果。
- 单个进程查询失败不应清空其他端口。
- V6 不做自动刷新、通知中心、历史记录和图表。

## 12. 响应式和可访问性

必须验收 80x24、100x30、120x40、160x50：

1. 不重叠、不截断关键字段、不因内容变化改变行高。
2. 颜色不是唯一状态表达；选中、错误、成功均有文本标记。
3. 长进程名在列表截断，详情页换行展示。
4. 终端无法使用颜色时仍能完成完整流程。

## 13. 非目标

V6 不做：

- 重写 TUI 框架、CLI 架构或平台扫描器。
- 首屏自动进入 Connections/All。
- 自动刷新、批量 Kill、无确认 Kill。
- Service Detection、Docker/Kubernetes、Web UI、数据库、AI/MCP。
- 通过硬编码把进程名映射成服务名。

## 14. 技术约束

- 保持 Bubble Tea 和现有接口注入方式。
- Scope 能力放在 `internal/port`；页面状态、过滤和导航放在 `internal/tui`。
- `internal/tui` 不直接调用 Windows API。
- 不新增 Repository、Service、DAO 或数据库层。
- 所有状态转移用 fake Scanner/Manager 测试；真实机器只用于 Windows 验收。
- 优先修正命令解析和信息架构，再实现高级 Scope；不得为了 UI 猜测扫描结果。

## 15. 验收标准

### 首次使用

- [ ] `portwatch tui` 明确进入 TUI，不再报“port must be a number”。
- [ ] 用户不看文档也能看懂首屏是在展示监听端口。
- [ ] 首屏能找到端口、进程和下一步操作。

### 核心流程

- [ ] 默认 Listening。
- [ ] 上下键选择，Enter 详情，Esc 返回。
- [ ] `/` 实时过滤端口/PID/进程名。
- [ ] K 必须确认，取消不调用 Terminate。
- [ ] Kill 后验证 PID 和端口，并给出成功/失败反馈。

### 稳定性

- [ ] 刷新尽可能保留选择和过滤。
- [ ] Unknown/权限错误可解释。
- [ ] `go test ./...`、`go vet ./...`、`go build ./...` 通过。
- [ ] Windows 实机完成 Listening、空结果、详情、取消 Kill、成功 Kill、权限失败场景。

## 16. 实施顺序

建议按以下小任务串行/并行执行：

1. 命令解析：修复 `tui` 子命令和可选端口参数。
2. 首屏信息架构：Listening 默认、Header、状态栏、选中标记。
3. 导航状态：List/Details/Confirm/Result 四态及 Esc 返回。
4. 过滤与刷新：实时过滤、稳定选择、刷新保留。
5. 进程信息：Unknown 和错误分类。
6. Kill：确认、保护、PID/端口验证。
7. 高级视图：View 菜单中的 Connections/All。
8. 终端尺寸验收、README 和版本升级提示。

任务之间必须按文件所有权拆分；修改统一 `app.go`/命令根入口的任务串行化，测试和文档在最后集成。

