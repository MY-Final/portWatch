# PortWatch TUI Polish PRD

## 1. 文档状态

- 状态：提案，供下一轮 TUI 实现使用
- 目标版本：`v0.4.0`（TUI 体验版本）
- 影响范围：`portwatch tui` 及其直接依赖的扫描能力
- 不改变：现有 CLI、JSON schema、`free`/`kill` 命令和默认端口扫描接口

这不是一次重写 PortWatch 的计划。它解决的是当前 TUI 的产品表达和交互问题：用户应该一打开就看到开发端口，能够在几次按键内完成“端口 → 进程 → 详情 → 确认终止 → 验证”的完整流程。

## 2. 产品定位

PortWatch TUI 是一个面向开发者的端口与进程管理工具，不是 `netstat` 的终端包装，也不是监控大屏。

核心承诺：

> **默认展示可用于开发服务诊断的监听端口，并用清晰、可恢复的键盘流程安全处理占用进程。**

设计关键词：终端原生、信息密集、键盘优先、低装饰、可解释、安全。

## 3. 当前问题与证据

当前 `internal/tui` 已具备 Bubble Tea 启动、列表、搜索、详情字符串、刷新和 Kill，但状态仍然混在一个简单 Model 中：

- 页面没有明确的产品 Header、当前模式和更新时间
- `Enter` 只在列表下方追加详情文本，不是真正的详情视图；`Esc` 也没有返回详情的语义
- `Detail`、`ConfirmKill` 和列表状态可以同时出现，容易产生层级混乱
- 刷新按索引保留选择，排序或端口变化后可能选中另一条记录
- 进程信息失败显示 `-`，用户无法区分权限不足、进程已退出和系统进程
- 当前 `WindowsScanner.List()` 已经只返回 TCP `LISTENING`；因此 Connections 必须通过明确的可选扫描接口实现，不能在 TUI 中猜测或过滤不存在的连接记录
- 底部帮助只覆盖 `R / / / Q`，无法发现详情、Kill 和视图切换

## 4. 目标与成功标准

### 4.1 目标

1. TUI 默认进入 `Listening` 视图，优先解决“端口被谁占用”。
2. 用明确的页面状态隔离列表、详情、过滤和 Kill 确认。
3. 让选择、刷新、过滤和返回行为可预测。
4. 让进程信息缺失成为可解释状态，而不是无意义的 `-`。
5. Kill 前展示上下文，Kill 后验证进程和端口状态。
6. 在 80 列宽终端仍保持关键字段可读。

### 4.2 成功标准

在 Windows 真实终端中，用户执行：

```powershell
portwatch tui
```

应能完成：

```text
默认 Listening
  -> 上下键选中端口
  -> Enter 进入详情
  -> Esc 返回列表
  -> K 打开确认
  -> Enter 确认或 Esc 取消
  -> 终止后验证 PID 和端口
  -> 返回列表并保留可理解的结果状态
```

自动化测试必须覆盖上述状态转移；真实终端验收必须覆盖 80x24、100x30、120x40 和 160x50 四种尺寸。

## 5. 视图模式

### 5.1 模式定义

TUI 使用一个明确的 `ViewScope`：

| 按键 | 模式 | 记录语义 |
| --- | --- | --- |
| `L` | Listening | TCP `LISTENING`；未来可包含 UDP `BOUND` |
| `C` | Connections | TCP 非 `LISTENING` 状态，例如 `ESTABLISHED`、`TIME_WAIT` |
| `A` | All | 当前扫描器实际返回的全部记录 |

默认模式必须是 `Listening`。模式判断使用 `PortInfo.State`，不能通过协议名称或临时端口范围猜测。

### 5.2 扫描能力边界

现有 `port.Scanner.List()` 继续表示默认监听视图，保持 CLI 兼容。为 TUI 增加可选能力接口，例如：

```go
type ScopedScanner interface {
    ListScope(context.Context, Scope) ([]model.PortInfo, error)
}
```

要求：

- 已实现该接口的平台使用真实状态扫描
- 未实现的平台在 `C`/`A` 模式显示 `Connections unavailable on this platform`，不崩溃、不伪造结果
- `L` 模式始终优先复用现有 `List()` 结果
- 不为了 TUI 修改现有 CLI 的默认 TCP/UDP 输出合同

