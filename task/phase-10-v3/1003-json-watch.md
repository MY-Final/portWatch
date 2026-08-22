# V3-1003 JSON Watch Events

## 目标

让 `watch` 输出可被脚本逐行消费的 JSON Lines：

```powershell
portwatch --json watch
portwatch --json watch 8080
```

## 负责文件

- `internal/command/watch.go`
- `internal/command/watch_test.go`
- `pkg/model/json.go`
- `README.md`、V3 文档和任务索引

## 数据约定

每行必须是独立 JSON 对象，包含：`schema_version`、`event`、`observed_at`、`port`、`protocol`、`state`、`pid`、`process_name`。

`event` 只能是 `added` 或 `removed`。错误写 stderr，不得污染 stdout。Ctrl+C 正常退出。

## 验收

```powershell
go test ./internal/command
go test ./...
go vet ./...
go build ./...
```
