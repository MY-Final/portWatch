# Phase 11: Configuration

来源：`backlog/002-config.md`（已排期）。目标版本 v0.9.0。

## 当前范围

```text
%AppData%\portwatch\config.json            (Windows)
$XDG_CONFIG_HOME/portwatch/config.json     (Linux/macOS，缺省 ~/.config)
PORTWATCH_CONFIG=<path>                    (显式覆盖，测试与高级用法)
```

支持键：`interval`（watch/wait 轮询周期）、`process`（默认进程名过滤）。
格式用 JSON（标准库），不引入 TOML 依赖，遵守「标准库优先」规则。

## 执行顺序

`1101 -> 1102`，串行；`1102` 依赖 `1101` 的加载接口。

- [`1101-config-discovery-and-load.md`](1101-config-discovery-and-load.md)
- [`1102-config-defaults-wiring.md`](1102-config-defaults-wiring.md)
