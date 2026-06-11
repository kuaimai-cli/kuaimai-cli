# registry 执行规则（scm 域）

## 域名

- `web call scm.*` 使用 registry 中 scm 条目的 `baseUrl`（`https://scm.superboss.cc/`）
- 与 `config api.url`（erp1）无关；同一 profile 可同时调 item 与 scm

## contentType / pageable / write

与 item 域相同，见 [`kuaimai-item-meta-execution.md`](kuaimai-item-meta-execution.md)。

## 日志时间范围

多数 `scm.logging-*` 需要：

```json
{"startTime":"2026-06-01 00:00:00","endTime":"2026-06-08 23:59:59","pageNo":1,"pageSize":20}
```

用户未给时间时，默认近 7 天并向用户确认。

## 发现接口

```bash
kuaimai-cli capabilities --output json
kuaimai-cli schema scm.logging-publish-log --output json
```

不要用 jq 过滤已废弃的 `.data.operations[]` 结构；以当前 `schema` 输出的 `apis` / `api` 为准。