V5 不要求 TUI 同时增加 UDP 协议切换；状态模式和协议切换应是两个独立概念。

## 6. 页面结构

页面从上到下固定为四个区域：

```text
PortWatch                         v0.4.0
Port & Process Manager

LISTENING · TCP                   18 ports · Updated 1s ago
PORT   PROTOCOL   PID      PROCESS
>8080  TCP        18232    java.exe
 3000  TCP        12884    node.exe

Filter: -

↑↓ Navigate  Enter Details  K Kill  / Filter
L Listening  C Connections   A All   R Refresh  Q Quit
```

### Header

- 必须显示 `PortWatch`
- 必须显示 `Port & Process Manager`
- 版本号可显示，但不能占用关键字段空间
- 禁止巨大 ASCII Logo、渐变、装饰性卡片和监控大屏布局

### 状态栏

显示当前模式、协议、记录数量和最近一次成功刷新时间。扫描失败时保留旧列表，并把错误放在状态栏；不能因为一次进程详情失败而清空整个页面。

### 列表

最低字段：`PORT`、`PROTOCOL`、`PID`、`PROCESS`。`STATE` 在宽度允许时显示；`SERVICE` 不是本版本必需字段。

排序固定为端口、协议、PID、本地地址、远端地址。选中行使用 `>` 等纯文本标记，颜色只能作为辅助，不能作为唯一标识。

### 底部帮助

快捷键名称和大小写统一，至少包含：

```text
↑↓ Navigate  Enter Details  K Kill  / Filter
L Listening  C Connections   A All   R Refresh  Q Quit
```

## 7. 详情视图

按 `Enter` 后进入独立详情视图，列表不再与详情文本混排。详情至少尝试展示：

- Port、Protocol、State
- Local Address、Remote Address
- PID、Process Name
- Executable Path、Command Line、Working Directory

字段读取不到时显示 `Unknown` 或 `Unavailable`，不得猜测。按 `Esc` 返回列表并恢复之前的模式、过滤条件和选择。

详情页支持 `K`，其确认流程与列表触发的 Kill 完全一致。

## 8. 进程信息状态

列表和详情必须区分“没有进程名”和“查询失败”：

| 状态 | 列表显示 | 详情显示 |
| --- | --- | --- |
| 成功 | 真实进程名 | 完整可用字段 |
| 权限不足 | `Unknown` | `Process information unavailable. Access denied.` |
| 进程已退出 | `Unknown` | `Process information unavailable. Process exited.` |
| 无法分类 | `Unknown` | `Process information unavailable.` |

只有平台能力明确返回系统进程身份时才显示 `System`；不能因为 PID 小、名字为空或访问失败而猜测。

进程详情失败不应阻止其他端口显示。状态信息应由 TUI 保存的 lookup error 分类产生，不能通过错误字符串模糊匹配。

## 9. Kill 流程

列表和详情均可按 `K` 进入确认状态。确认状态必须明确显示：PID、进程名（或 `Unknown`）、关联端口和警告文本。

交互要求：

- `Enter` 确认
- `Esc` 取消
- `y`/`n` 可作为兼容快捷键，但不能跳过确认
- 确认状态下禁止上下移动和重复触发
- PID 4、当前 PortWatch 进程和平台标记的关键进程必须拒绝终止

终止后：

1. 验证 PID 已退出
2. 对列表关联端口重新扫描并验证目标记录消失
3. 显示 `Process terminated` 和端口释放结果
4. 刷新列表，尽可能保留模式、过滤和合理选择

失败时保留列表和详情上下文，显示可操作的错误；权限错误必须提示管理员权限，但不能吞掉原始错误分类。

## 10. Filter

按 `/` 进入过滤输入，输入期间实时匹配：

- Port
- PID
- Process Name
- 当前已提供的 Service 名称

过滤不区分大小写；空过滤显示全部当前模式记录。`Esc` 清空并退出过滤，`Enter` 保留过滤并回到导航。

过滤后选择必须自动归一化到第一个可见记录；没有匹配时不能触发 `Enter` 或 `K` 操作。

## 11. Refresh 与选择保持

`R` 触发一次扫描。刷新期间显示 `Refreshing...`，但保留旧列表，避免页面闪空。

