# scm 域 meta 执行规则

## baseUrl

- `services.scm.baseUrl` = `https://scm.superboss.cc/`
- CLI `service scm *` 自动使用该域名，与全局 `api.url`（erp1）无关
- 同一 profile 下可同时调用 `service item`（erp1）与 `service scm`（scm）

## contentType

| contentType | HTTP | Agent `--body` |
|-------------|------|----------------|
| `get_query` | GET + query | JSON → URL 参数 |
| `post_form` | POST form | JSON → form 字段 |
| `post_json` | POST JSON | JSON body |

## pageable

- `pageable:true` 的接口（如 `logging-publish-log`、`item-base-page`）支持 `--page-all`
- 默认单页；用户未要求全量时不加 `--page-all`
- Agent/管道续查：`--page-confirm yes`；限条数：`--page-limit N`

## write / dry-run

- `write:true`：保存、编辑、删除类接口，必须先 `--dry-run --verbose`
- `write:false` 查询接口不支持 `--dry-run`

## 日志时间参数

多数 `/logging/*` 分页接口需要：

```json
{
  "startTime": "2026-06-01 00:00:00",
  "endTime": "2026-06-08 23:59:59",
  "pageNo": 1,
  "pageSize": 20
}
```

## schema 自省

```bash
kuaimai-cli schema --output json | jq '.data.operations[] | select(.service=="scm" and .operation=="logging-publish-log")'
```
