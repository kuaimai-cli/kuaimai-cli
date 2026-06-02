# item +list — 按条件搜索商品列表（库存页）

## 适用场景

用户要**列出、搜索、查找**商品（库存 ARCHIVE_V2 页），或只有标题需要先拿 `sysItemId` 时。

对应 meta operation：`stock-list`（`post_form`，`pageable:true`，`write:false`）。

## 前置检查

```bash
kuaimai-cli auth status --output json
```

未登录则按 [`kuaimai-shared`](../../kuaimai-shared/SKILL.md) 引导 `auth login`。

## 命令

将 `<关键词>` 替换为用户提供的值。

```bash
kuaimai-cli item +list \
  --body '{"title":"<关键词>","pageNo":1,"pageSize":50}' \
  --output json --no-color
```

`item list` 与 `item +list` **实现相同**；Agent 优先使用 `+list`。

### 全量翻页（--page-all）

仅在用户**明确要求全部/导出/所有页**时使用：

```bash
kuaimai-cli item +list \
  --body '{"title":"<关键词>","pageNo":1,"pageSize":50}' \
  --page-all --output json --no-color
```

**CLI 海量数据防护（与 Agent 规则对齐）**：

| 参数 | 作用 |
|------|------|
| `--page-all` | 自动翻页 |
| `--page-limit N` | 最多拉取 N 条 |
| `--page-confirm prompt` | 达 500 条/预估>1000 时交互 `[y/N]`（默认） |
| `--page-confirm yes` | Agent/脚本自动继续 |
| `--page-confirm no` | 达阈值静默停止 |

非交互环境达阈值会停止并返回已拉取数据；Agent 需继续时加 `--page-confirm yes`。

## 请求说明

- **contentType**：`post_form` — `--body` 传 JSON，CLI 转为 `application/x-www-form-urlencoded`
- 常用筛选：`title`、`pageNo`、`pageSize`
- shortcut 默认 body 含 ARCHIVE_V2 页字段（`pageType`、`subPageType`、`api_name` 等）；只改筛选时保留其余默认
- **`item +list` 不支持 `--dry-run`**（`write:false` 查询接口）

### 多值参数

逗号分隔字符串，非数组：`"userIds": "100,200"`

### 等价 service 调用

```bash
kuaimai-cli service item stock-list \
  --body '{"title":"<关键词>","pageNo":1,"pageSize":50}' \
  --output json --no-color
```

service 层无 ARCHIVE_V2 默认 body；若报错，参考 `item +list --help` 默认 body 补全字段。

## 响应解析

- 成功：`ok === true`，列表在 `data`
- 从列表记录中取 `sysItemId` 供 `get-detail` / `update-title` 使用

## 与 item-query-list-v2 的区别

| 场景 | 用哪个 |
|------|--------|
| 按标题搜库存页商品 | **本命令** `item +list` |
| 商品档案 V2、商家编码/类目筛选 | [`item-query-list-v2`](kuaimai-item-query-list-v2.md) |

## 禁止

- 不要用 `+list` 结果人工数条数来回答「有多少个」——应使用 [`kuaimai-item-count.md`](kuaimai-item-count.md)
- 不要手写 URL 或 curl
- 不要对查询接口加 `--dry-run`

## 失败处理

读 `error` 与 `hint`；必要时加 `--verbose` 重试。
