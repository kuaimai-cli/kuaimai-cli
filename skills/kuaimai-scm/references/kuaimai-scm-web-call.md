# web call — scm 域原子 API

scm 域全部走 `web call scm.<operation>`，无 shortcuts。

## 基本流程

```bash
kuaimai-cli capabilities --output json
kuaimai-cli schema scm.staff-query --output json

kuaimai-cli web call scm.staff-query \
  --body '{"pageNo":1,"pageSize":20}' --output json --no-color
```

## 注意

- **baseUrl** 固定 `https://scm.superboss.cc/`，勿改 `api.url`
- 日志类接口多数需 `startTime` / `endTime`
- 写操作先 `--dry-run --verbose`

执行规则详见 [`kuaimai-scm-meta-execution.md`](kuaimai-scm-meta-execution.md)。
