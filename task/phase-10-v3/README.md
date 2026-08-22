# Phase 10: V3 Port Selectors

V3 的第一项功能是端口范围查询。默认执行顺序为先完成 `1001`，再执行后续端口选择器任务。

## 当前范围

```text
portwatch 3000-4000
portwatch --json 3000-4000
```

UDP、端口集合和 JSON watch 不属于当前任务。
