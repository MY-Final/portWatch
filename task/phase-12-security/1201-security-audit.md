# 1201 Security Audit

## 目标

对全部安全敏感面做一次系统审计，输出 `docs/security-review.md`；只有附复现证据的问题才修改生产代码。

## 负责文件

- `docs/security-review.md`（新增，报告本体）
- 仅在有问题时：对应生产文件 + 回归测试

不增加防御性抽象、不改动无证据项。

## 审计清单（按风险面）

1. **Windows 进程内存读取**（`internal/process/windows_peb.go`）
   - OpenProcess 权限位是否最小（PROCESS_QUERY_LIMITED_INFORMATION|PROCESS_VM_READ）
   - ReadProcessMemory 长度上限（maxRemoteStringLength）与恶意长度的防御复核
   - PEB 偏移表对 32/64 位的正确性（offsets 文件）
2. **PID 重用 TOCTOU**（`kill` / `free`）
   - Info 展示 → 用户确认 → Terminate 之间 PID 被回收重用的窗口
   - 现有 validateKillTarget（PID 4、自身）覆盖是否充分；评估终止前后二次校验进程名的成本收益
3. **卸载自删除链**（`internal/command/uninstall*.go`）
   - 生成的批处理内容中路径引用（quoteWindowsPath 的双引号转义）是否可被文件名注入逃逸
   - 注册表 PATH 写入的值类型保持与 %VAR% 保留复核
4. **安装脚本**（`install.ps1` / `install.sh`）
   - 下载与 SHA256 校验的绑定强度（checksums 与资产同源 GitHub Release 的信任模型说明）
   - `irm | iex` 管道执行的固有风险与 scriptblock 变体的等价性
   - rc 文件写入的标记行注入面（恶意目录名包含换行的可能性）
5. **外部命令调用**（linux kill/ps、darwin lsof/ps、windows cmd）
   - 全部参数化传递，确认无 shell 字符串拼接；确认无用户输入进入路径搜索
6. **wait/watch 长驻路径**
   - ctx 取消传播、无 goroutine 泄漏、`--timeout` 无上限时的资源占用说明

## 报告格式

每项含：风险等级（高/中/低/信息）、复现步骤或代码位置、当前状态（已防护/需修复/接受风险）、
修复对应的 commit 或任务卡。结论汇总裁决表。

## 验收

```powershell
go test ./...
go vet ./...
```

报告覆盖上述 6 项全部条目；如产生代码修改，Windows CI 全绿且每项修改都有对应复现测试。