选择使用稳定记录键保存：协议、端口、PID、本地地址和远端地址的组合。刷新后：

- 原记录仍存在：恢复到原记录
- 原记录消失：选择相邻位置
- 当前过滤无匹配：选择状态保持但禁止操作
- 不得每次刷新无条件跳到第一行

自动刷新不是本版本范围。`A` 已分配给 All 模式，不得再复用为 Auto Refresh；自动刷新进入后续独立 PRD。

## 12. 响应式终端布局

必须验证：80x24、100x30、120x40、160x50。

宽度不足时按以下优先级隐藏或截断：

1. 保留 Port、PID、Process
2. 保留 Protocol、State
3. 截断或隐藏 Service、地址和长路径

任何截断都不能改变行高、选择位置或使文字覆盖相邻区域。详情页允许多行展示长命令和路径。

## 13. 非目标

本版本不做：

- 重写 CLI、PortInfo、ProcessInfo 或现有 TUI 框架
- 把 TUI 设为无参数默认入口
- TUI 自动刷新、通知中心、历史记录或监控图表
- TUI UDP 协议切换
- 新的 Service Detection 规则或硬编码猜测
- 颜色主题、渐变、ASCII 艺术和 Web Dashboard
- 批量 Kill、无确认 Kill、数据库、后台服务、Web、AI、MCP

## 14. 技术约束

- 继续使用 Bubble Tea 和现有接口注入
- 视图状态放在 `internal/tui`，平台扫描能力放在 `internal/port`
- `internal/tui` 不直接调用 Windows API
- 不新增 Repository、Service、DAO 或数据库层
- 现有 `port.Scanner.List()` 行为保持兼容
- 所有状态转移和渲染逻辑必须可用 fake Scanner/Manager 测试
- 任务按文件所有权拆分；最终集成阶段才修改 `internal/tui/app.go`

## 15. 建议任务与依赖

详细任务卡位于 `task/phase-05-tui-polish/`。依赖关系：

```text
001 -> (002, 003, 004, 005, 006, 007, 008)
(002, 003, 004, 005, 006, 007, 008) -> 009 -> 010
```

其中：

- `001` 冻结状态模型、Scope 和 lookup error 分类
- `002` 只负责可选 Connections 扫描契约和 Windows 实现
- `003` 负责 Header、状态栏和表格布局
- `004` 负责稳定选择和刷新恢复
- `005` 负责详情视图
- `006` 负责 Kill 确认、验证和反馈
- `007` 负责实时 Filter
- `008` 负责帮助栏和模式提示
- `009` 负责窄终端布局验收与修正
- `010` 负责最终 TUI 集成测试、手测清单和文档

`003`、`004`、`007`、`008` 只在不修改 `app.go` 的前提下并行；需要修改统一 Update/View 的任务必须串行进入 `009`/`010`。

## 16. 验收清单

### 默认体验

- [ ] `portwatch tui` 默认显示 Listening
- [ ] Header、模式、数量、更新时间清晰
- [ ] 选中行有不依赖颜色的明显标记
- [ ] 底部帮助覆盖所有核心快捷键

### 交互

- [ ] `L/C/A` 模式语义正确，未支持模式给出明确提示
- [ ] `Enter` 进入详情，`Esc` 返回列表
- [ ] `/` 实时匹配 Port/PID/Process，空结果不能操作
- [ ] `R` 刷新后尽可能保持稳定选择
- [ ] `Q` 和 Ctrl+C 正常退出

### 安全

- [ ] 列表和详情的 `K` 都必须先确认
- [ ] `Esc`/`n` 取消不会调用 Terminate
- [ ] Kill 后验证 PID 和端口
- [ ] 权限错误和进程信息缺失可解释

### 质量

- [ ] 80x24、100x30、120x40、160x50 无明显重叠或错位
- [ ] `go test ./...`
- [ ] `go vet ./...`
- [ ] `go build ./...`
- [ ] Windows 真实终端手测 Listening、Connections、详情、Kill 和权限错误

## 17. 发布与回滚

V5 TUI Polish 建议作为 `v0.4.0` 发布。CLI 默认输出、JSON schema 和原有 Kill 命令是回归门槛；若 Connections 扫描在某平台不稳定，只回滚该可选 Scope 能力，不回滚 Listening、详情和安全确认流程。
