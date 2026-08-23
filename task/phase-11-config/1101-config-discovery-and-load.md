# 1101 Config Discovery And Load

## 目标

新建 `internal/config` 包：按平台发现配置文件、解析并校验，产出零值等价于"无配置"的结果。

## 负责文件

- `internal/config/config.go`
- `internal/config/path_windows.go`（`//go:build windows`）
- `internal/config/path_unix.go`（`//go:build !windows`）
- `internal/config/config_test.go`

不得修改 `internal/command`、平台扫描器或任何现有行为。

## 行为要求

- 配置结构仅含 `interval`（duration 字符串，如 `"2s"`）与 `process`（字符串）；未知键报未知字段错误
- 路径解析顺序：`PORTWATCH_CONFIG` 环境变量 > 平台默认路径
- Windows 默认 `%AppData%\portwatch\config.json`；unix 默认 `$XDG_CONFIG_HOME/portwatch/config.json`，未设置时 `~/.config/portwatch/config.json`
- 文件不存在 → 返回零值 Config 与 nil error，调用方行为与现状完全一致
- JSON 语法错误 → 错误信息含行号（`json.SyntaxError` 的 Offset 换算）；字段类型错误 → 含字段名
- `interval` 非法 duration 或 `<= 0` → 字段级错误；`process` 空串合法（等价未设置）
- 包不打印、不退出；错误上抛由 CLI 决定处理方式
- 单测用 `t.TempDir` + `PORTWATCH_CONFIG` 注入，不触碰真实用户目录

## 验收

```powershell
go test ./internal/config
go test ./...
go vet ./...
GOOS=windows go build ./... ; GOOS=linux go build ./... ; GOOS=darwin go build ./...
```

测试覆盖：不存在文件、合法文件、语法错误（断言行号出现）、未知键、非法 interval、空 process、XDG 未设置回退。
