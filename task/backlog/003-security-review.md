# Backlog 003 Security Review

> 状态：**scheduled**。已排期为 Phase 12，范围按现状重写（PowerShell 路径已被 PEB 取代），任务卡见 [`../phase-12-security/`](../phase-12-security/)。以下为原始卡片内容。


## 目标
对 Windows API 权限、PowerShell 调用、PID 重用、命令注入和 kill 安全确认做一次独立审计。

## 文件边界
输出 `docs/security-review.md` 与必要的回归测试；只有有复现证据时才修改生产代码，不增加无用抽象。

## 依赖
Phase 1 MVP 和 Phase 2 kill 命令完成。

## 验收与测试
报告包含复现步骤、风险等级和修复状态；回归测试在 Windows CI 通过；确认用户输入不会进入 shell 拼接。
